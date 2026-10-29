// Package config loads runtime configuration from environment variables
// (optionally pre-populated from a .env file) with development-friendly defaults.
package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime settings.
type Config struct {
	AppEnv      string
	Port        string
	PublicURL   string
	WebDir      string
	UploadDir   string
	DatabaseURL string

	AdminEmail        string
	CookieSecure      bool
	SessionTTL        time.Duration
	SessionIdle       time.Duration
	SessionMaxPerUser int
	OTPSecret         string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	CORSOrigin string

	// TrustProxy makes the client IP come from the first X-Forwarded-For entry.
	// Default false: only enable it when the API sits behind a reverse proxy
	// that overwrites X-Forwarded-For, otherwise clients can spoof their IP.
	TrustProxy bool
}

const defaultOTPSecret = "change-me-to-a-long-random-string"

// IsDev reports whether the app runs in development mode.
func (c *Config) IsDev() bool { return c.AppEnv != "production" }

// Load reads .env files (if present, without overriding real env vars) and the environment.
func Load() (*Config, error) {
	for _, f := range []string{".env", "../.env"} {
		loadDotEnv(f)
	}
	c := &Config{
		AppEnv:            get("APP_ENV", "development"),
		Port:              get("PORT", "8080"),
		PublicURL:         get("PUBLIC_URL", "http://localhost:5173"),
		WebDir:            get("WEB_DIR", "../frontend/dist"),
		UploadDir:         get("UPLOAD_DIR", "./uploads"),
		DatabaseURL:       get("DATABASE_URL", "postgres://sawyun:sawyun@localhost:5433/sawyun?sslmode=disable"),
		AdminEmail:        strings.ToLower(strings.TrimSpace(get("ADMIN_EMAIL", "admin@example.com"))),
		CookieSecure:      getBool("COOKIE_SECURE", false),
		SessionTTL:        time.Duration(getInt("SESSION_TTL_HOURS", 720)) * time.Hour,
		SessionIdle:       time.Duration(getInt("SESSION_IDLE_HOURS", 168)) * time.Hour,
		SessionMaxPerUser: getInt("SESSION_MAX_PER_USER", 5),
		OTPSecret:         get("OTP_SECRET", defaultOTPSecret),
		SMTPHost:          get("SMTP_HOST", ""),
		SMTPPort:          getInt("SMTP_PORT", 587),
		SMTPUsername:      get("SMTP_USERNAME", ""),
		SMTPPassword:      get("SMTP_PASSWORD", ""),
		SMTPFrom:          get("SMTP_FROM", "Sawyun Ecommerce <no-reply@example.com>"),
		CORSOrigin:        get("CORS_ORIGIN", ""),
		TrustProxy:        getBool("TRUST_PROXY", false),
	}
	if c.SessionMaxPerUser < 1 {
		c.SessionMaxPerUser = 1
	}
	if !c.IsDev() && (len(c.OTPSecret) < 16 || c.OTPSecret == defaultOTPSecret) {
		return nil, errors.New("OTP_SECRET must be set to a long random value in production")
	}
	return c, nil
}

func get(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return def
}

func getInt(k string, def int) int {
	if n, err := strconv.Atoi(get(k, "")); err == nil {
		return n
	}
	return def
}

func getBool(k string, def bool) bool {
	if b, err := strconv.ParseBool(get(k, "")); err == nil {
		return b
	}
	return def
}

// loadDotEnv sets KEY=VALUE pairs from file for keys not already in the environment.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		k, v, ok := parseLine(s.Text())
		if !ok {
			continue
		}
		if _, exists := os.LookupEnv(k); !exists {
			_ = os.Setenv(k, v)
		}
	}
}

func parseLine(line string) (key, val string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimPrefix(line, "export ")
	k, v, found := strings.Cut(line, "=")
	if !found {
		return "", "", false
	}
	k = strings.TrimSpace(k)
	v = strings.TrimSpace(v)
	if len(v) > 0 && (v[0] == '"' || v[0] == '\'') {
		q := v[0]
		if end := strings.IndexByte(v[1:], q); end >= 0 {
			v = v[1 : end+1]
		}
	} else if strings.HasPrefix(v, "#") {
		v = "" // KEY=   # only a comment
	} else if i := strings.Index(v, " #"); i >= 0 {
		v = strings.TrimSpace(v[:i])
	}
	return k, v, k != ""
}
