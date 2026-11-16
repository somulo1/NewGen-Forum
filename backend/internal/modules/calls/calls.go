// Package calls is the API side of voice and video calls: who may start or join a call, ringing, the history, and the pass
// that lets a device into a room on the calls service. The room itself (signaling) runs on the separate calls service.
package calls

import (
	"context"
	"encoding/json"
	"errors"
	"hash/fnv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/calltoken"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/orgsettings"
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
	Call   *Call       `json:"call"`
	Token  string      `json:"token,omitempty"`
	URL    string      `json:"url,omitempty"`
	Mode   string      `json:"mode"` // "mesh": small peer-to-peer room on the calls service. "sfu": big room on a media server.
	SFU    *SFUInfo    `json:"sfu,omitempty"`
	Online *bool       `json:"online,omitempty"` // private calls: whether the other person is signed in right now (false: they will only see a missed call)
	ICE    []ICEServer `json:"ice,omitempty"`    // STUN and TURN servers the browsers use to find each other across networks
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
	// five minutes, because the browser asks for the camera and microphone before it uses the pass (a permission prompt, a slow
	// network); the pass still works once
	return calltoken.Mint(m.A.Cfg.CallsSecret, claims, 5*time.Minute)
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
	if err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		b, err := orgsettings.BlocksOf(ctx, m.A, tx, p.Org)
		if err != nil {
			return err
		}
		if blocked := map[string]bool{"dm/voice": b.BlockPrivateVoice, "dm/video": b.BlockPrivateVideo, "forum/voice": b.BlockForumVoice, "forum/video": b.BlockForumVideo}[in.Context+"/"+in.Kind]; blocked {
			return httpx.Err(403, "blocked", "These calls are switched off for this organization.")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if id, ok := m.runningRoom(ctx, p, in.Conversation, in.Context); ok {
		return m.Join(ctx, p, id)
	}
	var out *Entry
	var invited []uuid.UUID
	var accts []uuid.UUID
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, right[0], right[1]); err != nil {
			return err
		}
		if err := m.checkOpen(ctx, tx, in.Context, in.Conversation, time.Now()); err != nil {
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
		// A private call is made only to someone who is signed in right now and free, and only by someone who is free themselves: nobody
		// is rung (and nothing is created) for a person who is offline or already on a call.
		if in.Context == "dm" {
			if err := m.checkReachable(ctx, tx, p, invited); err != nil {
				return err
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
		if _, err := tx.Exec(ctx, `insert into hub.call_participants (org_id, call_id, person_id, invited) select $1, $2, x, true from unnest($3::uuid[]) x`, p.Org, id, invited); err != nil {
			return err
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
		out = &Entry{Call: c, Mode: "mesh", Token: tok, URL: m.roomURL(id.String()), ICE: m.iceServers(p)}
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
		missed := false
		_ = m.A.DB.WithOrg(context.Background(), p.Org, func(tx pgx.Tx) error {
			tag, err := tx.Exec(context.Background(), `update hub.calls set status='missed', ended_at=now(), `+integrity.Bump+` where id=$1 and status='ringing'`, id)
			if err == nil && tag.RowsAffected() == 1 {
				err = m.A.Sign.Seal(context.Background(), tx, "calls", "id=$1", id)
				missed = err == nil
			}
			return err
		})
		if missed { // nobody answered: the caller stops ringing and everyone's ring clears
			m.tell(p.Org, id, "call.ended", uuid.Nil)
		}
	}(out.Call.ID)
	return out, nil
}

// Join gives someone invited (or in the conversation) a pass for a running call.
func (m *Module) Join(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*Entry, error) {
	var out *Entry
	answers := false
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		c, err := m.load(ctx, tx, id, p.Person)
		if err != nil {
			return err
		}
		answers = c.Status == "ringing" && c.Host != p.Person
		if c.Status == "ended" || c.Status == "missed" || c.Status == "declined" {
			return httpx.Err(410, "call_over", "This call has ended.")
		}
		if err := m.checkOpen(ctx, tx, c.Context, c.Conversation, time.Now()); err != nil {
			return err
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
		out = &Entry{Call: c, Mode: "mesh", Token: tok, URL: m.roomURL(id.String()), ICE: m.iceServers(p)}
		return nil
	})
	// Picking up is told at once, from here: the caller (and my other devices) stop ringing now, not after my microphone and the media
	// connection are ready (which can take seconds, or as long as the permission prompt stays open). The room's own "joined" event
	// still records the answer in the call history, a little later.
	if err == nil && answers {
		go m.tell(p.Org, id, "call.answered", p.Person)
	}
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

// DeleteFromLog removes entries from the caller's own call history for good. Only calls the caller was part of, and only ones that
// are over; the other people keep their copies.
func (m *Module) DeleteFromLog(ctx context.Context, p *tenant.Principal, ids []uuid.UUID) (int, error) {
	if len(ids) == 0 || len(ids) > 500 {
		return 0, httpx.Bad("Choose between 1 and 500 calls.")
	}
	n := 0
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `insert into hub.call_log_deletes (org_id, person_id, call_id)
			select $1, $2, c.id from hub.calls c where c.id = any($3) and c.status in ('ended','missed','declined')
			  and exists (select 1 from hub.call_participants cp where cp.call_id = c.id and cp.person_id = $2) on conflict do nothing`, p.Org, p.Person, ids)
		n = int(tag.RowsAffected())
		return err
	})
	return n, err
}

const sel = `select c.id, c.kind, c.context, c.conversation_id, c.host, h.first_name || ' ' || h.last_name, c.status, c.started_at, c.answered_at, c.ended_at,
	coalesce(extract(epoch from (coalesce(c.ended_at, now()) - c.answered_at))::int, 0)
	from hub.calls c join hub.people h on h.id = c.host`

// Leave is the caller hanging up (or anyone leaving). If the person who started the call leaves before anyone has answered, the
// call is over at once: it is marked missed and everyone still being rung is told on the spot, so their phones stop ringing
// without waiting for the room to notice. If someone has already joined, the call goes on without the caller.
func (m *Module) Leave(ctx context.Context, p *tenant.Principal, id uuid.UUID) (cancelled bool, err error) {
	var ended bool
	err = m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var host uuid.UUID
		var status string
		var answered *time.Time
		if err := tx.QueryRow(ctx, `select host, status, answered_at from hub.calls where id=$1 for update`, id).Scan(&host, &status, &answered); err != nil {
			return httpx.ErrNotFound
		}
		var member bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.call_participants where call_id=$1 and person_id=$2)`, id, p.Person).Scan(&member); err != nil || !member {
			return httpx.ErrNotFound
		}
		if _, err := tx.Exec(ctx, `update hub.call_participants set left_at=coalesce(left_at, now()) where call_id=$1 and person_id=$2`, id, p.Person); err != nil {
			return err
		}
		if alone, err := m.endIfAlone(ctx, tx, id); err != nil || alone {
			ended = alone
			return err
		}
		if host != p.Person || answered != nil || (status != "ringing" && status != "active") {
			return nil
		}
		tag, err := tx.Exec(ctx, `update hub.calls set status='missed', ended_at=now(), `+integrity.Bump+` where id=$1 and status in ('ringing','active') and answered_at is null`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 1 {
			cancelled = true
			return m.A.Sign.Seal(ctx, tx, "calls", "id=$1", id)
		}
		return nil
	})
	if err == nil && (cancelled || ended) {
		go m.tell(p.Org, id, "call.ended", p.Person)
	}
	return cancelled, err
}

// endIfAlone ends an answered private or forum call once only one person is left in it: nobody is there to talk to, so the call is
// over for them too, and its connection and media are released instead of running on for no one. It returns true when it ended it.
func (m *Module) endIfAlone(ctx context.Context, tx pgx.Tx, id uuid.UUID) (bool, error) {
	var kind, status string
	var answered *time.Time
	if err := tx.QueryRow(ctx, `select context, status, answered_at from hub.calls where id=$1 for update`, id).Scan(&kind, &status, &answered); err != nil {
		return false, nil
	}
	if (kind != "dm" && kind != "forum") || status != "active" || answered == nil {
		return false, nil
	}
	var n int
	if err := tx.QueryRow(ctx, `select count(*) from hub.call_participants where call_id=$1 and joined_at is not null and left_at is null`, id).Scan(&n); err != nil || n != 1 {
		return false, err
	}
	tag, err := tx.Exec(ctx, `update hub.calls set status='ended', ended_at=now(), `+integrity.Bump+` where id=$1 and status='active'`, id)
	if err != nil || tag.RowsAffected() != 1 {
		return false, err
	}
	return true, m.A.Sign.Seal(ctx, tx, "calls", "id=$1", id)
}

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
