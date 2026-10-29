// Package inventory handles stock levels and inventory movements.
package inventory

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

// Row is an inventory list row.
type Row struct {
	VariantID         int64  `json:"variant_id"`
	SKU               string `json:"sku"`
	ProductID         int64  `json:"product_id"`
	ProductNameEN     string `json:"product_name_en"`
	VariantNameEN     string `json:"variant_name_en"`
	Stock             int64  `json:"stock"`
	LowStockThreshold int64  `json:"low_stock_threshold"`
	IsLow             bool   `json:"is_low"`
	ImageURL          string `json:"image_url"`
}

const rowSelect = `SELECT v.id, v.sku, p.id, p.name_en, v.name_en, v.stock, v.low_stock_threshold, v.stock <= v.low_stock_threshold,
	COALESCE((SELECT url FROM product_images i WHERE i.product_id = p.id ORDER BY sort_order, id LIMIT 1), '')
	FROM product_variants v JOIN products p ON p.id = v.product_id`

// LowCondition selects low-stock active variants of non-archived products.
const LowCondition = `v.is_active AND p.status <> 'archived' AND v.stock <= v.low_stock_threshold`

func scanRow(r pgx.Row) (Row, error) {
	var x Row
	err := r.Scan(&x.VariantID, &x.SKU, &x.ProductID, &x.ProductNameEN, &x.VariantNameEN, &x.Stock, &x.LowStockThreshold, &x.IsLow, &x.ImageURL)
	return x, err
}

// Query returns inventory rows for a where clause and tail (ORDER/LIMIT).
func Query(ctx context.Context, q db.DBTX, where, tail string, args ...any) ([]Row, error) {
	rows, err := q.Query(ctx, rowSelect+" WHERE "+where+" "+tail, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Row, error) { return scanRow(r) })
}

// Move changes a variant's stock (row must already be locked or the update is atomic) and records a movement.
// It returns the new stock; a result below zero is rejected by the CHECK constraint.
func Move(ctx context.Context, tx pgx.Tx, variantID, change int64, reason, note string, orderID, userID *int64) (int64, error) {
	var after int64
	if err := tx.QueryRow(ctx, `UPDATE product_variants SET stock = stock + $2 WHERE id = $1 RETURNING stock`, variantID, change).Scan(&after); err != nil {
		return 0, err
	}
	_, err := tx.Exec(ctx, `INSERT INTO inventory_movements (variant_id, change, reason, note, order_id, user_id, stock_after)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`, variantID, change, reason, note, orderID, userID, after)
	return after, err
}

// Handler serves inventory admin endpoints.
type Handler struct{ DB *pgxpool.Pool }

// List handles GET /admin/inventory.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	p := httpx.ParsePage(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	like := ""
	if q != "" {
		like = db.LikePattern(q)
	}
	where := `p.status <> 'archived' AND ($1 = '' OR v.sku ILIKE $1 OR p.name_en ILIKE $1 OR v.name_en ILIKE $1)`
	if r.URL.Query().Get("low") == "1" || r.URL.Query().Get("low") == "true" {
		where += " AND " + LowCondition
	}
	ctx := r.Context()
	var total int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM product_variants v JOIN products p ON p.id = v.product_id WHERE `+where, like).Scan(&total); err != nil {
		return err
	}
	items, err := Query(ctx, h.DB, where, `ORDER BY (v.stock <= v.low_stock_threshold) DESC, v.stock, p.name_en, v.sort_order LIMIT $2 OFFSET $3`,
		like, p.Size, p.Offset())
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

// Adjust handles POST /admin/inventory/{variantId}/adjust.
func (h *Handler) Adjust(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "variantId")
	if err != nil {
		return err
	}
	var req struct {
		Change int64  `json:"change"`
		Reason string `json:"reason"`
		Note   string `json:"note"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	req.Note = strings.TrimSpace(req.Note)
	f := httpx.Fields{}
	f.Check(req.Change == 0 || req.Change > 1_000_000 || req.Change < -1_000_000, "change", "must be a non-zero number")
	f.Check(req.Reason != "restock" && req.Reason != "adjust" && req.Reason != "damage", "reason", "must be restock, adjust or damage")
	f.Check(httpx.Len(req.Note) > 500, "note", "max 500 characters")
	if err := f.Err(); err != nil {
		return err
	}
	ctx := r.Context()
	u := auth.UserFrom(ctx)
	var out Row
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var stock int64
		if err := tx.QueryRow(ctx, `SELECT stock FROM product_variants WHERE id = $1 FOR UPDATE`, id).Scan(&stock); err != nil {
			return err
		}
		if stock+req.Change < 0 {
			return httpx.Validation(map[string]string{"change": "stock cannot go below 0 (current " + strconv.FormatInt(stock, 10) + ")"})
		}
		after, err := Move(ctx, tx, id, req.Change, req.Reason, req.Note, nil, &u.ID)
		if err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "inventory.adjust", "variant", id,
			map[string]any{"change": req.Change, "reason": req.Reason, "stock_after": after}); err != nil {
			return err
		}
		out, err = scanRow(tx.QueryRow(ctx, rowSelect+" WHERE v.id = $1", id))
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

type movement struct {
	ID            int64     `json:"id"`
	VariantID     int64     `json:"variant_id"`
	SKU           string    `json:"sku"`
	ProductNameEN string    `json:"product_name_en"`
	Change        int64     `json:"change"`
	Reason        string    `json:"reason"`
	Note          string    `json:"note"`
	OrderCode     string    `json:"order_code"`
	UserEmail     string    `json:"user_email"`
	StockAfter    int64     `json:"stock_after"`
	CreatedAt     time.Time `json:"created_at"`
}

// Movements handles GET /admin/inventory/movements.
func (h *Handler) Movements(w http.ResponseWriter, r *http.Request) error {
	p := httpx.ParsePage(r)
	var vid int64
	if n, err := strconv.ParseInt(r.URL.Query().Get("variant_id"), 10, 64); err == nil {
		vid = n
	}
	ctx := r.Context()
	var total int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM inventory_movements WHERE ($1 = 0 OR variant_id = $1)`, vid).Scan(&total); err != nil {
		return err
	}
	rows, err := h.DB.Query(ctx, `SELECT m.id, m.variant_id, v.sku, p.name_en, m.change, m.reason, m.note,
		COALESCE(o.code, ''), COALESCE(u.email, ''), m.stock_after, m.created_at
		FROM inventory_movements m JOIN product_variants v ON v.id = m.variant_id JOIN products p ON p.id = v.product_id
		LEFT JOIN orders o ON o.id = m.order_id LEFT JOIN users u ON u.id = m.user_id
		WHERE ($1 = 0 OR m.variant_id = $1) ORDER BY m.created_at DESC, m.id DESC LIMIT $2 OFFSET $3`, vid, p.Size, p.Offset())
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (movement, error) {
		var m movement
		err := r.Scan(&m.ID, &m.VariantID, &m.SKU, &m.ProductNameEN, &m.Change, &m.Reason, &m.Note, &m.OrderCode, &m.UserEmail, &m.StockAfter, &m.CreatedAt)
		return m, err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}
