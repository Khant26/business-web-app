// Package users handles the customer profile, addresses and admin user management.
package users

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/geo"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

// Handler serves user endpoints.
type Handler struct{ DB *pgxpool.Pool }

var (
	phoneRe    = regexp.MustCompile(`^\+?[0-9][0-9 \-]{5,19}$`)
	postcodeRe = regexp.MustCompile(`^[0-9]{5}$`)
)

// ValidPhone reports whether s looks like a phone number.
func ValidPhone(s string) bool { return phoneRe.MatchString(s) }

// UpdateMe handles PATCH /me.
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	var req struct {
		Name  *string `json:"name"`
		Phone *string `json:"phone"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	name, phone := u.Name, u.Phone
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}
	if req.Phone != nil {
		phone = strings.TrimSpace(*req.Phone)
	}
	f := httpx.Fields{}
	f.Check(httpx.Len(name) > 100, "name", "max 100 characters")
	f.Check(phone != "" && !ValidPhone(phone), "phone", "invalid phone number")
	if err := f.Err(); err != nil {
		return err
	}
	out, err := auth.ScanUser(h.DB.QueryRow(r.Context(), `UPDATE users SET name = $2, phone = $3, updated_at = now() WHERE id = $1
		RETURNING `+auth.UserColumns, u.ID, name, phone))
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"user": out})
}

// Address is a saved shipping address.
type Address struct {
	ID            int64  `json:"id"`
	Label         string `json:"label"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	Line1         string `json:"line1"`
	Subdistrict   string `json:"subdistrict"`
	District      string `json:"district"`
	Province      string `json:"province"`
	Postcode      string `json:"postcode"`
	IsDefault     bool   `json:"is_default"`
}

// Validate trims and validates an address (province is canonicalized). prefix is added to field names.
func (a *Address) Validate(prefix string) httpx.Fields {
	a.Label = strings.TrimSpace(a.Label)
	a.RecipientName = strings.TrimSpace(a.RecipientName)
	a.Phone = strings.TrimSpace(a.Phone)
	a.Line1 = strings.TrimSpace(a.Line1)
	a.Subdistrict = strings.TrimSpace(a.Subdistrict)
	a.District = strings.TrimSpace(a.District)
	a.Postcode = strings.TrimSpace(a.Postcode)
	f := httpx.Fields{}
	f.Check(httpx.Len(a.Label) > 50, prefix+"label", "max 50 characters")
	f.Check(a.RecipientName == "" || httpx.Len(a.RecipientName) > 100, prefix+"recipient_name", "required")
	f.Check(!ValidPhone(a.Phone), prefix+"phone", "invalid phone number")
	f.Check(a.Line1 == "" || httpx.Len(a.Line1) > 300, prefix+"line1", "required")
	f.Check(a.Subdistrict == "" || httpx.Len(a.Subdistrict) > 100, prefix+"subdistrict", "required")
	f.Check(a.District == "" || httpx.Len(a.District) > 100, prefix+"district", "required")
	f.Check(!postcodeRe.MatchString(a.Postcode), prefix+"postcode", "must be 5 digits")
	if p, ok := geo.Canonical(a.Province); ok {
		a.Province = p
	} else {
		f.Check(true, prefix+"province", "unknown province")
	}
	return f
}

const addrCols = `id, label, recipient_name, phone, line1, subdistrict, district, province, postcode, is_default`

func scanAddr(r pgx.Row) (Address, error) {
	var a Address
	err := r.Scan(&a.ID, &a.Label, &a.RecipientName, &a.Phone, &a.Line1, &a.Subdistrict, &a.District, &a.Province, &a.Postcode, &a.IsDefault)
	return a, err
}

// GetAddress loads an address owned by userID.
func GetAddress(ctx context.Context, q db.DBTX, userID, id int64) (Address, error) {
	return scanAddr(q.QueryRow(ctx, `SELECT `+addrCols+` FROM addresses WHERE id = $1 AND user_id = $2`, id, userID))
}

// InsertAddress saves a new address (the first address becomes default).
func InsertAddress(ctx context.Context, tx pgx.Tx, userID int64, a Address) (Address, error) {
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM addresses WHERE user_id = $1`, userID).Scan(&count); err != nil {
		return a, err
	}
	if count >= 20 {
		return a, httpx.Conflict("conflict", "You can save up to 20 addresses")
	}
	if count == 0 {
		a.IsDefault = true
	}
	if a.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1`, userID); err != nil {
			return a, err
		}
	}
	return scanAddr(tx.QueryRow(ctx, `INSERT INTO addresses (user_id, label, recipient_name, phone, line1, subdistrict, district, province, postcode, is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING `+addrCols,
		userID, a.Label, a.RecipientName, a.Phone, a.Line1, a.Subdistrict, a.District, a.Province, a.Postcode, a.IsDefault))
}

// Addresses handles GET /me/addresses.
func (h *Handler) Addresses(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	rows, err := h.DB.Query(r.Context(), `SELECT `+addrCols+` FROM addresses WHERE user_id = $1 ORDER BY is_default DESC, created_at, id`, u.ID)
	if err != nil {
		return err
	}
	out, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Address, error) { return scanAddr(r) })
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

// CreateAddress handles POST /me/addresses.
func (h *Handler) CreateAddress(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	var a Address
	if err := httpx.Decode(w, r, &a); err != nil {
		return err
	}
	if err := a.Validate("").Err(); err != nil {
		return err
	}
	ctx := r.Context()
	var out Address
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var err error
		out, err = InsertAddress(ctx, tx, u.ID, a)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.Created(w, out)
}

// UpdateAddress handles PUT /me/addresses/{id}.
func (h *Handler) UpdateAddress(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var a Address
	if err := httpx.Decode(w, r, &a); err != nil {
		return err
	}
	if err := a.Validate("").Err(); err != nil {
		return err
	}
	ctx := r.Context()
	var out Address
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		cur, err := GetAddress(ctx, tx, u.ID, id)
		if err != nil {
			return err
		}
		if cur.IsDefault {
			a.IsDefault = true // keep one default; set another address as default instead
		}
		if a.IsDefault {
			if _, err := tx.Exec(ctx, `UPDATE addresses SET is_default = false WHERE user_id = $1 AND id <> $2`, u.ID, id); err != nil {
				return err
			}
		}
		out, err = scanAddr(tx.QueryRow(ctx, `UPDATE addresses SET label = $3, recipient_name = $4, phone = $5, line1 = $6,
			subdistrict = $7, district = $8, province = $9, postcode = $10, is_default = $11 WHERE id = $1 AND user_id = $2 RETURNING `+addrCols,
			id, u.ID, a.Label, a.RecipientName, a.Phone, a.Line1, a.Subdistrict, a.District, a.Province, a.Postcode, a.IsDefault))
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

// DeleteAddress handles DELETE /me/addresses/{id}.
func (h *Handler) DeleteAddress(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	return httpx.NoContentAfter(w, db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var wasDefault bool
		if err := tx.QueryRow(ctx, `DELETE FROM addresses WHERE id = $1 AND user_id = $2 RETURNING is_default`, id, u.ID).Scan(&wasDefault); err != nil {
			return err
		}
		if wasDefault {
			_, err := tx.Exec(ctx, `UPDATE addresses SET is_default = true WHERE id = (
				SELECT id FROM addresses WHERE user_id = $1 ORDER BY created_at, id LIMIT 1)`, u.ID)
			return err
		}
		return nil
	}))
}

// ---- Admin users ----

type adminUser struct {
	auth.User
	IsActive   bool  `json:"is_active"`
	OrderCount int64 `json:"order_count"`
}

const adminUserSelect = `SELECT u.id, u.email, u.name, u.phone, u.role, u.created_at, u.is_active,
	(SELECT count(*) FROM orders o WHERE o.user_id = u.id) FROM users u`

func scanAdminUser(r pgx.Row) (adminUser, error) {
	var a adminUser
	err := r.Scan(&a.ID, &a.Email, &a.Name, &a.Phone, &a.Role, &a.CreatedAt, &a.IsActive, &a.OrderCount)
	return a, err
}

// AdminList handles GET /admin/users.
func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) error {
	p := httpx.ParsePage(r)
	role := r.URL.Query().Get("role")
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	like := ""
	if q != "" {
		like = db.LikePattern(q)
	}
	where := ` WHERE ($1 = '' OR u.role = $1) AND ($2 = '' OR u.email ILIKE $2 OR u.name ILIKE $2 OR u.phone ILIKE $2)`
	ctx := r.Context()
	var total int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM users u`+where, role, like).Scan(&total); err != nil {
		return err
	}
	rows, err := h.DB.Query(ctx, adminUserSelect+where+` ORDER BY u.created_at DESC, u.id DESC LIMIT $3 OFFSET $4`, role, like, p.Size, p.Offset())
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (adminUser, error) { return scanAdminUser(r) })
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

// AdminCreate handles POST /admin/users (create or promote staff/admin).
func (h *Handler) AdminCreate(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	email, ok := auth.NormalizeEmail(req.Email)
	req.Name = strings.TrimSpace(req.Name)
	f := httpx.Fields{}
	f.Check(!ok, "email", "invalid")
	f.Check(req.Role != "staff" && req.Role != "admin", "role", "must be staff or admin")
	f.Check(httpx.Len(req.Name) > 100, "name", "max 100 characters")
	if err := f.Err(); err != nil {
		return err
	}
	ctx := r.Context()
	var out *auth.User
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var err error
		out, err = auth.ScanUser(tx.QueryRow(ctx, `INSERT INTO users (email, name, role) VALUES ($1, $2, $3)
			ON CONFLICT (email) DO UPDATE SET role = EXCLUDED.role,
				name = CASE WHEN EXCLUDED.name <> '' THEN EXCLUDED.name ELSE users.name END, updated_at = now()
			RETURNING `+auth.UserColumns, email, req.Name, req.Role))
		if err != nil {
			return err
		}
		return audit.Log(ctx, tx, r, "user.grant_role", "user", out.ID, map[string]any{"email": email, "role": req.Role})
	})
	if err != nil {
		return err
	}
	return httpx.Created(w, out)
}

// AdminUpdate handles PATCH /admin/users/{id}.
func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var req struct {
		Role     *string `json:"role"`
		IsActive *bool   `json:"is_active"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	me := auth.UserFrom(r.Context())
	if id == me.ID && ((req.Role != nil && *req.Role != me.Role) || (req.IsActive != nil && !*req.IsActive)) {
		return httpx.Forbidden("You cannot demote or disable your own account")
	}
	if req.Role != nil && *req.Role != "customer" && *req.Role != "staff" && *req.Role != "admin" {
		return httpx.Validation(map[string]string{"role": "must be customer, staff or admin"})
	}
	ctx := r.Context()
	var out adminUser
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE users SET role = COALESCE($2, role), is_active = COALESCE($3, is_active), updated_at = now() WHERE id = $1`,
			id, req.Role, req.IsActive)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.NotFound()
		}
		if req.IsActive != nil && !*req.IsActive {
			if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, id); err != nil {
				return err
			}
		}
		if err := audit.Log(ctx, tx, r, "user.update", "user", id, req); err != nil {
			return err
		}
		out, err = scanAdminUser(tx.QueryRow(ctx, adminUserSelect+` WHERE u.id = $1`, id))
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}
