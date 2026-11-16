package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/notify"
	"hub/internal/tenant"
)

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
		if isNew {
			// an id used by an item of another project must never reach across
			var elsewhere bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.project_items where id=$1)`, id).Scan(&elsewhere); err != nil {
				return err
			}
			if elsewhere {
				return httpx.Err(409, "other_project", "That item belongs to another project.")
			}
		}
		if err := checkLinks(ctx, tx, project, coll, data); err != nil {
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
			case len(ch) == 1 && ch["files"] && droppedFile(old, data):
				narrow = "Remove attachments"
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
		case "page_comment":
			pages, err := loadPages(ctx, tx, project)
			if err != nil {
				return err
			}
			if !pageAllowed(pages, text(data, "page"), pr, p.Person, "view") {
				return httpx.ErrNotFound
			}
		case "page":
			pages, err := loadPages(ctx, tx, project)
			if err != nil {
				return err
			}
			if !isNew && !pageAllowed(pages, id.String(), pr, p.Person, "view") {
				return httpx.ErrNotFound
			}
			if !isNew && narrow != "View documents" && !pageAllowed(pages, id.String(), pr, p.Person, "edit") {
				return httpx.Err(403, "page_restricted", "This page's restrictions do not let you change it.")
			}
			// taking a file off a page has its own right, whichever way the save arrives
			if droppedFile(old, data) {
				if err := authz.Require(ctx, tx, p.Person, "projects", "Remove attachments"); err != nil {
					return err
				}
			}
			oldR, _ := json.Marshal(old["restrict"])
			newR, _ := json.Marshal(data["restrict"])
			if string(oldR) != string(newR) {
				if err := authz.Require(ctx, tx, p.Person, "projects", "Restrict pages"); err != nil {
					return err
				}
			}
			if !isNew && data["createdBy"] == nil {
				data["createdBy"] = old["createdBy"]
			}
			if isNew {
				data["createdBy"] = p.Person.String()
			}
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

// checkLinks keeps a project to itself: what an item points at has to be in the same project.
func checkLinks(ctx context.Context, tx pgx.Tx, project uuid.UUID, coll string, data map[string]any) error {
	links := map[string][]string{
		"task":         {"sprint:sprint", "milestone:milestone"},
		"sprint":       {"milestone:milestone"},
		"page":         {"parent:page"},
		"page_comment": {"page:page"},
	}[coll]
	for _, l := range links {
		field, want, _ := strings.Cut(l, ":")
		v := text(data, field)
		if v == "" {
			continue
		}
		id, err := uuid.Parse(v)
		if err != nil {
			return httpx.Bad("That " + want + " is not in this project.")
		}
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.project_items where id=$1 and project_id=$2 and collection=$3)`, id, project, want).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return httpx.Bad("That " + want + " is not in this project.")
		}
	}
	return nil
}
