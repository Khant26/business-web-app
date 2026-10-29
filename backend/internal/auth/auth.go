// Package auth implements email OTP login, cookie sessions and role checks.
package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/config"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
	"github.com/Khant26/business-web-app/backend/internal/ratelimit"
)

// CookieName is the session cookie name.
const CookieName = "sy_session"

// User is the public user shape.
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UserColumns selects the columns scanned by ScanUser.
const UserColumns = "id, email, name, phone, role, created_at"

// ScanUser scans UserColumns.
func ScanUser(row pgx.Row) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// IsStaff reports whether the user has staff or admin role.
func (u *User) IsStaff() bool { return u != nil && (u.Role == "staff" || u.Role == "admin") }

type ctxKey struct{}

type sessionCtx struct {
	user      *User
	sessionID string
}

// UserFrom returns the signed-in user or nil.
func UserFrom(ctx context.Context) *User {
	if s, ok := ctx.Value(ctxKey{}).(*sessionCtx); ok {
		return s.user
	}
	return nil
}

// SessionIDFrom returns the current session id or "".
func SessionIDFrom(ctx context.Context) string {
	if s, ok := ctx.Value(ctxKey{}).(*sessionCtx); ok {
		return s.sessionID
	}
	return ""
}

// WithUser stores a user in ctx (used by tests and the session middleware).
func WithUser(ctx context.Context, u *User, sessionID string) context.Context {
	return context.WithValue(ctx, ctxKey{}, &sessionCtx{user: u, sessionID: sessionID})
}

// NormalizeEmail lower-cases and validates an email address.
func NormalizeEmail(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) < 3 || len(s) > 254 {
		return "", false
	}
	a, err := mail.ParseAddress(s)
	if err != nil || a.Address != s || !strings.Contains(s[strings.LastIndexByte(s, '@')+1:], ".") {
		return "", false
	}
	return s, true
}

// Service handles OTP login and sessions.
type Service struct {
	DB     *pgxpool.Pool
	Cfg    *config.Config
	Mailer Mailer

	otpPerEmailMinute *ratelimit.Limiter
	otpPerEmailHour   *ratelimit.Limiter
	otpPerIP          *ratelimit.Limiter
	verifyPerIP       *ratelimit.Limiter
}

// NewService builds the auth service with its rate limiters.
func NewService(pool *pgxpool.Pool, cfg *config.Config, m Mailer) *Service {
	return &Service{
		DB: pool, Cfg: cfg, Mailer: m,
		otpPerEmailMinute: ratelimit.New(1, time.Minute),
		otpPerEmailHour:   ratelimit.New(5, time.Hour),
		otpPerIP:          ratelimit.New(20, time.Hour),
		verifyPerIP:       ratelimit.New(10, 15*time.Minute),
	}
}

// Limiters returns limiters that need periodic cleanup.
func (s *Service) Limiters() []*ratelimit.Limiter {
	return []*ratelimit.Limiter{s.otpPerEmailMinute, s.otpPerEmailHour, s.otpPerIP, s.verifyPerIP}
}

func allow(l *ratelimit.Limiter, key string) error {
	if ok, wait := l.Allow(key); !ok {
		return httpx.RateLimited(wait)
	}
	return nil
}

// RequestOTP handles POST /auth/otp/request.
func (s *Service) RequestOTP(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	email, ok := NormalizeEmail(req.Email)
	if !ok {
		return httpx.Validation(map[string]string{"email": "invalid"})
	}
	ip := httpx.ClientIP(r)
	if err := allow(s.otpPerIP, ip); err != nil {
		return err
	}
	if err := allow(s.otpPerEmailMinute, email); err != nil {
		return err
	}
	if err := allow(s.otpPerEmailHour, email); err != nil {
		return err
	}
	code, err := GenerateCode()
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(r.Context(),
		`INSERT INTO otp_codes (email, code_hash, ip, expires_at) VALUES ($1, $2, $3, now() + $4::interval)`,
		email, HashCode(s.Cfg.OTPSecret, email, code), ip, OTPTTL)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	if err := s.Mailer.SendOTP(ctx, email, code); err != nil {
		slog.Error("send OTP email failed", "to", email, "err", err)
		return httpx.E(http.StatusInternalServerError, "internal", "Could not send the email, please try again")
	}
	return httpx.NoContent(w)
}

var (
	errOTPInvalid = httpx.E(http.StatusBadRequest, "otp_invalid", "The code is incorrect")
	errOTPExpired = httpx.E(http.StatusBadRequest, "otp_expired", "The code has expired, please request a new one")
	errOTPTooMany = httpx.E(http.StatusBadRequest, "otp_too_many_attempts", "Too many wrong attempts, please request a new code")
	errAccountOff = httpx.E(http.StatusForbidden, "account_disabled", "This account has been disabled")
)

// VerifyOTP handles POST /auth/otp/verify.
func (s *Service) VerifyOTP(w http.ResponseWriter, r *http.Request) error {
	if err := allow(s.verifyPerIP, httpx.ClientIP(r)); err != nil {
		return err
	}
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := httpx.Decode(w, r, &req); err != nil {
		return err
	}
	email, ok := NormalizeEmail(req.Email)
	code := strings.TrimSpace(req.Code)
	if !ok || len(code) != 6 {
		return errOTPInvalid
	}
	ctx := r.Context()
	var apiErr error
	var user *User
	var active bool
	err := db.Tx(ctx, s.DB, func(tx pgx.Tx) error {
		var id int64
		var hash string
		var attempts int
		var expires time.Time
		var consumed *time.Time
		err := tx.QueryRow(ctx, `SELECT id, code_hash, attempts, expires_at, consumed_at FROM otp_codes
			WHERE email = $1 ORDER BY created_at DESC, id DESC LIMIT 1 FOR UPDATE`, email).
			Scan(&id, &hash, &attempts, &expires, &consumed)
		if errors.Is(err, pgx.ErrNoRows) {
			apiErr = errOTPInvalid
			return nil
		}
		if err != nil {
			return err
		}
		switch {
		case attempts >= OTPMaxAttempts:
			apiErr = errOTPTooMany
			return nil
		case consumed != nil:
			apiErr = errOTPInvalid
			return nil
		case time.Now().After(expires):
			apiErr = errOTPExpired
			return nil
		}
		if !VerifyCode(s.Cfg.OTPSecret, email, code, hash) {
			attempts++
			if attempts >= OTPMaxAttempts {
				apiErr = errOTPTooMany
				_, err = tx.Exec(ctx, `UPDATE otp_codes SET attempts = $2, consumed_at = now() WHERE id = $1`, id, attempts)
			} else {
				apiErr = errOTPInvalid
				_, err = tx.Exec(ctx, `UPDATE otp_codes SET attempts = $2 WHERE id = $1`, id, attempts)
			}
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE otp_codes SET consumed_at = now() WHERE id = $1`, id); err != nil {
			return err
		}
		u := &User{}
		err = tx.QueryRow(ctx, `INSERT INTO users (email) VALUES ($1)
			ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
			RETURNING `+UserColumns+`, is_active`, email).
			Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.CreatedAt, &active)
		user = u
		return err
	})
	if err != nil {
		return err
	}
	if apiErr != nil {
		return apiErr
	}
	if !active {
		return errAccountOff
	}
	if err := s.createSession(ctx, w, r, user.ID); err != nil {
		return err
	}
	return httpx.OK(w, map[string]any{"user": user})
}

func (s *Service) createSession(ctx context.Context, w http.ResponseWriter, r *http.Request, userID int64) error {
	token, id, err := NewToken()
	if err != nil {
		return err
	}
	ua := r.UserAgent()
	if len(ua) > 255 {
		ua = ua[:255]
	}
	err = db.Tx(ctx, s.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO sessions (id, user_id, ip, user_agent, expires_at)
			VALUES ($1, $2, $3, $4, now() + $5::interval)`,
			id, userID, httpx.ClientIP(r), ua, s.Cfg.SessionTTL); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `DELETE FROM sessions WHERE id IN (
			SELECT id FROM sessions WHERE user_id = $1 ORDER BY created_at DESC, id OFFSET $2)`,
			userID, s.Cfg.SessionMaxPerUser)
		return err
	})
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name: CookieName, Value: token, Path: "/", HttpOnly: true, Secure: s.Cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode, MaxAge: int(s.Cfg.SessionTTL.Seconds()),
	})
	return nil
}

func (s *Service) clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: "", Path: "/", HttpOnly: true,
		Secure: s.Cfg.CookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1})
}

// Logout handles POST /auth/logout.
func (s *Service) Logout(w http.ResponseWriter, r *http.Request) error {
	if id := SessionIDFrom(r.Context()); id != "" {
		if _, err := s.DB.Exec(r.Context(), `DELETE FROM sessions WHERE id = $1`, id); err != nil {
			return err
		}
	}
	s.clearCookie(w)
	return httpx.NoContent(w)
}

// Me handles GET /auth/me.
func (s *Service) Me(w http.ResponseWriter, r *http.Request) error {
	u := UserFrom(r.Context())
	if u == nil {
		return httpx.Unauthorized()
	}
	return httpx.OK(w, map[string]any{"user": u})
}

// Middleware loads the session (if any) into the request context.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(CookieName)
		if err != nil || c.Value == "" || len(c.Value) > 100 {
			next.ServeHTTP(w, r)
			return
		}
		id := HashToken(c.Value)
		ctx := r.Context()
		var lastSeen time.Time
		u := &User{}
		err = s.DB.QueryRow(ctx, `SELECT u.id, u.email, u.name, u.phone, u.role, u.created_at, s.last_seen_at
			FROM sessions s JOIN users u ON u.id = s.user_id
			WHERE s.id = $1 AND s.expires_at > now() AND s.last_seen_at > now() - $2::interval AND u.is_active`,
			id, s.Cfg.SessionIdle).
			Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.CreatedAt, &lastSeen)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				slog.Error("load session", "err", err)
			} else {
				s.clearCookie(w)
			}
			next.ServeHTTP(w, r)
			return
		}
		if time.Since(lastSeen) > time.Minute {
			if _, err := s.DB.Exec(ctx, `UPDATE sessions SET last_seen_at = now(), ip = $2 WHERE id = $1`, id, httpx.ClientIP(r)); err != nil {
				slog.Error("touch session", "err", err)
			}
		}
		next.ServeHTTP(w, r.WithContext(WithUser(ctx, u, id)))
	})
}

// RequireAuth rejects anonymous requests with 401.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFrom(r.Context()) == nil {
			httpx.WriteError(w, r, httpx.Unauthorized())
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole rejects users without one of roles (401 when anonymous, 403 otherwise).
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFrom(r.Context())
			if u == nil {
				httpx.WriteError(w, r, httpx.Unauthorized())
				return
			}
			for _, role := range roles {
				if u.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			httpx.WriteError(w, r, httpx.Forbidden("You do not have permission to do this"))
		})
	}
}

// EnsureAdmin creates or promotes email to an active admin.
func EnsureAdmin(ctx context.Context, q db.DBTX, email string) error {
	email, ok := NormalizeEmail(email)
	if !ok {
		return errors.New("ADMIN_EMAIL is not a valid email address")
	}
	_, err := q.Exec(ctx, `INSERT INTO users (email, role) VALUES ($1, 'admin')
		ON CONFLICT (email) DO UPDATE SET role = 'admin', is_active = true, updated_at = now()`, email)
	return err
}
