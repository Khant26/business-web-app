package orders

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/inventory"
)

// AdminSummary is the admin order list item.
type AdminSummary struct {
	OrderSummary
	CustomerEmail  string `json:"customer_email"`
	RecipientName  string `json:"recipient_name"`
	HasPendingSlip bool   `json:"has_pending_slip"`
}

const adminSummaryCols = summaryCols + `, u.email, o.recipient_name,
	EXISTS (SELECT 1 FROM payment_slips s WHERE s.order_id = o.id AND s.status = 'pending')`

func scanAdminSummary(r pgx.CollectableRow) (AdminSummary, error) {
	var s AdminSummary
	return s, r.Scan(append(scanSummary(&s.OrderSummary), &s.CustomerEmail, &s.RecipientName, &s.HasPendingSlip)...)
}

// AdminList handles GET /admin/orders.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	p := httpx.ParsePage(r)
	var conds []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, strings.ReplaceAll(cond, "?", "$"+strconv.Itoa(len(args))))
	}
	conds = append(conds, "true")
	if s := q.Get("status"); s != "" {
		add("o.status = ?", s)
	}
	if s := q.Get("payment_method"); s != "" {
		add("o.payment_method = ?", s)
	}
	if s := strings.TrimSpace(q.Get("q")); s != "" {
		add("(o.code ILIKE ? OR u.email ILIKE ? OR o.email ILIKE ? OR o.recipient_name ILIKE ? OR o.phone ILIKE ?)", db.LikePattern(s))
	}
	for _, k := range []string{"from", "to"} {
		s := q.Get(k)
		if s == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", s); err != nil {
			return httpx.Validation(map[string]string{k: "must be YYYY-MM-DD"})
		}
		if k == "from" {
			add("o.created_at >= (?::date)::timestamp AT TIME ZONE 'Asia/Bangkok'", s)
		} else {
			add("o.created_at < (?::date + 1)::timestamp AT TIME ZONE 'Asia/Bangkok'", s)
		}
	}
	where := strings.Join(conds, " AND ")
	ctx := r.Context()
	var total int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM orders o JOIN users u ON u.id = o.user_id WHERE `+where, args...).Scan(&total); err != nil {
		return err
	}
	n := len(args)
	rows, err := h.DB.Query(ctx, `SELECT `+adminSummaryCols+` FROM orders o JOIN users u ON u.id = o.user_id WHERE `+where+
		` ORDER BY o.created_at DESC, o.id DESC LIMIT $`+strconv.Itoa(n+1)+` OFFSET $`+strconv.Itoa(n+2), append(args, p.Size, p.Offset())...)
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, scanAdminSummary)
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

// AdminGet handles GET /admin/orders/{code}.
func (h *Handler) AdminGet(w http.ResponseWriter, r *http.Request) error {
	o, err := loadOrder(r.Context(), h.DB, true, "o.code = $1", r.PathValue("code"))
	if err != nil {
		return err
	}
	return httpx.OK(w, o)
}

// AdminStatus handles POST /admin/orders/{code}/status.
func (h *Handler) AdminStatus(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	staff := auth.UserFrom(ctx)
	var req struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	req.Note = strings.TrimSpace(req.Note)
	if !ValidStatus(req.Status) {
		return httpx.Validation(map[string]string{"status": "unknown status"})
	}
	if httpx.Len(req.Note) > 500 {
		return httpx.Validation(map[string]string{"note": "max 500 characters"})
	}
	code := r.PathValue("code")
	var out *Order
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var id int64
		var from, tracking string
		if err := tx.QueryRow(ctx, `SELECT id, status, tracking_number FROM orders WHERE code = $1 FOR UPDATE`, code).Scan(&id, &from, &tracking); err != nil {
			return err
		}
		if !AdminCanTransition(from, req.Status) {
			return errInvalidTransition
		}
		var err error
		switch req.Status {
		case StatusShipped:
			if tracking == "" {
				return httpx.Validation(map[string]string{"tracking_number": "set the tracking number before shipping"})
			}
			err = setStatus(ctx, tx, id, req.Status, req.Note, &staff.ID, "shipped_at = now()")
		case StatusDelivered:
			if _, err = tx.Exec(ctx, `UPDATE products p SET sold_count = p.sold_count + x.qty FROM (
				SELECT product_id, sum(quantity) AS qty FROM order_items WHERE order_id = $1 AND product_id IS NOT NULL GROUP BY product_id
			) x WHERE p.id = x.product_id`, id); err != nil {
				return err
			}
			err = setStatus(ctx, tx, id, req.Status, req.Note, &staff.ID, "delivered_at = now()")
		case StatusCancelled:
			reason := req.Note
			if reason == "" {
				reason = "Cancelled by shop"
			}
			err = cancelOrder(ctx, tx, id, reason, &staff.ID)
		default:
			err = setStatus(ctx, tx, id, req.Status, req.Note, &staff.ID, "")
		}
		if err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "order.status", "order", code, map[string]any{"from": from, "to": req.Status, "note": req.Note}); err != nil {
			return err
		}
		out, err = loadOrder(ctx, tx, true, "o.id = $1", id)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

// AdminTracking handles PUT /admin/orders/{code}/tracking.
func (h *Handler) AdminTracking(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	var req struct {
		Courier        string `json:"courier"`
		TrackingNumber string `json:"tracking_number"`
		TrackingURL    string `json:"tracking_url"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	req.Courier = strings.TrimSpace(req.Courier)
	req.TrackingNumber = strings.TrimSpace(req.TrackingNumber)
	req.TrackingURL = strings.TrimSpace(req.TrackingURL)
	f := httpx.Fields{}
	f.Check(req.Courier == "" || httpx.Len(req.Courier) > 100, "courier", "required, max 100 characters")
	f.Check(req.TrackingNumber == "" || httpx.Len(req.TrackingNumber) > 100, "tracking_number", "required, max 100 characters")
	if req.TrackingURL != "" {
		u, err := url.Parse(req.TrackingURL)
		f.Check(err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(req.TrackingURL) > 500, "tracking_url", "must be an http(s) URL")
	}
	if err := f.Err(); err != nil {
		return err
	}
	code := r.PathValue("code")
	var out *Order
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var id int64
		var status string
		if err := tx.QueryRow(ctx, `SELECT id, status FROM orders WHERE code = $1 FOR UPDATE`, code).Scan(&id, &status); err != nil {
			return err
		}
		if status == StatusCancelled || status == StatusExpired || status == StatusPendingPayment || status == StatusPaymentReview {
			return errInvalidTransition
		}
		if _, err := tx.Exec(ctx, `UPDATE orders SET courier = $2, tracking_number = $3, tracking_url = $4, updated_at = now() WHERE id = $1`,
			id, req.Courier, req.TrackingNumber, req.TrackingURL); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "order.tracking", "order", code, req); err != nil {
			return err
		}
		var err error
		out, err = loadOrder(ctx, tx, true, "o.id = $1", id)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

type dayStat struct {
	Date        string `json:"date"`
	TotalSatang int64  `json:"total_satang"`
	Orders      int64  `json:"orders"`
}

// salesStatuses are statuses counted as sales.
const salesStatuses = `('paid', 'confirmed', 'packing', 'shipped', 'delivered')`

// Dashboard handles GET /admin/dashboard.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	var out struct {
		TodaySalesSatang int64           `json:"today_sales_satang"`
		TodayOrders      int64           `json:"today_orders"`
		PendingSlips     int64           `json:"pending_slips"`
		ToPack           int64           `json:"to_pack"`
		LowStockCount    int64           `json:"low_stock_count"`
		Sales7d          []dayStat       `json:"sales_7d"`
		RecentOrders     []AdminSummary  `json:"recent_orders"`
		LowStock         []inventory.Row `json:"low_stock"`
	}
	const today = `(now() AT TIME ZONE 'Asia/Bangkok')::date`
	err := h.DB.QueryRow(ctx, `SELECT
		COALESCE((SELECT sum(total_satang) FROM orders WHERE status IN `+salesStatuses+` AND (created_at AT TIME ZONE 'Asia/Bangkok')::date = `+today+`), 0),
		(SELECT count(*) FROM orders WHERE (created_at AT TIME ZONE 'Asia/Bangkok')::date = `+today+`),
		(SELECT count(*) FROM payment_slips WHERE status = 'pending'),
		(SELECT count(*) FROM orders WHERE status IN ('paid', 'confirmed')),
		(SELECT count(*) FROM product_variants v JOIN products p ON p.id = v.product_id WHERE `+inventory.LowCondition+`)`).
		Scan(&out.TodaySalesSatang, &out.TodayOrders, &out.PendingSlips, &out.ToPack, &out.LowStockCount)
	if err != nil {
		return err
	}
	rows, err := h.DB.Query(ctx, `SELECT to_char(d, 'YYYY-MM-DD'),
		COALESCE(sum(o.total_satang) FILTER (WHERE o.status IN `+salesStatuses+`), 0),
		count(o.id) FILTER (WHERE o.status IN `+salesStatuses+`)
		FROM generate_series(`+today+` - 6, `+today+`, interval '1 day') d
		LEFT JOIN orders o ON (o.created_at AT TIME ZONE 'Asia/Bangkok')::date = d::date
		GROUP BY d ORDER BY d`)
	if err != nil {
		return err
	}
	out.Sales7d, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (dayStat, error) {
		var d dayStat
		return d, r.Scan(&d.Date, &d.TotalSatang, &d.Orders)
	})
	if err != nil {
		return err
	}
	rows, err = h.DB.Query(ctx, `SELECT `+adminSummaryCols+` FROM orders o JOIN users u ON u.id = o.user_id ORDER BY o.created_at DESC, o.id DESC LIMIT 10`)
	if err != nil {
		return err
	}
	if out.RecentOrders, err = pgx.CollectRows(rows, scanAdminSummary); err != nil {
		return err
	}
	if out.LowStock, err = inventory.Query(ctx, h.DB, inventory.LowCondition, "ORDER BY v.stock, p.name_en LIMIT 10"); err != nil {
		return err
	}
	return httpx.OK(w, out)
}
