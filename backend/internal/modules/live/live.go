// Package live is the realtime connection: one WebSocket per device, with organization and account channels attached at
// connect time and conversation channels attached on request after a membership check.
package live

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Module struct{ A *app.App }

func New(a *app.App) *Module { return &Module{A: a} }

const maxChannelsPerConn = 500

type ticket struct {
	Account, Org, Person, Session uuid.UUID
	Name                          string
}

func (m *Module) Private(r chi.Router) {
	// A ticket is a 30-second, single-use pass. Browsers cannot set headers on a WebSocket, and putting the access token
	// in a URL would leak it into logs, so the socket is opened with a ticket instead.
	r.Post("/realtime/ticket", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id := uuid.NewString() + uuid.NewString()
		raw, _ := json.Marshal(ticket{p.Account, p.Org, p.Person, p.Session, p.Name})
		if err := m.A.Rdb.Set(r.Context(), "wst:"+id, raw, 30*time.Second).Err(); err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"ticket": id, "expires_in": 30})
		return nil
	}))
	r.Get("/presence", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		ids := r.URL.Query()["id"]
		if len(ids) > 200 {
			ids = ids[:200]
		}
		out := map[string]bool{}
		for _, id := range ids {
			if _, err := uuid.Parse(id); err != nil {
				continue
			}
			n, _ := m.A.Rdb.Exists(r.Context(), presenceKey(p.Org, id)).Result()
			out[id] = n > 0
		}
		httpx.JSON(w, 200, map[string]any{"online": out})
		return nil
	}))
}

func presenceKey(org uuid.UUID, person string) string { return "pres:" + org.String() + ":" + person }

func (m *Module) Public(r chi.Router) {
	r.Get("/realtime", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("ticket")
		raw, err := m.A.Rdb.GetDel(r.Context(), "wst:"+id).Bytes() // single use
		if err != nil || id == "" {
			httpx.JSON(w, 401, map[string]any{"error": httpx.ErrUnauthorized})
			return
		}
		var t ticket
		if json.Unmarshal(raw, &t) != nil {
			httpx.JSON(w, 401, map[string]any{"error": httpx.ErrUnauthorized})
			return
		}
		if n, _ := m.A.Rdb.Exists(r.Context(), "revoked:"+t.Session.String()).Result(); n > 0 {
			httpx.JSON(w, 401, map[string]any{"error": httpx.ErrUnauthorized})
			return
		}
		up := realtime.Upgrade{W: w, R: r, Origins: m.A.Cfg.AllowedOrigins}
		channels := []string{realtime.OrgChannel(t.Org), realtime.AcctChannel(t.Account)}
		ctx := r.Context()
		pk := presenceKey(t.Org, t.Person.String())
		m.A.Rdb.Set(ctx, pk, 1, 70*time.Second)
		m.announce(ctx, t, true)
		defer func() { m.A.Rdb.Del(context.Background(), pk); m.announce(context.Background(), t, false) }()
		_ = m.A.Hub.Serve(ctx, up, t.Account, t.Org, channels, func(c *realtime.Conn, data []byte) { m.onFrame(ctx, c, t, pk, data) })
	})
}

func (m *Module) announce(ctx context.Context, t ticket, online bool) {
	d, _ := json.Marshal(map[string]any{"person": t.Person, "online": online})
	m.A.Publish(ctx, realtime.Event{Type: "presence", Channel: realtime.OrgChannel(t.Org), Org: t.Org.String(), Actor: t.Person.String(), Data: d})
}

type frame struct {
	Op   string    `json:"op"`
	Conv uuid.UUID `json:"conv"`
	Data any       `json:"data,omitempty"`
}

func (m *Module) onFrame(ctx context.Context, c *realtime.Conn, t ticket, pk string, data []byte) {
	var f frame
	if json.Unmarshal(data, &f) != nil {
		return
	}
	switch f.Op {
	case "ping":
		m.A.Rdb.Expire(ctx, pk, 70*time.Second)
		c.Send(map[string]any{"type": "pong"})
	case "sub":
		if m.A.Hub.Count(c) >= maxChannelsPerConn {
			c.Send(map[string]any{"type": "error", "code": "too_many_channels", "conv": f.Conv})
			return
		}
		ok, err := m.isMember(ctx, t, f.Conv)
		if err != nil || !ok {
			c.Send(map[string]any{"type": "error", "code": "forbidden", "conv": f.Conv})
			return
		}
		m.A.Hub.Subscribe(c, realtime.ConvChannel(f.Conv))
		c.Send(map[string]any{"type": "subscribed", "conv": f.Conv})
	case "unsub":
		m.A.Hub.Unsubscribe(c, realtime.ConvChannel(f.Conv))
	case "typing":
		if !m.A.Hub.Subscribed(c, realtime.ConvChannel(f.Conv)) {
			return
		}
		d, _ := json.Marshal(map[string]any{"person": t.Person, "name": t.Name})
		m.A.Publish(ctx, realtime.Event{Type: "typing", Channel: realtime.ConvChannel(f.Conv), Org: t.Org.String(), Actor: t.Person.String(), Data: d})
	}
}

func (m *Module) isMember(ctx context.Context, t ticket, conv uuid.UUID) (bool, error) {
	var ok bool
	err := m.A.DB.WithOrg(ctx, t.Org, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select exists (select 1 from hub.conversation_members where conversation_id=$1 and account_id=$2 and left_at is null)`, conv, t.Account).Scan(&ok)
	})
	return ok, err
}
