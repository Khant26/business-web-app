// Command seed inserts idempotent demo data (categories, products, events, zones, coupons, settings).
package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/config"
	"github.com/Khant26/business-web-app/backend/internal/db"
)

type variant struct {
	sku, en, my    string
	price, compare int64 // THB; compare 0 = none
	stock          int
}

type product struct {
	slug, cat, en, my, descEN, descMY string
	featured                          bool
	sold                              int
	variants                          []variant
}

var categories = []struct {
	slug, en, my string
}{
	{"fashion", "Fashion", "ဖက်ရှင်"},
	{"shoes", "Shoes", "ဖိနပ်"},
	{"electronics", "Electronics", "အီလက်ထရောနစ်"},
	{"home-living", "Home & Living", "အိမ်နှင့် နေထိုင်မှု"},
	{"beauty", "Beauty", "အလှအပ"},
	{"food-snacks", "Food & Snacks", "အစားအစာနှင့် မုန့်များ"},
}

var products = []product{
	{"classic-cotton-tshirt", "fashion", "Classic Cotton T-Shirt", "ချည်သား တီရှပ်",
		"Soft 100% cotton crew-neck tee for everyday wear.", "နေ့စဉ်ဝတ်ဆင်ရန် ချည်သား ၁၀၀% ပျော့ပျောင်းသော တီရှပ်။", true, 320,
		[]variant{{"TS-WHT-M", "White / M", "အဖြူ / M", 290, 390, 40}, {"TS-WHT-L", "White / L", "အဖြူ / L", 290, 390, 35},
			{"TS-BLK-M", "Black / M", "အနက် / M", 290, 390, 30}, {"TS-BLK-L", "Black / L", "အနက် / L", 290, 390, 4}}},
	{"linen-button-shirt", "fashion", "Linen Button Shirt", "လင်နင် ကြယ်သီးအင်္ကျီ",
		"Breathable linen shirt, perfect for hot Bangkok days.", "ဘန်ကောက်၏ ပူပြင်းသောနေ့များအတွက် လေဝင်လေထွက်ကောင်းသော လင်နင်အင်္ကျီ။", false, 85,
		[]variant{{"LS-BEG-M", "Beige / M", "နို့နှစ်ရောင် / M", 890, 0, 15}, {"LS-BEG-L", "Beige / L", "နို့နှစ်ရောင် / L", 890, 0, 12}}},
	{"slim-fit-denim-jeans", "fashion", "Slim Fit Denim Jeans", "ကိုယ်ကျ ဂျင်းဘောင်းဘီ",
		"Stretch denim with a modern slim fit.", "ခေတ်မီ ကိုယ်ကျပုံစံ ဆန့်နိုင်သော ဂျင်းဘောင်းဘီ။", true, 140,
		[]variant{{"DJ-30", "Waist 30", "ခါး ၃၀", 1290, 1590, 10}, {"DJ-32", "Waist 32", "ခါး ၃၂", 1290, 1590, 14}, {"DJ-34", "Waist 34", "ခါး ၃၄", 1290, 1590, 6}}},
	{"thai-silk-scarf", "fashion", "Thai Silk Scarf", "ထိုင်းပိုး ပဝါ",
		"Hand-woven Thai silk scarf, a lovely gift.", "လက်ဖြင့်ရက်လုပ်ထားသော ထိုင်းပိုးပဝါ၊ လက်ဆောင်ပေးရန် အကောင်းဆုံး။", false, 60,
		[]variant{{"SS-RED", "Red", "အနီ", 650, 0, 20}, {"SS-BLU", "Blue", "အပြာ", 650, 0, 18}}},
	{"everyday-running-sneakers", "shoes", "Everyday Running Sneakers", "နေ့စဉ်သုံး ပြေးဖိနပ်",
		"Lightweight cushioned sneakers for running and walking.", "ပြေးရန်နှင့် လမ်းလျှောက်ရန် ပေါ့ပါးပြီး ခုံညှိထားသော ဖိနပ်။", true, 210,
		[]variant{{"RS-40", "EU 40", "EU 40", 1890, 2290, 8}, {"RS-41", "EU 41", "EU 41", 1890, 2290, 12},
			{"RS-42", "EU 42", "EU 42", 1890, 2290, 10}, {"RS-43", "EU 43", "EU 43", 1890, 2290, 3}}},
	{"leather-sandals", "shoes", "Leather Sandals", "သားရေ ညှပ်ဖိနပ်",
		"Genuine leather sandals with a comfortable footbed.", "စီးရလွယ်ကူသော ခြေဖဝါးခံပါသည့် စစ်မှန်သော သားရေ ညှပ်ဖိနပ်။", false, 95,
		[]variant{{"LSD-39", "EU 39", "EU 39", 790, 0, 9}, {"LSD-40", "EU 40", "EU 40", 790, 0, 11}, {"LSD-41", "EU 41", "EU 41", 790, 0, 7}}},
	{"wireless-earbuds-pro", "electronics", "Wireless Earbuds Pro", "ကြိုးမဲ့ နားကြပ် Pro",
		"Active noise cancelling earbuds with 30-hour battery.", "ဆူညံသံ ဖယ်ရှားနိုင်ပြီး ဘက်ထရီ ၃၀ နာရီ ခံသော ကြိုးမဲ့ နားကြပ်။", true, 450,
		[]variant{{"EB-BLK", "Black", "အနက်", 1490, 1990, 25}, {"EB-WHT", "White", "အဖြူ", 1490, 1990, 18}}},
	{"power-bank-20000", "electronics", "20000mAh Power Bank", "20000mAh အားသွင်းဘဏ်",
		"Fast-charging power bank with USB-C PD.", "USB-C PD ဖြင့် အမြန်အားသွင်းနိုင်သော အားသွင်းဘဏ်။", false, 380,
		[]variant{{"PB-20K", "Standard", "ပုံမှန်", 690, 0, 50}}},
	{"smart-fitness-watch", "electronics", "Smart Fitness Watch", "စမတ် ကျန်းမာရေး နာရီ",
		"Heart rate, sleep and step tracking with 7-day battery.", "နှလုံးခုန်နှုန်း၊ အိပ်စက်မှုနှင့် ခြေလှမ်းများကို မှတ်တမ်းတင်ပေးပြီး ဘက်ထရီ ၇ ရက် ခံသည်။", true, 175,
		[]variant{{"FW-BLK", "Black", "အနက်", 2490, 0, 12}, {"FW-PNK", "Pink", "ပန်းရောင်", 2490, 0, 0}}},
	{"usb-c-charger-65w", "electronics", "USB-C Fast Charger 65W", "USB-C အမြန်အားသွင်းကိရိယာ 65W",
		"Compact GaN charger for phones and laptops.", "ဖုန်းနှင့် လက်ပ်တော့များအတွက် သေးငယ်သော GaN အားသွင်းကိရိယာ။", false, 260,
		[]variant{{"CH-65W", "Standard", "ပုံမှန်", 590, 790, 40}}},
	{"ceramic-coffee-mug-set", "home-living", "Ceramic Coffee Mug Set", "ကြွေ ကော်ဖီခွက် အစုံ",
		"Hand-glazed ceramic mugs, 350 ml each.", "လက်ဖြင့် ကြွေသုတ်ထားသော ကော်ဖီခွက်များ၊ တစ်ခွက်လျှင် ၃၅၀ မီလီလီတာ။", false, 130,
		[]variant{{"MUG-2", "Set of 2", "၂ ခွက်ပါ အစုံ", 350, 0, 30}, {"MUG-4", "Set of 4", "၄ ခွက်ပါ အစုံ", 650, 0, 15}}},
	{"bamboo-bed-sheet-set", "home-living", "Bamboo Bed Sheet Set", "ဝါးမျှင် အိပ်ရာခင်း အစုံ",
		"Cool, silky bamboo fibre sheets with two pillowcases.", "ခေါင်းအုံးစွပ် ၂ ခုပါသော အေးမြချောမွေ့သည့် ဝါးမျှင် အိပ်ရာခင်း။", true, 70,
		[]variant{{"BS-Q", "Queen", "Queen အရွယ်", 1590, 0, 10}, {"BS-K", "King", "King အရွယ်", 1890, 0, 8}}},
	{"aroma-diffuser", "home-living", "Aroma Diffuser", "အနံ့မွှေး ပျံ့နှံ့စက်",
		"Ultrasonic diffuser with soft LED light.", "အလင်းရောင် ညင်သာသော LED ပါသည့် အနံ့မွှေး ပျံ့နှံ့စက်။", false, 110,
		[]variant{{"AD-STD", "Standard", "ပုံမှန်", 990, 1290, 22}}},
	{"vitamin-c-serum", "beauty", "Vitamin C Serum 30ml", "ဗီတာမင်စီ ဆီရမ် 30ml",
		"Brightening serum with 15% vitamin C.", "ဗီတာမင်စီ ၁၅% ပါဝင်သော အသားအရေ ကြည်လင်စေသည့် ဆီရမ်။", true, 290,
		[]variant{{"VC-30", "30 ml", "30 မီလီလီတာ", 450, 0, 35}}},
	{"sunscreen-spf50", "beauty", "Sunscreen SPF50 PA++++", "နေလောင်ကာ ခရင်မ် SPF50 PA++++",
		"Light, non-greasy daily sunscreen.", "ပေါ့ပါးပြီး အဆီမပြန်သော နေ့စဉ်သုံး နေလောင်ကာ ခရင်မ်။", false, 410,
		[]variant{{"SC-50", "50 ml", "50 မီလီလီတာ", 390, 0, 45}, {"SC-100", "100 ml", "100 မီလီလီတာ", 690, 0, 20}}},
	{"dried-mango", "food-snacks", "Dried Mango", "သရက်သီးခြောက်",
		"Sweet Thai dried mango, no preservatives.", "ထိန်းသိမ်းဆေး မပါသော ချိုမြိန်သည့် ထိုင်း သရက်သီးခြောက်။", false, 520,
		[]variant{{"DM-200", "200 g", "200 ဂရမ်", 120, 0, 80}, {"DM-500", "500 g", "500 ဂရမ်", 270, 0, 40}}},
}

var runningEventProducts = []string{"classic-cotton-tshirt", "slim-fit-denim-jeans", "everyday-running-sneakers",
	"wireless-earbuds-pro", "usb-c-charger-65w", "aroma-diffuser"}

var upcomingEventProducts = []string{"bamboo-bed-sheet-set", "smart-fitness-watch", "vitamin-c-serum"}

var centralEast = []string{"Phra Nakhon Si Ayutthaya", "Ang Thong", "Lopburi", "Sing Buri", "Chai Nat", "Saraburi",
	"Chon Buri", "Rayong", "Chanthaburi", "Trat", "Chachoengsao", "Prachin Buri", "Nakhon Nayok", "Sa Kaeo",
	"Suphan Buri", "Ratchaburi", "Kanchanaburi", "Samut Songkhram", "Phetchaburi", "Prachuap Khiri Khan"}

var bangkokVicinity = []string{"Bangkok", "Nonthaburi", "Pathum Thani", "Samut Prakan", "Samut Sakhon", "Nakhon Pathom"}

func main() {
	time.Local = time.UTC
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}
	if err := auth.EnsureAdmin(ctx, pool, cfg.AdminEmail); err != nil {
		slog.Error("admin", "err", err)
		os.Exit(1)
	}
	if err := db.Tx(ctx, pool, func(tx pgx.Tx) error { return seed(ctx, tx) }); err != nil {
		slog.Error("seed failed", "err", err)
		os.Exit(1)
	}
	slog.Info("seed complete", "categories", len(categories), "products", len(products))
}

func seed(ctx context.Context, tx pgx.Tx) error {
	catIDs := map[string]int64{}
	for i, c := range categories {
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO categories (slug, name_en, name_my, sort_order) VALUES ($1, $2, $3, $4)
			ON CONFLICT (slug) DO UPDATE SET name_en = EXCLUDED.name_en, name_my = EXCLUDED.name_my RETURNING id`,
			c.slug, c.en, c.my, i).Scan(&id); err != nil {
			return err
		}
		catIDs[c.slug] = id
	}
	prodIDs := map[string]int64{}
	for i, p := range products {
		var id int64
		// Stagger created_at so "new arrivals" has a stable order.
		created := time.Now().Add(-time.Duration(len(products)-i) * 36 * time.Hour)
		if err := tx.QueryRow(ctx, `INSERT INTO products (slug, category_id, name_en, name_my, description_en, description_my,
			status, featured, sold_count, created_at) VALUES ($1, $2, $3, $4, $5, $6, 'active', $7, $8, $9)
			ON CONFLICT (slug) DO UPDATE SET name_en = EXCLUDED.name_en, name_my = EXCLUDED.name_my RETURNING id`,
			p.slug, catIDs[p.cat], p.en, p.my, p.descEN, p.descMY, p.featured, p.sold, created).Scan(&id); err != nil {
			return err
		}
		prodIDs[p.slug] = id
		for j, v := range p.variants {
			var compare *int64
			if v.compare > 0 {
				c := v.compare * 100
				compare = &c
			}
			var vid int64
			err := tx.QueryRow(ctx, `INSERT INTO product_variants (product_id, sku, name_en, name_my, price_satang, compare_at_satang, stock, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (sku) DO NOTHING RETURNING id`,
				id, v.sku, v.en, v.my, v.price*100, compare, v.stock, j).Scan(&vid)
			if err == pgx.ErrNoRows {
				continue // already seeded; keep current stock
			}
			if err != nil {
				return err
			}
			if v.stock > 0 {
				if _, err := tx.Exec(ctx, `INSERT INTO inventory_movements (variant_id, change, reason, note, stock_after)
					VALUES ($1, $2, 'restock', 'seed', $2)`, vid, v.stock); err != nil {
					return err
				}
			}
		}
	}

	now := time.Now().UTC()
	events := []struct {
		slug, en, my, descEN, descMY string
		start, end                   time.Time
		products                     []string
	}{
		{"11-11-sale", "11.11 Sale", "၁၁.၁၁ အထူးလျှော့စျေး", "Big discounts on best sellers — this week only.",
			"ရောင်းအားအကောင်းဆုံး ပစ္စည်းများကို ဤတစ်ပတ်အတွင်းသာ အထူးလျှော့စျေးဖြင့် ရယူပါ။",
			now.Add(-48 * time.Hour), now.Add(12 * 24 * time.Hour), runningEventProducts},
		{"12-12-year-end-sale", "12.12 Year-End Sale", "၁၂.၁၂ နှစ်ကုန် လျှော့စျေး", "Celebrate the year end with special prices.",
			"နှစ်ကုန်ကို အထူးဈေးနှုန်းများဖြင့် ကျင်းပလိုက်ပါ။",
			now.Add(20 * 24 * time.Hour), now.Add(27 * 24 * time.Hour), upcomingEventProducts},
	}
	for i, e := range events {
		var id int64
		if err := tx.QueryRow(ctx, `INSERT INTO events (slug, title_en, title_my, description_en, description_my, starts_at, ends_at, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (slug) DO UPDATE SET starts_at = EXCLUDED.starts_at, ends_at = EXCLUDED.ends_at, is_active = true RETURNING id`,
			e.slug, e.en, e.my, e.descEN, e.descMY, e.start, e.end, i).Scan(&id); err != nil {
			return err
		}
		for j, slug := range e.products {
			if _, err := tx.Exec(ctx, `INSERT INTO event_products (event_id, product_id, sort_order) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
				id, prodIDs[slug], j); err != nil {
				return err
			}
		}
	}

	ptr := func(v int64) *int64 { return &v }
	zones := []struct {
		en, my    string
		fee       int64
		freeOver  *int64
		isDefault bool
		provinces []string
	}{
		{"Bangkok & Vicinity", "ဘန်ကောက်နှင့် အနီးတဝိုက်", 4000, ptr(100000), false, bangkokVicinity},
		{"Central & East", "အလယ်ပိုင်းနှင့် အရှေ့ပိုင်း", 6000, ptr(150000), false, centralEast},
		{"North / Northeast / South", "မြောက်ပိုင်း / အရှေ့မြောက်ပိုင်း / တောင်ပိုင်း", 8000, ptr(200000), true, nil},
	}
	for i, z := range zones {
		var id int64
		err := tx.QueryRow(ctx, `SELECT id FROM shipping_zones WHERE name_en = $1`, z.en).Scan(&id)
		if err == pgx.ErrNoRows {
			if z.isDefault {
				var hasDefault bool
				if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM shipping_zones WHERE is_default)`).Scan(&hasDefault); err != nil {
					return err
				}
				if hasDefault {
					continue
				}
			}
			if err := tx.QueryRow(ctx, `INSERT INTO shipping_zones (name_en, name_my, fee_satang, free_over_satang, is_default, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`, z.en, z.my, z.fee, z.freeOver, z.isDefault, i).Scan(&id); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		for _, p := range z.provinces {
			if _, err := tx.Exec(ctx, `INSERT INTO shipping_zone_provinces (zone_id, province) VALUES ($1, $2) ON CONFLICT DO NOTHING`, id, p); err != nil {
				return err
			}
		}
	}

	if _, err := tx.Exec(ctx, `INSERT INTO coupons (code, type, value, min_order_satang, max_discount_satang, per_user_limit)
		VALUES ('WELCOME10', 'percent', 10, 30000, 20000, 1), ('SAVE50', 'fixed', 5000, 50000, NULL, 3)
		ON CONFLICT (code) DO NOTHING`); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `UPDATE settings SET value = CASE key
			WHEN 'promptpay_id' THEN '"0812345678"'::jsonb
			WHEN 'promptpay_name' THEN '"Sawyun Shop"'::jsonb
			WHEN 'support_email' THEN '"support@example.com"'::jsonb
			WHEN 'support_phone' THEN '"02-123-4567"'::jsonb END
		WHERE key IN ('promptpay_id', 'promptpay_name', 'support_email', 'support_phone') AND value = '""'::jsonb`)
	return err
}
