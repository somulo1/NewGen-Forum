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

	"hub/internal/access"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/notify"
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
		if _, err := s.member(ctx, tx, p, id, false); err != nil {
			return err
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
			changes = append(changes, c)
		}
		return hs.Err()
	})
	return items, changes, err
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

// checkTask validates a task against the rest of the plan and fills in its category.
func (s *Service) checkTask(ctx context.Context, tx pgx.Tx, pr *Project, id uuid.UUID, t map[string]any) error {
	title := strings.TrimSpace(text(t, "title"))
	if title == "" || len(title) > 300 {
		return httpx.Bad("Give the task a title of up to 300 characters.")
	}
	kind := text(t, "kind")
	if kind == "" {
		kind = "task"
		t["kind"] = kind
	}
	if !taskKinds[kind] {
		return httpx.Bad("A task is an epic, story, task, bug or subtask.")
	}
	if pts, has := t["points"]; has && pts != nil {
		n, ok := pts.(float64)
		if !ok || n < 0 || n > 1000 {
			return httpx.Bad("Story points are a number from 0 to 1000.")
		}
	}
	for _, f := range []string{"start", "due"} {
		if v := text(t, f); v != "" {
			if _, err := time.Parse("2006-01-02", v); err != nil {
				return httpx.Bad("Dates look like 2026-12-15.")
			}
		}
	}
	wf := workflowOf(pr)
	status := text(t, "status")
	if status == "" && len(wf) > 0 {
		status = wf[0].Key
		t["status"] = status
	}
	cat := -1
	for _, st := range wf {
		if st.Key == status {
			cat = st.Cat
		}
	}
	if cat < 0 {
		return httpx.Bad("That status is not in this project's workflow.")
	}
	t["col"] = cat

	// the rest of the plan, for parents and dependencies
	rs, err := tx.Query(ctx, `select id, data from hub.project_items where project_id=$1 and collection='task'`, pr.ID)
	if err != nil {
		return err
	}
	all := map[string]map[string]any{}
	for rs.Next() {
		var oid uuid.UUID
		var b []byte
		if err := rs.Scan(&oid, &b); err != nil {
			rs.Close()
			return err
		}
		var d map[string]any
		_ = json.Unmarshal(b, &d)
		all[oid.String()] = d
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	me := id.String()
	all[me] = t

	parent := text(t, "parent")
	if kind == "subtask" && parent == "" {
		return httpx.Bad("A subtask needs a parent task.")
	}
	if kind == "epic" && parent != "" {
		return httpx.Bad("An epic is the top of its tree, so it has no parent.")
	}
	for cur, n := parent, 0; cur != ""; n++ {
		if cur == me || n > 50 {
			return httpx.Bad("A task cannot sit under itself.")
		}
		up, ok := all[cur]
		if !ok {
			return httpx.Bad("The parent task is not in this project.")
		}
		cur = text(up, "parent")
	}
	deps := strList(t["deps"])
	for _, d := range deps {
		if _, ok := all[d]; !ok || d == me {
			return httpx.Bad("A task can only depend on other tasks in this project.")
		}
	}
	// a dependency loop would mean neither task can ever start
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(n string) bool {
		if n == me {
			return true
		}
		if seen[n] {
			return false
		}
		seen[n] = true
		for _, d := range strList(all[n]["deps"]) {
			if walk(d) {
				return true
			}
		}
		return false
	}
	for _, d := range deps {
		if walk(d) {
			return httpx.Bad("That dependency would make a loop: these tasks would wait for each other forever.")
		}
	}
	if cat == 3 {
		for _, d := range deps {
			if c, _ := all[d]["col"].(float64); c != 3 {
				return httpx.Err(409, "blocked", fmt.Sprintf("This task is blocked: %q is not finished yet.", text(all[d], "title")))
			}
		}
	}
	return nil
}

func historyValue(v any) *string {
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return &s
	}
	b, _ := json.Marshal(v)
	out := string(b)
	return &out
}

// recordChanges writes one history row per field that changed.
func recordChanges(ctx context.Context, tx pgx.Tx, p *tenant.Principal, project, item uuid.UUID, old, next map[string]any) error {
	fields := map[string]bool{}
	for k := range old {
		fields[k] = true
	}
	for k := range next {
		fields[k] = true
	}
	for f := range fields {
		if f == "history" { // the move comments already keep their own record
			continue
		}
		ob, _ := json.Marshal(old[f])
		nb, _ := json.Marshal(next[f])
		if string(ob) == string(nb) {
			continue
		}
		if _, err := tx.Exec(ctx, `insert into hub.project_item_history (org_id, item_id, project_id, person_id, who, field, old_value, new_value) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
			p.Org, item, project, p.Person, p.Name, f, historyValue(old[f]), historyValue(next[f])); err != nil {
			return err
		}
	}
	return nil
}

// Save creates or updates one item.
func (s *Service) Save(ctx context.Context, p *tenant.Principal, project, id uuid.UUID, coll string, data map[string]any) (*Item, error) {
	if !collection[coll] {
		return nil, httpx.Bad("Unknown kind of item.")
	}
	if id.Version() != 7 {
		return nil, httpx.Bad("Items need a time-ordered id.")
	}
	if data == nil {
		return nil, httpx.Bad("Nothing to save.")
	}
	out := &Item{}
	var nb notify.Batch
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		if !pr.Imported {
			return httpx.Err(409, "plan_moving", "This project's older plan is still being moved over. Reload in a moment.")
		}
		var oldB []byte
		var oldColl string
		var author *uuid.UUID
		err = tx.QueryRow(ctx, `select collection, data, created_by from hub.project_items where id=$1 and project_id=$2`, id, project).Scan(&oldColl, &oldB, &author)
		isNew := errors.Is(err, pgx.ErrNoRows)
		if err != nil && !isNew {
			return err
		}
		if !isNew && oldColl != coll {
			return httpx.Bad("An item cannot change its kind.")
		}
		old := map[string]any{}
		if !isNew {
			_ = json.Unmarshal(oldB, &old)
		}
		// a narrow change needs only its own right
		narrow := ""
		if coll == "task" && !isNew {
			narrow = narrowRight(old, data)
		}
		if coll == "page" && !isNew {
			ch := changedFields(old, data, "baseAt", "author")
			switch {
			case len(ch) == 1 && ch["watchers"]:
				narrow = "View documents"
			case len(ch) == 1 && ch["locked"]:
				narrow = "Lock pages"
			case len(ch) == 1 && ch["files"]:
				narrow = "Attach to pages"
			}
		}
		if coll == "page_comment" && !isNew {
			// anyone who may comment can resolve; the comment stays its writer's
			if ch := changedFields(old, data, "who"); len(ch) == 1 && ch["resolved"] {
				narrow = "Comment on pages"
			}
		}
		if narrow != "" {
			if err := authz.Require(ctx, tx, p.Person, "projects", narrow); err != nil {
				return err
			}
		} else if err := mayChange(ctx, tx, p, pr, coll, isNew, author); err != nil {
			return err
		}
		switch coll {
		case "task":
			if err := s.checkTask(ctx, tx, pr, id, data); err != nil {
				return err
			}
			if !isNew && narrow == "" {
				if err := taskRules(ctx, tx, p, pr, old, data); err != nil {
					return err
				}
				if text(old, "status") != text(data, "status") {
					for _, st := range workflowOf(pr) {
						if st.Key == text(old, "status") && len(st.To) > 0 && !contains(st.To, text(data, "status")) {
							return httpx.Err(409, "transition", "This workflow does not allow that move.")
						}
					}
				}
			}
			if c, _ := data["col"].(int); c > 0 {
				if was, _ := old["col"].(float64); was == 0 {
					name, err := phaseLocked(ctx, tx, pr, data)
					if err != nil {
						return err
					}
					if name != "" {
						return httpx.Err(409, "phase_locked", fmt.Sprintf("Phase %q has to pass its gate before this work starts.", name))
					}
				}
			}
		case "page":
			if t := strings.TrimSpace(text(data, "title")); t == "" || len(t) > 200 {
				return httpx.Bad("Give the page a title of up to 200 characters.")
			}
			if !isNew {
				// someone else saved while this was being edited: never overwrite them silently
				if base := text(data, "baseAt"); base != "" {
					var at time.Time
					if err := tx.QueryRow(ctx, `select updated_at from hub.project_items where id=$1`, id).Scan(&at); err != nil {
						return err
					}
					if t, err := time.Parse(time.RFC3339Nano, base); err == nil && !t.Equal(at) {
						return httpx.Err(409, "page_changed", "Someone else saved this page while you were editing. Reload it and add your changes again.")
					}
				}
				canLock, err := authz.Can(ctx, tx, p.Person, "projects", "Lock pages")
				if err != nil {
					return err
				}
				if locked, _ := old["locked"].(bool); locked && !canLock && pr.Owner != p.Person && (author == nil || *author != p.Person) && narrow != "View documents" {
					return httpx.Err(403, "page_locked", "This page is locked.")
				}
			}
			delete(data, "baseAt")
		}
		if coll == "page_comment" {
			data["who"] = p.Name // who wrote it is never taken from the browser
			if w := text(old, "who"); w != "" {
				data["who"] = w
			}
			if strings.TrimSpace(text(data, "text")) == "" {
				return httpx.Bad("Write the comment first.")
			}
		}
		b, _ := json.Marshal(data)
		if isNew {
			var ref *int
			if coll == "task" {
				var n int
				if err := tx.QueryRow(ctx, `select coalesce(max(ref),0)+1 from hub.project_items where project_id=$1`, project).Scan(&n); err != nil {
					return err
				}
				ref = &n
			}
			if _, err := tx.Exec(ctx, `insert into hub.project_items (org_id, id, project_id, collection, ref, data, created_by) values ($1,$2,$3,$4,$5,$6,$7)`, p.Org, id, project, coll, ref, b, p.Person); err != nil {
				return err
			}
		} else if _, err := tx.Exec(ctx, `update hub.project_items set data=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, id, b); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "project_items", "id=$1", id); err != nil {
			return err
		}
		if coll != "activity" && coll != "page_comment" {
			if err := recordChanges(ctx, tx, p, project, id, old, data); err != nil {
				return err
			}
		}
		if coll == "task" {
			// history and parents read the row as stored
			if err := s.rollUp(ctx, tx, p.Org, pr, text(data, "parent")); err != nil {
				return err
			}
			if old := text(old, "parent"); old != "" && old != text(data, "parent") {
				if err := s.rollUp(ctx, tx, p.Org, pr, old); err != nil {
					return err
				}
			}
			if who, err := uuid.Parse(text(data, "assignee")); err == nil && who != p.Person && text(old, "assignee") != who.String() {
				if err := nb.Add(ctx, tx, p.Org, who, "project", fmt.Sprintf("%s assigned %q in %s to you.", p.Name, text(data, "title"), pr.Code), "#7CC4FF", pr.Code); err != nil {
					return err
				}
			}
		}
		if coll == "page" && !isNew && text(old, "body") != text(data, "body") {
			for _, w := range strList(data["watchers"]) {
				if who, err := uuid.Parse(w); err == nil && who != p.Person {
					if err := nb.Add(ctx, tx, p.Org, who, "project", fmt.Sprintf("%s changed the page %q in %s.", p.Name, text(data, "title"), pr.Code), "#7CC4FF", pr.Code); err != nil {
						return err
					}
				}
			}
		}
		return tx.QueryRow(ctx, `select id, collection, ref, data, created_at, updated_at from hub.project_items where id=$1`, id).Scan(&out.ID, &out.Collection, &out.Ref, &out.Data, &out.CreatedAt, &out.UpdatedAt)
	})
	if err == nil {
		nb.Send(ctx, s.A)
		s.planChanged(ctx, p, project)
	}
	return out, err
}

func contains(l []string, v string) bool {
	for _, x := range l {
		if x == v {
			return true
		}
	}
	return false
}

// Delete removes an item. The history stays. A task others depend on, or with tasks under it, has to be untied first.
func (s *Service) Delete(ctx context.Context, p *tenant.Principal, project, id uuid.UUID) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		var coll string
		var author *uuid.UUID
		if err := tx.QueryRow(ctx, `select collection, created_by from hub.project_items where id=$1 and project_id=$2`, id, project).Scan(&coll, &author); err != nil {
			return httpx.ErrNotFound
		}
		switch coll {
		case "task":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete tasks")
		case "page":
			err = authz.Require(ctx, tx, p.Person, "projects", "Delete pages")
		default:
			err = mayChange(ctx, tx, p, pr, coll, false, author)
		}
		if err != nil {
			return err
		}
		if coll == "task" {
			var used bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.project_items where project_id=$1 and collection='task' and (data->>'parent' = $2 or data->'deps' ? $2))`, project, id.String()).Scan(&used); err != nil {
				return err
			}
			if used {
				return httpx.Err(409, "in_use", "Other tasks sit under this one or depend on it. Change them first.")
			}
		}
		if _, err := tx.Exec(ctx, `insert into hub.project_item_history (org_id, item_id, project_id, person_id, who, field) values ($1,$2,$3,$4,$5,'deleted')`, p.Org, id, project, p.Person, p.Name); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `delete from hub.project_items where id=$1`, id)
		return err
	})
	if err == nil {
		s.planChanged(ctx, p, project)
	}
	return err
}

// ImportItem is one record of an older encrypted plan, opened by a member's device.
type ImportItem struct {
	ID         uuid.UUID      `json:"id"`
	Collection string         `json:"collection"`
	Data       map[string]any `json:"data"`
	CreatedAt  *time.Time     `json:"created_at"`
}

// Import moves an older encrypted plan over, once. Any team member's device may do it: they can all read it already.
func (s *Service) Import(ctx context.Context, p *tenant.Principal, project uuid.UUID, items []ImportItem) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		pr, err := s.member(ctx, tx, p, project, true)
		if err != nil {
			return err
		}
		if pr.Imported {
			return nil // another member's device got there first
		}
		ref := 0
		for _, it := range items {
			if !collection[it.Collection] || it.Data == nil {
				return httpx.Bad("Unknown kind of item.")
			}
			var r *int
			if it.Collection == "task" {
				ref++
				n := ref
				r = &n
			}
			at := time.Now()
			if it.CreatedAt != nil {
				at = *it.CreatedAt
			}
			b, _ := json.Marshal(it.Data)
			if _, err := tx.Exec(ctx, `insert into hub.project_items (org_id, id, project_id, collection, ref, data, created_at) values ($1,$2,$3,$4,$5,$6,$7) on conflict do nothing`, p.Org, it.ID, project, it.Collection, r, b, at); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "project_items", "id=$1", it.ID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `update hub.projects set plan_imported_at=now(), updated_at=now(), `+integrity.Bump+` where conversation_id=$1`, project); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", project); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Moved the plan of %s out of end-to-end encryption", pr.Code), "Projects", p.IP, map[string]any{"project": project, "items": len(items)})
	})
	if err == nil {
		s.planChanged(ctx, p, project)
	}
	return err
}

// patchPlan changes a project's type and workflow. A status still holding tasks cannot be removed.
func (s *Service) patchPlan(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID, in Patch) error {
	if in.Method != nil {
		if err := authz.Require(ctx, tx, p.Person, "projects", "Change project type"); err != nil {
			return err
		}
		if !methods[*in.Method] {
			return httpx.Bad("Choose agile, waterfall or hybrid.")
		}
		if _, err := tx.Exec(ctx, `update hub.projects set method=$2, `+integrity.Bump+` where conversation_id=$1`, id, *in.Method); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", id); err != nil {
			return err
		}
	}
	if in.Workflow == nil {
		return nil
	}
	if err := authz.Require(ctx, tx, p.Person, "projects", "Manage workflow"); err != nil {
		return err
	}
	w := in.Workflow
	if len(w) < 2 || len(w) > 12 {
		return httpx.Bad("A workflow has 2 to 12 statuses.")
	}
	if w[0].Cat == 3 || w[len(w)-1].Cat != 3 {
		return httpx.Bad("A workflow starts with an unfinished status and ends with a finished one.")
	}
	keys := map[string]int{}
	for i := range w {
		w[i].Name = strings.TrimSpace(w[i].Name)
		if w[i].Name == "" || len(w[i].Name) > 40 || w[i].Key == "" || w[i].Cat < 0 || w[i].Cat > 3 || w[i].WIP < 0 {
			return httpx.Bad("Every status needs a name of up to 40 characters.")
		}
		if _, dup := keys[w[i].Key]; dup {
			return httpx.Bad("Two statuses are the same.")
		}
		keys[w[i].Key] = w[i].Cat
	}
	rs, err := tx.Query(ctx, `select id, data from hub.project_items where project_id=$1 and collection='task'`, id)
	if err != nil {
		return err
	}
	type row struct {
		id   uuid.UUID
		data map[string]any
	}
	var tasks []row
	for rs.Next() {
		var r row
		var b []byte
		if err := rs.Scan(&r.id, &b); err != nil {
			rs.Close()
			return err
		}
		_ = json.Unmarshal(b, &r.data)
		tasks = append(tasks, r)
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	for _, t := range tasks {
		cat, ok := keys[text(t.data, "status")]
		if !ok {
			return httpx.Err(409, "status_in_use", fmt.Sprintf("Move %q out of its status before removing that status.", text(t.data, "title")))
		}
		// a status's meaning changed: its tasks follow
		if c, _ := t.data["col"].(float64); int(c) != cat {
			t.data["col"] = cat
			b, _ := json.Marshal(t.data)
			if _, err := tx.Exec(ctx, `update hub.project_items set data=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, t.id, b); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "project_items", "id=$1", t.id); err != nil {
				return err
			}
		}
	}
	b, _ := json.Marshal(w)
	_, err = tx.Exec(ctx, `update hub.projects set workflow=$2, `+integrity.Bump+` where conversation_id=$1`, id, b)
	return err
}

func (s *Service) planChanged(ctx context.Context, p *tenant.Principal, project uuid.UUID) {
	d, _ := json.Marshal(map[string]any{"project": project, "by": p.Person})
	s.A.Publish(ctx, realtime.Event{Type: "project.plan", Channel: realtime.ConvChannel(project), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
}

func (s *Service) planRoutes(r chi.Router, org func(func(http.ResponseWriter, *http.Request, *tenant.Principal) error) http.HandlerFunc) {
	ids := func(r *http.Request, name string) (uuid.UUID, error) {
		id, err := uuid.Parse(chi.URLParam(r, name))
		if err != nil {
			return id, httpx.ErrNotFound
		}
		return id, nil
	}
	s.automationRoutes(r, org, ids)
	r.Get("/projects/{id}/plan", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		items, changes, err := s.Plan(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"items": items, "changes": changes})
		return nil
	}))
	r.Put("/projects/{id}/items/{item}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		var in struct {
			Collection string         `json:"collection"`
			Data       map[string]any `json:"data"`
		}
		if err := httpx.Decode(r, &in, 8<<20); err != nil {
			return err
		}
		out, err := s.Save(r.Context(), p, id, item, in.Collection, in.Data)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	r.Delete("/projects/{id}/items/{item}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		if err := s.Delete(r.Context(), p, id, item); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/projects/{id}/import", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		var in struct {
			Items []ImportItem `json:"items"`
		}
		if err := httpx.Decode(r, &in, 32<<20); err != nil {
			return err
		}
		if err := s.Import(r.Context(), p, id, in.Items); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}
