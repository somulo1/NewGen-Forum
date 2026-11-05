package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/notify"
	"hub/internal/tenant"
)

// What runs by itself, depending on the project's type:
//   every type: a parent follows its children; assignees are told; reminders before and after a due date; the project's health
//   agile, hybrid: a sprint past its end closes itself, records what it finished, and carries the rest forward
//   waterfall: work in a phase cannot start before the phases before it have passed their gate; a gate needs sign-off

const gateFlow = "f20"

var nairobi = func() *time.Location {
	l, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		return time.FixedZone("EAT", 3*3600)
	}
	return l
}()

type planRow struct {
	id   uuid.UUID
	ref  *int
	data map[string]any
}

func loadItems(ctx context.Context, tx pgx.Tx, project uuid.UUID, coll string) ([]planRow, error) {
	rs, err := tx.Query(ctx, `select id, ref, data from hub.project_items where project_id=$1 and collection=$2 order by created_at, id`, project, coll)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	var out []planRow
	for rs.Next() {
		var r planRow
		var b []byte
		if err := rs.Scan(&r.id, &r.ref, &b); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(b, &r.data)
		out = append(out, r)
	}
	return out, rs.Err()
}

// writeBySystem changes an item on Hub's own account, with its history.
func (s *Service) writeBySystem(ctx context.Context, tx pgx.Tx, org, project uuid.UUID, r planRow, field string, old, next any) error {
	b, _ := json.Marshal(r.data)
	if _, err := tx.Exec(ctx, `update hub.project_items set data=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, r.id, b); err != nil {
		return err
	}
	if err := s.A.Sign.Seal(ctx, tx, "project_items", "id=$1", r.id); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `insert into hub.project_item_history (org_id, item_id, project_id, who, field, old_value, new_value) values ($1,$2,$3,'Hub',$4,$5,$6)`, org, r.id, project, field, historyValue(old), historyValue(next))
	return err
}

func setStatus(r planRow, st Status) (bool, any) {
	old := r.data["status"]
	if old == st.Key {
		return false, old
	}
	r.data["status"], r.data["col"] = st.Key, st.Cat
	return true, old
}

// rollUp makes a parent follow its children: all finished finishes it, any work started starts it, and a reopened child reopens it.
func (s *Service) rollUp(ctx context.Context, tx pgx.Tx, org uuid.UUID, pr *Project, parent string) error {
	wf := workflowOf(pr)
	if len(wf) == 0 {
		return nil
	}
	rows, err := loadItems(ctx, tx, pr.ID, "task")
	if err != nil {
		return err
	}
	byID := map[string]planRow{}
	for _, r := range rows {
		byID[r.id.String()] = r
	}
	firstActive, last := wf[0], wf[len(wf)-1]
	for _, st := range wf {
		if st.Cat == 1 {
			firstActive = st
			break
		}
	}
	for n := 0; parent != "" && n < 50; n++ {
		p, ok := byID[parent]
		if !ok {
			return nil
		}
		total, done, started := 0, 0, 0
		for _, r := range rows {
			if text(r.data, "parent") != parent {
				continue
			}
			total++
			c, _ := r.data["col"].(float64)
			if c == 3 {
				done++
			}
			if c > 0 {
				started++
			}
		}
		cat, _ := p.data["col"].(float64)
		var to *Status
		switch {
		case total > 0 && done == total && cat != 3:
			to = &last
		case total > 0 && done < total && cat == 3:
			to = &firstActive
		case started > 0 && cat == 0:
			to = &firstActive
		}
		if to != nil {
			if changed, old := setStatus(p, *to); changed {
				if err := s.writeBySystem(ctx, tx, org, pr.ID, p, "status", old, to.Key); err != nil {
					return err
				}
			}
			p.data["col"] = float64(to.Cat)
		}
		parent = text(p.data, "parent")
	}
	return nil
}

// phaseOpen stops work starting in a waterfall phase before the earlier phases have passed their gate.
func phaseLocked(ctx context.Context, tx pgx.Tx, pr *Project, task map[string]any) (string, error) {
	mid := text(task, "milestone")
	if pr.Method != "waterfall" || mid == "" {
		return "", nil
	}
	rows, err := loadItems(ctx, tx, pr.ID, "milestone")
	if err != nil {
		return "", err
	}
	mine := ""
	for _, m := range rows {
		if m.id.String() == mid {
			mine = text(m.data, "iso")
		}
	}
	for _, m := range rows {
		if closed, _ := m.data["closed"].(bool); !closed && text(m.data, "iso") < mine {
			return text(m.data, "name"), nil
		}
	}
	return "", nil
}

func (s *Service) closeSprint(ctx context.Context, tx pgx.Tx, org uuid.UUID, pr *Project, sp planRow, sprints, tasks []planRow, who string) error {
	var next *planRow
	sort.Slice(sprints, func(i, j int) bool { return text(sprints[i].data, "start") < text(sprints[j].data, "start") })
	for i := range sprints {
		o := sprints[i]
		if closed, _ := o.data["closed"].(bool); !closed && o.id != sp.id && text(o.data, "start") >= text(sp.data, "start") {
			next = &o
			break
		}
	}
	size := func(t map[string]any) float64 { v, _ := t["points"].(float64); return v }
	var completed, committed, doneCount, openCount float64
	for _, t := range tasks {
		if text(t.data, "sprint") != sp.id.String() {
			continue
		}
		committed += size(t.data)
		if c, _ := t.data["col"].(float64); c == 3 {
			completed += size(t.data)
			doneCount++
			continue
		}
		openCount++
		to := ""
		if next != nil {
			to = next.id.String()
		}
		t.data["sprint"] = to
		if err := s.writeBySystem(ctx, tx, org, pr.ID, t, "sprint", sp.id.String(), to); err != nil {
			return err
		}
	}
	sp.data["closed"], sp.data["closedAt"], sp.data["closedBy"] = true, time.Now().UTC().Format(time.RFC3339), who
	sp.data["committed"], sp.data["completed"], sp.data["doneCount"], sp.data["carried"] = committed, completed, doneCount, openCount
	return s.writeBySystem(ctx, tx, org, pr.ID, sp, "closed", false, true)
}

// CloseSprint is someone closing a sprint before its end.
func (s *Service) CloseSprint(ctx context.Context, p *tenant.Principal, project, sprint uuid.UUID) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		if err := mayChange(ctx, tx, p, pr, "sprint", false, nil); err != nil {
			return err
		}
		sprints, err := loadItems(ctx, tx, project, "sprint")
		if err != nil {
			return err
		}
		tasks, err := loadItems(ctx, tx, project, "task")
		if err != nil {
			return err
		}
		for _, sp := range sprints {
			if sp.id == sprint {
				if c, _ := sp.data["closed"].(bool); c {
					return httpx.Err(409, "closed", "This sprint is already closed.")
				}
				return s.closeSprint(ctx, tx, p.Org, pr, sp, sprints, tasks, p.Name)
			}
		}
		return httpx.ErrNotFound
	})
	if err == nil {
		s.planChanged(ctx, p, project)
	}
	return err
}

func (s *Service) ensureGateFlow(ctx context.Context, tx pgx.Tx, org uuid.UUID) error {
	tag, err := tx.Exec(ctx, `insert into hub.approval_flows (org_id, key, module, action, levels) values ($1,$2,'projects','Pass a project phase gate',1) on conflict do nothing`, org, gateFlow)
	if err != nil || tag.RowsAffected() == 0 {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "approval_flows", "org_id=$1 and key=$2", org, gateFlow)
}

// PassGate asks for sign-off on a phase whose tasks are all finished. It closes when the approver agrees.
func (s *Service) PassGate(ctx context.Context, p *tenant.Principal, project, phase uuid.UUID, comment string) (*approvals.Request, error) {
	comment = strings.TrimSpace(comment)
	if len(comment) < 8 {
		return nil, httpx.Bad("Say what this phase achieved. It is kept with the gate.")
	}
	var req *approvals.Request
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		if err := mayChange(ctx, tx, p, pr, "milestone", false, nil); err != nil {
			return err
		}
		ms, err := loadItems(ctx, tx, project, "milestone")
		if err != nil {
			return err
		}
		name := ""
		for _, m := range ms {
			if m.id == phase {
				name = text(m.data, "name")
				if c, _ := m.data["closed"].(bool); c {
					return httpx.Err(409, "closed", "This phase has already passed its gate.")
				}
			}
		}
		if name == "" {
			return httpx.ErrNotFound
		}
		tasks, err := loadItems(ctx, tx, project, "task")
		if err != nil {
			return err
		}
		for _, t := range tasks {
			if c, _ := t.data["col"].(float64); text(t.data, "milestone") == phase.String() && c != 3 {
				return httpx.Err(409, "phase_open", fmt.Sprintf("%q is not finished yet.", text(t.data, "title")))
			}
		}
		if err := s.ensureGateFlow(ctx, tx, p.Org); err != nil {
			return err
		}
		r, applied, err := s.Ap.Submit(ctx, tx, p, gateFlow, "project.gate", "Pass gate: "+name+" ("+pr.Code+")", comment, map[string]any{"project": project, "phase": phase, "comment": comment, "by": p.Name})
		if err != nil {
			return err
		}
		if !applied {
			req = r
		}
		return nil
	})
	if err == nil {
		s.planChanged(ctx, p, project)
	}
	return req, err
}

func (s *Service) applyGate(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var pl struct {
		Project, Phase uuid.UUID
		Comment, By    string
	}
	if err := json.Unmarshal(a.Payload, &pl); err != nil {
		return err
	}
	ms, err := loadItems(ctx, tx, pl.Project, "milestone")
	if err != nil {
		return err
	}
	for _, m := range ms {
		if m.id == pl.Phase {
			now := time.Now().UTC().Format("2006-01-02")
			m.data["closed"], m.data["closedOn"], m.data["closedBy"], m.data["closedComment"] = true, now, pl.By, pl.Comment
			return s.writeBySystem(ctx, tx, a.Org, pl.Project, m, "closed", false, true)
		}
	}
	return nil
}

// Sweep runs the timed automation for one organization.
func (s *Service) Sweep(ctx context.Context, org uuid.UUID, now time.Time) error {
	var nb notify.Batch
	today := now.In(nairobi).Format("2006-01-02")
	soon := now.In(nairobi).AddDate(0, 0, 2).Format("2006-01-02")
	err := s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		var got bool
		if err := tx.QueryRow(ctx, `select pg_try_advisory_xact_lock(hashtextextended('projects-sweep:'||$1::text, 0))`, org).Scan(&got); err != nil || !got {
			return err
		}
		var ids []uuid.UUID
		rs, err := tx.Query(ctx, `select conversation_id from hub.projects where status not in ('Pending approval','Done','Cancelled') and plan_imported_at is not null`)
		if err != nil {
			return err
		}
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				rs.Close()
				return err
			}
			ids = append(ids, id)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}
		for _, id := range ids {
			pr, err := scan(tx.QueryRow(ctx, sel+` where p.conversation_id=$1 for update of p`, id))
			if err != nil {
				return err
			}
			tasks, err := loadItems(ctx, tx, id, "task")
			if err != nil {
				return err
			}
			if pr.Method != "waterfall" {
				sprints, err := loadItems(ctx, tx, id, "sprint")
				if err != nil {
					return err
				}
				for _, sp := range sprints {
					if c, _ := sp.data["closed"].(bool); !c && text(sp.data, "end") != "" && text(sp.data, "end") < today {
						if err := s.closeSprint(ctx, tx, org, pr, sp, sprints, tasks, "Hub"); err != nil {
							return err
						}
						if err := nb.Add(ctx, tx, org, pr.Owner, "project", text(sp.data, "name")+" in "+pr.Code+" ended and was closed. Unfinished work moved on.", "#7CC4FF", pr.Code); err != nil {
							return err
						}
					}
				}
			}
			open, late := 0, 0
			for _, t := range tasks {
				if c, _ := t.data["col"].(float64); c == 3 {
					continue
				}
				open++
				due := text(t.data, "due")
				if due != "" && due < today {
					late++
				}
				who, err := uuid.Parse(text(t.data, "assignee"))
				if err != nil || due == "" || due > soon || text(t.data, "remindedOn") == today {
					continue
				}
				msg := fmt.Sprintf("%q in %s is due %s.", text(t.data, "title"), pr.Code, due)
				color := "#FFB86B"
				if due < today {
					msg, color = fmt.Sprintf("%q in %s is overdue (was due %s).", text(t.data, "title"), pr.Code, due), "#FF9B8A"
				}
				if err := nb.Add(ctx, tx, org, who, "project", msg, color, pr.Code); err != nil {
					return err
				}
				t.data["remindedOn"] = today
				b, _ := json.Marshal(t.data)
				if _, err := tx.Exec(ctx, `update hub.project_items set data=$2, `+integrity.Bump+` where id=$1`, t.id, b); err != nil {
					return err
				}
				if err := s.A.Sign.Seal(ctx, tx, "project_items", "id=$1", t.id); err != nil {
					return err
				}
			}
			// health follows the work
			health := "On track"
			if pr.DueOn != nil && *pr.DueOn < today && open > 0 || open > 0 && late*4 >= open {
				health = "Off track"
			} else if late > 0 {
				health = "At risk"
			}
			if health != pr.Status {
				if _, err := tx.Exec(ctx, `update hub.projects set status=$2, updated_at=now(), `+integrity.Bump+` where conversation_id=$1`, id, health); err != nil {
					return err
				}
				if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", id); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err == nil {
		nb.Send(ctx, s.A)
	}
	return err
}

// SweepAll runs for every active organization (the worker calls this on a timer).
func (s *Service) SweepAll(ctx context.Context) {
	var orgs []uuid.UUID
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where status='active'`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				return err
			}
			orgs = append(orgs, id)
		}
		return rs.Err()
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("project sweep", "err", err)
	}
	for _, o := range orgs {
		if err := s.Sweep(ctx, o, time.Now()); err != nil {
			slog.Error("project sweep", "org", o, "err", err)
		}
	}
}

func (s *Service) automationRoutes(r chi.Router, org func(func(http.ResponseWriter, *http.Request, *tenant.Principal) error) http.HandlerFunc, ids func(*http.Request, string) (uuid.UUID, error)) {
	r.Post("/projects/{id}/sprints/{item}/close", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		if err := s.CloseSprint(r.Context(), p, id, item); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/projects/{id}/gates/{item}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		var in struct {
			Comment string `json:"comment"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		req, err := s.PassGate(r.Context(), p, id, item, in.Comment)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"approval": req})
		return nil
	}))
}
