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
	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

type adminProductRow struct {
	ID             int64     `json:"id"`
	Slug           string    `json:"slug"`
	NameEN         string    `json:"name_en"`
	NameMY         string    `json:"name_my"`
	Status         string    `json:"status"`
	Featured       bool      `json:"featured"`
	CategoryNameEN string    `json:"category_name_en"`
	ImageURL       string    `json:"image_url"`
	VariantCount   int64     `json:"variant_count"`
	TotalStock     int64     `json:"total_stock"`
	MinPriceSatang int64     `json:"min_price_satang"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AdminProducts handles GET /admin/products.
func (h *Handler) AdminProducts(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	p := httpx.ParsePage(r)
	b := &whereBuilder{conds: []string{"true"}}
	if s := strings.TrimSpace(q.Get("q")); s != "" {
		b.add("(p.name_en ILIKE ? OR p.name_my ILIKE ? OR p.slug ILIKE ? OR EXISTS (SELECT 1 FROM product_variants v WHERE v.product_id = p.id AND v.sku ILIKE ?))", db.LikePattern(s))
	}
	if s := q.Get("status"); s != "" {
		b.add("p.status = ?", s)
	}
	if n, err := strconv.ParseInt(q.Get("category"), 10, 64); err == nil {
		b.add("p.category_id = ?", n)
	}
	ctx := r.Context()
	var total int
	if err := h.DB.QueryRow(ctx, "SELECT count(*) FROM products p WHERE "+b.sql(), b.args...).Scan(&total); err != nil {
		return err
	}
	n := len(b.args)
	rows, err := h.DB.Query(ctx, `SELECT p.id, p.slug, p.name_en, p.name_my, p.status, p.featured, COALESCE(c.name_en, ''),
		COALESCE((SELECT url FROM product_images WHERE product_id = p.id ORDER BY sort_order, id LIMIT 1), ''),
		(SELECT count(*) FROM product_variants v WHERE v.product_id = p.id),
		COALESCE((SELECT sum(stock) FROM product_variants v WHERE v.product_id = p.id), 0),
		COALESCE((SELECT min(price_satang) FROM product_variants v WHERE v.product_id = p.id AND v.is_active), 0),
		p.updated_at
		FROM products p LEFT JOIN categories c ON c.id = p.category_id
		WHERE `+b.sql()+` ORDER BY p.updated_at DESC, p.id DESC LIMIT $`+strconv.Itoa(n+1)+` OFFSET $`+strconv.Itoa(n+2),
		append(b.args, p.Size, p.Offset())...)
	if err != nil {
		return err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (adminProductRow, error) {
		var a adminProductRow
		err := row.Scan(&a.ID, &a.Slug, &a.NameEN, &a.NameMY, &a.Status, &a.Featured, &a.CategoryNameEN, &a.ImageURL,
			&a.VariantCount, &a.TotalStock, &a.MinPriceSatang, &a.UpdatedAt)
		return a, err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}

func (h *Handler) loadAdminProduct(ctx context.Context, q db.DBTX, id int64) (*AdminProduct, error) {
	d, status, catID, err := loadDetail(ctx, q, "p.id = $1", true, id)
	if err != nil {
		return nil, err
	}
	rows, err := q.Query(ctx, `SELECT event_id FROM event_products WHERE product_id = $1 ORDER BY event_id`, id)
	if err != nil {
		return nil, err
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return nil, err
	}
	return &AdminProduct{ProductDetail: *d, Status: status, CategoryID: catID, EventIDs: ids}, nil
}

// AdminProduct handles GET /admin/products/{id}.
func (h *Handler) AdminProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	p, err := h.loadAdminProduct(r.Context(), h.DB, id)
	if err != nil {
		return err
	}
	return httpx.OK(w, p)
}

type productInput struct {
	Slug          string   `json:"slug"`
	CategoryID    *int64   `json:"category_id"`
	NameEN        string   `json:"name_en"`
	NameMY        string   `json:"name_my"`
	DescriptionEN string   `json:"description_en"`
	DescriptionMY string   `json:"description_my"`
	Status        string   `json:"status"`
	Featured      bool     `json:"featured"`
	EventIDs      *[]int64 `json:"event_ids"`
}

func (in *productInput) validate() error {
	in.Slug = strings.ToLower(strings.TrimSpace(in.Slug))
	in.NameEN = strings.TrimSpace(in.NameEN)
	in.NameMY = strings.TrimSpace(in.NameMY)
	if in.Status == "" {
		in.Status = "draft"
	}
	f := httpx.Fields{}
	f.Check(!ValidSlug(in.Slug), "slug", "lowercase letters, numbers and dashes only")
	f.Check(in.NameEN == "" || httpx.Len(in.NameEN) > 200, "name_en", "required, max 200 characters")
	f.Check(httpx.Len(in.NameMY) > 200, "name_my", "max 200 characters")
	f.Check(httpx.Len(in.DescriptionEN) > 10000, "description_en", "max 10000 characters")
	f.Check(httpx.Len(in.DescriptionMY) > 10000, "description_my", "max 10000 characters")
	f.Check(in.Status != "draft" && in.Status != "active" && in.Status != "archived", "status", "must be draft, active or archived")
	return f.Err()
}

func setProductEvents(ctx context.Context, tx pgx.Tx, productID int64, ids *[]int64) error {
	if ids == nil {
		return nil
	}
	if _, err := tx.Exec(ctx, `DELETE FROM event_products WHERE product_id = $1`, productID); err != nil {
		return err
	}
	if len(*ids) == 0 {
		return nil
	}
	tag, err := tx.Exec(ctx, `INSERT INTO event_products (event_id, product_id)
		SELECT id, $1 FROM events WHERE id = ANY($2) ON CONFLICT DO NOTHING`, productID, *ids)
	if err != nil {
		return err
	}
	if int(tag.RowsAffected()) != len(uniq(*ids)) {
		return httpx.Validation(map[string]string{"event_ids": "unknown event"})
	}
	return nil
}

func uniq(ids []int64) []int64 {
	seen := map[int64]bool{}
	var out []int64
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func mapWriteErr(err error, field string) error {
	if db.IsUniqueViolation(err) {
		return httpx.Conflict("conflict", field+" already exists").WithFields(map[string]string{field: "already exists"})
	}
	if db.IsForeignKeyViolation(err) {
		return httpx.Validation(map[string]string{"category_id": "unknown category"})
	}
	return err
}

// CreateProduct handles POST /admin/products.
func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) error {
	var in productInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	if err := in.validate(); err != nil {
		return err
	}
	ctx := r.Context()
	var out *AdminProduct
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var id int64
		err := tx.QueryRow(ctx, `INSERT INTO products (slug, category_id, name_en, name_my, description_en, description_my, status, featured)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
			in.Slug, in.CategoryID, in.NameEN, in.NameMY, in.DescriptionEN, in.DescriptionMY, in.Status, in.Featured).Scan(&id)
		if err != nil {
			return mapWriteErr(err, "slug")
		}
		if err := setProductEvents(ctx, tx, id, in.EventIDs); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "product.create", "product", id, map[string]any{"slug": in.Slug, "status": in.Status}); err != nil {
			return err
		}
		out, err = h.loadAdminProduct(ctx, tx, id)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.Created(w, out)
}

// UpdateProduct handles PUT /admin/products/{id}.
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in productInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	if err := in.validate(); err != nil {
		return err
	}
	ctx := r.Context()
	var out *AdminProduct
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE products SET slug = $2, category_id = $3, name_en = $4, name_my = $5,
			description_en = $6, description_my = $7, status = $8, featured = $9, updated_at = now() WHERE id = $1`,
			id, in.Slug, in.CategoryID, in.NameEN, in.NameMY, in.DescriptionEN, in.DescriptionMY, in.Status, in.Featured)
		if err != nil {
			return mapWriteErr(err, "slug")
		}
		if tag.RowsAffected() == 0 {
			return httpx.NotFound()
		}
		if err := setProductEvents(ctx, tx, id, in.EventIDs); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "product.update", "product", id, map[string]any{"slug": in.Slug, "status": in.Status, "featured": in.Featured}); err != nil {
			return err
		}
		out, err = h.loadAdminProduct(ctx, tx, id)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, out)
}

// DeleteProduct handles DELETE /admin/products/{id} (archives).
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	return httpx.NoContentAfter(w, db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `UPDATE products SET status = 'archived', featured = false, updated_at = now() WHERE id = $1`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.NotFound()
		}
		if err := audit.Log(ctx, tx, r, "product.archive", "product", id, nil); err != nil {
			return err
		}
		return nil
	}))
}

// UploadProductImage handles POST /admin/products/{id}/images.
func (h *Handler) UploadProductImage(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	var exists bool
	if err := h.DB.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)`, id).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return httpx.NotFound()
	}
	name, err := httpx.SaveImage(w, r, filepath.Join(h.UploadDir, "public", "products"))
	if err != nil {
		return err
	}
	img := Image{URL: "/uploads/products/" + name}
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO product_images (product_id, url, sort_order)
			VALUES ($1, $2, COALESCE((SELECT max(sort_order) + 1 FROM product_images WHERE product_id = $1), 0))
			RETURNING id, sort_order`, id, img.URL).Scan(&img.ID, &img.SortOrder); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE products SET updated_at = now() WHERE id = $1`, id); err != nil {
			return err
		}
		return audit.Log(ctx, tx, r, "product.image_add", "product", id, map[string]any{"image_id": img.ID, "url": img.URL})
	})
	if err != nil {
		removePublicFile(h.UploadDir, img.URL)
		return err
	}
	return httpx.Created(w, img)
}

// ReorderImages handles PUT /admin/products/{id}/images/order.
func (h *Handler) ReorderImages(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var req struct {
		ImageIDs []int64 `json:"image_ids"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	ctx := r.Context()
	return httpx.NoContentAfter(w, db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		for i, imgID := range req.ImageIDs {
			if _, err := tx.Exec(ctx, `UPDATE product_images SET sort_order = $3 WHERE id = $1 AND product_id = $2`, imgID, id, i); err != nil {
				return err
			}
		}
		if err := audit.Log(ctx, tx, r, "product.image_reorder", "product", id, map[string]any{"image_ids": req.ImageIDs}); err != nil {
			return err
		}
		return nil
	}))
}

// DeleteImage handles DELETE /admin/products/{id}/images/{imageId}.
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	imgID, err := httpx.PathInt64(r, "imageId")
	if err != nil {
		return err
	}
	ctx := r.Context()
	var url string
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `DELETE FROM product_images WHERE id = $1 AND product_id = $2 RETURNING url`, imgID, id).Scan(&url); err != nil {
			return err
		}
		return audit.Log(ctx, tx, r, "product.image_delete", "product", id, map[string]any{"image_id": imgID})
	})
	if err != nil {
		return err
	}
	removePublicFile(h.UploadDir, url)
	return httpx.NoContent(w)
}

type variantInput struct {
	SKU               string `json:"sku"`
	NameEN            string `json:"name_en"`
	NameMY            string `json:"name_my"`
	PriceSatang       int64  `json:"price_satang"`
	CompareAtSatang   *int64 `json:"compare_at_satang"`
	Stock             *int64 `json:"stock"`
	LowStockThreshold int64  `json:"low_stock_threshold"`
	IsActive          bool   `json:"is_active"`
	SortOrder         int    `json:"sort_order"`
}

func (in *variantInput) validate() error {
	in.SKU = strings.ToUpper(strings.TrimSpace(in.SKU))
	in.NameEN = strings.TrimSpace(in.NameEN)
	in.NameMY = strings.TrimSpace(in.NameMY)
	f := httpx.Fields{}
	f.Check(in.SKU == "" || len(in.SKU) > 64, "sku", "required, max 64 characters")
	f.Check(in.NameEN == "" || httpx.Len(in.NameEN) > 100, "name_en", "required, max 100 characters")
	f.Check(httpx.Len(in.NameMY) > 100, "name_my", "max 100 characters")
	f.Check(in.PriceSatang < 0 || in.PriceSatang > 1_000_000_000, "price_satang", "must be 0 or more")
	f.Check(in.CompareAtSatang != nil && (*in.CompareAtSatang < 0 || *in.CompareAtSatang > 1_000_000_000), "compare_at_satang", "must be 0 or more")
	f.Check(in.Stock != nil && (*in.Stock < 0 || *in.Stock > 1_000_000), "stock", "must be 0 or more")
	f.Check(in.LowStockThreshold < 0, "low_stock_threshold", "must be 0 or more")
	return f.Err()
}

func loadVariant(ctx context.Context, q db.DBTX, id int64) (Variant, error) {
	return scanVariant(q.QueryRow(ctx, `SELECT `+variantCols+` FROM product_variants WHERE id = $1`, id), true)
}

// CreateVariant handles POST /admin/products/{id}/variants.
func (h *Handler) CreateVariant(w http.ResponseWriter, r *http.Request) error {
	pid, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in variantInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	if err := in.validate(); err != nil {
		return err
	}
	stock := int64(0)
	if in.Stock != nil {
		stock = *in.Stock
	}
	ctx := r.Context()
	var v Variant
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var id int64
		err := tx.QueryRow(ctx, `INSERT INTO product_variants (product_id, sku, name_en, name_my, price_satang, compare_at_satang,
			stock, low_stock_threshold, is_active, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
			pid, in.SKU, in.NameEN, in.NameMY, in.PriceSatang, in.CompareAtSatang, stock, in.LowStockThreshold, in.IsActive, in.SortOrder).Scan(&id)
		if db.IsForeignKeyViolation(err) {
			return httpx.NotFound()
		}
		if err != nil {
			return mapWriteErr(err, "sku")
		}
		if stock > 0 {
			var uid *int64
			if u := auth.UserFrom(ctx); u != nil {
				uid = &u.ID
			}
			if _, err := tx.Exec(ctx, `INSERT INTO inventory_movements (variant_id, change, reason, note, user_id, stock_after)
				VALUES ($1, $2, 'restock', 'initial stock', $3, $2)`, id, stock, uid); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE products SET updated_at = now() WHERE id = $1`, pid); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "variant.create", "variant", id, map[string]any{"product_id": pid, "sku": in.SKU, "stock": stock}); err != nil {
			return err
		}
		v, err = loadVariant(ctx, tx, id)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.Created(w, v)
}

// UpdateVariant handles PUT /admin/variants/{id} (stock is not changed here).
func (h *Handler) UpdateVariant(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	var in variantInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	in.Stock = nil
	if err := in.validate(); err != nil {
		return err
	}
	ctx := r.Context()
	var v Variant
	err = db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var pid int64
		err := tx.QueryRow(ctx, `UPDATE product_variants SET sku = $2, name_en = $3, name_my = $4, price_satang = $5,
			compare_at_satang = $6, low_stock_threshold = $7, is_active = $8, sort_order = $9 WHERE id = $1 RETURNING product_id`,
			id, in.SKU, in.NameEN, in.NameMY, in.PriceSatang, in.CompareAtSatang, in.LowStockThreshold, in.IsActive, in.SortOrder).Scan(&pid)
		if err != nil {
			return mapWriteErr(err, "sku")
		}
		if _, err := tx.Exec(ctx, `UPDATE products SET updated_at = now() WHERE id = $1`, pid); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, "variant.update", "variant", id, map[string]any{"sku": in.SKU, "price_satang": in.PriceSatang, "is_active": in.IsActive}); err != nil {
			return err
		}
		v, err = loadVariant(ctx, tx, id)
		return err
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, v)
}

// DeleteVariant handles DELETE /admin/variants/{id}.
func (h *Handler) DeleteVariant(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	return httpx.NoContentAfter(w, db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var pid int64
		if err := tx.QueryRow(ctx, `SELECT product_id FROM product_variants WHERE id = $1 FOR UPDATE`, id).Scan(&pid); err != nil {
			return err
		}
		var used bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM order_items WHERE variant_id = $1)`, id).Scan(&used); err != nil {
			return err
		}
		action := "variant.delete"
		if used {
			action = "variant.deactivate"
			if _, err := tx.Exec(ctx, `UPDATE product_variants SET is_active = false WHERE id = $1`, id); err != nil {
				return err
			}
		} else if _, err := tx.Exec(ctx, `DELETE FROM product_variants WHERE id = $1`, id); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE products SET updated_at = now() WHERE id = $1`, pid); err != nil {
			return err
		}
		if err := audit.Log(ctx, tx, r, action, "variant", id, map[string]any{"product_id": pid}); err != nil {
			return err
		}
		return nil
	}))
}
