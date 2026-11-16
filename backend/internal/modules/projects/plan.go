package projects

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// The plan of a project: tasks (epics, stories, tasks, bugs, subtasks), milestones or phases, sprints, risks, activity and documents.
// All plain rows. The server keeps the rules and records every change to every field.

var (
	methods    = map[string]bool{"agile": true, "waterfall": true, "hybrid": true}
	taskKinds  = map[string]bool{"epic": true, "story": true, "task": true, "bug": true, "subtask": true}
	collection = map[string]bool{"task": true, "milestone": true, "sprint": true, "risk": true, "activity": true, "page": true, "page_comment": true}
)

// Status is one column of a project's workflow. Cat says what it means for progress: 0 not started, 1 in progress, 2 being checked, 3 finished.
type Status struct {
	Key  string   `json:"key"`
	Name string   `json:"name"`
	Cat  int      `json:"cat"`
	WIP  int      `json:"wip,omitempty"` // most tasks wanted in this column at once; 0 means no limit
	To   []string `json:"to,omitempty"`  // the statuses a task may go to from here; empty means any
}

type Item struct {
	ID         uuid.UUID       `json:"id"`
	Collection string          `json:"collection"`
	Ref        *int            `json:"ref,omitempty"`
	Data       json.RawMessage `json:"data"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type Change struct {
	Item  uuid.UUID `json:"item"`
	Who   string    `json:"who"`
	At    time.Time `json:"at"`
	Field string    `json:"field"`
	Old   *string   `json:"old"`
	New   *string   `json:"new"`
}

// member returns the project, locked when asked, for someone on its team.
func (s *Service) member(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID, lock bool) (*Project, error) {
	if err := s.ensureRights(ctx, tx, p.Org); err != nil {
		return nil, err
	}
	var ok bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.conversation_members where conversation_id=$1 and account_id=$2 and left_at is null)`, id, p.Account).Scan(&ok); err != nil {
		return nil, err
	}
	if !ok {
		return nil, httpx.ErrNotFound
	}
	q := sel + ` where p.conversation_id=$1`
	if lock {
		q += ` for update of p`
	}
	pr, err := scan(tx.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	return pr, err
}

// Plan is everything in a project's plan with its whole change history.
func (s *Service) Plan(ctx context.Context, p *tenant.Principal, id uuid.UUID) ([]Item, []Change, error) {
	items, changes := []Item{}, []Change{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		_, err := s.member(ctx, tx, p, id, false)
		if err != nil && !errors.Is(err, httpx.ErrNotFound) {
			return err
		}
		onTeam := err == nil
		all, err := authz.Can(ctx, tx, p.Person, "projects", "View all projects")
		if err != nil {
			return err
		}
		own, err := authz.Can(ctx, tx, p.Person, "projects", "View own projects")
		if err != nil {
			return err
		}
		if !all && !(onTeam && own) {
			return httpx.ErrNotFound
		}
		if err := authz.Require(ctx, tx, p.Person, "projects", "View"); err != nil {
			return err
		}
		docs, err := authz.Can(ctx, tx, p.Person, "projects", "View documents")
		if err != nil {
			return err
		}
		history, err := authz.Can(ctx, tx, p.Person, "projects", "View change history")
		if err != nil {
			return err
		}
		pr, err := scan(tx.QueryRow(ctx, sel+` where p.conversation_id=$1`, id))
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		pages, err := loadPages(ctx, tx, id)
		if err != nil {
			return err
		}
		hidden := map[string]bool{}
		for pid := range pages {
			if !pageAllowed(pages, pid, pr, p.Person, "view") {
				hidden[pid] = true
			}
		}
		rs, err := tx.Query(ctx, `select id, collection, ref, data, created_at, updated_at from hub.project_items where project_id=$1 and ($2 or collection not in ('page','page_comment')) order by created_at, id`, id, docs)
		if err != nil {
			return err
		}
		for rs.Next() {
			var it Item
			if err := rs.Scan(&it.ID, &it.Collection, &it.Ref, &it.Data, &it.CreatedAt, &it.UpdatedAt); err != nil {
				rs.Close()
				return err
			}
			if it.Collection == "page" && hidden[it.ID.String()] {
				continue
			}
			if it.Collection == "page_comment" {
				var c map[string]any
				_ = json.Unmarshal(it.Data, &c)
				if hidden[text(c, "page")] {
					continue
				}
			}
			items = append(items, it)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}
		// burndown needs when work finished, so the status record stays; the rest needs the right
		hs, err := tx.Query(ctx, `select item_id, who, at, field, old_value, new_value from hub.project_item_history h where project_id=$1 and ($2 or field='col') and ($3 or not exists (select 1 from hub.project_items i where i.id=h.item_id and i.collection='page')) order by at, id`, id, history, docs)
		if err != nil {
			return err
		}
		defer hs.Close()
		for hs.Next() {
			var c Change
			if err := hs.Scan(&c.Item, &c.Who, &c.At, &c.Field, &c.Old, &c.New); err != nil {
				return err
			}
			if hidden[c.Item.String()] {
				continue
			}
			changes = append(changes, c)
		}
		return hs.Err()
	})
	return items, changes, err
}

// loadPages reads every page of a project, for its restrictions.
func loadPages(ctx context.Context, tx pgx.Tx, project uuid.UUID) (map[string]map[string]any, error) {
	rows, err := loadItems(ctx, tx, project, "page")
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[string]any, len(rows))
	for _, r := range rows {
		out[r.id.String()] = r.data
	}
	return out, nil
}

// pageAllowed: a page's restrictions list who may view or edit it, and a restriction on a page also applies to everything under it. Its
// creator and the project manager are never locked out. No list means everyone on the team.
func pageAllowed(pages map[string]map[string]any, id string, pr *Project, person uuid.UUID, what string) bool {
	if pr.Owner == person {
		return true
	}
	me := person.String()
	for cur, n := id, 0; cur != "" && n < 50; n++ {
		pg, ok := pages[cur]
		if !ok {
			return true
		}
		if text(pg, "createdBy") == me {
			return true
		}
		if rs, _ := pg["restrict"].(map[string]any); rs != nil {
			view, edit := strList(rs["view"]), strList(rs["edit"])
			// someone listed to edit may also read; a page you cannot read you cannot edit
			if len(view) > 0 && !contains(view, me) && !contains(edit, me) {
				return false
			}
			if what == "edit" && len(edit) > 0 && !contains(edit, me) {
				return false
			}
		}
		cur = text(pg, "parent")
	}
	return true
}

// LiveIn is one message between people editing a page together. The server only checks who may send it and passes it on.
type LiveIn struct {
	Kind string `json:"kind"` // hello | state | update | bye
	Data string `json:"data"`
	SID  string `json:"sid"`
}

// Live relays co-editing messages to the others on the project. Nothing is stored: the page is saved the usual way.
func (s *Service) Live(ctx context.Context, p *tenant.Principal, project, page uuid.UUID, in LiveIn) error {
	if !contains([]string{"hello", "state", "update", "bye"}, in.Kind) || in.SID == "" || len(in.SID) > 40 {
		return httpx.Bad("Unknown message.")
	}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, false)
		if err != nil {
			return err
		}
		if err := authz.Require(ctx, tx, p.Person, "projects", "Edit pages"); err != nil {
			return err
		}
		pages, err := loadPages(ctx, tx, project)
		if err != nil {
			return err
		}
		if !pageAllowed(pages, page.String(), pr, p.Person, "edit") {
			return httpx.ErrForbidden
		}
		return nil
	})
	if err != nil {
		return err
	}
	d, _ := json.Marshal(map[string]any{"project": project, "page": page, "kind": in.Kind, "data": in.Data, "sid": in.SID, "name": p.Name})
	s.A.Publish(ctx, realtime.Event{Type: "project.live", Channel: realtime.ConvChannel(project), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	return nil
}

func (s *Service) ensureRights(ctx context.Context, tx pgx.Tx, org uuid.UUID) error {
	return access.EnsureProjectRights(ctx, tx, org, func(table, where string, args ...any) error { return s.A.Sign.Seal(ctx, tx, table, where, args...) })
}

func text(m map[string]any, k string) string { v, _ := m[k].(string); return v }

func workflowOf(pr *Project) []Status {
	var w []Status
	_ = json.Unmarshal(pr.Workflow, &w)
	return w
}

func strList(v any) []string {
	l, _ := v.([]any)
	out := make([]string, 0, len(l))
	for _, x := range l {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// mayChange applies the rights of a collection. author is who created the item (nil on a new one).
func mayChange(ctx context.Context, tx pgx.Tx, p *tenant.Principal, pr *Project, coll string, isNew bool, author *uuid.UUID) error {
	mine := author != nil && *author == p.Person
	switch coll {
	case "task":
		if isNew {
			return authz.Require(ctx, tx, p.Person, "projects", "Edit tasks")
		}
		if mine {
			return authz.RequireOwn(ctx, tx, p.Person, "projects", "Move tasks", "Edit own tasks", true)
		}
		return authz.Require(ctx, tx, p.Person, "projects", "Move tasks")
	case "milestone":
		return authz.Require(ctx, tx, p.Person, "projects", "Manage milestones")
	case "sprint":
		return authz.Require(ctx, tx, p.Person, "projects", "Plan sprints")
	case "risk":
		return authz.Require(ctx, tx, p.Person, "projects", "Manage team")
	case "page":
		if isNew {
			return authz.Require(ctx, tx, p.Person, "projects", "Create pages")
		}
		return authz.Require(ctx, tx, p.Person, "projects", "Edit pages")
	case "page_comment":
		if isNew {
			return authz.Require(ctx, tx, p.Person, "projects", "Comment on pages")
		}
		if mine {
			return nil
		}
		return httpx.ErrForbidden
	case "activity":
		if isNew || mine {
			return nil
		}
		return httpx.ErrForbidden
	}
	return httpx.Bad("Unknown kind of item.")
}

// A change that touches only these fields needs only its own right, not the right to edit the task.
var narrowTask = []struct {
	right  string
	fields map[string]bool
}{
	{"Rank backlog", map[string]bool{"rank": true}},
	{"Plan sprints", map[string]bool{"sprint": true, "due": true}},
	{"Reschedule tasks", map[string]bool{"start": true, "due": true}},
	{"Save baseline", map[string]bool{"baseStart": true, "baseDue": true}},
}

func changedFields(old, next map[string]any, skip ...string) map[string]bool {
	out := map[string]bool{}
	for _, m := range []map[string]any{old, next} {
		for k := range m {
			ob, _ := json.Marshal(old[k])
			nb, _ := json.Marshal(next[k])
			if string(ob) != string(nb) {
				out[k] = true
			}
		}
	}
	for _, k := range skip {
		delete(out, k)
	}
	return out
}

// narrowRight finds the single right that covers this task change, if there is one.
func narrowRight(old, next map[string]any) string {
	ch := changedFields(old, next, "col")
	if len(ch) == 0 {
		return ""
	}
	for _, n := range narrowTask {
		ok := true
		for f := range ch {
			ok = ok && n.fields[f]
		}
		if ok {
			return n.right
		}
	}
	return ""
}

// taskRules: the project manager moves a task into or out of a finished status; its owners move it between the others (the manager may
// too) and are the only ones who edit it. Owners are read from the stored task, never from what is being sent.
func taskRules(ctx context.Context, tx pgx.Tx, p *tenant.Principal, pr *Project, old, next map[string]any) error {
	isManager := pr.Owner == p.Person
	finish, err := authz.Can(ctx, tx, p.Person, "projects", "Finish tasks")
	if err != nil {
		return err
	}
	editAny, err := authz.Can(ctx, tx, p.Person, "projects", "Edit any task")
	if err != nil {
		return err
	}
	me := p.Person.String()
	isOwner := text(old, "assignee") == me || text(old, "secondary_id") == me
	cat := map[string]int{}
	for _, st := range workflowOf(pr) {
		cat[st.Key] = st.Cat
	}
	from, to := text(old, "status"), text(next, "status")
	switch {
	case from != to && (cat[from] == 3 || cat[to] == 3):
		if !isManager && !finish {
			return httpx.Err(403, "manager_only", "Only the project manager can move a task to Done or back from Done.")
		}
	case from != to:
		if !isManager && !isOwner && !editAny {
			return httpx.Err(403, "owners_only", "Only the people the task is assigned to, or the project manager, can move it.")
		}
	default:
		if text(old, "assignee") == "" && text(old, "secondary_id") == "" {
			return nil
		}
		if !isOwner && !isManager && !editAny {
			return httpx.Err(403, "owners_only", "Only the people the task is assigned to can edit it.")
		}
	}
	return nil
}
