package auth

import (
	"net/http"
	"time"

	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

type sessionView struct {
	ID         string    `json:"id"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	Current    bool      `json:"current"`
}

// ListSessions handles GET /me/sessions.
func (s *Service) ListSessions(w http.ResponseWriter, r *http.Request) error {
	u := UserFrom(r.Context())
	cur := SessionIDFrom(r.Context())
	rows, err := s.DB.Query(r.Context(), `SELECT id, ip, user_agent, created_at, last_seen_at FROM sessions
		WHERE user_id = $1 AND expires_at > now() AND last_seen_at > now() - $2::interval
		ORDER BY last_seen_at DESC`, u.ID, s.Cfg.SessionIdle)
	if err != nil {
		return err
	}
	defer rows.Close()
	out := []sessionView{}
	for rows.Next() {
		var v sessionView
		var id string
		if err := rows.Scan(&id, &v.IP, &v.UserAgent, &v.CreatedAt, &v.LastSeenAt); err != nil {
			return err
		}
		v.ID = ShortID(id)
		v.Current = id == cur
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return httpx.OK(w, out)
}

// DeleteSession handles DELETE /me/sessions/{id}.
func (s *Service) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	u := UserFrom(r.Context())
	short := r.PathValue("id")
	if len(short) != 16 {
		return httpx.NotFound()
	}
	tag, err := s.DB.Exec(r.Context(), `DELETE FROM sessions WHERE user_id = $1 AND left(id, 16) = $2`, u.ID, short)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.NotFound()
	}
	if ShortID(SessionIDFrom(r.Context())) == short {
		s.clearCookie(w)
	}
	return httpx.NoContent(w)
}

// RevokeOthers handles POST /me/sessions/revoke-others.
func (s *Service) RevokeOthers(w http.ResponseWriter, r *http.Request) error {
	u := UserFrom(r.Context())
	if _, err := s.DB.Exec(r.Context(), `DELETE FROM sessions WHERE user_id = $1 AND id <> $2`, u.ID, SessionIDFrom(r.Context())); err != nil {
		return err
	}
	return httpx.NoContent(w)
}
