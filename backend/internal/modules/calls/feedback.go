package calls

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// Feedback: everyone who was in a call, meeting or standup that someone else also joined rates it once, 1 to 5 stars and an
// optional remark. A call nobody joined (missed, declined, alone in the room) asks for nothing.

// FeedbackWindow is how long after a call people are still asked for their feedback.
const FeedbackWindow = 7 * 24 * time.Hour

// Due is a call the person was in and has not rated yet.
type Due struct {
	Call    uuid.UUID `json:"call"`
	Kind    string    `json:"kind"`
	Context string    `json:"context"`
	Started time.Time `json:"started_at"`
	People  []string  `json:"people"`
}

// Rating is one person's feedback (shown to the person who started the call).
type Rating struct {
	Person uuid.UUID `json:"person"`
	Name   string    `json:"name"`
	Stars  int       `json:"stars"`
	Remark string    `json:"remark,omitempty"`
	At     time.Time `json:"at"`
}

func remarkCtx(org, id uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "call_feedback", Column: "remark_enc", RowID: id.String()}
}

// rateable: the person joined the call, at least one other person joined it too, and the person is no longer in it.
const rateable = `exists (select 1 from hub.call_participants me where me.call_id = c.id and me.person_id = $1 and me.joined_at is not null
		and (me.left_at is not null or c.ended_at is not null))
	and exists (select 1 from hub.call_participants o where o.call_id = c.id and o.person_id <> $1 and o.joined_at is not null)`

// FeedbackDue lists the caller's recent calls still waiting for their feedback, newest first.
func (m *Module) FeedbackDue(ctx context.Context, p *tenant.Principal) ([]Due, error) {
	out := []Due{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select c.id, c.kind, c.context, c.started_at,
			coalesce((select array_agg(x.first_name || ' ' || x.last_name order by x.first_name) from hub.call_participants o join hub.people x on x.id = o.person_id
				where o.call_id = c.id and o.person_id <> $1 and o.joined_at is not null), '{}')
			from hub.calls c where c.started_at > now() - make_interval(secs => $2) and `+rateable+`
			and not exists (select 1 from hub.call_feedback f where f.call_id = c.id and f.person_id = $1)
			order by c.started_at desc limit 20`, p.Person, FeedbackWindow.Seconds())
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var d Due
			if err := rs.Scan(&d.Call, &d.Kind, &d.Context, &d.Started, &d.People); err != nil {
				return err
			}
			out = append(out, d)
		}
		return rs.Err()
	})
	return out, err
}

// Rate records the caller's feedback on a call they were in. It is given once.
func (m *Module) Rate(ctx context.Context, p *tenant.Principal, call uuid.UUID, stars int, remark string) error {
	remark = strings.TrimSpace(remark)
	if stars < 1 || stars > 5 {
		return httpx.Bad("Rate it from 1 to 5 stars.")
	}
	if len(remark) > 2000 {
		return httpx.Bad("Keep the remark under 2000 characters.")
	}
	return m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.calls c where c.id = $2 and `+rateable+`)`, p.Person, call).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return httpx.Bad("Only people who were in this call, with someone else, can rate it, once they have left it.")
		}
		id := app.NewID()
		var enc []byte
		if remark != "" {
			dek, _, err := m.A.OrgDEK(ctx, tx, p.Org)
			if err != nil {
				return err
			}
			if enc, err = crypto.Field(dek, remarkCtx(p.Org, id), []byte(remark)); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `insert into hub.call_feedback (org_id, id, call_id, person_id, stars, remark_enc) values ($1,$2,$3,$4,$5,$6)`, p.Org, id, call, p.Person, stars, enc); err != nil {
			var pe *pgconn.PgError
			if errors.As(err, &pe) && pe.Code == "23505" {
				return httpx.Err(409, "conflict", "You have already rated this call.")
			}
			return err
		}
		return m.A.Sign.Seal(ctx, tx, "call_feedback", "id=$1", id)
	})
}

// Feedback is everything said about a call: only the person who started it reads it.
func (m *Module) Feedback(ctx context.Context, p *tenant.Principal, call uuid.UUID) ([]Rating, error) {
	out := []Rating{}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var host uuid.UUID
		if err := tx.QueryRow(ctx, `select host from hub.calls where id = $1`, call).Scan(&host); err != nil || host != p.Person {
			return httpx.ErrNotFound
		}
		dek, _, err := m.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select f.id, f.person_id, x.first_name || ' ' || x.last_name, f.stars, f.remark_enc, f.created_at
			from hub.call_feedback f join hub.people x on x.id = f.person_id where f.call_id = $1 order by f.created_at`, call)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			var r Rating
			var enc []byte
			if err := rs.Scan(&id, &r.Person, &r.Name, &r.Stars, &enc, &r.At); err != nil {
				return err
			}
			if len(enc) > 0 {
				if b, err := crypto.FieldOpen(dek, remarkCtx(p.Org, id), enc); err == nil {
					r.Remark = string(b)
				}
			}
			out = append(out, r)
		}
		return rs.Err()
	})
	return out, err
}

// RatedCall is one call in the organization's feedback overview, with every rating given on it.
type RatedCall struct {
	Call     uuid.UUID `json:"call"`
	Kind     string    `json:"kind"`
	Context  string    `json:"context"`
	Host     string    `json:"host_name"`
	Started  time.Time `json:"started_at"`
	Duration int       `json:"duration_seconds"`
	Average  float64   `json:"average"`
	Ratings  []Rating  `json:"ratings"`
}

type ContextStat struct {
	Context string  `json:"context"`
	Count   int     `json:"count"`
	Average float64 `json:"average"`
}

type OrgFeedbackPage struct {
	Calls    []RatedCall   `json:"calls"`
	Total    int           `json:"total"` // rated calls matching the filter
	Ratings  int           `json:"ratings"`
	Average  float64       `json:"average"`
	Contexts []ContextStat `json:"contexts"`
	Page     int           `json:"page"`
	Size     int           `json:"size"`
}

// OrgFeedback is the organization-wide overview of call, meeting and standup feedback (Settings: View settings), newest first.
// context: "" for all, or dm, forum, meeting, standup, project.
func (m *Module) OrgFeedback(ctx context.Context, p *tenant.Principal, context_ string, page, size int) (*OrgFeedbackPage, error) {
	if size <= 0 || size > 50 {
		size = 10
	}
	if page < 1 {
		page = 1
	}
	out := &OrgFeedbackPage{Calls: []RatedCall{}, Contexts: []ContextStat{}, Page: page, Size: size}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "View settings"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select c.context, count(*), avg(f.stars)::float8 from hub.call_feedback f join hub.calls c on c.id = f.call_id group by c.context order by c.context`)
		if err != nil {
			return err
		}
		for rs.Next() {
			var cs ContextStat
			if err := rs.Scan(&cs.Context, &cs.Count, &cs.Average); err != nil {
				rs.Close()
				return err
			}
			out.Contexts = append(out.Contexts, cs)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}
		filter := ` where ($1 = '' or c.context = $1) and exists (select 1 from hub.call_feedback f where f.call_id = c.id)`
		if err := tx.QueryRow(ctx, `select count(*), coalesce((select count(*) from hub.call_feedback f join hub.calls c on c.id = f.call_id where $1 = '' or c.context = $1), 0),
			coalesce((select avg(f.stars) from hub.call_feedback f join hub.calls c on c.id = f.call_id where $1 = '' or c.context = $1), 0)::float8
			from hub.calls c`+filter, context_).Scan(&out.Total, &out.Ratings, &out.Average); err != nil {
			return err
		}
		rs, err = tx.Query(ctx, `select c.id, c.kind, c.context, h.first_name || ' ' || h.last_name, c.started_at,
			coalesce(extract(epoch from (c.ended_at - coalesce(c.answered_at, c.started_at)))::int, 0),
			(select avg(f.stars)::float8 from hub.call_feedback f where f.call_id = c.id)
			from hub.calls c join hub.people h on h.id = c.host`+filter+` order by c.started_at desc, c.id limit $2 offset $3`, context_, size, (page-1)*size)
		if err != nil {
			return err
		}
		at := map[uuid.UUID]int{}
		var ids []uuid.UUID
		for rs.Next() {
			var c RatedCall
			if err := rs.Scan(&c.Call, &c.Kind, &c.Context, &c.Host, &c.Started, &c.Duration, &c.Average); err != nil {
				rs.Close()
				return err
			}
			if c.Duration < 0 {
				c.Duration = 0
			}
			c.Ratings = []Rating{}
			at[c.Call] = len(out.Calls)
			ids = append(ids, c.Call)
			out.Calls = append(out.Calls, c)
		}
		rs.Close()
		if err := rs.Err(); err != nil || len(ids) == 0 {
			return err
		}
		dek, _, err := m.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		rs, err = tx.Query(ctx, `select f.id, f.call_id, f.person_id, x.first_name || ' ' || x.last_name, f.stars, f.remark_enc, f.created_at
			from hub.call_feedback f join hub.people x on x.id = f.person_id where f.call_id = any($1) order by f.created_at`, ids)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id, call uuid.UUID
			var r Rating
			var enc []byte
			if err := rs.Scan(&id, &call, &r.Person, &r.Name, &r.Stars, &enc, &r.At); err != nil {
				return err
			}
			if len(enc) > 0 {
				if b, err := crypto.FieldOpen(dek, remarkCtx(p.Org, id), enc); err == nil {
					r.Remark = string(b)
				}
			}
			i := at[call]
			out.Calls[i].Ratings = append(out.Calls[i].Ratings, r)
		}
		return rs.Err()
	})
	return out, err
}

func (m *Module) feedbackRoutes(r chi.Router, org func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc) {
	r.Get("/calls/feedback", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := m.FeedbackDue(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"due": l})
		return nil
	}))
	r.Get("/calls/feedback/all", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		q := r.URL.Query()
		pg, _ := strconv.Atoi(q.Get("page"))
		n, _ := strconv.Atoi(q.Get("size"))
		c := q.Get("context")
		if c != "" && c != "dm" && c != "forum" && c != "meeting" && c != "standup" && c != "project" {
			return httpx.Bad("Unknown kind of call.")
		}
		out, err := m.OrgFeedback(r.Context(), p, c, pg, n)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	// {"stars": 1-5, "remark": "..."}
	r.Post("/calls/{id}/feedback", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in struct {
			Stars  int    `json:"stars"`
			Remark string `json:"remark"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		if err := m.Rate(r.Context(), p, id, in.Stars, in.Remark); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/calls/{id}/feedback", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		l, err := m.Feedback(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"feedback": l})
		return nil
	}))
}
