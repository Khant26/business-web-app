// Package settings reads and writes the key/value shop settings.
package settings

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/promptpay"
)

// Settings is the full settings object (admin view).
type Settings struct {
	ShopName             string `json:"shop_name"`
	PromptPayID          string `json:"promptpay_id"`
	PromptPayName        string `json:"promptpay_name"`
	VATRateBP            int64  `json:"vat_rate_bp"`
	PaymentWindowMinutes int64  `json:"payment_window_minutes"`
	CODMaxSatang         int64  `json:"cod_max_satang"`
	SupportEmail         string `json:"support_email"`
	SupportPhone         string `json:"support_phone"`
}

// Public is the subset shown to everyone.
type Public struct {
	ShopName             string `json:"shop_name"`
	VATRateBP            int64  `json:"vat_rate_bp"`
	PaymentWindowMinutes int64  `json:"payment_window_minutes"`
	CODMaxSatang         int64  `json:"cod_max_satang"`
	SupportEmail         string `json:"support_email"`
	SupportPhone         string `json:"support_phone"`
}

// Load reads all settings.
func Load(ctx context.Context, q db.DBTX) (Settings, error) {
	s := Settings{VATRateBP: 700, PaymentWindowMinutes: 30, CODMaxSatang: 1000000}
	rows, err := q.Query(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	m := map[string]json.RawMessage{}
	for rows.Next() {
		var k string
		var v []byte
		if err := rows.Scan(&k, &v); err != nil {
			return s, err
		}
		m[k] = v
	}
	if err := rows.Err(); err != nil {
		return s, err
	}
	b, _ := json.Marshal(m)
	err = json.Unmarshal(b, &s)
	return s, err
}

// Save writes all settings.
func Save(ctx context.Context, q db.DBTX, s Settings) error {
	b, _ := json.Marshal(s)
	var m map[string]json.RawMessage
	_ = json.Unmarshal(b, &m)
	for k, v := range m {
		if _, err := q.Exec(ctx, `INSERT INTO settings (key, value) VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, k, []byte(v)); err != nil {
			return err
		}
	}
	return nil
}

// Handler serves settings endpoints.
type Handler struct{ DB *pgxpool.Pool }

// GetPublic handles GET /settings/public.
func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) error {
	s, err := Load(r.Context(), h.DB)
	if err != nil {
		return err
	}
	return httpx.OK(w, Public{s.ShopName, s.VATRateBP, s.PaymentWindowMinutes, s.CODMaxSatang, s.SupportEmail, s.SupportPhone})
}

// GetAdmin handles GET /admin/settings.
func (h *Handler) GetAdmin(w http.ResponseWriter, r *http.Request) error {
	s, err := Load(r.Context(), h.DB)
	if err != nil {
		return err
	}
	return httpx.OK(w, s)
}

// Put handles PUT /admin/settings.
func (h *Handler) Put(w http.ResponseWriter, r *http.Request) error {
	var s Settings
	if err := httpx.Decode(w, r, &s); err != nil {
		return err
	}
	s.ShopName = strings.TrimSpace(s.ShopName)
	s.PromptPayID = strings.TrimSpace(s.PromptPayID)
	s.PromptPayName = strings.TrimSpace(s.PromptPayName)
	s.SupportEmail = strings.TrimSpace(s.SupportEmail)
	s.SupportPhone = strings.TrimSpace(s.SupportPhone)
	f := map[string]string{}
	if s.ShopName == "" || len(s.ShopName) > 100 {
		f["shop_name"] = "required, max 100 characters"
	}
	if s.PromptPayID != "" {
		if _, _, err := promptpay.NormalizeID(s.PromptPayID); err != nil {
			f["promptpay_id"] = "must be a 10-digit phone number or 13-digit tax ID"
		}
	}
	if len(s.PromptPayName) > 100 {
		f["promptpay_name"] = "max 100 characters"
	}
	if s.VATRateBP < 0 || s.VATRateBP > 10000 {
		f["vat_rate_bp"] = "must be between 0 and 10000"
	}
	if s.PaymentWindowMinutes < 5 || s.PaymentWindowMinutes > 10080 {
		f["payment_window_minutes"] = "must be between 5 and 10080"
	}
	if s.CODMaxSatang < 0 {
		f["cod_max_satang"] = "must be 0 or more"
	}
	if s.SupportEmail != "" {
		if _, ok := auth.NormalizeEmail(s.SupportEmail); !ok {
			f["support_email"] = "invalid"
		}
	}
	if len(s.SupportPhone) > 30 {
		f["support_phone"] = "max 30 characters"
	}
	if len(f) > 0 {
		return httpx.Validation(f)
	}
	err := db.Tx(r.Context(), h.DB, func(tx pgx.Tx) error {
		if err := Save(r.Context(), tx, s); err != nil {
			return err
		}
		return audit.Log(r.Context(), tx, r, "settings.update", "settings", "", s)
	})
	if err != nil {
		return err
	}
	return httpx.OK(w, s)
}
