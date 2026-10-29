// Package httpx contains JSON helpers, the API error type, pagination and uploads.
package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

// MaxJSONBody is the request body limit for JSON endpoints.
const MaxJSONBody = 1 << 20

// Error is an API error rendered as {"error": {...}}.
type Error struct {
	Status     int               `json:"-"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Fields     map[string]string `json:"fields,omitempty"`
	RetryAfter time.Duration     `json:"-"`
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// E builds an Error.
func E(status int, code, msg string) *Error { return &Error{Status: status, Code: code, Message: msg} }

// WithFields returns a copy with field errors attached.
func (e *Error) WithFields(f map[string]string) *Error {
	c := *e
	c.Fields = f
	return &c
}

// BadRequest is a 400 error.
func BadRequest(msg string) *Error { return E(http.StatusBadRequest, "bad_request", msg) }

// NotFound is a 404 error.
func NotFound() *Error { return E(http.StatusNotFound, "not_found", "Not found") }

// Unauthorized is a 401 error.
func Unauthorized() *Error { return E(http.StatusUnauthorized, "unauthorized", "Please sign in") }

// Forbidden is a 403 error.
func Forbidden(msg string) *Error { return E(http.StatusForbidden, "forbidden", msg) }

// Conflict is a 409 error with a specific code.
func Conflict(code, msg string) *Error { return E(http.StatusConflict, code, msg) }

// Validation returns a 422 with per-field messages.
func Validation(fields map[string]string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: "validation", Message: "Please check the highlighted fields", Fields: fields}
}

// RateLimited returns a 429 error with Retry-After.
func RateLimited(after time.Duration) *Error {
	return &Error{Status: http.StatusTooManyRequests, Code: "rate_limited", Message: "Too many requests, please try again later", RetryAfter: after}
}

// HandlerFunc is an http handler that returns an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// H adapts a HandlerFunc to http.HandlerFunc, rendering any returned error.
func H(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			WriteError(w, r, err)
		}
	}
}

// WriteJSON writes v as JSON with the given status.
func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

// OK writes a 200 JSON response.
func OK(w http.ResponseWriter, v any) error { return WriteJSON(w, http.StatusOK, v) }

// Created writes a 201 JSON response.
func Created(w http.ResponseWriter, v any) error { return WriteJSON(w, http.StatusCreated, v) }

// NoContent writes 204.
func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

// NoContentAfter writes 204 when err is nil, otherwise returns err.
func NoContentAfter(w http.ResponseWriter, err error) error {
	if err != nil {
		return err
	}
	return NoContent(w)
}

// WriteError renders err as the API error envelope.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	var ae *Error
	var mbe *http.MaxBytesError
	switch {
	case errors.As(err, &ae):
	case errors.Is(err, pgx.ErrNoRows):
		ae = NotFound()
	case errors.As(err, &mbe):
		ae = E(http.StatusRequestEntityTooLarge, "too_large", "Request body is too large")
	case errors.Is(err, context.Canceled):
		return
	default:
		slog.ErrorContext(r.Context(), "internal error", "err", err, "method", r.Method, "path", r.URL.Path)
		ae = E(http.StatusInternalServerError, "internal", "Something went wrong")
	}
	if ae.RetryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(ae.RetryAfter.Seconds()))))
	}
	_ = WriteJSON(w, ae.Status, map[string]*Error{"error": ae})
}

// Decode reads a JSON body (max 1 MB) into v.
func Decode(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxJSONBody)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return E(http.StatusRequestEntityTooLarge, "too_large", "Request body is too large")
		}
		if errors.Is(err, io.EOF) {
			return BadRequest("Request body is required")
		}
		return BadRequest("Invalid JSON: " + err.Error())
	}
	return nil
}

// DecodeOptional is like Decode but accepts an empty body.
func DecodeOptional(w http.ResponseWriter, r *http.Request, v any) error {
	err := Decode(w, r, v)
	var ae *Error
	if errors.As(err, &ae) && ae.Message == "Request body is required" {
		return nil
	}
	return err
}

// PathInt64 parses a path parameter as a positive int64 or returns 404.
func PathInt64(r *http.Request, name string) (int64, error) {
	n, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || n <= 0 {
		return 0, NotFound()
	}
	return n, nil
}

// Page holds pagination parameters.
type Page struct {
	Page, Size int
}

// Offset returns the SQL offset.
func (p Page) Offset() int { return (p.Page - 1) * p.Size }

// ParsePage reads ?page and ?page_size (default 20, max 100).
func ParsePage(r *http.Request) Page {
	p := Page{Page: 1, Size: 20}
	if n, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && n > 0 {
		p.Page = min(n, 100000)
	}
	if n, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && n > 0 {
		p.Size = min(n, 100)
	}
	return p
}

// Paged is the paginated response envelope.
type Paged[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// NewPaged builds a Paged response, never with null items.
func NewPaged[T any](items []T, total int, p Page) Paged[T] {
	if items == nil {
		items = []T{}
	}
	return Paged[T]{Items: items, Total: total, Page: p.Page, PageSize: p.Size}
}

type ctxKey int

const ipKey ctxKey = 1

// WithClientIP stores the client IP in ctx.
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, ipKey, ip)
}

// ClientIP returns the client IP stored by middleware.
func ClientIP(r *http.Request) string {
	if ip, ok := r.Context().Value(ipKey).(string); ok {
		return ip
	}
	return ""
}

// Fields collects validation errors.
type Fields map[string]string

// Check records msg for field when bad is true (first error per field wins).
func (f Fields) Check(bad bool, field, msg string) {
	if bad {
		if _, exists := f[field]; !exists {
			f[field] = msg
		}
	}
}

// Err returns a validation error or nil.
func (f Fields) Err() error {
	if len(f) == 0 {
		return nil
	}
	return Validation(f)
}

// Len returns the number of characters (runes) in s.
func Len(s string) int { return utf8.RuneCountInString(s) }
