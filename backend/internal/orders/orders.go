package orders

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/inventory"
	"github.com/Khant26/business-web-app/backend/internal/settings"
	"github.com/Khant26/business-web-app/backend/internal/users"
)

// OrderSummary is the order list item.
type OrderSummary struct {
	Code           string    `json:"code"`
	Status         string    `json:"status"`
	PaymentMethod  string    `json:"payment_method"`
	TotalSatang    int64     `json:"total_satang"`
	ItemCount      int64     `json:"item_count"`
	CreatedAt      time.Time `json:"created_at"`
	FirstItemImage string    `json:"first_item_image"`
}

// ShippingAddress is the address snapshot stored on the order.
type ShippingAddress struct {
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	Line1         string `json:"line1"`
	Subdistrict   string `json:"subdistrict"`
	District      string `json:"district"`
	Province      string `json:"province"`
	Postcode      string `json:"postcode"`
}

// Item is an order line.
type Item struct {
	ID              int64  `json:"id"`
	ProductID       *int64 `json:"product_id"`
	ProductSlug     string `json:"product_slug"`
	VariantID       *int64 `json:"variant_id"`
	ProductNameEN   string `json:"product_name_en"`
	ProductNameMY   string `json:"product_name_my"`
	VariantNameEN   string `json:"variant_name_en"`
	VariantNameMY   string `json:"variant_name_my"`
	SKU             string `json:"sku"`
	ImageURL        string `json:"image_url"`
	UnitPriceSatang int64  `json:"unit_price_satang"`
	Quantity        int64  `json:"quantity"`
	LineTotalSatang int64  `json:"line_total_satang"`
	CanReview       bool   `json:"can_review"`
}

// HistoryEntry is one status change.
type HistoryEntry struct {
	Status    string    `json:"status"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// Slip is a payment slip. ImageURL is only set for staff.
type Slip struct {
	ID         int64      `json:"id"`
	Status     string     `json:"status"`
	Note       string     `json:"note"`
	UploadedAt time.Time  `json:"uploaded_at"`
	ReviewedAt *time.Time `json:"reviewed_at"`
	ImageURL   string     `json:"image_url,omitempty"`
}

// Order is the full order.
type Order struct {
	OrderSummary
	SubtotalSatang  int64           `json:"subtotal_satang"`
	DiscountSatang  int64           `json:"discount_satang"`
	ShippingSatang  int64           `json:"shipping_satang"`
	VATRateBP       int64           `json:"vat_rate_bp"`
	VATSatang       int64           `json:"vat_satang"`
	CouponCode      string          `json:"coupon_code"`
	ShippingAddress ShippingAddress `json:"shipping_address"`
	Email           string          `json:"email"`
	Note            string          `json:"note"`
	Courier         string          `json:"courier"`
	TrackingNumber  string          `json:"tracking_number"`
	TrackingURL     string          `json:"tracking_url"`
	PaymentDueAt    *time.Time      `json:"payment_due_at"`
	PaidAt          *time.Time      `json:"paid_at"`
	ShippedAt       *time.Time      `json:"shipped_at"`
	DeliveredAt     *time.Time      `json:"delivered_at"`
	CancelledAt     *time.Time      `json:"cancelled_at"`
	CancelReason    string          `json:"cancel_reason"`
	Items           []Item          `json:"items"`
	History         []HistoryEntry  `json:"history"`
	Slips           []Slip          `json:"slips"`
	CanCancel       bool            `json:"can_cancel"`
	CanUploadSlip   bool            `json:"can_upload_slip"`
	Customer        *auth.User      `json:"customer,omitempty"`

	id     int64
	userID int64
}

const codeAlphabet = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

func newOrderCode() string {
	b := make([]byte, 6)
	max := big.NewInt(int64(len(codeAlphabet)))
	for i := range b {
		n, _ := rand.Int(rand.Reader, max)
		b[i] = codeAlphabet[n.Int64()]
	}
	return "SY-" + string(b)
}

const summaryCols = `o.code, o.status, o.payment_method, o.total_satang,
	(SELECT COALESCE(sum(quantity), 0) FROM order_items WHERE order_id = o.id), o.created_at,
	COALESCE((SELECT image_url FROM order_items WHERE order_id = o.id ORDER BY id LIMIT 1), '')`

func scanSummary(dest *OrderSummary) []any {
	return []any{&dest.Code, &dest.Status, &dest.PaymentMethod, &dest.TotalSatang, &dest.ItemCount, &dest.CreatedAt, &dest.FirstItemImage}
}

// loadOrder loads a full order by where (on alias o). Staff views include slip image URLs and the customer.
func loadOrder(ctx context.Context, q db.DBTX, staff bool, where string, args ...any) (*Order, error) {
	o := &Order{}
	a := &o.ShippingAddress
	dest := append(scanSummary(&o.OrderSummary), &o.id, &o.userID, &o.SubtotalSatang, &o.DiscountSatang, &o.ShippingSatang,
		&o.VATRateBP, &o.VATSatang, &o.CouponCode, &a.RecipientName, &a.Phone, &a.Line1, &a.Subdistrict, &a.District, &a.Province,
		&a.Postcode, &o.Email, &o.Note, &o.Courier, &o.TrackingNumber, &o.TrackingURL, &o.PaymentDueAt, &o.PaidAt, &o.ShippedAt,
		&o.DeliveredAt, &o.CancelledAt, &o.CancelReason)
	err := q.QueryRow(ctx, `SELECT `+summaryCols+`, o.id, o.user_id, o.subtotal_satang, o.discount_satang, o.shipping_satang,
		o.vat_rate_bp, o.vat_satang, o.coupon_code, o.recipient_name, o.phone, o.line1, o.subdistrict, o.district, o.province,
		o.postcode, o.email, o.note, o.courier, o.tracking_number, o.tracking_url, o.payment_due_at, o.paid_at, o.shipped_at,
		o.delivered_at, o.cancelled_at, o.cancel_reason FROM orders o WHERE `+where, args...).Scan(dest...)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `SELECT oi.id, oi.product_id, COALESCE(p.slug, ''), oi.variant_id, oi.product_name_en, oi.product_name_my,
		oi.variant_name_en, oi.variant_name_my, oi.sku, oi.image_url, oi.unit_price_satang, oi.quantity,
		($2 = 'delivered' AND oi.product_id IS NOT NULL AND NOT EXISTS (
			SELECT 1 FROM reviews r WHERE r.product_id = oi.product_id AND r.user_id = $3))
		FROM order_items oi LEFT JOIN products p ON p.id = oi.product_id WHERE oi.order_id = $1 ORDER BY oi.id`, o.id, o.Status, o.userID)
	if err != nil {
		return nil, err
	}
	o.Items, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Item, error) {
		var it Item
		err := r.Scan(&it.ID, &it.ProductID, &it.ProductSlug, &it.VariantID, &it.ProductNameEN, &it.ProductNameMY, &it.VariantNameEN,
			&it.VariantNameMY, &it.SKU, &it.ImageURL, &it.UnitPriceSatang, &it.Quantity, &it.CanReview)
		it.LineTotalSatang = it.UnitPriceSatang * it.Quantity
		return it, err
	})
	if err != nil {
		return nil, err
	}
	rows, err = q.Query(ctx, `SELECT status, note, created_at FROM order_status_history WHERE order_id = $1 ORDER BY created_at, id`, o.id)
	if err != nil {
		return nil, err
	}
	o.History, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (HistoryEntry, error) {
		var h HistoryEntry
		return h, r.Scan(&h.Status, &h.Note, &h.CreatedAt)
	})
	if err != nil {
		return nil, err
	}
	rows, err = q.Query(ctx, `SELECT id, status, note, uploaded_at, reviewed_at FROM payment_slips WHERE order_id = $1 ORDER BY uploaded_at DESC, id DESC`, o.id)
	if err != nil {
		return nil, err
	}
	o.Slips, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Slip, error) {
		var s Slip
		err := r.Scan(&s.ID, &s.Status, &s.Note, &s.UploadedAt, &s.ReviewedAt)
		if staff {
			s.ImageURL = "/api/v1/admin/slips/" + itoa(s.ID) + "/file"
		}
		return s, err
	})
	if err != nil {
		return nil, err
	}
	o.CanCancel = CustomerCanCancel(o.Status)
	o.CanUploadSlip = CanUploadSlip(o.Status, o.PaymentMethod)
	if staff {
		o.Customer, err = auth.ScanUser(q.QueryRow(ctx, `SELECT `+auth.UserColumns+` FROM users WHERE id = $1`, o.userID))
		if err != nil {
			return nil, err
		}
	}
	return o, nil
}

// setStatus updates the order status (plus optional extra SET clause) and writes history.
func setStatus(ctx context.Context, tx pgx.Tx, orderID int64, status, note string, userID *int64, extraSet string, extraArgs ...any) error {
	sql := `UPDATE orders SET status = $2, updated_at = now()`
	if extraSet != "" {
		sql += ", " + extraSet
	}
	args := append([]any{orderID, status}, extraArgs...)
	if _, err := tx.Exec(ctx, sql+` WHERE id = $1`, args...); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO order_status_history (order_id, status, note, user_id) VALUES ($1, $2, $3, $4)`, orderID, status, note, userID)
	return err
}

// releaseOrder restores stock and releases the coupon redemption of an order.
func releaseOrder(ctx context.Context, tx pgx.Tx, orderID int64, reason string, userID *int64) error {
	rows, err := tx.Query(ctx, `SELECT variant_id, quantity FROM order_items WHERE order_id = $1 AND variant_id IS NOT NULL ORDER BY variant_id`, orderID)
	if err != nil {
		return err
	}
	type iq struct{ v, q int64 }
	items, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (iq, error) {
		var x iq
		return x, r.Scan(&x.v, &x.q)
	})
	if err != nil {
		return err
	}
	for _, it := range items {
		if _, err := inventory.Move(ctx, tx, it.v, it.q, reason, "", &orderID, userID); err != nil {
			return err
		}
	}
	var couponID int64
	err = tx.QueryRow(ctx, `DELETE FROM coupon_redemptions WHERE order_id = $1 RETURNING coupon_id`, orderID).Scan(&couponID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE coupons SET used_count = GREATEST(used_count - 1, 0) WHERE id = $1`, couponID)
	return err
}

// cancelOrder cancels a locked order: restores stock, releases coupon, rejects pending slips.
func cancelOrder(ctx context.Context, tx pgx.Tx, orderID int64, reason string, userID *int64) error {
	if err := releaseOrder(ctx, tx, orderID, "cancel", userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE payment_slips SET status = 'rejected', note = 'Order cancelled', reviewed_at = now()
		WHERE order_id = $1 AND status = 'pending'`, orderID); err != nil {
		return err
	}
	return setStatus(ctx, tx, orderID, StatusCancelled, reason, userID, "cancelled_at = now(), cancel_reason = $3", reason)
}

// ExpireDue expires pending_payment orders past their due time. It returns the number expired.
func ExpireDue(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	n := 0
	err := pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `SELECT id FROM orders WHERE status = 'pending_payment' AND payment_due_at < now()
			ORDER BY payment_due_at LIMIT 100 FOR UPDATE SKIP LOCKED`)
		if err != nil {
			return err
		}
		ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
		if err != nil {
			return err
		}
		for _, id := range ids {
			if err := releaseOrder(ctx, tx, id, "expire", nil); err != nil {
				return err
			}
			if err := setStatus(ctx, tx, id, StatusExpired, "Payment window expired", nil, ""); err != nil {
				return err
			}
		}
		n = len(ids)
		return nil
	})
	return n, err
}

type placeReq struct {
	Items         []ItemReq      `json:"items"`
	AddressID     *int64         `json:"address_id"`
	Address       *users.Address `json:"address"`
	PaymentMethod string         `json:"payment_method"`
	CouponCode    string         `json:"coupon_code"`
	Note          string         `json:"note"`
	SaveAddress   bool           `json:"save_address"`
}

// Place handles POST /orders.
func (h *Handler) Place(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	u := auth.UserFrom(ctx)
	var req placeReq
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	items, err := normalizeItems(req.Items)
	if err != nil {
		return err
	}
	req.Note = strings.TrimSpace(req.Note)
	f := httpx.Fields{}
	f.Check(req.PaymentMethod != "promptpay" && req.PaymentMethod != "cod", "payment_method", "must be promptpay or cod")
	f.Check(httpx.Len(req.Note) > 500, "note", "max 500 characters")
	f.Check(req.AddressID == nil && req.Address == nil, "address", "required")
	if req.AddressID == nil && req.Address != nil {
		for k, v := range req.Address.Validate("address.") {
			f[k] = v
		}
	}
	if err := f.Err(); err != nil {
		return err
	}

	var order *Order
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var addr users.Address
		if req.AddressID != nil {
			a, err := users.GetAddress(ctx, tx, u.ID, *req.AddressID)
			if errors.Is(err, pgx.ErrNoRows) {
				return httpx.Validation(map[string]string{"address_id": "unknown address"})
			}
			if err != nil {
				return err
			}
			addr = a
		} else {
			addr = *req.Address
			if req.SaveAddress {
				if _, err := users.InsertAddress(ctx, tx, u.ID, addr); err != nil {
					return err
				}
			}
		}

		quote, err := buildQuote(ctx, tx, u.ID, items, addr.Province, req.CouponCode, true)
		if err != nil {
			return err
		}
		if len(quote.Problems) > 0 {
			return httpx.Conflict("out_of_stock", "Some items are out of stock or unavailable").WithFields(problemFields(quote.Problems))
		}
		if quote.Coupon != nil && !quote.Coupon.Valid {
			return httpx.E(http.StatusUnprocessableEntity, "coupon_invalid", "The coupon cannot be used").
				WithFields(map[string]string{"coupon_code": quote.Coupon.MessageCode})
		}
		if req.PaymentMethod == "cod" && !quote.CODAllowed {
			return httpx.E(http.StatusUnprocessableEntity, "cod_not_allowed", "Cash on delivery is not available for this order total")
		}
		st, err := settings.Load(ctx, tx)
		if err != nil {
			return err
		}

		status, note := StatusConfirmed, "Order placed (cash on delivery)"
		var due *time.Time
		if req.PaymentMethod == "promptpay" {
			status, note = StatusPendingPayment, "Order placed, waiting for PromptPay payment"
			d := time.Now().Add(time.Duration(st.PaymentWindowMinutes) * time.Minute)
			due = &d
		}
		couponCode := ""
		if quote.Coupon != nil && quote.couponID != nil {
			couponCode = quote.Coupon.Code
		}
		var code string
		for i := 0; ; i++ {
			code = newOrderCode()
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM orders WHERE code = $1)`, code).Scan(&exists); err != nil {
				return err
			}
			if !exists || i > 10 {
				break
			}
		}
		var orderID int64
		err = tx.QueryRow(ctx, `INSERT INTO orders (code, user_id, status, payment_method, subtotal_satang, discount_satang,
			shipping_satang, vat_rate_bp, vat_satang, total_satang, coupon_id, coupon_code, recipient_name, phone, email,
			line1, subdistrict, district, province, postcode, note, payment_due_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) RETURNING id`,
			code, u.ID, status, req.PaymentMethod, quote.SubtotalSatang, quote.DiscountSatang, quote.ShippingSatang, quote.VATRateBP,
			quote.VATSatang, quote.TotalSatang, quote.couponID, couponCode, addr.RecipientName, addr.Phone, u.Email,
			addr.Line1, addr.Subdistrict, addr.District, addr.Province, addr.Postcode, req.Note, due).Scan(&orderID)
		if err != nil {
			return err
		}
		for _, l := range quote.Lines {
			if _, err := tx.Exec(ctx, `INSERT INTO order_items (order_id, product_id, variant_id, product_name_en, product_name_my,
				variant_name_en, variant_name_my, sku, image_url, unit_price_satang, quantity)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`, orderID, l.productID, l.VariantID, l.ProductNameEN, l.ProductNameMY,
				l.VariantNameEN, l.VariantNameMY, l.sku, l.ImageURL, l.UnitPriceSatang, l.Quantity); err != nil {
				return err
			}
			if _, err := inventory.Move(ctx, tx, l.VariantID, -l.Quantity, "order", "", &orderID, &u.ID); err != nil {
				return err
			}
		}
		if quote.couponID != nil {
			if _, err := tx.Exec(ctx, `INSERT INTO coupon_redemptions (coupon_id, order_id, user_id) VALUES ($1, $2, $3)`, *quote.couponID, orderID, u.ID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `UPDATE coupons SET used_count = used_count + 1 WHERE id = $1`, *quote.couponID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `INSERT INTO order_status_history (order_id, status, note, user_id) VALUES ($1, $2, $3, $4)`,
			orderID, status, note, u.ID); err != nil {
			return err
		}
		order, err = loadOrder(ctx, tx, false, "o.id = $1", orderID)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.Created(w, order)
}

// MyOrders handles GET /me/orders.
func (h *Handler) MyOrders(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	p := httpx.ParsePage(r)
	status := r.URL.Query().Get("status")
	ctx := r.Context()
	var total int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM orders o WHERE o.user_id = $1 AND ($2 = '' OR o.status = $2)`, u.ID, status).Scan(&total); err != nil {
		return err
	}
	rows, err := h.DB.Query(ctx, `SELECT `+summaryCols+` FROM orders o WHERE o.user_id = $1 AND ($2 = '' OR o.status = $2)
		ORDER BY o.created_at DESC, o.id DESC LIMIT $3 OFFSET $4`, u.ID, status, p.Size, p.Offset())
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (OrderSummary, error) {
		var s OrderSummary
		return s, r.Scan(scanSummary(&s)...)
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

// MyOrder handles GET /me/orders/{code}.
func (h *Handler) MyOrder(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	o, err := loadOrder(r.Context(), h.DB, false, "o.code = $1 AND o.user_id = $2", r.PathValue("code"), u.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, o)
}

// Cancel handles POST /me/orders/{code}/cancel.
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	u := auth.UserFrom(ctx)
	var req struct {
		Reason string `json:"reason"`
	}
	if err := httpx.DecodeOptional(w, r, &req); err != nil {
		return err
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if httpx.Len(req.Reason) > 500 {
		return httpx.Validation(map[string]string{"reason": "max 500 characters"})
	}
	if req.Reason == "" {
		req.Reason = "Cancelled by customer"
	}
	var out *Order
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var id int64
		var status string
		if err := tx.QueryRow(ctx, `SELECT id, status FROM orders WHERE code = $1 AND user_id = $2 FOR UPDATE`, r.PathValue("code"), u.ID).Scan(&id, &status); err != nil {
			return err
		}
		if !CustomerCanCancel(status) {
			return errInvalidTransition
		}
		if err := cancelOrder(ctx, tx, id, req.Reason, &u.ID); err != nil {
			return err
		}
		var err error
		out, err = loadOrder(ctx, tx, false, "o.id = $1", id)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

var errInvalidTransition = httpx.Conflict("invalid_transition", "This order cannot be changed to that status")

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
