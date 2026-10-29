package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Khant26/business-web-app/backend/internal/config"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/orders"
)

type captureMailer struct {
	mu    sync.Mutex
	codes map[string]string
}

func (m *captureMailer) SendOTP(_ context.Context, to, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[to] = code
	return nil
}

// TestIntegrationCheckoutAndExpiry runs only when TEST_DATABASE_URL points at a disposable PostgreSQL database.
func TestIntegrationCheckoutAndExpiry(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}

	suffix := fmt.Sprint(time.Now().UnixNano())
	var variantID, couponID int64
	err = pool.QueryRow(ctx, `WITH p AS (INSERT INTO products (slug, name_en, name_my, status) VALUES ($1, 'IT Product', '', 'active') RETURNING id)
		INSERT INTO product_variants (product_id, sku, name_en, name_my, price_satang, stock) SELECT id, $2, 'Default', '', 50000, 10 FROM p RETURNING id`,
		"it-"+suffix, "IT-"+suffix).Scan(&variantID)
	if err != nil {
		t.Fatal(err)
	}
	code := "IT" + suffix[len(suffix)-8:]
	if err := pool.QueryRow(ctx, `INSERT INTO coupons (code, type, value, usage_limit) VALUES ($1, 'fixed', 1000, 5) RETURNING id`, code).Scan(&couponID); err != nil {
		t.Fatal(err)
	}

	mailer := &captureMailer{codes: map[string]string{}}
	cfg := &config.Config{AppEnv: "development", UploadDir: t.TempDir(), OTPSecret: "test-secret", SessionTTL: time.Hour,
		SessionIdle: time.Hour, SessionMaxPerUser: 5, WebDir: t.TempDir()}
	ts := httptest.NewServer(New(ctx, cfg, pool, mailer))
	defer ts.Close()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}
	call := func(method, path string, body any, out any) int {
		var buf bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&buf).Encode(body)
		}
		req, _ := http.NewRequest(method, ts.URL+"/api/v1"+path, &buf)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Requested-With", "sawyun")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if out != nil {
			_ = json.NewDecoder(resp.Body).Decode(out)
		}
		return resp.StatusCode
	}

	email := "it-" + suffix + "@example.com"
	if s := call("POST", "/auth/otp/request", map[string]string{"email": email}, nil); s != 204 {
		t.Fatalf("otp request: %d", s)
	}
	if s := call("POST", "/auth/otp/verify", map[string]string{"email": email, "code": mailer.codes[email]}, nil); s != 200 {
		t.Fatalf("otp verify: %d", s)
	}
	var order struct {
		Code   string `json:"code"`
		Status string `json:"status"`
		Total  int64  `json:"total_satang"`
	}
	s := call("POST", "/orders", map[string]any{
		"items":          []map[string]int64{{"variant_id": variantID, "quantity": 3}},
		"address":        map[string]string{"recipient_name": "IT", "phone": "0812345678", "line1": "1 Road", "subdistrict": "A", "district": "B", "province": "Bangkok", "postcode": "10110"},
		"payment_method": "promptpay", "coupon_code": code,
	}, &order)
	if s != 201 || order.Status != "pending_payment" {
		t.Fatalf("place order: %d %+v", s, order)
	}
	assertState := func(wantStock, wantUsed int64) {
		t.Helper()
		var stock, used int64
		if err := pool.QueryRow(ctx, `SELECT stock FROM product_variants WHERE id = $1`, variantID).Scan(&stock); err != nil {
			t.Fatal(err)
		}
		if err := pool.QueryRow(ctx, `SELECT used_count FROM coupons WHERE id = $1`, couponID).Scan(&used); err != nil {
			t.Fatal(err)
		}
		if stock != wantStock || used != wantUsed {
			t.Fatalf("stock=%d used=%d, want %d/%d", stock, used, wantStock, wantUsed)
		}
	}
	assertState(7, 1)

	if _, err := pool.Exec(ctx, `UPDATE orders SET payment_due_at = now() - interval '1 minute' WHERE code = $1`, order.Code); err != nil {
		t.Fatal(err)
	}
	if _, err := orders.ExpireDue(ctx, pool); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM orders WHERE code = $1`, order.Code).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "expired" {
		t.Fatalf("status = %s, want expired", status)
	}
	assertState(10, 0)
	var moves int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM inventory_movements WHERE variant_id = $1 AND reason = 'expire'`, variantID).Scan(&moves)
	if moves != 1 {
		t.Fatalf("expected 1 expire movement, got %d", moves)
	}
}
