// Package coupons manages discount coupons.
package coupons

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/pricing"
)

// Coupon is the admin coupon shape.
type Coupon struct {
	ID                int64      `json:"id"`
	Code              string     `json:"code"`
	Type              string     `json:"type"`
	Value             int64      `json:"value"`
	MinOrderSatang    int64      `json:"min_order_satang"`
	MaxDiscountSatang *int64     `json:"max_discount_satang"`
	StartsAt          *time.Time `json:"starts_at"`
	EndsAt            *time.Time `json:"ends_at"`
	UsageLimit        *int64     `json:"usage_limit"`
	PerUserLimit      int64      `json:"per_user_limit"`
	UsedCount         int64      `json:"used_count"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
}

const couponSelect = `SELECT id, code, type, value, min_order_satang, max_discount_satang, starts_at, ends_at,
	usage_limit, per_user_limit, used_count, is_active, created_at FROM coupons`

func scan(r pgx.Row) (Coupon, error) {
	var c Coupon
	err := r.Scan(&c.ID, &c.Code, &c.Type, &c.Value, &c.MinOrderSatang, &c.MaxDiscountSatang, &c.StartsAt, &c.EndsAt,
		&c.UsageLimit, &c.PerUserLimit, &c.UsedCount, &c.IsActive, &c.CreatedAt)
	return c, err
}

// ToPricing converts to the pricing coupon.
func (c *Coupon) ToPricing() *pricing.Coupon {
	return &pricing.Coupon{Code: c.Code, Type: c.Type, Value: c.Value, MinOrder: c.MinOrderSatang, MaxDiscount: c.MaxDiscountSatang,
		StartsAt: c.StartsAt, EndsAt: c.EndsAt, UsageLimit: c.UsageLimit, PerUserLimit: c.PerUserLimit, UsedCount: c.UsedCount, IsActive: c.IsActive}
}

// Lookup finds a coupon by code (optionally locking it) and the user's redemption count.
// It returns nil when the code does not exist.
func Lookup(ctx context.Context, q db.DBTX, code string, userID int64, forUpdate bool) (*Coupon, int64, error) {
	code = NormalizeCode(code)
	sql := couponSelect + " WHERE code = $1"
	if forUpdate {
		sql += " FOR UPDATE"
	}
	c, err := scan(q.QueryRow(ctx, sql, code))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	var n int64
	if userID > 0 {
		if err := q.QueryRow(ctx, `SELECT count(*) FROM coupon_redemptions WHERE coupon_id = $1 AND user_id = $2`, c.ID, userID).Scan(&n); err != nil {
			return nil, 0, err
		}
	}
	return &c, n, nil
}

// NormalizeCode trims and upper-cases a coupon code.
func NormalizeCode(s string) string { return strings.ToUpper(strings.TrimSpace(s)) }

// Handler serves admin coupon endpoints.
type Handler struct{ DB *pgxpool.Pool }

// List handles GET /admin/coupons.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	rows, err := h.DB.Query(r.Context(), couponSelect+" ORDER BY created_at DESC, id DESC")
	if err != nil {
		return err
	}
	out, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Coupon, error) { return scan(r) })
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

var codeRe = regexp.MustCompile(`^[A-Z0-9_-]{3,32}$`)

type input struct {
	Code              string     `json:"code"`
	Type              string     `json:"type"`
	Value             int64      `json:"value"`
	MinOrderSatang    int64      `json:"min_order_satang"`
	MaxDiscountSatang *int64     `json:"max_discount_satang"`
	StartsAt          *time.Time `json:"starts_at"`
	EndsAt            *time.Time `json:"ends_at"`
	UsageLimit        *int64     `json:"usage_limit"`
	PerUserLimit      *int64     `json:"per_user_limit"`
	IsActive          *bool      `json:"is_active"`
}

func (in *input) validate() error {
	in.Code = NormalizeCode(in.Code)
	if in.PerUserLimit == nil {
		one := int64(1)
		in.PerUserLimit = &one
	}
	if in.IsActive == nil {
		t := true
		in.IsActive = &t
	}
	f := httpx.Fields{}
	f.Check(!codeRe.MatchString(in.Code), "code", "3-32 letters, numbers, dash or underscore")
	f.Check(in.Type != "percent" && in.Type != "fixed", "type", "must be percent or fixed")
	f.Check(in.Value <= 0, "value", "must be greater than 0")
	f.Check(in.Type == "percent" && in.Value > 100, "value", "percent must be 1-100")
	f.Check(in.MinOrderSatang < 0, "min_order_satang", "must be 0 or more")
	f.Check(in.MaxDiscountSatang != nil && *in.MaxDiscountSatang < 0, "max_discount_satang", "must be 0 or more")
	f.Check(in.UsageLimit != nil && *in.UsageLimit < 0, "usage_limit", "must be 0 or more")
	f.Check(*in.PerUserLimit < 0, "per_user_limit", "must be 0 or more")
	f.Check(in.StartsAt != nil && in.EndsAt != nil && !in.EndsAt.After(*in.StartsAt), "ends_at", "must be after starts_at")
	return f.Err()
}

// Save handles POST /admin/coupons and PUT /admin/coupons/{id}.
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) error {
	var id int64
	if r.PathValue("id") != "" {
		var err error
		if id, err = httpx.PathInt64(r, "id"); err != nil {
			return err
		}
	}
	var in input
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	if err := in.validate(); err != nil {
		return err
	}
	ctx := r.Context()
	var out Coupon
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		action := "coupon.update"
		var err error
		if id == 0 {
			action = "coupon.create"
			err = tx.QueryRow(ctx, `INSERT INTO coupons (code, type, value, min_order_satang, max_discount_satang, starts_at, ends_at,
				usage_limit, per_user_limit, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
				in.Code, in.Type, in.Value, in.MinOrderSatang, in.MaxDiscountSatang, in.StartsAt, in.EndsAt, in.UsageLimit,
				*in.PerUserLimit, *in.IsActive).Scan(&id)
		} else {
			err = tx.QueryRow(ctx, `UPDATE coupons SET code = $2, type = $3, value = $4, min_order_satang = $5, max_discount_satang = $6,
				starts_at = $7, ends_at = $8, usage_limit = $9, per_user_limit = $10, is_active = $11 WHERE id = $1 RETURNING id`,
				id, in.Code, in.Type, in.Value, in.MinOrderSatang, in.MaxDiscountSatang, in.StartsAt, in.EndsAt, in.UsageLimit,
				*in.PerUserLimit, *in.IsActive).Scan(&id)
		}
		if db.IsUniqueViolation(err) {
			return httpx.Conflict("conflict", "Coupon code already exists").WithFields(map[string]string{"code": "already exists"})
		}
		if err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, action, "coupon", id, in); err != nil {
			return err
		}
		out, err = scan(tx.QueryRow(ctx, couponSelect+" WHERE id = $1", id))
		return err
	})
	if err != nil {
		return err
	}
	if r.Method == http.MethodPost {
		return httpx.Created(w, out)
	}
	return httpx.OK(w, out)
}

// Delete handles DELETE /admin/coupons/{id}: deactivates used coupons, deletes unused ones.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	return httpx.NoContentAfter(w, db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var used bool
		if err := tx.QueryRow(ctx, `SELECT used_count > 0 OR EXISTS (SELECT 1 FROM orders WHERE coupon_id = $1)
			FROM coupons WHERE id = $1 FOR UPDATE`, id).Scan(&used); err != nil {
			return err
		}
		action := "coupon.delete"
		if used {
			action = "coupon.deactivate"
			if _, err := tx.Exec(ctx, `UPDATE coupons SET is_active = false WHERE id = $1`, id); err != nil {
				return err
			}
		} else if _, err := tx.Exec(ctx, `DELETE FROM coupons WHERE id = $1`, id); err != nil {
			return err
		}
		return audit.Log(ctx, tx, r, action, "coupon", id, nil)
	}))
}
