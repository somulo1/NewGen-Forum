// Package calls is the API side of voice and video calls: who may start or join a call, ringing, the history, and the pass
// that lets a device into a room on the calls service. The room itself (signaling) runs on the separate calls service.
package calls

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/callsvc"
	"hub/internal/calltoken"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/sfu"
	"hub/internal/tenant"
)

type Module struct {
	A    *app.App
	sfuc *sfu.Client
}

func New(a *app.App) *Module { return &Module{A: a} }

type Call struct {
	ID           uuid.UUID  `json:"id"`
	Kind         string     `json:"kind"`
	Context      string     `json:"context"`
	Conversation *uuid.UUID `json:"conversation,omitempty"`
	Host         uuid.UUID  `json:"host"`
	HostName     string     `json:"host_name"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	AnsweredAt   *time.Time `json:"answered_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	Duration     int        `json:"duration_seconds"`
	Direction    string     `json:"direction,omitempty"`
	People       []string   `json:"people,omitempty"`
}

type Entry struct {
	Call  *Call    `json:"call"`
	Token string   `json:"token,omitempty"`
	URL   string   `json:"url,omitempty"`
	Mode  string   `json:"mode"` // "mesh": small peer-to-peer room on the calls service. "sfu": big room on a media server.
	SFU   *SFUInfo `json:"sfu,omitempty"`
}

type StartIn struct {
	Conversation *uuid.UUID `json:"conversation"`
	Kind         string     `json:"kind"`
	Context      string     `json:"context"`
}

// roomURL pins a room to one calls-service instance by rendezvous hashing, so any API node picks the same one.
func (m *Module) roomURL(room string) string {
	best, bestScore := "", uint64(0)
	for _, u := range m.A.Cfg.CallsURLs {
		h := fnv.New64a()
		h.Write([]byte(u + "|" + room))
		if s := h.Sum64(); best == "" || s > bestScore {
			best, bestScore = u, s
		}
	}
	return strings.Replace(strings.Replace(best, "http://", "ws://", 1), "https://", "wss://", 1) + "/ws"
}

func (m *Module) token(p *tenant.Principal, c *Call, role string) (string, error) {
	claims := calltoken.Claims{Room: c.ID.String(), Org: p.Org.String(), Person: p.Person.String(), Name: p.Name, Role: role, Kind: c.Kind, MaxPeer: 8}
	claims.ID = uuid.NewString()
	return calltoken.Mint(m.A.Cfg.CallsSecret, claims, 90*time.Second)
}

// Start opens a room. For a direct message the other person's devices ring; for a forum or a meeting everyone in the
// conversation is told a call has begun.
func (m *Module) Start(ctx context.Context, p *tenant.Principal, in StartIn) (*Entry, error) {
	if in.Kind != "voice" && in.Kind != "video" {
		return nil, httpx.Bad("Choose voice or video.")
	}
	if in.Context == "" {
		in.Context = "meeting"
	}
	right := map[string][2]string{"dm": {"chat", "Private calls"}, "forum": {"forums", "Start calls"}, "meeting": {"sessions", "Start meeting"}, "standup": {"sessions", "Start meeting"}, "project": {"sessions", "Start meeting"}}[in.Context]
	if right[0] == "" {
		return nil, httpx.Bad("Unknown kind of call.")
	}
	var out *Entry
	var invited []uuid.UUID
	var accts []uuid.UUID
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, right[0], right[1]); err != nil {
			return err
		}
		if in.Conversation != nil {
			rs, err := tx.Query(ctx, `select p.id, p.account_id from hub.conversation_members mine
				join hub.conversation_members m on m.conversation_id = mine.conversation_id and m.left_at is null
				join hub.people p on p.account_id = m.account_id and p.org_id = m.org_id
				where mine.conversation_id=$1 and mine.account_id=$2 and mine.left_at is null and m.account_id <> $2 limit 500`, *in.Conversation, p.Account)
			if err != nil {
				return err
			}
			for rs.Next() {
				var id, a uuid.UUID
				if err := rs.Scan(&id, &a); err != nil {
					rs.Close()
					return err
				}
				invited, accts = append(invited, id), append(accts, a)
			}
			rs.Close()
			if len(invited) == 0 {
				return httpx.ErrNotFound
			}
		}
		id := app.NewID()
		status := "active"
		if in.Context == "dm" {
			status = "ringing"
		}
		if _, err := tx.Exec(ctx, `insert into hub.calls (org_id, id, kind, context, conversation_id, host, status) values ($1,$2,$3,$4,$5,$6,$7)`, p.Org, id, in.Kind, in.Context, in.Conversation, p.Person, status); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "calls", "id=$1 and org_id=$2", id, p.Org); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.call_participants (org_id, call_id, person_id, invited, joined_at) values ($1,$2,$3,false,null)`, p.Org, id, p.Person); err != nil {
			return err
		}
		for _, pid := range invited {
			if _, err := tx.Exec(ctx, `insert into hub.call_participants (org_id, call_id, person_id, invited) values ($1,$2,$3,true)`, p.Org, id, pid); err != nil {
				return err
			}
		}
		c := &Call{ID: id, Kind: in.Kind, Context: in.Context, Conversation: in.Conversation, Host: p.Person, HostName: p.Name, Status: status, StartedAt: app.Now()}
		if m.sfuFor(in.Context) {
			if err := m.reserveRoom(ctx, p.Org, id); err != nil {
				return err
			}
			info, err := m.sfuEntry(ctx, p, c, "host", true)
			if err != nil {
				m.releaseRoom(ctx, p.Org, id)
				return err
			}
			out = &Entry{Call: c, Mode: "sfu", SFU: info}
			return nil
		}
		tok, err := m.token(p, c, "host")
		if err != nil {
			return err
		}
		out = &Entry{Call: c, Mode: "mesh", Token: tok, URL: m.roomURL(id.String())}
		return nil
	})
	if err != nil {
		return nil, err
	}
	d, _ := json.Marshal(map[string]any{"call": out.Call.ID, "kind": in.Kind, "context": in.Context, "conversation": in.Conversation, "from": p.Person, "from_name": p.Name})
	typ := "call.started"
	if in.Context == "dm" {
		typ = "call.incoming"
	}
	for _, a := range accts {
		m.A.Publish(ctx, realtime.Event{Type: typ, Channel: realtime.AcctChannel(a), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	}
	// a ring that nobody answers becomes a missed call
	go func(id uuid.UUID) {
		time.Sleep(45 * time.Second)
		_ = m.A.DB.WithOrg(context.Background(), p.Org, func(tx pgx.Tx) error {
			tag, err := tx.Exec(context.Background(), `update hub.calls set status='missed', ended_at=now(), `+integrity.Bump+` where id=$1 and status='ringing'`, id)
			if err == nil && tag.RowsAffected() == 1 {
				err = m.A.Sign.Seal(context.Background(), tx, "calls", "id=$1", id)
			}
			return err
		})
	}(out.Call.ID)
	return out, nil
}

// Join gives someone invited (or in the conversation) a pass for a running call.
func (m *Module) Join(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Entry, error) {
	var out *Entry
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		c, err := m.load(ctx, tx, id, p.Person)
		if err != nil {
			return err
		}
		if c.Status == "ended" || c.Status == "missed" || c.Status == "declined" {
			return httpx.Err(410, "call_over", "This call has ended.")
		}
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.call_participants where call_id=$1 and person_id=$2)
			or exists (select 1 from hub.conversation_members where conversation_id=$3 and account_id=$4 and left_at is null)`, id, p.Person, c.Conversation, p.Account).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return httpx.ErrNotFound
		}
		role := m.roleOf(ctx, tx, c, p)
		if m.sfuFor(c.Context) {
			info, err := m.sfuEntry(ctx, p, c, role, true)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `insert into hub.call_participants (org_id, call_id, person_id, invited) values ($1,$2,$3,false) on conflict do nothing`, p.Org, id, p.Person); err != nil {
				return err
			}
			out = &Entry{Call: c, Mode: "sfu", SFU: info}
			return nil
		}
		if _, err := tx.Exec(ctx, `insert into hub.call_participants (org_id, call_id, person_id, invited) values ($1,$2,$3,false) on conflict do nothing`, p.Org, id, p.Person); err != nil {
			return err
		}
		tok, err := m.token(p, c, role)
		if err != nil {
			return err
		}
		out = &Entry{Call: c, Mode: "mesh", Token: tok, URL: m.roomURL(id.String())}
		return nil
	})
	return out, err
}

func (m *Module) Decline(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	var host uuid.UUID
	var hostAcct *uuid.UUID
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		c, err := m.load(ctx, tx, id, p.Person)
		if err != nil {
			return err
		}
		host = c.Host
		if c.Context != "dm" || c.Status != "ringing" || c.Host == p.Person {
			return httpx.Err(409, "cannot_decline", "Only a ringing direct call can be declined.")
		}
		if _, err := tx.Exec(ctx, `update hub.calls set status='declined', ended_at=now(), `+integrity.Bump+` where id=$1`, id); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "calls", "id=$1", id); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select account_id from hub.people where id=$1`, host).Scan(&hostAcct)
	})
	if err == nil && hostAcct != nil {
		d, _ := json.Marshal(map[string]any{"call": id, "by": p.Person, "by_name": p.Name})
		m.A.Publish(ctx, realtime.Event{Type: "call.declined", Channel: realtime.AcctChannel(*hostAcct), Org: p.Org.String(), Data: d})
	}
	return err
}

const sel = `select c.id, c.kind, c.context, c.conversation_id, c.host, h.first_name || ' ' || h.last_name, c.status, c.started_at, c.answered_at, c.ended_at,
	coalesce(extract(epoch from (coalesce(c.ended_at, now()) - c.answered_at))::int, 0)
	from hub.calls c join hub.people h on h.id = c.host`

func scan(r pgx.Row) (*Call, error) {
	c := &Call{}
	err := r.Scan(&c.ID, &c.Kind, &c.Context, &c.Conversation, &c.Host, &c.HostName, &c.Status, &c.StartedAt, &c.AnsweredAt, &c.EndedAt, &c.Duration)
	return c, err
}

func (m *Module) load(ctx context.Context, tx pgx.Tx, id, me uuid.UUID) (*Call, error) {
	c, err := scan(tx.QueryRow(ctx, sel+` where c.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	return c, err
}

// History lists the caller's calls, newest first: ones they made, ones they received, and ones they missed.
func (m *Module) History(ctx context.Context, p *tenant.Principal, limit int, before *time.Time) ([]*Call, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := []*Call{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		args := []any{p.Person}
		q := sel + ` where exists (select 1 from hub.call_participants cp where cp.call_id = c.id and cp.person_id = $1)`
		if before != nil {
			args = append(args, *before)
			q += ` and c.started_at < $2`
		}
		rs, err := tx.Query(ctx, q+` order by c.started_at desc limit `+strconv.Itoa(limit), args...)
		if err != nil {
			return err
		}
		for rs.Next() {
			c, err := scan(rs)
			if err != nil {
				rs.Close()
				return err
			}
			c.Direction = "in"
			if c.Host == p.Person {
				c.Direction = "out"
			}
			out = append(out, c)
		}
		rs.Close()
		for _, c := range out {
			ps, err := tx.Query(ctx, `select p.first_name || ' ' || p.last_name from hub.call_participants cp join hub.people p on p.id = cp.person_id where cp.call_id=$1 and cp.person_id <> $2 order by p.first_name limit 8`, c.ID, p.Person)
			if err != nil {
				return err
			}
			for ps.Next() {
				var n string
				if err := ps.Scan(&n); err != nil {
					ps.Close()
					return err
				}
				c.People = append(c.People, n)
			}
			ps.Close()
		}
		return nil
	})
	return out, err
}

// ---- events from the calls service ----

// Consume applies the calls service's events (someone joined, someone left, the room closed) to the history. The worker
// runs it. The queue is a Redis list, so events are not lost if the worker is briefly down.
func (m *Module) Consume(ctx context.Context) {
	for ctx.Err() == nil {
		res, err := m.A.Rdb.BLPop(ctx, 5*time.Second, callsvc.EventsKey).Result()
		if err != nil || len(res) < 2 {
			continue
		}
		var e struct {
			Type, Call, Org, Person string
			At                      time.Time
		}
		if json.Unmarshal([]byte(res[1]), &e) != nil {
			continue
		}
		org, e1 := uuid.Parse(e.Org)
		call, e2 := uuid.Parse(e.Call)
		if e1 != nil || e2 != nil {
			continue
		}
		person, _ := uuid.Parse(e.Person)
		_ = m.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
			switch e.Type {
			case "joined":
				if _, err := tx.Exec(ctx, `update hub.call_participants set joined_at=coalesce(joined_at,$3), left_at=null where call_id=$1 and person_id=$2`, call, person, e.At); err != nil {
					return err
				}
				// the first person other than the host to join answers the call
				tag, err := tx.Exec(ctx, `update hub.calls set status='active', answered_at=coalesce(answered_at,$2), `+integrity.Bump+` where id=$1 and status in ('ringing','active') and (status='ringing' and host <> $3 or answered_at is null and host <> $3)`, call, e.At, person)
				if err == nil && tag.RowsAffected() == 1 {
					err = m.A.Sign.Seal(ctx, tx, "calls", "id=$1", call)
					defer m.tell(org, call, "call.answered", person)
				}
				return err
			case "left":
				_, err := tx.Exec(ctx, `update hub.call_participants set left_at=$3 where call_id=$1 and person_id=$2`, call, person, e.At)
				return err
			case "ended":
				tag, err := tx.Exec(ctx, `update hub.calls set status = case when answered_at is null then 'missed' else 'ended' end, ended_at=coalesce(ended_at,$2), `+integrity.Bump+` where id=$1 and status in ('ringing','active')`, call, e.At)
				if err == nil && tag.RowsAffected() == 1 {
					err = m.A.Sign.Seal(ctx, tx, "calls", "id=$1", call)
					defer m.tell(org, call, "call.ended", person)
				}
				return err
			}
			return nil
		})
	}
}

// tell lets everyone on a call know it was answered or ended, so a phone that is still ringing (another tab, another device,
// or the caller who hung up first) stops. It runs after the update commits.
func (m *Module) tell(org, call uuid.UUID, typ string, by uuid.UUID) {
	ctx := context.Background()
	var accts []uuid.UUID
	_ = m.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select p.account_id from hub.call_participants cp join hub.people p on p.id = cp.person_id where cp.call_id=$1 and p.account_id is not null`, call)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var a uuid.UUID
			if rs.Scan(&a) == nil {
				accts = append(accts, a)
			}
		}
		return rs.Err()
	})
	d, _ := json.Marshal(map[string]any{"call": call, "by": by})
	for _, a := range accts {
		m.A.Publish(ctx, realtime.Event{Type: typ, Channel: realtime.AcctChannel(a), Org: org.String(), Data: d})
	}
}

func (m *Module) Public(r chi.Router) {
	m.sfuPublic(r)
}

func (m *Module) Private(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	m.sfuPrivate(r, org)
	r.Post("/calls", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in StartIn
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		e, err := m.Start(r.Context(), p, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, e)
		return nil
	}))
	r.Post("/calls/{id}/join", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		e, err := m.Join(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, e)
		return nil
	}))
	r.Post("/calls/{id}/decline", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		if err := m.Decline(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/calls/history", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		var before *time.Time
		if v := r.URL.Query().Get("before"); v != "" {
			t, err := time.Parse(time.RFC3339Nano, v)
			if err != nil {
				return httpx.Bad("Bad time.")
			}
			before = &t
		}
		l, err := m.History(r.Context(), p, n, before)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"calls": l})
		return nil
	}))
}

var _ = binary.BigEndian

// RecordPast writes one finished call into the history (loading history; no media was involved).
func (m *Module) RecordPast(ctx context.Context, org uuid.UUID, host uuid.UUID, kind, context_ string, conv *uuid.UUID, others []uuid.UUID, status string, startedAt time.Time, seconds int) error {
	return m.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		id := app.NewID()
		var answered, ended any
		if status == "ended" {
			a := startedAt.Add(8 * time.Second)
			answered, ended = a, a.Add(time.Duration(seconds)*time.Second)
		} else {
			ended = startedAt.Add(30 * time.Second)
		}
		if _, err := tx.Exec(ctx, `insert into hub.calls (org_id, id, kind, context, conversation_id, host, status, started_at, answered_at, ended_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			org, id, kind, context_, conv, host, status, startedAt, answered, ended); err != nil {
			return err
		}
		if err := m.A.Sign.Seal(ctx, tx, "calls", "id=$1 and org_id=$2", id, org); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.call_participants (org_id, call_id, person_id, invited, joined_at, left_at) values ($1,$2,$3,false,$4,$5)`, org, id, host, startedAt, ended); err != nil {
			return err
		}
		for _, o := range others {
			var joined any
			if status == "ended" {
				joined = answered
			}
			if _, err := tx.Exec(ctx, `insert into hub.call_participants (org_id, call_id, person_id, invited, joined_at, left_at) values ($1,$2,$3,true,$4,$5)`, org, id, o, joined, ended); err != nil {
				return err
			}
		}
		return nil
	})
}
