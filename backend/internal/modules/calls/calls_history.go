package calls

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/callsvc"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// History lists the caller's calls, newest first: ones they made, ones they received, and ones they missed.
func (m *Module) History(ctx context.Context, p *tenant.Principal, limit int, before *time.Time) ([]*Call, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := []*Call{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		args := []any{p.Person}
		q := sel + ` where exists (select 1 from hub.call_participants cp where cp.call_id = c.id and cp.person_id = $1)
			and not exists (select 1 from hub.call_log_deletes d where d.call_id = c.id and d.person_id = $1)`
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
		if len(out) > 0 {
			ids := make([]uuid.UUID, len(out))
			at := make(map[uuid.UUID]*Call, len(out))
			for i, c := range out {
				ids[i] = c.ID
				at[c.ID] = c
			}
			ps, err := tx.Query(ctx, `select call_id, name from (
				select cp.call_id, p.first_name || ' ' || p.last_name as name, row_number() over (partition by cp.call_id order by p.first_name) as n
				from hub.call_participants cp join hub.people p on p.id = cp.person_id
				where cp.call_id = any($1) and cp.person_id <> $2) x where n <= 8 order by call_id, n`, ids, p.Person)
			if err != nil {
				return err
			}
			for ps.Next() {
				var cid uuid.UUID
				var n string
				if err := ps.Scan(&cid, &n); err != nil {
					ps.Close()
					return err
				}
				if c := at[cid]; c != nil {
					c.People = append(c.People, n)
				}
			}
			ps.Close()
		}
		return nil
	})
	return out, err
}

// LiveCall is a call happening right now that the person can see: one they are on, are being rung for, or that is running in a
// forum or private chat they belong to. Shown after sign-in (and after a reload) so a ringing call or a call in progress is never
// missed while the rest of the app is still loading.
type LiveCall struct {
	*Call
	Invited   bool `json:"invited"`    // I am on the call's list (rung, or invited to the meeting)
	InCall    bool `json:"in_call"`    // I am connected to it right now (another tab or device)
	RingingMe bool `json:"ringing_me"` // it is still ringing me and nobody has hung up on my side
	Joined    int  `json:"joined"`     // people connected now
}

// Live lists the calls in progress the person may see or join, newest first.
func (m *Module) Live(ctx context.Context, p *tenant.Principal) ([]*LiveCall, error) {
	out := []*LiveCall{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, sel+` where c.status in ('ringing','active') and c.started_at > now() - interval '12 hours'
			and (exists (select 1 from hub.call_participants cp where cp.call_id = c.id and cp.person_id = $1)
			     or (c.conversation_id is not null and exists (select 1 from hub.conversation_members cm where cm.conversation_id = c.conversation_id and cm.account_id = $2 and cm.left_at is null)))
			order by c.started_at desc limit 20`, p.Person, p.Account)
		if err != nil {
			return err
		}
		for rs.Next() {
			c, err := scan(rs)
			if err != nil {
				rs.Close()
				return err
			}
			out = append(out, &LiveCall{Call: c})
		}
		rs.Close()
		for _, l := range out {
			var invited, inCall, declined bool
			var joined int
			if err := tx.QueryRow(ctx, `select
				coalesce(bool_or(cp.person_id = $2 and cp.invited), false),
				coalesce(bool_or(cp.person_id = $2 and cp.joined_at is not null and cp.left_at is null), false),
				coalesce(bool_or(cp.person_id = $2 and cp.joined_at is null and cp.left_at is not null), false),
				count(*) filter (where cp.joined_at is not null and cp.left_at is null)
				from hub.call_participants cp where cp.call_id = $1`, l.ID, p.Person).Scan(&invited, &inCall, &declined, &joined); err != nil {
				return err
			}
			l.Invited, l.InCall, l.Joined = invited, inCall, joined
			l.RingingMe = l.Status == "ringing" && l.Host != p.Person && invited && !inCall && !declined
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
				if _, err := tx.Exec(ctx, `update hub.call_participants set left_at=$3 where call_id=$1 and person_id=$2`, call, person, e.At); err != nil {
					return err
				}
				// only one person left in a private or forum call: it ends for them too
				alone, err := m.endIfAlone(ctx, tx, call)
				if alone {
					defer m.tell(org, call, "call.ended", person)
				}
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
	m.feedbackRoutes(r, org)
	r.Post("/calls/quality", org(handleQuality))
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
	r.Post("/calls/{id}/leave", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		cancelled, err := m.Leave(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"cancelled": cancelled})
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
	// {"ids": [...]}: remove calls from my own history for good
	r.Post("/calls/history/delete", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in struct {
			IDs []uuid.UUID `json:"ids"`
		}
		if err := httpx.Decode(r, &in, 64<<10); err != nil {
			return err
		}
		n, err := m.DeleteFromLog(r.Context(), p, in.IDs)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"deleted": n})
		return nil
	}))
	r.Get("/calls/live", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := m.Live(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"calls": l})
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
