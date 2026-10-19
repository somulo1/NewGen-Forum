// Package notify writes the bell notifications and pushes them to people's open devices.
package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Note struct {
	ID        uuid.UUID  `json:"id"`
	Kind      string     `json:"kind"`
	Text      string     `json:"text"`
	Color     string     `json:"color"`
	Ref       string     `json:"ref"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

type Pending struct {
	acct uuid.UUID
	org  uuid.UUID
	note Note
}

type Batch struct{ items []Pending }

// Add records a notification inside the caller's transaction and queues its push; call Send after the transaction commits.
func (b *Batch) Add(ctx context.Context, tx pgx.Tx, org, person uuid.UUID, kind, text, color, ref string) error {
	n := Note{ID: app.NewID(), Kind: kind, Text: text, Color: color, Ref: ref, CreatedAt: app.Now()}
	var acct *uuid.UUID
	if err := tx.QueryRow(ctx, `with ins as (insert into hub.notifications (org_id, id, person_id, kind, text, color, ref) values ($1,$2,$3,$4,$5,$6,$7))
		select account_id from hub.people where id=$3`, org, n.ID, person, kind, text, color, ref).Scan(&acct); err != nil {
		return err
	}
	if acct != nil {
		b.items = append(b.items, Pending{*acct, org, n})
	}
	return nil
}

func (b *Batch) Send(ctx context.Context, a *app.App) {
	for _, p := range b.items {
		d, _ := json.Marshal(p.note)
		a.Publish(ctx, realtime.Event{Type: "notification", Channel: realtime.AcctChannel(p.acct), Org: p.org.String(), Data: d})
	}
}

type Module struct{ A *app.App }

func New(a *app.App) *Module { return &Module{A: a} }

func (m *Module) Public(chi.Router) {}

func (m *Module) Private(r chi.Router) {
	r.Get("/notifications", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if n <= 0 || n > 100 {
			n = 30
		}
		out := []Note{}
		unread := 0
		err = m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			rs, err := tx.Query(r.Context(), `select id, kind, text, color, ref, created_at, read_at from hub.notifications where person_id=$1 order by created_at desc, id desc limit `+strconv.Itoa(n), p.Person)
			if err != nil {
				return err
			}
			for rs.Next() {
				var x Note
				if err := rs.Scan(&x.ID, &x.Kind, &x.Text, &x.Color, &x.Ref, &x.CreatedAt, &x.ReadAt); err != nil {
					rs.Close()
					return err
				}
				out = append(out, x)
			}
			rs.Close()
			return tx.QueryRow(r.Context(), `select count(*) from hub.notifications where person_id=$1 and read_at is null`, p.Person).Scan(&unread)
		})
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"notifications": out, "unread": unread})
		return nil
	}))
	r.Post("/notifications/read", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		return m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			_, err := tx.Exec(r.Context(), `update hub.notifications set read_at=now() where person_id=$1 and read_at is null`, p.Person)
			if err == nil {
				w.WriteHeader(204)
			}
			return err
		})
	}))
}
