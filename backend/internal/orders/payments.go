package orders

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/promptpay"
	"github.com/Khant26/business-web-app/backend/internal/settings"
)

var errNotPayable = httpx.Conflict("not_payable", "This order is not waiting for payment")

// PromptPay handles GET /me/orders/{code}/promptpay.
func (h *Handler) PromptPay(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	u := auth.UserFrom(ctx)
	var status, method string
	var total int64
	var due *time.Time
	if err := h.DB.QueryRow(ctx, `SELECT status, payment_method, total_satang, payment_due_at FROM orders WHERE code = $1 AND user_id = $2`,
		r.PathValue("code"), u.ID).Scan(&status, &method, &total, &due); err != nil {
		return err
	}
	if status != StatusPendingPayment || method != "promptpay" {
		return errNotPayable
	}
	st, err := settings.Load(ctx, h.DB)
	if err != nil {
		return err
	}
	if st.PromptPayID == "" {
		return httpx.Conflict("promptpay_not_configured", "PromptPay is not configured")
	}
	payload, err := promptpay.Payload(st.PromptPayID, total)
	if err != nil {
		return httpx.Conflict("promptpay_not_configured", "PromptPay ID is invalid")
	}
	return httpx.OK(w, map[string]any{"payload": payload, "amount_satang": total, "promptpay_name": st.PromptPayName, "due_at": due})
}

// UploadSlip handles POST /me/orders/{code}/slip.
func (h *Handler) UploadSlip(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	u := auth.UserFrom(ctx)
	code := r.PathValue("code")
	var status, method string
	if err := h.DB.QueryRow(ctx, `SELECT status, payment_method FROM orders WHERE code = $1 AND user_id = $2`, code, u.ID).Scan(&status, &method); err != nil {
		return err
	}
	if !CanUploadSlip(status, method) {
		return errNotPayable
	}
	dir := filepath.Join(h.UploadDir, "private", "slips")
	name, err := httpx.SaveImage(w, r, dir)
	if err != nil {
		return err
	}
	var out *Order
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var id int64
		if err := tx.QueryRow(ctx, `SELECT id, status, payment_method FROM orders WHERE code = $1 AND user_id = $2 FOR UPDATE`, code, u.ID).
			Scan(&id, &status, &method); err != nil {
			return err
		}
		if !CanUploadSlip(status, method) {
			return errNotPayable
		}
		if _, err := tx.Exec(ctx, `INSERT INTO payment_slips (order_id, file_path) VALUES ($1, $2)`, id, "slips/"+name); err != nil {
			return err
		}
		if err := setStatus(ctx, tx, id, StatusPaymentReview, "Payment slip uploaded", &u.ID, ""); err != nil {
			return err
		}
		var err error
		out, err = loadOrder(ctx, tx, false, "o.id = $1", id)
		return err
	})
	if err != nil {
		_ = os.Remove(filepath.Join(dir, name))
		return err
	}
	return httpx.OK(w, out)
}

// SlipFile handles GET /admin/slips/{id}/file.
func (h *Handler) SlipFile(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var rel string
	if err := h.DB.QueryRow(r.Context(), `SELECT file_path FROM payment_slips WHERE id = $1`, id).Scan(&rel); err != nil {
		return err
	}
	if strings.Contains(rel, "..") {
		return httpx.NotFound()
	}
	data, err := os.ReadFile(filepath.Join(h.UploadDir, "private", filepath.FromSlash(rel)))
	if os.IsNotExist(err) {
		return httpx.NotFound()
	} else if err != nil {
		return err
	}
	w.Header().Set("Content-Type", httpx.ContentTypeForExt(rel))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Disposition", "inline")
	_, err = w.Write(data)
	return err
}

// ReviewSlip handles POST /admin/slips/{id}/approve and /reject.
func (h *Handler) ReviewSlip(approve bool) httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()
		staff := auth.UserFrom(ctx)
		id, err := httpx.PathInt64(r, "id")
		if err != nil {
			return err
		}
		var req struct {
			Note string `json:"note"`
		}
		if err := httpx.DecodeOptional(w, r, &req); err != nil {
			return err
		}
		req.Note = strings.TrimSpace(req.Note)
		if !approve && req.Note == "" {
			return httpx.Validation(map[string]string{"note": "required"})
		}
		if httpx.Len(req.Note) > 500 {
			return httpx.Validation(map[string]string{"note": "max 500 characters"})
		}
		var out *Order
		err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
			var orderID int64
			var slipStatus, orderStatus, code string
			if err := tx.QueryRow(ctx, `SELECT s.order_id, s.status, o.status, o.code FROM payment_slips s JOIN orders o ON o.id = s.order_id
				WHERE s.id = $1 FOR UPDATE OF s, o`, id).Scan(&orderID, &slipStatus, &orderStatus, &code); err != nil {
				return err
			}
			if slipStatus != "pending" || orderStatus != StatusPaymentReview {
				return errInvalidTransition
			}
			newSlip, action := "approved", "slip.approve"
			if !approve {
				newSlip, action = "rejected", "slip.reject"
			}
			if _, err := tx.Exec(ctx, `UPDATE payment_slips SET status = $2, note = $3, reviewed_by = $4, reviewed_at = now() WHERE id = $1`,
				id, newSlip, req.Note, staff.ID); err != nil {
				return err
			}
			if approve {
				note := "Payment approved"
				if req.Note != "" {
					note += ": " + req.Note
				}
				if err := setStatus(ctx, tx, orderID, StatusPaid, note, &staff.ID, "paid_at = now()"); err != nil {
					return err
				}
			} else {
				st, err := settings.Load(ctx, tx)
				if err != nil {
					return err
				}
				if err := setStatus(ctx, tx, orderID, StatusPendingPayment, "Payment slip rejected: "+req.Note, &staff.ID,
					"payment_due_at = now() + make_interval(mins => $3::int)", st.PaymentWindowMinutes); err != nil {
					return err
				}
			}
			if err := audit.Log(ctx, tx, r, action, "order", code, map[string]any{"slip_id": id, "note": req.Note}); err != nil {
				return err
			}
			var err error
			out, err = loadOrder(ctx, tx, true, "o.id = $1", orderID)
			return err
		})
		if err != nil {
			return err
		}
		return httpx.OK(w, out)
	}
}
