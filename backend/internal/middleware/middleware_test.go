package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Khant26/business-web-app/backend/internal/ratelimit"
)

var ok = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

func TestCSRF(t *testing.T) {
	h := CSRF(ok)
	cases := []struct {
		method, path, header string
		want                 int
	}{
		{"GET", "/api/v1/products", "", 200},
		{"POST", "/api/v1/orders", "", 403},
		{"POST", "/api/v1/orders", "other", 403},
		{"POST", "/api/v1/orders", "sawyun", 200},
		{"DELETE", "/api/v1/me/addresses/1", "", 403},
	}
	for _, c := range cases {
		r := httptest.NewRequest(c.method, c.path, nil)
		if c.header != "" {
			r.Header.Set("X-Requested-With", c.header)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != c.want {
			t.Errorf("%s %s (%q) = %d, want %d", c.method, c.path, c.header, w.Code, c.want)
		}
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "10.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1")
	if got := clientIP(r, false); got != "10.0.0.1" {
		t.Errorf("untrusted proxy: %s", got)
	}
	if got := clientIP(r, true); got != "203.0.113.9" {
		t.Errorf("trusted proxy: %s", got)
	}
}

func TestRateLimitIP(t *testing.T) {
	h := ClientIP(false)(RateLimitIP(ratelimit.New(2, time.Minute))(ok))
	codes := []int{}
	for i := 0; i < 3; i++ {
		r := httptest.NewRequest("GET", "/api/v1/products", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		codes = append(codes, w.Code)
		if w.Code == 429 && w.Header().Get("Retry-After") == "" {
			t.Error("missing Retry-After")
		}
	}
	if codes[0] != 200 || codes[1] != 200 || codes[2] != 429 {
		t.Fatalf("codes %v", codes)
	}
}

func TestSecurityHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	SecurityHeaders(true)(ok).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	for _, k := range []string{"X-Content-Type-Options", "Referrer-Policy", "X-Frame-Options", "Content-Security-Policy", "Strict-Transport-Security"} {
		if w.Header().Get(k) == "" {
			t.Errorf("missing %s", k)
		}
	}
}

func TestCORSAllowsListedOrigins(t *testing.T) {
	h := CORS([]string{"https://shop.example.com", "https://preview.vercel.app"})(ok)
	for origin, want := range map[string]string{
		"https://shop.example.com":   "https://shop.example.com",
		"https://preview.vercel.app": "https://preview.vercel.app",
		"https://evil.example.com":   "",
	} {
		r := httptest.NewRequest("GET", "/api/v1/home", nil)
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != want {
			t.Errorf("origin %s: allow-origin = %q, want %q", origin, got, want)
		}
	}

	r := httptest.NewRequest("OPTIONS", "/api/v1/orders", nil)
	r.Header.Set("Origin", "https://shop.example.com")
	r.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent || w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("preflight: code %d, credentials %q", w.Code, w.Header().Get("Access-Control-Allow-Credentials"))
	}
}
