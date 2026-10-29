// Package server wires handlers, middleware, static uploads and the SPA.
package server

import (
	"context"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/catalog"
	"github.com/Khant26/business-web-app/backend/internal/config"
	"github.com/Khant26/business-web-app/backend/internal/coupons"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/inventory"
	"github.com/Khant26/business-web-app/backend/internal/middleware"
	"github.com/Khant26/business-web-app/backend/internal/orders"
	"github.com/Khant26/business-web-app/backend/internal/ratelimit"
	"github.com/Khant26/business-web-app/backend/internal/settings"
	"github.com/Khant26/business-web-app/backend/internal/shipping"
	"github.com/Khant26/business-web-app/backend/internal/users"
)

// New builds the root HTTP handler. Limiter cleanup goroutines stop when ctx is done.
func New(ctx context.Context, cfg *config.Config, pool *pgxpool.Pool, mailer auth.Mailer) http.Handler {
	authSvc := auth.NewService(pool, cfg, mailer)
	cat := &catalog.Handler{DB: pool, UploadDir: cfg.UploadDir}
	ord := &orders.Handler{DB: pool, UploadDir: cfg.UploadDir}
	set := &settings.Handler{DB: pool}
	ship := &shipping.Handler{DB: pool}
	cpn := &coupons.Handler{DB: pool}
	usr := &users.Handler{DB: pool}
	inv := &inventory.Handler{DB: pool}
	aud := &audit.Handler{DB: pool}

	globalLimit := ratelimit.New(300, time.Minute)
	orderLimit := ratelimit.New(10, time.Hour)
	uploadLimit := ratelimit.New(20, time.Hour)
	for _, l := range append(authSvc.Limiters(), globalLimit, orderLimit, uploadLimit) {
		l.StartCleanup(ctx, 5*time.Minute)
	}

	mux := http.NewServeMux()
	const api = "/api/v1"
	type mw = func(http.Handler) http.Handler
	handle := func(pattern string, h httpx.HandlerFunc, mws ...mw) {
		method, p, _ := strings.Cut(pattern, " ")
		mux.Handle(method+" "+api+p, middleware.Chain(httpx.H(h), mws...))
	}
	authed := auth.RequireAuth
	staff := auth.RequireRole("staff", "admin")
	admin := auth.RequireRole("admin")
	orderRL := middleware.RateLimitUser(orderLimit)
	uploadRL := middleware.RateLimitUser(uploadLimit)

	// Public
	handle("GET /health", func(w http.ResponseWriter, r *http.Request) error {
		return httpx.OK(w, map[string]string{"status": "ok"})
	})
	handle("GET /settings/public", set.GetPublic)
	handle("GET /home", cat.Home)
	handle("GET /categories", cat.Categories)
	handle("GET /categories/{slug}", cat.Category)
	handle("GET /products", cat.Products)
	handle("GET /products/{slug}", cat.Product)
	handle("GET /products/{slug}/reviews", cat.Reviews)
	handle("GET /events", cat.Events)
	handle("GET /events/{slug}", cat.EventDetail)
	handle("GET /provinces", ship.Provinces)
	handle("GET /shipping/zones", ship.List)
	handle("POST /cart/validate", ord.ValidateCart)

	// Auth
	handle("POST /auth/otp/request", authSvc.RequestOTP)
	handle("POST /auth/otp/verify", authSvc.VerifyOTP)
	handle("POST /auth/logout", authSvc.Logout)
	handle("GET /auth/me", authSvc.Me)

	// Customer
	handle("PATCH /me", usr.UpdateMe, authed)
	handle("GET /me/sessions", authSvc.ListSessions, authed)
	handle("DELETE /me/sessions/{id}", authSvc.DeleteSession, authed)
	handle("POST /me/sessions/revoke-others", authSvc.RevokeOthers, authed)
	handle("GET /me/addresses", usr.Addresses, authed)
	handle("POST /me/addresses", usr.CreateAddress, authed)
	handle("PUT /me/addresses/{id}", usr.UpdateAddress, authed)
	handle("DELETE /me/addresses/{id}", usr.DeleteAddress, authed)
	handle("GET /me/wishlist", cat.Wishlist, authed)
	handle("GET /me/wishlist/ids", cat.WishlistIDs, authed)
	handle("PUT /me/wishlist/{productId}", cat.WishlistAdd, authed)
	handle("DELETE /me/wishlist/{productId}", cat.WishlistRemove, authed)
	handle("POST /checkout/quote", ord.QuoteHandler, authed)
	handle("POST /orders", ord.Place, authed, orderRL)
	handle("GET /me/orders", ord.MyOrders, authed)
	handle("GET /me/orders/{code}", ord.MyOrder, authed)
	handle("POST /me/orders/{code}/cancel", ord.Cancel, authed)
	handle("GET /me/orders/{code}/promptpay", ord.PromptPay, authed)
	handle("POST /me/orders/{code}/slip", ord.UploadSlip, authed, uploadRL)
	handle("POST /products/{slug}/reviews", cat.CreateReview, authed)

	// Admin (staff or admin)
	handle("GET /admin/dashboard", ord.Dashboard, staff)
	handle("GET /admin/orders", ord.AdminList, staff)
	handle("GET /admin/orders/{code}", ord.AdminGet, staff)
	handle("POST /admin/orders/{code}/status", ord.AdminStatus, staff)
	handle("PUT /admin/orders/{code}/tracking", ord.AdminTracking, staff)
	handle("POST /admin/slips/{id}/approve", ord.ReviewSlip(true), staff)
	handle("POST /admin/slips/{id}/reject", ord.ReviewSlip(false), staff)
	handle("GET /admin/slips/{id}/file", ord.SlipFile, staff)
	handle("GET /admin/products", cat.AdminProducts, staff)
	handle("POST /admin/products", cat.CreateProduct, staff)
	handle("GET /admin/products/{id}", cat.AdminProduct, staff)
	handle("PUT /admin/products/{id}", cat.UpdateProduct, staff)
	handle("DELETE /admin/products/{id}", cat.DeleteProduct, staff)
	handle("POST /admin/products/{id}/images", cat.UploadProductImage, staff, uploadRL)
	handle("PUT /admin/products/{id}/images/order", cat.ReorderImages, staff)
	handle("DELETE /admin/products/{id}/images/{imageId}", cat.DeleteImage, staff)
	handle("POST /admin/products/{id}/variants", cat.CreateVariant, staff)
	handle("PUT /admin/variants/{id}", cat.UpdateVariant, staff)
	handle("DELETE /admin/variants/{id}", cat.DeleteVariant, staff)
	handle("GET /admin/inventory", inv.List, staff)
	handle("GET /admin/inventory/movements", inv.Movements, staff)
	handle("POST /admin/inventory/{variantId}/adjust", inv.Adjust, staff)
	handle("GET /admin/categories", cat.AdminCategories, staff)
	handle("POST /admin/categories", cat.SaveCategory, staff)
	handle("PUT /admin/categories/{id}", cat.SaveCategory, staff)
	handle("DELETE /admin/categories/{id}", cat.DeleteCategory, staff)
	handle("POST /admin/categories/{id}/image", cat.UploadCategoryImage, staff, uploadRL)
	handle("GET /admin/events", cat.AdminEvents, staff)
	handle("POST /admin/events", cat.SaveEvent, staff)
	handle("PUT /admin/events/{id}", cat.SaveEvent, staff)
	handle("DELETE /admin/events/{id}", cat.DeleteEvent, staff)
	handle("POST /admin/events/{id}/banner", cat.UploadEventBanner, staff, uploadRL)
	handle("GET /admin/coupons", cpn.List, staff)
	handle("POST /admin/coupons", cpn.Save, staff)
	handle("PUT /admin/coupons/{id}", cpn.Save, staff)
	handle("DELETE /admin/coupons/{id}", cpn.Delete, staff)
	handle("GET /admin/reviews", cat.AdminReviews, staff)
	handle("POST /admin/reviews/{id}/approve", cat.ModerateReview("approved"), staff)
	handle("POST /admin/reviews/{id}/reject", cat.ModerateReview("rejected"), staff)

	// Admin only
	handle("GET /admin/shipping-zones", ship.List, admin)
	handle("POST /admin/shipping-zones", ship.Save, admin)
	handle("PUT /admin/shipping-zones/{id}", ship.Save, admin)
	handle("DELETE /admin/shipping-zones/{id}", ship.Delete, admin)
	handle("GET /admin/settings", set.GetAdmin, admin)
	handle("PUT /admin/settings", set.Put, admin)
	handle("GET /admin/users", usr.AdminList, admin)
	handle("POST /admin/users", usr.AdminCreate, admin)
	handle("PATCH /admin/users/{id}", usr.AdminUpdate, admin)
	handle("GET /admin/audit-logs", aud.List, admin)

	// Unknown API routes → JSON 404.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteError(w, r, httpx.NotFound())
	})
	// Public uploads (never the private slips folder).
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", staticFiles(filepath.Join(cfg.UploadDir, "public"), "public, max-age=31536000, immutable")))
	// SPA (production builds).
	if !cfg.IsDev() || dirExists(cfg.WebDir) {
		mux.Handle("/", spa(cfg.WebDir))
	}

	mws := []func(http.Handler) http.Handler{
		middleware.Recover,
		middleware.ClientIP(cfg.TrustProxy),
		middleware.Logger,
		middleware.SecurityHeaders(cfg.CookieSecure),
	}
	if cfg.IsDev() && cfg.CORSOrigin != "" {
		mws = append(mws, middleware.CORS(cfg.CORSOrigin))
	}
	mws = append(mws, middleware.RateLimitIP(globalLimit), middleware.CSRF, authSvc.Middleware)
	return middleware.Chain(mux, mws...)
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// staticFiles serves regular files from root (no directory listings).
func staticFiles(root, cacheControl string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := path.Clean("/" + r.URL.Path)
		full := filepath.Join(root, filepath.FromSlash(name))
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", cacheControl)
		http.ServeFile(w, r, full)
	})
}

// spa serves the built frontend: real files when present, otherwise index.html (history fallback).
func spa(webDir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/uploads/") {
			http.NotFound(w, r)
			return
		}
		name := path.Clean("/" + r.URL.Path)
		if name != "/" {
			full := filepath.Join(webDir, filepath.FromSlash(name))
			if st, err := os.Stat(full); err == nil && !st.IsDir() {
				if strings.HasPrefix(name, "/assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "public, max-age=3600")
				}
				http.ServeFile(w, r, full)
				return
			}
			if strings.HasPrefix(name, "/assets/") {
				http.NotFound(w, r)
				return
			}
		}
		index := filepath.Join(webDir, "index.html")
		if _, err := os.Stat(index); err != nil {
			http.Error(w, "frontend not built", http.StatusNotFound)
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, index)
	})
}
