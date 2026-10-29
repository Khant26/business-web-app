package catalog

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

// ---- Categories ----

// AdminCategories handles GET /admin/categories.
func (h *Handler) AdminCategories(w http.ResponseWriter, r *http.Request) error {
	cats, err := queryCategories(r.Context(), h.DB, "")
	if err != nil {
		return err
	}
	return httpx.OK(w, cats)
}

type categoryInput struct {
	Slug      string `json:"slug"`
	NameEN    string `json:"name_en"`
	NameMY    string `json:"name_my"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func (in *categoryInput) validate() error {
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	in.NameEN = strings.TrimSpace(in.NameEN)
	in.NameMY = strings.TrimSpace(in.NameMY)
	f := httpx.Fields{}
	f.Check(!ValidSlug(in.Slug), "slug", "lowercase letters, numbers and dashes only")
	f.Check(in.NameEN == "" || httpx.Len(in.NameEN) > 100, "name_en", "required, max 100 characters")
	f.Check(httpx.Len(in.NameMY) > 100, "name_my", "max 100 characters")
	return f.Err()
}

func loadCategory(ctx context.Context, q db.DBTX, id int64) (Category, error) {
	return scanCategory(q.QueryRow(ctx, categorySelect+" WHERE c.id = $1", id))
}

// SaveCategory handles POST /admin/categories and PUT /admin/categories/{id}.
func (h *Handler) SaveCategory(w http.ResponseWriter, r *http.Request) error {
	var id int64
	if r.PathValue("id") != "" {
		var err error
		if id, err = httpx.PathInt64(r, "id"); err != nil {
			return err
		}
	}
	var in categoryInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	if err := in.validate(); err != nil {
		return err
	}
	ctx := r.Context()
	var c Category
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		action := "category.update"
		if id == 0 {
			action = "category.create"
			if err := tx.QueryRow(ctx, `INSERT INTO categories (slug, name_en, name_my, sort_order, is_active)
				VALUES ($1, $2, $3, $4, $5) RETURNING id`, in.Slug, in.NameEN, in.NameMY, in.SortOrder, in.IsActive).Scan(&id); err != nil {
				return mapWriteErr(err, "slug")
			}
		} else {
			tag, err := tx.Exec(ctx, `UPDATE categories SET slug = $2, name_en = $3, name_my = $4, sort_order = $5, is_active = $6 WHERE id = $1`,
				id, in.Slug, in.NameEN, in.NameMY, in.SortOrder, in.IsActive)
			if err != nil {
				return mapWriteErr(err, "slug")
			}
			if tag.RowsAffected() == 0 {
				return httpx.NotFound()
			}
		}
		if err := audit.Log(ctx, tx, r, action, "category", id, in); err != nil {
			return err
		}
		var err error
		c, err = loadCategory(ctx, tx, id)
		return err
	})
	if err != nil {
		return err
	}
	if r.Method == http.MethodPost {
		return httpx.Created(w, c)
	}
	return httpx.OK(w, c)
}

// DeleteCategory handles DELETE /admin/categories/{id}.
func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	var url string
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var used bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM products WHERE category_id = $1)`, id).Scan(&used); err != nil {
			return err
		}
		if used {
			return httpx.Conflict("in_use", "This category still has products")
		}
		if err := tx.QueryRow(ctx, `DELETE FROM categories WHERE id = $1 RETURNING image_url`, id).Scan(&url); err != nil {
			return err
		}
		return audit.Log(ctx, tx, r, "category.delete", "category", id, nil)
	})
	if err != nil {
		return err
	}
	removePublicFile(h.UploadDir, url)
	return httpx.NoContent(w)
}

// UploadCategoryImage handles POST /admin/categories/{id}/image.
func (h *Handler) UploadCategoryImage(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	old, err := loadCategory(ctx, h.DB, id)
	if err != nil {
		return err
	}
	name, err := httpx.SaveImage(w, r, filepath.Join(h.UploadDir, "public", "categories"))
	if err != nil {
		return err
	}
	url := "/uploads/categories/" + name
	var c Category
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE categories SET image_url = $2 WHERE id = $1`, id, url); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "category.image", "category", id, map[string]any{"url": url}); err != nil {
			return err
		}
		c, err = loadCategory(ctx, tx, id)
		return err
	})
	if err != nil {
		removePublicFile(h.UploadDir, url)
		return err
	}
	removePublicFile(h.UploadDir, old.ImageURL)
	return httpx.OK(w, c)
}

// ---- Events ----

type adminEvent struct {
	Event
	ProductIDs []int64 `json:"product_ids"`
}

func loadAdminEvents(ctx context.Context, q db.DBTX, where string, args ...any) ([]adminEvent, error) {
	evs, err := queryEvents(ctx, q, where, args...)
	if err != nil {
		return nil, err
	}
	out := make([]adminEvent, 0, len(evs))
	for _, e := range evs {
		rows, err := q.Query(ctx, `SELECT product_id FROM event_products WHERE event_id = $1 ORDER BY sort_order, product_id`, e.ID)
		if err != nil {
			return nil, err
		}
		ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
		if err != nil {
			return nil, err
		}
		out = append(out, adminEvent{Event: e, ProductIDs: ids})
	}
	return out, nil
}

// AdminEvents handles GET /admin/events.
func (h *Handler) AdminEvents(w http.ResponseWriter, r *http.Request) error {
	evs, err := loadAdminEvents(r.Context(), h.DB, "")
	if err != nil {
		return err
	}
	return httpx.OK(w, evs)
}

type eventInput struct {
	Slug          string    `json:"slug"`
	TitleEN       string    `json:"title_en"`
	TitleMY       string    `json:"title_my"`
	DescriptionEN string    `json:"description_en"`
	DescriptionMY string    `json:"description_my"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	IsActive      bool      `json:"is_active"`
	SortOrder     int       `json:"sort_order"`
	ProductIDs    []int64   `json:"product_ids"`
}

func (in *eventInput) validate() error {
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	in.TitleEN = strings.TrimSpace(in.TitleEN)
	in.TitleMY = strings.TrimSpace(in.TitleMY)
	f := httpx.Fields{}
	f.Check(!ValidSlug(in.Slug), "slug", "lowercase letters, numbers and dashes only")
	f.Check(in.TitleEN == "" || httpx.Len(in.TitleEN) > 200, "title_en", "required, max 200 characters")
	f.Check(httpx.Len(in.TitleMY) > 200, "title_my", "max 200 characters")
	f.Check(httpx.Len(in.DescriptionEN) > 5000, "description_en", "max 5000 characters")
	f.Check(httpx.Len(in.DescriptionMY) > 5000, "description_my", "max 5000 characters")
	f.Check(in.StartsAt.IsZero(), "starts_at", "required")
	f.Check(in.EndsAt.IsZero() || !in.EndsAt.After(in.StartsAt), "ends_at", "must be after starts_at")
	return f.Err()
}

// SaveEvent handles POST /admin/events and PUT /admin/events/{id}.
func (h *Handler) SaveEvent(w http.ResponseWriter, r *http.Request) error {
	var id int64
	if r.PathValue("id") != "" {
		var err error
		if id, err = httpx.PathInt64(r, "id"); err != nil {
			return err
		}
	}
	var in eventInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	if err := in.validate(); err != nil {
		return err
	}
	ctx := r.Context()
	var out adminEvent
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		action := "event.update"
		if id == 0 {
			action = "event.create"
			if err := tx.QueryRow(ctx, `INSERT INTO events (slug, title_en, title_my, description_en, description_my, starts_at, ends_at, is_active, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`, in.Slug, in.TitleEN, in.TitleMY, in.DescriptionEN, in.DescriptionMY,
				in.StartsAt, in.EndsAt, in.IsActive, in.SortOrder).Scan(&id); err != nil {
				return mapWriteErr(err, "slug")
			}
		} else {
			tag, err := tx.Exec(ctx, `UPDATE events SET slug = $2, title_en = $3, title_my = $4, description_en = $5, description_my = $6,
				starts_at = $7, ends_at = $8, is_active = $9, sort_order = $10 WHERE id = $1`, id, in.Slug, in.TitleEN, in.TitleMY,
				in.DescriptionEN, in.DescriptionMY, in.StartsAt, in.EndsAt, in.IsActive, in.SortOrder)
			if err != nil {
				return mapWriteErr(err, "slug")
			}
			if tag.RowsAffected() == 0 {
				return httpx.NotFound()
			}
		}
		if in.ProductIDs != nil {
			if _, err := tx.Exec(ctx, `DELETE FROM event_products WHERE event_id = $1`, id); err != nil {
				return err
			}
			for i, pid := range uniq(in.ProductIDs) {
				tag, err := tx.Exec(ctx, `INSERT INTO event_products (event_id, product_id, sort_order)
					SELECT $1, id, $3 FROM products WHERE id = $2`, id, pid, i)
				if err != nil {
					return err
				}
				if tag.RowsAffected() == 0 {
					return httpx.Validation(map[string]string{"product_ids": "unknown product " + strconv.FormatInt(pid, 10)})
				}
			}
		}
		if err := audit.Log(ctx, tx, r, action, "event", id, map[string]any{"slug": in.Slug, "product_ids": in.ProductIDs}); err != nil {
			return err
		}
		evs, err := loadAdminEvents(ctx, tx, "id = $1", id)
		if err == nil && len(evs) == 1 {
			out = evs[0]
		}
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

// DeleteEvent handles DELETE /admin/events/{id}.
func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	var url string
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `DELETE FROM events WHERE id = $1 RETURNING banner_url`, id).Scan(&url); err != nil {
			return err
		}
		return audit.Log(ctx, tx, r, "event.delete", "event", id, nil)
	})
	if err != nil {
		return err
	}
	removePublicFile(h.UploadDir, url)
	return httpx.NoContent(w)
}

// UploadEventBanner handles POST /admin/events/{id}/banner.
func (h *Handler) UploadEventBanner(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	var old string
	if err := h.DB.QueryRow(ctx, `SELECT banner_url FROM events WHERE id = $1`, id).Scan(&old); err != nil {
		return err
	}
	name, err := httpx.SaveImage(w, r, filepath.Join(h.UploadDir, "public", "events"))
	if err != nil {
		return err
	}
	url := "/uploads/events/" + name
	var out adminEvent
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE events SET banner_url = $2 WHERE id = $1`, id, url); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "event.banner", "event", id, map[string]any{"url": url}); err != nil {
			return err
		}
		evs, err := loadAdminEvents(ctx, tx, "id = $1", id)
		if err == nil && len(evs) == 1 {
			out = evs[0]
		}
		return err
	})
	if err != nil {
		removePublicFile(h.UploadDir, url)
		return err
	}
	removePublicFile(h.UploadDir, old)
	return httpx.OK(w, out)
}

// ---- Reviews moderation ----

type adminReview struct {
	ID            int64     `json:"id"`
	ProductID     int64     `json:"product_id"`
	ProductNameEN string    `json:"product_name_en"`
	UserEmail     string    `json:"user_email"`
	Rating        int       `json:"rating"`
	Comment       string    `json:"comment"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// AdminReviews handles GET /admin/reviews.
func (h *Handler) AdminReviews(w http.ResponseWriter, r *http.Request) error {
	p := httpx.ParsePage(r)
	status := r.URL.Query().Get("status")
	ctx := r.Context()
	var total int
	if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM reviews WHERE ($1 = '' OR status = $1)`, status).Scan(&total); err != nil {
		return err
	}
	rows, err := h.DB.Query(ctx, `SELECT rv.id, rv.product_id, p.name_en, u.email, rv.rating, rv.comment, rv.status, rv.created_at
		FROM reviews rv JOIN products p ON p.id = rv.product_id JOIN users u ON u.id = rv.user_id
		WHERE ($1 = '' OR rv.status = $1) ORDER BY rv.created_at DESC, rv.id DESC LIMIT $2 OFFSET $3`, status, p.Size, p.Offset())
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (adminReview, error) {
		var a adminReview
		err := row.Scan(&a.ID, &a.ProductID, &a.ProductNameEN, &a.UserEmail, &a.Rating, &a.Comment, &a.Status, &a.CreatedAt)
		return a, err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

// ModerateReview handles POST /admin/reviews/{id}/approve and /reject.
func (h *Handler) ModerateReview(status string) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		id, err := httpx.PathInt64(r, "id")
		if err != nil {
			return err
		}
		ctx := r.Context()
		err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
			var pid int64
			if err := tx.QueryRow(ctx, `UPDATE reviews SET status = $2 WHERE id = $1 RETURNING product_id`, id, status).Scan(&pid); err != nil {
				return err
			}
			if err := RecomputeRating(ctx, tx, pid); err != nil {
				return err
			}
			action := "review.approve"
			if status == "rejected" {
				action = "review.reject"
			}
			return audit.Log(ctx, tx, r, action, "review", id, map[string]any{"product_id": pid})
		})
		if err != nil {
			return err
		}
		return httpx.NoContent(w)
	}
}

// RecomputeRating refreshes products.rating_avg / rating_count from approved reviews.
func RecomputeRating(ctx context.Context, q db.DBTX, productID int64) error {
	_, err := q.Exec(ctx, `UPDATE products SET
		rating_avg = COALESCE((SELECT round(avg(rating)::numeric, 2) FROM reviews WHERE product_id = $1 AND status = 'approved'), 0),
		rating_count = (SELECT count(*) FROM reviews WHERE product_id = $1 AND status = 'approved')
		WHERE id = $1`, productID)
	return err
}
