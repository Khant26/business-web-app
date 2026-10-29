package orders

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/coupons"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/geo"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/pricing"
	"github.com/Khant26/business-web-app/backend/internal/settings"
	"github.com/Khant26/business-web-app/backend/internal/shipping"
)

// Handler serves order endpoints.
type Handler struct {
	DB        *pgxpool.Pool
	UploadDir string
}

// ItemReq is one requested cart item.
type ItemReq struct {
	VariantID int64 `json:"variant_id"`
	Quantity  int64 `json:"quantity"`
}

const maxLines = 50

// normalizeItems merges duplicate variants, validates quantities and sorts by variant id.
func normalizeItems(items []ItemReq) ([]ItemReq, error) {
	if len(items) == 0 {
		return nil, httpx.E(http.StatusUnprocessableEntity, "empty_cart", "Your cart is empty")
	}
	m := map[int64]int64{}
	for _, it := range items {
		if it.VariantID <= 0 || it.Quantity <= 0 || it.Quantity > 999 {
			return nil, httpx.Validation(map[string]string{"items": "each item needs a variant_id and a quantity between 1 and 999"})
		}
		m[it.VariantID] += it.Quantity
	}
	if len(m) > maxLines {
		return nil, httpx.Validation(map[string]string{"items": "too many different items"})
	}
	out := make([]ItemReq, 0, len(m))
	for id, q := range m {
		out = append(out, ItemReq{VariantID: id, Quantity: min(q, 999)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].VariantID < out[j].VariantID })
	return out, nil
}

// QuoteLine is a priced cart line.
type QuoteLine struct {
	VariantID       int64  `json:"variant_id"`
	ProductSlug     string `json:"product_slug"`
	ProductNameEN   string `json:"product_name_en"`
	ProductNameMY   string `json:"product_name_my"`
	VariantNameEN   string `json:"variant_name_en"`
	VariantNameMY   string `json:"variant_name_my"`
	ImageURL        string `json:"image_url"`
	UnitPriceSatang int64  `json:"unit_price_satang"`
	Quantity        int64  `json:"quantity"`
	LineTotalSatang int64  `json:"line_total_satang"`
	Available       int64  `json:"available"`

	productID int64
	sku       string
}

// Problem describes an unavailable cart line.
type Problem struct {
	VariantID int64  `json:"variant_id"`
	Code      string `json:"code"`
	Available int64  `json:"available"`
}

// resolveLines loads variants for items (locking them when forUpdate) and reports problems.
func resolveLines(ctx context.Context, q db.DBTX, items []ItemReq, forUpdate bool) ([]QuoteLine, []Problem, error) {
	ids := make([]int64, len(items))
	for i, it := range items {
		ids[i] = it.VariantID
	}
	sql := `SELECT v.id, p.id, p.slug, p.name_en, p.name_my, v.name_en, v.name_my, v.sku,
		COALESCE((SELECT url FROM product_images i WHERE i.product_id = p.id ORDER BY sort_order, id LIMIT 1), ''),
		v.price_satang, v.stock, v.is_active AND p.status = 'active'
		FROM product_variants v JOIN products p ON p.id = v.product_id WHERE v.id = ANY($1) ORDER BY v.id`
	if forUpdate {
		sql += " FOR UPDATE OF v"
	}
	rows, err := q.Query(ctx, sql, ids)
	if err != nil {
		return nil, nil, err
	}
	type found struct {
		line   QuoteLine
		stock  int64
		active bool
	}
	byID := map[int64]found{}
	for rows.Next() {
		var f found
		l := &f.line
		if err := rows.Scan(&l.VariantID, &l.productID, &l.ProductSlug, &l.ProductNameEN, &l.ProductNameMY, &l.VariantNameEN,
			&l.VariantNameMY, &l.sku, &l.ImageURL, &l.UnitPriceSatang, &f.stock, &f.active); err != nil {
			rows.Close()
			return nil, nil, err
		}
		byID[l.VariantID] = f
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	lines := []QuoteLine{}
	problems := []Problem{}
	for _, it := range items {
		f, ok := byID[it.VariantID]
		if !ok {
			problems = append(problems, Problem{VariantID: it.VariantID, Code: "unavailable"})
			continue
		}
		l := f.line
		l.Quantity = it.Quantity
		l.LineTotalSatang = l.UnitPriceSatang * l.Quantity
		if f.active {
			l.Available = f.stock
		}
		switch {
		case !f.active:
			problems = append(problems, Problem{VariantID: l.VariantID, Code: "unavailable"})
		case f.stock <= 0:
			problems = append(problems, Problem{VariantID: l.VariantID, Code: "out_of_stock"})
		case f.stock < l.Quantity:
			problems = append(problems, Problem{VariantID: l.VariantID, Code: "insufficient_stock", Available: f.stock})
		}
		lines = append(lines, l)
	}
	return lines, problems, nil
}

// ValidateCart handles POST /cart/validate.
func (h *Handler) ValidateCart(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Items []ItemReq `json:"items"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	if len(req.Items) == 0 {
		return httpx.OK(w, map[string]any{"lines": []QuoteLine{}, "problems": []Problem{}})
	}
	items, err := normalizeItems(req.Items)
	if err != nil {
		return err
	}
	lines, problems, err := resolveLines(r.Context(), h.DB, items, false)
	if err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"lines": lines, "problems": problems})
}

// CouponResult is the quote coupon status.
type CouponResult struct {
	Code        string `json:"code"`
	Valid       bool   `json:"valid"`
	MessageCode string `json:"message_code"`
}

// ZoneInfo is the quote shipping zone summary.
type ZoneInfo struct {
	NameEN         string `json:"name_en"`
	NameMY         string `json:"name_my"`
	FreeOverSatang *int64 `json:"free_over_satang"`
}

// Quote is the checkout quote.
type Quote struct {
	Lines          []QuoteLine   `json:"lines"`
	SubtotalSatang int64         `json:"subtotal_satang"`
	DiscountSatang int64         `json:"discount_satang"`
	ShippingSatang int64         `json:"shipping_satang"`
	VATRateBP      int64         `json:"vat_rate_bp"`
	VATSatang      int64         `json:"vat_satang"`
	TotalSatang    int64         `json:"total_satang"`
	Coupon         *CouponResult `json:"coupon"`
	Zone           *ZoneInfo     `json:"zone"`
	Problems       []Problem     `json:"problems"`
	CODAllowed     bool          `json:"cod_allowed"`

	couponID *int64
	result   pricing.Result
}

// buildQuote prices items for user/province. When forUpdate, variants and coupon rows are locked.
func buildQuote(ctx context.Context, q db.DBTX, userID int64, items []ItemReq, province, couponCode string, forUpdate bool) (*Quote, error) {
	lines, problems, err := resolveLines(ctx, q, items, forUpdate)
	if err != nil {
		return nil, err
	}
	st, err := settings.Load(ctx, q)
	if err != nil {
		return nil, err
	}
	bad := map[int64]bool{}
	for _, p := range problems {
		bad[p.VariantID] = true
	}
	var pl []pricing.Line
	var subtotal int64
	for _, l := range lines {
		if !bad[l.VariantID] {
			pl = append(pl, pricing.Line{UnitPrice: l.UnitPriceSatang, Qty: l.Quantity})
			subtotal += l.LineTotalSatang
		}
	}
	quote := &Quote{Lines: lines, Problems: problems, VATRateBP: st.VATRateBP}
	in := pricing.Input{Lines: pl, VATRateBP: st.VATRateBP, CODMax: st.CODMaxSatang}
	if province != "" {
		z, err := shipping.ZoneFor(ctx, q, province)
		if err != nil {
			return nil, err
		}
		if z != nil {
			in.Zone = &pricing.Zone{Fee: z.FeeSatang, FreeOver: z.FreeOverSatang}
			quote.Zone = &ZoneInfo{NameEN: z.NameEN, NameMY: z.NameMY, FreeOverSatang: z.FreeOverSatang}
		}
	}
	if code := coupons.NormalizeCode(couponCode); code != "" {
		c, used, err := coupons.Lookup(ctx, q, code, userID, forUpdate)
		if err != nil {
			return nil, err
		}
		var pc *pricing.Coupon
		if c != nil {
			pc = c.ToPricing()
		}
		msg := pricing.ValidateCoupon(pc, subtotal, used, time.Now())
		quote.Coupon = &CouponResult{Code: code, Valid: msg == "", MessageCode: msg}
		if msg == "" {
			in.Coupon = pc
			quote.couponID = &c.ID
		}
	}
	res := pricing.Compute(in)
	quote.result = res
	quote.SubtotalSatang, quote.DiscountSatang, quote.ShippingSatang = res.Subtotal, res.Discount, res.Shipping
	quote.VATSatang, quote.TotalSatang, quote.CODAllowed = res.VAT, res.Total, res.CODAllowed
	return quote, nil
}

// QuoteHandler handles POST /checkout/quote.
func (h *Handler) QuoteHandler(w http.ResponseWriter, r *http.Request) error {
	u := auth.UserFrom(r.Context())
	var req struct {
		Items      []ItemReq `json:"items"`
		Province   string    `json:"province"`
		CouponCode string    `json:"coupon_code"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	items, err := normalizeItems(req.Items)
	if err != nil {
		return err
	}
	province := ""
	if req.Province != "" {
		p, ok := geo.Canonical(req.Province)
		if !ok {
			return httpx.Validation(map[string]string{"province": "unknown province"})
		}
		province = p
	}
	q, err := buildQuote(r.Context(), h.DB, u.ID, items, province, req.CouponCode, false)
	if err != nil {
		return err
	}
	return httpx.OK(w, q)
}

func problemFields(ps []Problem) map[string]string {
	f := map[string]string{}
	for _, p := range ps {
		f[strconv.FormatInt(p.VariantID, 10)] = p.Code
	}
	return f
}
