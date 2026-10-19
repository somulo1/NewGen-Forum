// Package tones serves the sounds the apps play (message chime, ringing, call connected, hang up). They are stored as
// parameters, not audio files, and every organization gets the same set.
package tones

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/httpx"
)

type Tone struct {
	Key         string          `json:"key"`
	Kind        string          `json:"kind"`
	Description string          `json:"description"`
	Params      json.RawMessage `json:"params"`
}

type Module struct{ A *app.App }

func New(a *app.App) *Module { return &Module{A: a} }

func (m *Module) List(ctx context.Context) ([]Tone, error) {
	out := []Tone{}
	err := m.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select key, kind, description, params from hub.tones order by key`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var t Tone
			if err := rs.Scan(&t.Key, &t.Kind, &t.Description, &t.Params); err != nil {
				return err
			}
			out = append(out, t)
		}
		return rs.Err()
	})
	return out, err
}

// Put stores one tone (used by the seeder).
func (m *Module) Put(ctx context.Context, tx pgx.Tx, t Tone) error {
	if _, err := tx.Exec(ctx, `insert into hub.tones (key, kind, description, params) values ($1,$2,$3,$4)
		on conflict (key) do update set kind=excluded.kind, description=excluded.description, params=excluded.params, ver = hub.tones.ver + 1, mac = hub.pending_mac()`, t.Key, t.Kind, t.Description, t.Params); err != nil {
		return err
	}
	return m.A.Sign.Seal(ctx, tx, "tones", "key=$1", t.Key)
}

func (m *Module) Public(r chi.Router) {
	// tones are the same for everyone and contain nothing private, so the sign-in screen can use them too
	r.Get("/tones", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		l, err := m.List(r.Context())
		if err != nil {
			return err
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		httpx.JSON(w, 200, map[string]any{"tones": l})
		return nil
	}))
}

func (m *Module) Private(chi.Router) {}
