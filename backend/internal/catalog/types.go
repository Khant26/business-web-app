// Package catalog serves categories, products, events, reviews and wishlists.
package catalog

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/db"
)

// Handler serves catalog endpoints.
type Handler struct {
	DB        *pgxpool.Pool
	UploadDir string
}

// CategoryRef is the category summary embedded in product cards.
type CategoryRef struct {
	Slug   string `json:"slug"`
	NameEN string `json:"name_en"`
	NameMY string `json:"name_my"`
}

// ProductCard is the product list item.
type ProductCard struct {
	ID              int64        `json:"id"`
	Slug            string       `json:"slug"`
	NameEN          string       `json:"name_en"`
	NameMY          string       `json:"name_my"`
	ImageURL        string       `json:"image_url"`
	PriceSatang     int64        `json:"price_satang"`
	CompareAtSatang *int64       `json:"compare_at_satang"`
	RatingAvg       float64      `json:"rating_avg"`
	RatingCount     int64        `json:"rating_count"`
	SoldCount       int64        `json:"sold_count"`
	InStock         bool         `json:"in_stock"`
	Featured        bool         `json:"featured"`
	Category        *CategoryRef `json:"category"`
}

// Image is a product image.
type Image struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

// Variant is a product variant. LowStockThreshold is only set for admin responses.
type Variant struct {
	ID                int64  `json:"id"`
	SKU               string `json:"sku"`
	NameEN            string `json:"name_en"`
	NameMY            string `json:"name_my"`
	PriceSatang       int64  `json:"price_satang"`
	CompareAtSatang   *int64 `json:"compare_at_satang"`
	Stock             int64  `json:"stock"`
	IsActive          bool   `json:"is_active"`
	SortOrder         int    `json:"sort_order"`
	LowStockThreshold *int64 `json:"low_stock_threshold,omitempty"`
}

// ProductDetail is the product page payload.
type ProductDetail struct {
	ProductCard
	DescriptionEN string    `json:"description_en"`
	DescriptionMY string    `json:"description_my"`
	Images        []Image   `json:"images"`
	Variants      []Variant `json:"variants"`
}

// AdminProduct is the admin product payload.
type AdminProduct struct {
	ProductDetail
	Status     string  `json:"status"`
	CategoryID *int64  `json:"category_id"`
	EventIDs   []int64 `json:"event_ids"`
}

// Category is a product category.
type Category struct {
	ID           int64  `json:"id"`
	Slug         string `json:"slug"`
	NameEN       string `json:"name_en"`
	NameMY       string `json:"name_my"`
	ImageURL     string `json:"image_url"`
	SortOrder    int    `json:"sort_order"`
	IsActive     bool   `json:"is_active"`
	ProductCount int64  `json:"product_count"`
}

// Event is a promotional event.
type Event struct {
	ID            int64     `json:"id"`
	Slug          string    `json:"slug"`
	TitleEN       string    `json:"title_en"`
	TitleMY       string    `json:"title_my"`
	DescriptionEN string    `json:"description_en"`
	DescriptionMY string    `json:"description_my"`
	BannerURL     string    `json:"banner_url"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	IsActive      bool      `json:"is_active"`
	SortOrder     int       `json:"sort_order"`
}

const cardSelect = `SELECT p.id, p.slug, p.name_en, p.name_my, COALESCE(img.url, ''),
	COALESCE(mv.price_satang, 0), mv.compare_at_satang, p.rating_avg::float8, p.rating_count, p.sold_count,
	COALESCE(st.total, 0) > 0, p.featured, c.slug, c.name_en, c.name_my`

const cardFrom = ` FROM products p
	LEFT JOIN categories c ON c.id = p.category_id
	LEFT JOIN LATERAL (SELECT url FROM product_images WHERE product_id = p.id ORDER BY sort_order, id LIMIT 1) img ON true
	LEFT JOIN LATERAL (SELECT price_satang, compare_at_satang FROM product_variants
		WHERE product_id = p.id AND is_active ORDER BY price_satang, sort_order, id LIMIT 1) mv ON true
	LEFT JOIN LATERAL (SELECT SUM(stock) AS total FROM product_variants WHERE product_id = p.id AND is_active) st ON true`

func scanCard(row pgx.Row) (ProductCard, error) {
	var c ProductCard
	var cs, cen, cmy *string
	err := row.Scan(&c.ID, &c.Slug, &c.NameEN, &c.NameMY, &c.ImageURL, &c.PriceSatang, &c.CompareAtSatang,
		&c.RatingAvg, &c.RatingCount, &c.SoldCount, &c.InStock, &c.Featured, &cs, &cen, &cmy)
	if cs != nil {
		c.Category = &CategoryRef{Slug: *cs, NameEN: *cen, NameMY: *cmy}
	}
	return c, err
}

// QueryCards returns product cards matching where (SQL after WHERE) with tail (ORDER/LIMIT).
func QueryCards(ctx context.Context, q db.DBTX, where, tail string, args ...any) ([]ProductCard, error) {
	sql := cardSelect + cardFrom
	if where != "" {
		sql += " WHERE " + where
	}
	rows, err := q.Query(ctx, sql+" "+tail, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (ProductCard, error) { return scanCard(r) })
}

const categorySelect = `SELECT c.id, c.slug, c.name_en, c.name_my, c.image_url, c.sort_order, c.is_active,
	(SELECT count(*) FROM products p WHERE p.category_id = c.id AND p.status = 'active') FROM categories c`

func scanCategory(r pgx.Row) (Category, error) {
	var c Category
	err := r.Scan(&c.ID, &c.Slug, &c.NameEN, &c.NameMY, &c.ImageURL, &c.SortOrder, &c.IsActive, &c.ProductCount)
	return c, err
}

func queryCategories(ctx context.Context, q db.DBTX, where string, args ...any) ([]Category, error) {
	sql := categorySelect
	if where != "" {
		sql += " WHERE " + where
	}
	rows, err := q.Query(ctx, sql+" ORDER BY c.sort_order, c.id", args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Category, error) { return scanCategory(r) })
}

const eventSelect = `SELECT id, slug, title_en, title_my, description_en, description_my, banner_url,
	starts_at, ends_at, is_active, sort_order FROM events`

func scanEvent(r pgx.Row) (Event, error) {
	var e Event
	err := r.Scan(&e.ID, &e.Slug, &e.TitleEN, &e.TitleMY, &e.DescriptionEN, &e.DescriptionMY, &e.BannerURL,
		&e.StartsAt, &e.EndsAt, &e.IsActive, &e.SortOrder)
	return e, err
}

func queryEvents(ctx context.Context, q db.DBTX, where string, args ...any) ([]Event, error) {
	sql := eventSelect
	if where != "" {
		sql += " WHERE " + where
	}
	rows, err := q.Query(ctx, sql+" ORDER BY sort_order, starts_at, id", args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Event, error) { return scanEvent(r) })
}

const runningEvents = "is_active AND starts_at <= now() AND ends_at > now()"

func scanVariant(r pgx.Row, admin bool) (Variant, error) {
	var v Variant
	var low int64
	err := r.Scan(&v.ID, &v.SKU, &v.NameEN, &v.NameMY, &v.PriceSatang, &v.CompareAtSatang, &v.Stock,
		&v.IsActive, &v.SortOrder, &low)
	if admin {
		v.LowStockThreshold = &low
	}
	return v, err
}

const variantCols = `id, sku, name_en, name_my, price_satang, compare_at_satang, stock, is_active, sort_order, low_stock_threshold`

// loadDetail loads a product by where clause (e.g. "p.slug = $1").
func loadDetail(ctx context.Context, q db.DBTX, where string, admin bool, arg any) (*ProductDetail, string, *int64, error) {
	sql := cardSelect + `, p.description_en, p.description_my, p.status, p.category_id` + cardFrom + " WHERE " + where
	var d ProductDetail
	var cs, cen, cmy *string
	var status string
	var catID *int64
	c := &d.ProductCard
	err := q.QueryRow(ctx, sql, arg).Scan(&c.ID, &c.Slug, &c.NameEN, &c.NameMY, &c.ImageURL, &c.PriceSatang,
		&c.CompareAtSatang, &c.RatingAvg, &c.RatingCount, &c.SoldCount, &c.InStock, &c.Featured, &cs, &cen, &cmy,
		&d.DescriptionEN, &d.DescriptionMY, &status, &catID)
	if err != nil {
		return nil, "", nil, err
	}
	if cs != nil {
		c.Category = &CategoryRef{Slug: *cs, NameEN: *cen, NameMY: *cmy}
	}
	rows, err := q.Query(ctx, `SELECT id, url, sort_order FROM product_images WHERE product_id = $1 ORDER BY sort_order, id`, c.ID)
	if err != nil {
		return nil, "", nil, err
	}
	d.Images, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Image, error) {
		var i Image
		return i, r.Scan(&i.ID, &i.URL, &i.SortOrder)
	})
	if err != nil {
		return nil, "", nil, err
	}
	vsql := `SELECT ` + variantCols + ` FROM product_variants WHERE product_id = $1`
	if !admin {
		vsql += " AND is_active"
	}
	rows, err = q.Query(ctx, vsql+" ORDER BY sort_order, id", c.ID)
	if err != nil {
		return nil, "", nil, err
	}
	d.Variants, err = pgx.CollectRows(rows, func(r pgx.CollectableRow) (Variant, error) { return scanVariant(r, admin) })
	if err != nil {
		return nil, "", nil, err
	}
	return &d, status, catID, nil
}

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// ValidSlug reports whether s is a lowercase URL slug (max 100 chars).
func ValidSlug(s string) bool { return len(s) <= 100 && slugRe.MatchString(s) }

// removePublicFile deletes a file referenced by a /uploads/... URL under uploadDir/public.
func removePublicFile(uploadDir, url string) {
	rel, ok := strings.CutPrefix(url, "/uploads/")
	if !ok || strings.Contains(rel, "..") || rel == "" {
		return
	}
	_ = os.Remove(filepath.Join(uploadDir, "public", filepath.FromSlash(rel)))
}
