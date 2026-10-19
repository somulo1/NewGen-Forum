// Package projects holds the plain facts about a project (owner, status, dates). Its name, goal, tasks, milestones and
// discussion are end-to-end encrypted in its conversation (see the chat package's records).
package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Service struct {
	A  *app.App
	Ap *approvals.Engine
}

func New(a *app.App, ap *approvals.Engine) *Service {
	s := &Service{A: a, Ap: ap}
	ap.Register("project.activate", s.applyActivate)
	return s
}

type Project struct {
	ID        uuid.UUID `json:"id"` // = its conversation
	Code      string    `json:"code"`
	Color     string    `json:"color"`
	Owner     uuid.UUID `json:"owner"`
	OwnerName string    `json:"owner_name"`
	Status    string    `json:"status"`
	StartOn   *string   `json:"start,omitempty"`
	DueOn     *string   `json:"due,omitempty"`
	Budget    *int      `json:"budget_pct,omitempty"`
	Sync      string    `json:"sync"`
	Links     []string  `json:"links"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateIn struct {
	Conversation uuid.UUID `json:"conversation"`
	Color        string    `json:"color"`
	Start        string    `json:"start"`
	Due          string    `json:"due"`
	Budget       *int      `json:"budget_pct"`
	Sync         string    `json:"sync"`
	Links        []string  `json:"links"`
}

const sel = `select p.conversation_id, p.code, p.color, p.owner, o.first_name || ' ' || o.last_name, p.status, p.start_on::text, p.due_on::text, p.budget_pct, p.sync, p.links, p.created_at
	from hub.projects p join hub.people o on o.id = p.owner`

func scan(r pgx.Row) (*Project, error) {
	x := &Project{}
	err := r.Scan(&x.ID, &x.Code, &x.Color, &x.Owner, &x.OwnerName, &x.Status, &x.StartOn, &x.DueOn, &x.Budget, &x.Sync, &x.Links, &x.CreatedAt)
	return x, err
}

func validDate(s string) (*string, error) {
	if s == "" {
		return nil, nil
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return nil, httpx.Bad("Dates look like 2026-12-15.")
	}
	return &s, nil
}

// Create registers a project for a conversation the creator's device already made. If the "create a project" flow needs
// approval, the project waits as "Pending approval" and goes live when the last approver agrees.
func (s *Service) Create(ctx context.Context, p *tenant.Principal, in CreateIn) (*Project, *approvals.Request, error) {
	start, err := validDate(in.Start)
	if err != nil {
		return nil, nil, err
	}
	due, err := validDate(in.Due)
	if err != nil {
		return nil, nil, err
	}
	if in.Budget != nil && (*in.Budget < 0 || *in.Budget > 100) {
		return nil, nil, httpx.Bad("Budget used is a percentage.")
	}
	if in.Color == "" {
		in.Color = "#7CC4FF"
	}
	if in.Links == nil {
		in.Links = []string{}
	}
	var out *Project
	var req *approvals.Request
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "projects", "Create project"); err != nil {
			return err
		}
		var kind string
		var creator uuid.UUID
		if err := tx.QueryRow(ctx, `select c.kind, c.created_by from hub.conversations c join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null where c.id=$1`, in.Conversation, p.Account).Scan(&kind, &creator); err != nil {
			return httpx.ErrNotFound
		}
		if kind != "project" || creator != p.Account {
			return httpx.Bad("Create the project conversation first.")
		}
		var n int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'project',1) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org).Scan(&n); err != nil {
			return err
		}
		code := fmt.Sprintf("PRJ-%02d", n)
		if _, err := tx.Exec(ctx, `insert into hub.projects (org_id, conversation_id, code, color, owner, status, start_on, due_on, budget_pct, sync, links) values ($1,$2,$3,$4,$5,'Pending approval',$6,$7,$8,$9,$10)`,
			p.Org, in.Conversation, code, in.Color, p.Person, start, due, in.Budget, in.Sync, in.Links); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", in.Conversation); err != nil {
			return err
		}
		r, applied, err := s.Ap.Submit(ctx, tx, p, "f2", "project.activate", "New project "+code, "", map[string]any{"project": in.Conversation})
		if err != nil {
			return err
		}
		if !applied {
			req = r
		}
		out, err = scan(tx.QueryRow(ctx, sel+` where p.conversation_id=$1`, in.Conversation))
		return err
	})
	return out, req, err
}

func (s *Service) applyActivate(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var pl struct {
		Project uuid.UUID `json:"project"`
	}
	if err := json.Unmarshal(a.Payload, &pl); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `update hub.projects set status='On track', updated_at=now(), `+integrity.Bump+` where conversation_id=$1 and status='Pending approval'`, pl.Project); err != nil {
		return err
	}
	if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", pl.Project); err != nil {
		return err
	}
	d, _ := json.Marshal(map[string]any{"project": pl.Project, "status": "On track"})
	go s.A.Publish(context.Background(), realtime.Event{Type: "project.updated", Channel: realtime.ConvChannel(pl.Project), Org: a.Org.String(), Data: d})
	return nil
}

func (s *Service) Mine(ctx context.Context, p *tenant.Principal) ([]*Project, error) {
	out := []*Project{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "projects", "View"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, sel+` where exists (select 1 from hub.conversation_members m where m.conversation_id = p.conversation_id and m.account_id=$1 and m.left_at is null) order by p.created_at desc`, p.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			x, err := scan(rs)
			if err != nil {
				return err
			}
			out = append(out, x)
		}
		return rs.Err()
	})
	return out, err
}

type Patch struct {
	Status *string    `json:"status"`
	Due    *string    `json:"due"`
	Budget *int       `json:"budget_pct"`
	Sync   *string    `json:"sync"`
	Owner  *uuid.UUID `json:"owner"`
}

var statuses = map[string]bool{"On track": true, "At risk": true, "Off track": true, "Delayed": true, "Done": true, "Cancelled": true}

func (s *Service) Update(ctx context.Context, p *tenant.Principal, id uuid.UUID, in Patch) (*Project, error) {
	if in.Status != nil && !statuses[*in.Status] {
		return nil, httpx.Bad("Choose a status from the list.")
	}
	var due *string
	if in.Due != nil {
		var err error
		if due, err = validDate(*in.Due); err != nil {
			return nil, err
		}
	}
	if in.Budget != nil && (*in.Budget < 0 || *in.Budget > 100) {
		return nil, httpx.Bad("Budget used is a percentage.")
	}
	var out *Project
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		cur, err := scan(tx.QueryRow(ctx, sel+` where p.conversation_id=$1 for update of p`, id))
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		if cur.Owner != p.Person {
			if err := authz.Require(ctx, tx, p.Person, "projects", "Manage team"); err != nil {
				return err
			}
		}
		if cur.Status == "Pending approval" {
			return httpx.Err(409, "pending", "This project is waiting for approval.")
		}
		if in.Owner != nil {
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.people p join hub.conversation_members m on m.account_id = p.account_id and m.conversation_id=$2 and m.left_at is null where p.id=$1)`, *in.Owner, id).Scan(&ok); err != nil || !ok {
				return httpx.Bad("The new owner must be on the project team.")
			}
		}
		if _, err := tx.Exec(ctx, `update hub.projects set status=coalesce($2,status), due_on=case when $3::boolean then $4::date else due_on end, budget_pct=coalesce($5,budget_pct), sync=coalesce($6,sync), owner=coalesce($7,owner), updated_at=now(), `+integrity.Bump+` where conversation_id=$1`,
			id, in.Status, in.Due != nil, due, in.Budget, in.Sync, in.Owner); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", id); err != nil {
			return err
		}
		if in.Status != nil {
			if err := s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Set %s to %s", cur.Code, *in.Status), "Projects", p.IP, map[string]any{"project": id}); err != nil {
				return err
			}
		}
		out, err = scan(tx.QueryRow(ctx, sel+` where p.conversation_id=$1`, id))
		return err
	})
	if err == nil {
		d, _ := json.Marshal(out)
		s.A.Publish(ctx, realtime.Event{Type: "project.updated", Channel: realtime.ConvChannel(id), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	}
	return out, err
}

func (s *Service) Public(chi.Router) {}

func (s *Service) Private(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	r.Post("/projects", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in CreateIn
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		pr, req, err := s.Create(r.Context(), p, in)
		if err != nil {
			return err
		}
		code := 201
		httpx.JSON(w, code, map[string]any{"project": pr, "approval": req})
		return nil
	}))
	r.Get("/projects", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := s.Mine(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"projects": l})
		return nil
	}))
	r.Patch("/projects/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in Patch
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		pr, err := s.Update(r.Context(), p, id, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, pr)
		return nil
	}))
}

var _ = strings.TrimSpace
