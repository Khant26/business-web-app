// Package audit records admin writes in audit_logs.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/auth"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

// Log writes one audit entry for the current user. entityID may be any printable value.
func Log(ctx context.Context, q db.DBTX, r *http.Request, action, entity string, entityID any, details any) error {
	var uid *int64
	if u := auth.UserFrom(r.Context()); u != nil {
		uid = &u.ID
	}
	if details == nil {
		details = map[string]any{}
	}
	b, err := json.Marshal(details)
	if err != nil {
		return err
	}
	_, err = q.Exec(ctx, `INSERT INTO audit_logs (user_id, action, entity, entity_id, details, ip)
		VALUES ($1, $2, $3, $4, $5, $6)`, uid, action, entity, fmt.Sprint(entityID), b, httpx.ClientIP(r))
	return err
}

type entry struct {
	ID        int64           `json:"id"`
	UserEmail string          `json:"user_email"`
	Action    string          `json:"action"`
	Entity    string          `json:"entity"`
	EntityID  string          `json:"entity_id"`
	Details   json.RawMessage `json:"details"`
	IP        string          `json:"ip"`
	CreatedAt time.Time       `json:"created_at"`
}

// Handler serves GET /admin/audit-logs.
type Handler struct{ DB *pgxpool.Pool }

// List handles GET /admin/audit-logs.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	p := httpx.ParsePage(r)
	var total int
	if err := h.DB.QueryRow(r.Context(), `SELECT count(*) FROM audit_logs`).Scan(&total); err != nil {
		return err
	}
	rows, err := h.DB.Query(r.Context(), `SELECT a.id, COALESCE(u.email, ''), a.action, a.entity, a.entity_id, a.details, a.ip, a.created_at
		FROM audit_logs a LEFT JOIN users u ON u.id = a.user_id
		ORDER BY a.created_at DESC, a.id DESC LIMIT $1 OFFSET $2`, p.Size, p.Offset())
	if err != nil {
		return err
	}
	defer rows.Close()
	var items []entry
	for rows.Next() {
		var e entry
		var details []byte
		if err := rows.Scan(&e.ID, &e.UserEmail, &e.Action, &e.Entity, &e.EntityID, &details, &e.IP, &e.CreatedAt); err != nil {
			return err
		}
		e.Details = details
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return httpx.OK(w, httpx.NewPaged(items, total, p))
}
