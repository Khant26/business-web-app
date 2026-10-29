package catalog

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

// Home handles GET /home.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	events, err := queryEvents(ctx, h.DB, runningEvents)
	if err != nil {
		return err
	}
	cats, err := queryCategories(ctx, h.DB, "c.is_active")
	if err != nil {
		return err
	}
	popular, err := QueryCards(ctx, h.DB, "p.status = 'active'", "ORDER BY p.sold_count DESC, p.id DESC LIMIT 8")
	if err != nil {
		return err
	}
	featured, err := QueryCards(ctx, h.DB, "p.status = 'active' AND p.featured", "ORDER BY p.updated_at DESC, p.id DESC LIMIT 8")
	if err != nil {
		return err
	}
	newest, err := QueryCards(ctx, h.DB, "p.status = 'active'", "ORDER BY p.created_at DESC, p.id DESC LIMIT 8")
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{
		"events": events, "categories": cats, "popular": popular, "featured": featured, "new_arrivals": newest,
	})
}

// Categories handles GET /categories.
func (h *Handler) Categories(w http.ResponseWriter, r *http.Request) error {
	cats, err := queryCategories(r.Context(), h.DB, "c.is_active")
	if err != nil {
		return err
	}
	return httpx.OK(w, cats)
}

// Category handles GET /categories/{slug}.
func (h *Handler) Category(w http.ResponseWriter, r *http.Request) error {
	c, err := scanCategory(h.DB.QueryRow(r.Context(), categorySelect+" WHERE c.slug = $1 AND c.is_active", r.PathValue("slug")))
	if err != nil {
		return err
	}
	return httpx.OK(w, c)
}

type whereBuilder struct {
	conds []string
	args  []any
}

// add appends cond with "?" replaced by the next positional parameter.
func (b *whereBuilder) add(cond string, v any) {
	b.args = append(b.args, v)
	b.conds = append(b.conds, strings.ReplaceAll(cond, "?", "$"+strconv.Itoa(len(b.args))))
}

func (b *whereBuilder) sql() string { return strings.Join(b.conds, " AND ") }

// Products handles GET /products.
func (h *Handler) Products(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	p := httpx.ParsePage(r)
	b := &whereBuilder{conds: []string{"p.status = 'active'"}}
	if s := strings.TrimSpace(q.Get("q")); s != "" {
		b.add("(p.name_en ILIKE ? OR p.name_my ILIKE ? OR p.description_en ILIKE ?)", db.LikePattern(s))
	}
	if s := q.Get("category"); s != "" {
		b.add("c.slug = ?", s)
	}
	if s := q.Get("event"); s != "" {
		b.add("EXISTS (SELECT 1 FROM event_products ep JOIN events e ON e.id = ep.event_id WHERE ep.product_id = p.id AND e.slug = ? AND e.is_active)", s)
	}
	if n, err := strconv.ParseInt(q.Get("min_price"), 10, 64); err == nil {
		b.add("mv.price_satang >= ?", n)
	}
	if n, err := strconv.ParseInt(q.Get("max_price"), 10, 64); err == nil {
		b.add("mv.price_satang <= ?", n)
	}
	if q.Get("in_stock") == "1" || q.Get("in_stock") == "true" {
		b.conds = append(b.conds, "COALESCE(st.total, 0) > 0")
	}
	order := "p.sold_count DESC, p.id DESC"
	switch q.Get("sort") {
	case "newest":
		order = "p.created_at DESC, p.id DESC"
	case "price_asc":
		order = "mv.price_satang ASC NULLS LAST, p.id DESC"
	case "price_desc":
		order = "mv.price_satang DESC NULLS LAST, p.id DESC"
	case "rating":
		order = "p.rating_avg DESC, p.rating_count DESC, p.id DESC"
	}
	var total int
	if err := h.DB.QueryRow(r.Context(), "SELECT count(*)"+cardFrom+" WHERE "+b.sql(), b.args...).Scan(&total); err != nil {
		return err
	}
	n := len(b.args)
	items, err := QueryCards(r.Context(), h.DB, b.sql(),
		"ORDER BY "+order+" LIMIT $"+strconv.Itoa(n+1)+" OFFSET $"+strconv.Itoa(n+2),
		append(b.args, p.Size, p.Offset())...)
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

// Product handles GET /products/{slug}.
func (h *Handler) Product(w http.ResponseWriter, r *http.Request) error {
	d, _, _, err := loadDetail(r.Context(), h.DB, "p.slug = $1 AND p.status = 'active'", false, r.PathValue("slug"))
	if err != nil {
		return err
	}
	return httpx.OK(w, d)
}

// Events handles GET /events.
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) error {
	evs, err := queryEvents(r.Context(), h.DB, runningEvents)
	if err != nil {
		return err
	}
	return httpx.OK(w, evs)
}

// EventDetail handles GET /events/{slug}.
func (h *Handler) EventDetail(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	e, err := scanEvent(h.DB.QueryRow(ctx, eventSelect+" WHERE slug = $1 AND is_active", r.PathValue("slug")))
	if err != nil {
		return err
	}
	products, err := QueryCards(ctx, h.DB,
		"p.status = 'active' AND EXISTS (SELECT 1 FROM event_products ep WHERE ep.product_id = p.id AND ep.event_id = $1)",
		"ORDER BY (SELECT sort_order FROM event_products ep WHERE ep.product_id = p.id AND ep.event_id = $1), p.id", e.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, struct {
		Event
		Products []ProductCard `json:"products"`
	}{e, products})
}

type review struct {
	ID        int64     `json:"id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	UserName  string    `json:"user_name"`
	CreatedAt time.Time `json:"created_at"`
}

// displayName returns the user's name or a masked email.
func displayName(name, email string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}
	local, _, _ := strings.Cut(email, "@")
	if local == "" {
		return "Customer"
	}
	return string([]rune(local)[:1]) + "***"
}

// Reviews handles GET /products/{slug}/reviews.
func (h *Handler) Reviews(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	p := httpx.ParsePage(r)
	var pid int64
	if err := h.DB.QueryRow(ctx, `SELECT id FROM products WHERE slug = $1 AND status = 'active'`, r.PathValue("slug")).Scan(&pid); err != nil {
		return err
	}
	var total int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM reviews WHERE product_id = $1 AND status = 'approved'`, pid).Scan(&total); err != nil {
		return err
	}
	rows, err := h.DB.Query(ctx, `SELECT r.id, r.rating, r.comment, u.name, u.email, r.created_at
		FROM reviews r JOIN users u ON u.id = r.user_id
		WHERE r.product_id = $1 AND r.status = 'approved' ORDER BY r.created_at DESC, r.id DESC LIMIT $2 OFFSET $3`,
		pid, p.Size, p.Offset())
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (review, error) {
		var v review
		var name, email string
		err := row.Scan(&v.ID, &v.Rating, &v.Comment, &name, &email, &v.CreatedAt)
		v.UserName = displayName(name, email)
		return v, err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

// CreateReview handles POST /products/{slug}/reviews.
func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	u := auth.UserFrom(ctx)
	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	req.Comment = strings.TrimSpace(req.Comment)
	f := httpx.Fields{}
	f.Check(req.Rating < 1 || req.Rating > 5, "rating", "must be between 1 and 5")
	f.Check(httpx.Len(req.Comment) > 1000, "comment", "max 1000 characters")
	if err := f.Err(); err != nil {
		return err
	}
	var pid int64
	if err := h.DB.QueryRow(ctx, `SELECT id FROM products WHERE slug = $1 AND status <> 'archived'`, r.PathValue("slug")).Scan(&pid); err != nil {
		return err
	}
	var orderID int64
	err := h.DB.QueryRow(ctx, `SELECT o.id FROM orders o JOIN order_items oi ON oi.order_id = o.id
		WHERE o.user_id = $1 AND o.status = 'delivered' AND oi.product_id = $2 ORDER BY o.delivered_at DESC LIMIT 1`,
		u.ID, pid).Scan(&orderID)
	if err == pgx.ErrNoRows {
		return httpx.E(http.StatusForbidden, "not_purchased", "You can review products from delivered orders only")
	} else if err != nil {
		return err
	}
	var id int64
	err = h.DB.QueryRow(ctx, `INSERT INTO reviews (product_id, user_id, order_id, rating, comment) VALUES ($1, $2, $3, $4, $5)
		RETURNING id`, pid, u.ID, orderID, req.Rating, req.Comment).Scan(&id)
	if db.IsUniqueViolation(err) {
		return httpx.Conflict("already_reviewed", "You have already reviewed this product")
	} else if err != nil {
		return err
	}
	return httpx.Created(w, map[string]int64{"id": id})
}

// Wishlist handles GET /me/wishlist.
func (h *Handler) Wishlist(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	items, err := QueryCards(r.Context(), h.DB,
		"p.status = 'active' AND EXISTS (SELECT 1 FROM wishlists wl WHERE wl.product_id = p.id AND wl.user_id = $1)",
		"ORDER BY (SELECT created_at FROM wishlists wl WHERE wl.product_id = p.id AND wl.user_id = $1) DESC", u.ID)
	if err != nil {
		return err
	}
	return httpx.OK(w, items)
}

// WishlistIDs handles GET /me/wishlist/ids.
func (h *Handler) WishlistIDs(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	rows, err := h.DB.Query(r.Context(), `SELECT product_id FROM wishlists WHERE user_id = $1 ORDER BY created_at DESC`, u.ID)
	if err != nil {
		return err
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return err
	}
	return httpx.OK(w, ids)
}

// WishlistAdd handles PUT /me/wishlist/{productId}.
func (h *Handler) WishlistAdd(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	pid, err := httpx.PathInt64(r, "productId")
	if err != nil {
		return err
	}
	tag, err := h.DB.Exec(r.Context(), `INSERT INTO wishlists (user_id, product_id)
		SELECT $1, id FROM products WHERE id = $2 AND status = 'active' ON CONFLICT DO NOTHING`, u.ID, pid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := h.DB.QueryRow(r.Context(), `SELECT EXISTS (SELECT 1 FROM wishlists WHERE user_id = $1 AND product_id = $2)`, u.ID, pid).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return httpx.NotFound()
		}
	}
	return httpx.NoContent(w)
}

// WishlistRemove handles DELETE /me/wishlist/{productId}.
func (h *Handler) WishlistRemove(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	pid, err := httpx.PathInt64(r, "productId")
	if err != nil {
		return err
	}
	if _, err := h.DB.Exec(r.Context(), `DELETE FROM wishlists WHERE user_id = $1 AND product_id = $2`, u.ID, pid); err != nil {
		return err
	}
	return httpx.NoContent(w)
}
