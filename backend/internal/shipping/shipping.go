// Package shipping manages province-based shipping zones.
package shipping

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Khant26/business-web-app/backend/internal/audit"
	"github.com/Khant26/business-web-app/backend/internal/db"
	"github.com/Khant26/business-web-app/backend/internal/geo"
	"github.com/Khant26/business-web-app/backend/internal/httpx"
)

// Zone is a shipping zone.
type Zone struct {
	ID             int64    `json:"id"`
	NameEN         string   `json:"name_en"`
	NameMY         string   `json:"name_my"`
	FeeSatang      int64    `json:"fee_satang"`
	FreeOverSatang *int64   `json:"free_over_satang"`
	IsDefault      bool     `json:"is_default"`
	SortOrder      int      `json:"sort_order"`
	Provinces      []string `json:"provinces"`
}

const zoneSelect = `SELECT z.id, z.name_en, z.name_my, z.fee_satang, z.free_over_satang, z.is_default, z.sort_order,
	COALESCE((SELECT array_agg(province ORDER BY province) FROM shipping_zone_provinces p WHERE p.zone_id = z.id), '{}')
	FROM shipping_zones z`

func scanZone(r pgx.Row) (Zone, error) {
	var z Zone
	err := r.Scan(&z.ID, &z.NameEN, &z.NameMY, &z.FeeSatang, &z.FreeOverSatang, &z.IsDefault, &z.SortOrder, &z.Provinces)
	if z.Provinces == nil {
		z.Provinces = []string{}
	}
	return z, err
}

// ListZones returns all zones ordered by sort_order.
func ListZones(ctx context.Context, q db.DBTX) ([]Zone, error) {
	rows, err := q.Query(ctx, zoneSelect+" ORDER BY z.sort_order, z.id")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Zone, error) { return scanZone(r) })
}

// ZoneFor returns the zone for an English province name, falling back to the default zone.
// It returns nil when no zone applies.
func ZoneFor(ctx context.Context, q db.DBTX, province string) (*Zone, error) {
	z, err := scanZone(q.QueryRow(ctx, zoneSelect+`
		WHERE z.id IN (SELECT zone_id FROM shipping_zone_provinces WHERE province = $1) OR z.is_default
		ORDER BY z.is_default ASC LIMIT 1`, province))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &z, nil
}

// Handler serves shipping endpoints.
type Handler struct{ DB *pgxpool.Pool }

// List handles GET /shipping/zones and GET /admin/shipping-zones.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	zs, err := ListZones(r.Context(), h.DB)
	if err != nil {
		return err
	}
	return httpx.OK(w, zs)
}

// Provinces handles GET /provinces.
func (h *Handler) Provinces(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "public, max-age=86400")
	return httpx.OK(w, geo.List())
}

type zoneInput struct {
	NameEN         string   `json:"name_en"`
	NameMY         string   `json:"name_my"`
	FeeSatang      int64    `json:"fee_satang"`
	FreeOverSatang *int64   `json:"free_over_satang"`
	IsDefault      bool     `json:"is_default"`
	SortOrder      int      `json:"sort_order"`
	Provinces      []string `json:"provinces"`
}

func (in *zoneInput) validate() error {
	in.NameEN = strings.TrimSpace(in.NameEN)
	in.NameMY = strings.TrimSpace(in.NameMY)
	f := httpx.Fields{}
	f.Check(in.NameEN == "" || httpx.Len(in.NameEN) > 100, "name_en", "required, max 100 characters")
	f.Check(httpx.Len(in.NameMY) > 100, "name_my", "max 100 characters")
	f.Check(in.FeeSatang < 0, "fee_satang", "must be 0 or more")
	f.Check(in.FreeOverSatang != nil && *in.FreeOverSatang < 0, "free_over_satang", "must be 0 or more")
	seen := map[string]bool{}
	clean := make([]string, 0, len(in.Provinces))
	for _, p := range in.Provinces {
		c, ok := geo.Canonical(p)
		if !ok {
			f.Check(true, "provinces", "unknown province: "+p)
			continue
		}
		if !seen[c] {
			seen[c] = true
			clean = append(clean, c)
		}
	}
	sort.Strings(clean)
	in.Provinces = clean
	return f.Err()
}

// Save handles POST /admin/shipping-zones and PUT /admin/shipping-zones/{id}.
func (h *Handler) Save(w http.ResponseWriter, r *http.Request) error {
	var id int64
	if r.PathValue("id") != "" {
		var err error
		if id, err = httpx.PathInt64(r, "id"); err != nil {
			return err
		}
	}
	var in zoneInput
	if err := httpx.Decode(w, r, &in); err != nil {
		return err
	}
	if err := in.validate(); err != nil {
		return err
	}
	ctx := r.Context()
	var out Zone
	err := db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `LOCK TABLE shipping_zone_provinces IN SHARE ROW EXCLUSIVE MODE`); err != nil {
			return err
		}
		var taken []string
		rows, err := tx.Query(ctx, `SELECT province FROM shipping_zone_provinces WHERE province = ANY($1) AND zone_id <> $2 ORDER BY province`, in.Provinces, id)
		if err != nil {
			return err
		}
		if taken, err = pgx.CollectRows(rows, pgx.RowTo[string]); err != nil {
			return err
		}
		if len(taken) > 0 {
			return httpx.Conflict("province_taken", "Province already belongs to another zone: "+strings.Join(taken, ", ")).
				WithFields(map[string]string{"provinces": strings.Join(taken, ", ")})
		}
		if in.IsDefault {
			if _, err := tx.Exec(ctx, `UPDATE shipping_zones SET is_default = false WHERE is_default AND id <> $1`, id); err != nil {
				return err
			}
		}
		action := "shipping_zone.update"
		if id == 0 {
			action = "shipping_zone.create"
			if err := tx.QueryRow(ctx, `INSERT INTO shipping_zones (name_en, name_my, fee_satang, free_over_satang, is_default, sort_order)
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`, in.NameEN, in.NameMY, in.FeeSatang, in.FreeOverSatang, in.IsDefault, in.SortOrder).Scan(&id); err != nil {
				return err
			}
		} else {
			tag, err := tx.Exec(ctx, `UPDATE shipping_zones SET name_en = $2, name_my = $3, fee_satang = $4, free_over_satang = $5,
				is_default = $6, sort_order = $7 WHERE id = $1`, id, in.NameEN, in.NameMY, in.FeeSatang, in.FreeOverSatang, in.IsDefault, in.SortOrder)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return httpx.NotFound()
			}
		}
		if _, err := tx.Exec(ctx, `DELETE FROM shipping_zone_provinces WHERE zone_id = $1`, id); err != nil {
			return err
		}
		if len(in.Provinces) > 0 {
			if _, err := tx.Exec(ctx, `INSERT INTO shipping_zone_provinces (zone_id, province) SELECT $1, unnest($2::text[])`, id, in.Provinces); err != nil {
				return err
			}
		}
		if err := audit.Log(ctx, tx, r, action, "shipping_zone", id, in); err != nil {
			return err
		}
		out, err = scanZone(tx.QueryRow(ctx, zoneSelect+" WHERE z.id = $1", id))
		return err
	})
	if err != nil {
		return err
	}
	if r.Method == http.MethodPost {
		return httpx.Created(w, out)
	}
	return httpx.OK(w, out)
}

// Delete handles DELETE /admin/shipping-zones/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := httpx.PathInt64(r, "id")
	if err != nil {
		return err
	}
	ctx := r.Context()
	return httpx.NoContentAfter(w, db.Tx(ctx, h.DB, func(tx pgx.Tx) error {
		var isDefault bool
		if err := tx.QueryRow(ctx, `SELECT is_default FROM shipping_zones WHERE id = $1 FOR UPDATE`, id).Scan(&isDefault); err != nil {
			return err
		}
		if isDefault {
			return httpx.Conflict("default_zone", "The default zone cannot be deleted")
		}
		if _, err := tx.Exec(ctx, `DELETE FROM shipping_zones WHERE id = $1`, id); err != nil {
			return err
		}
		return audit.Log(ctx, tx, r, "shipping_zone.delete", "shipping_zone", id, nil)
	}))
}
