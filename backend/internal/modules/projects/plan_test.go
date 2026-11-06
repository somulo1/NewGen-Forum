package projects_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

func TestPlanIsPlainTrackedAndFollowsItsRules(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	pm := e.Join(ownerU, slug, "Hana", "hod")
	amy := e.Join(ownerU, slug, "Amy", "hod")
	cy := e.Join(ownerU, slug, "Cy", "hod")
	out := e.Join(ownerU, slug, "Oz", "hod")
	for _, m := range []*testkit.Member{o, pm, amy, cy, out} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "Alpha", "members": []string{amy.Person, cy.Person, o.Person}, "keys": pm.WrapFor(e, key, 1, pm.Person, amy.Person, cy.Person, o.Person)}, pm.Access)["id"].(string)
	res := e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-09-01", "due": "2026-12-15", "method": "waterfall"}, pm.Access)
	if res["project"].(map[string]any)["method"] != "waterfall" || res["project"].(map[string]any)["plan_imported"] != true {
		t.Fatalf("new project: %v", res)
	}
	for i := 0; i < 3; i++ {
		e.MustDo(200, "POST", "/approvals/"+res["approval"].(map[string]any)["id"].(string)+"/approve", map[string]any{}, ownerU.Access)
	}
	url := func(id string) string { return "/projects/" + conv + "/items/" + id }
	newID := func() string { return uuid.Must(uuid.NewV7()).String() }
	task := func(title string, extra map[string]any) map[string]any {
		d := map[string]any{"title": title, "status": "todo", "assignee": amy.Person, "due": "2026-10-01"}
		for k, v := range extra {
			d[k] = v
		}
		return map[string]any{"collection": "task", "data": d}
	}
	a, b := newID(), newID()
	if got := e.MustDo(200, "PUT", url(a), task("Design", map[string]any{"kind": "epic", "points": 5}), pm.Access); got["ref"].(float64) != 1 {
		t.Fatalf("ref: %v", got)
	}
	e.MustDo(200, "PUT", url(b), task("Build", map[string]any{"deps": []string{a}, "parent": a}), pm.Access)

	do := func(id string, want int, what string, m *testkit.Member, body map[string]any) {
		t.Helper()
		if code, r := e.Do("PUT", url(id), body, m.Access); code != want {
			t.Fatalf("%s: want %d got %d %v", what, want, code, r)
		}
	}
	do(newID(), 400, "a subtask with no parent", pm, task("Sub", map[string]any{"kind": "subtask"}))
	do(newID(), 400, "a bad kind", pm, task("X", map[string]any{"kind": "saga"}))
	do(newID(), 400, "a dependency elsewhere", pm, task("X", map[string]any{"deps": []string{newID()}}))
	do(a, 400, "a dependency loop", pm, task("Design", map[string]any{"kind": "epic", "deps": []string{b}}))
	do(b, 409, "finishing a blocked task", pm, task("Build", map[string]any{"deps": []string{a}, "parent": a, "status": "done"}))
	do(a, 403, "an owner finishing a task", amy, task("Design", map[string]any{"kind": "epic", "points": 5, "status": "done"}))
	do(a, 403, "someone else moving it", cy, task("Design", map[string]any{"kind": "epic", "points": 5, "status": "doing"}))
	do(a, 200, "an owner moving it", amy, task("Design", map[string]any{"kind": "epic", "points": 8, "status": "doing"}))
	do(a, 200, "the manager finishing it", pm, task("Design", map[string]any{"kind": "epic", "points": 8, "status": "done"}))
	do(b, 200, "finishing it once unblocked", pm, task("Build", map[string]any{"deps": []string{a}, "parent": a, "status": "done"}))
	if code, _ := e.Do("DELETE", url(a), nil, pm.Access); code != 409 {
		t.Fatalf("deleted a task others need: %d", code)
	}

	// a document, with its versions in the history
	pg := newID()
	do(pg, 200, "a page", amy, map[string]any{"collection": "page", "data": map[string]any{"title": "Runbook", "body": "one"}})
	do(pg, 200, "a new version", amy, map[string]any{"collection": "page", "data": map[string]any{"title": "Runbook", "body": "two"}})

	// a custom workflow
	if code, _ := e.Do("PATCH", "/projects/"+conv, map[string]any{"workflow": []map[string]any{{"key": "todo", "name": "To do", "cat": 0}, {"key": "shipped", "name": "Shipped", "cat": 3}}}, pm.Access); code != 409 {
		t.Fatalf("removed a status in use: %d", code)
	}
	got := e.MustDo(200, "PATCH", "/projects/"+conv, map[string]any{"method": "hybrid", "workflow": []map[string]any{{"key": "todo", "name": "To do", "cat": 0}, {"key": "test", "name": "QA", "cat": 2, "wip": 3}, {"key": "done", "name": "Live", "cat": 3}}}, pm.Access)
	if got["method"] != "hybrid" {
		t.Fatalf("patch: %v", got)
	}

	// outsiders see nothing; the whole history is there for the team
	if code, _ := e.Do("GET", "/projects/"+conv+"/plan", nil, out.Access); code != 404 {
		t.Fatalf("an outsider read the plan: %d", code)
	}
	plan := e.MustDo(200, "GET", "/projects/"+conv+"/plan", nil, cy.Access)
	raw, _ := json.Marshal(plan["changes"])
	for _, want := range []string{`"field":"status"`, `"field":"points"`, `"new":"two"`, `"who":"Amy`} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("history is missing %s: %s", want, raw)
		}
	}
	if len(plan["items"].([]any)) != 3 {
		t.Fatalf("items: %v", plan["items"])
	}
}

func TestProjectAutomation(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	mk := func(method string) string {
		key, _ := e2ee.NewConversationKey()
		conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": method, "members": []string{amy.Person}, "keys": o.WrapFor(e, key, 1, o.Person, amy.Person)}, ownerU.Access)["id"].(string)
		e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2020-01-01", "due": "2030-01-01", "method": method}, ownerU.Access)
		return conv
	}
	id := func() string { return uuid.Must(uuid.NewV7()).String() }
	put := func(conv, item, coll string, d map[string]any) {
		t.Helper()
		e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+item, map[string]any{"collection": coll, "data": d}, ownerU.Access)
	}
	item := func(conv, item string) map[string]any {
		for _, x := range e.MustDo(200, "GET", "/projects/"+conv+"/plan", nil, ownerU.Access)["items"].([]any) {
			if m := x.(map[string]any); m["id"] == item {
				return m["data"].(map[string]any)
			}
		}
		t.Fatal("no item")
		return nil
	}
	// a parent follows its children
	ag := mk("agile")
	parent, c1, c2 := id(), id(), id()
	put(ag, parent, "task", map[string]any{"title": "P"})
	put(ag, c1, "task", map[string]any{"title": "c1", "kind": "subtask", "parent": parent, "status": "doing"})
	if item(ag, parent)["status"] != "doing" {
		t.Fatal("parent did not start")
	}
	put(ag, c2, "task", map[string]any{"title": "c2", "kind": "subtask", "parent": parent, "status": "done"})
	put(ag, c1, "task", map[string]any{"title": "c1", "kind": "subtask", "parent": parent, "status": "done"})
	if item(ag, parent)["status"] != "done" {
		t.Fatal("parent did not finish")
	}
	// closing a sprint records its result and carries the rest forward
	s1, s2, open := id(), id(), id()
	put(ag, s1, "sprint", map[string]any{"name": "S1", "start": "2020-01-01", "end": "2020-01-14"})
	put(ag, s2, "sprint", map[string]any{"name": "S2", "start": "2020-01-15", "end": "2029-01-01"})
	put(ag, open, "task", map[string]any{"title": "left", "sprint": s1, "points": 3})
	e.MustDo(204, "POST", "/projects/"+ag+"/sprints/"+s1+"/close", nil, ownerU.Access)
	if item(ag, open)["sprint"] != s2 || item(ag, s1)["closed"] != true || item(ag, s1)["carried"].(float64) != 1 {
		t.Fatalf("sprint close: %v %v", item(ag, open), item(ag, s1))
	}
	// waterfall: later phases wait for the gate, and a gate needs finished work
	wf := mk("waterfall")
	p1, p2, t1, t2 := id(), id(), id(), id()
	put(wf, p1, "milestone", map[string]any{"name": "Design", "iso": "2026-01-01"})
	put(wf, p2, "milestone", map[string]any{"name": "Build", "iso": "2026-02-01"})
	put(wf, t1, "task", map[string]any{"title": "d", "milestone": p1})
	put(wf, t2, "task", map[string]any{"title": "b", "milestone": p2})
	if code, _ := e.Do("PUT", "/projects/"+wf+"/items/"+t2, map[string]any{"collection": "task", "data": map[string]any{"title": "b", "milestone": p2, "status": "doing"}}, ownerU.Access); code != 409 {
		t.Fatalf("later phase started: %d", code)
	}
	if code, _ := e.Do("POST", "/projects/"+wf+"/gates/"+p1, map[string]any{"comment": "all designed now"}, ownerU.Access); code != 409 {
		t.Fatalf("gate with open work: %d", code)
	}
	put(wf, t1, "task", map[string]any{"title": "d", "milestone": p1, "status": "done"})
	req := e.MustDo(200, "POST", "/projects/"+wf+"/gates/"+p1, map[string]any{"comment": "all designed now"}, ownerU.Access)["approval"].(map[string]any)
	e.MustDo(200, "POST", "/approvals/"+req["id"].(string)+"/approve", map[string]any{}, amy.Access)
	if item(wf, p1)["closed"] != true {
		t.Fatal("gate did not close")
	}
	put(wf, t2, "task", map[string]any{"title": "b", "milestone": p2, "status": "doing"})
	// the older plan is moved once
	old := mk("agile")
	e.MakePlanLegacy(old)
	task := id()
	if code, _ := e.Do("PUT", "/projects/"+old+"/items/"+id(), map[string]any{"collection": "task", "data": map[string]any{"title": "x"}}, ownerU.Access); code != 409 {
		t.Fatalf("wrote before the move: %d", code)
	}
	e.MustDo(204, "POST", "/projects/"+old+"/import", map[string]any{"items": []map[string]any{{"id": task, "collection": "task", "data": map[string]any{"title": "kept", "status": "todo"}}}}, amy.Access)
	e.MustDo(204, "POST", "/projects/"+old+"/import", map[string]any{"items": []map[string]any{{"id": id(), "collection": "task", "data": map[string]any{"title": "second go"}}}}, amy.Access)
	if item(old, task)["title"] != "kept" || len(e.MustDo(200, "GET", "/projects/"+old+"/plan", nil, ownerU.Access)["items"].([]any)) != 1 {
		t.Fatal("the move did not happen exactly once")
	}
}

// Every part of the plan follows its own right, set on the Roles page; older rights carry over.
func TestPlanFollowsItsRights(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "R", "members": []string{amy.Person}, "keys": o.WrapFor(e, key, 1, o.Person, amy.Person)}, ownerU.Access)["id"].(string)
	e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-01-01", "due": "2027-01-01"}, ownerU.Access)
	rights := e.MustDo(200, "GET", "/me", nil, amy.Access)["rights"].(map[string]any)["projects"].([]any)
	has := func(n string) bool {
		for _, r := range rights {
			if r == n {
				return true
			}
		}
		return false
	}
	if !has("Edit pages") || !has("Rank backlog") || has("Finish tasks") {
		t.Fatalf("starting rights: %v", rights)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	set := func(right string, on bool) {
		t.Helper()
		e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "projects", "right": right, "allowed": on}, ownerU.Access)
	}
	id := func() string { return uuid.Must(uuid.NewV7()).String() }
	page := id()
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+page, map[string]any{"collection": "page", "data": map[string]any{"title": "Doc", "body": "x"}}, amy.Access)
	cm := id()
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+cm, map[string]any{"collection": "page_comment", "data": map[string]any{"page": page, "text": "why?", "quote": "x", "who": "Someone else"}}, amy.Access)
	got0 := e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+cm, map[string]any{"collection": "page_comment", "data": map[string]any{"page": page, "text": "why?", "quote": "x", "resolved": true}}, ownerU.Access)
	if d := got0["data"].(map[string]any); !strings.HasPrefix(d["who"].(string), "Amy") || d["resolved"] != true {
		t.Fatalf("resolving a comment: %v", d)
	}
	// restrictions: a page only the manager may see disappears for others, with what sits under it
	secret, under := id(), id()
	amyKid := map[string]any{"title": "Under", "body": "y", "parent": secret}
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+secret, map[string]any{"collection": "page", "data": map[string]any{"title": "Secret", "body": "z", "restrict": map[string]any{"view": []string{o.Person}}}}, ownerU.Access)
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+under, map[string]any{"collection": "page", "data": amyKid}, ownerU.Access)
	for _, x := range e.MustDo(200, "GET", "/projects/"+conv+"/plan", nil, amy.Access)["items"].([]any) {
		if i := x.(map[string]any)["id"]; i == secret || i == under {
			t.Fatal("a restricted page was sent")
		}
	}
	if code, _ := e.Do("PUT", "/projects/"+conv+"/items/"+secret, map[string]any{"collection": "page", "data": map[string]any{"title": "Mine", "body": ""}}, amy.Access); code != 404 {
		t.Fatalf("changed a restricted page: %d", code)
	}
	if code, _ := e.Do("POST", "/projects/"+conv+"/pages/"+secret+"/live", map[string]any{"kind": "update", "sid": "a", "data": "x"}, amy.Access); code != 403 {
		t.Fatalf("joined a restricted page: %d", code)
	}
	e.MustDo(204, "POST", "/projects/"+conv+"/pages/"+page+"/live", map[string]any{"kind": "hello", "sid": "a"}, amy.Access)
	set("View documents", false)
	if items := e.MustDo(200, "GET", "/projects/"+conv+"/plan", nil, amy.Access)["items"].([]any); len(items) != 0 {
		t.Fatalf("documents shown without the right: %v", items)
	}
	set("Delete pages", false)
	if code, _ := e.Do("DELETE", "/projects/"+conv+"/items/"+page, nil, amy.Access); code != 403 {
		t.Fatalf("deleted without the right: %d", code)
	}
	// ranking needs only its right, not ownership of the task
	task := id()
	base := map[string]any{"title": "T", "due": "2026-05-01"}
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+task, map[string]any{"collection": "task", "data": base}, ownerU.Access)
	got := e.MustDo(200, "GET", "/projects/"+conv+"/plan", nil, ownerU.Access)["items"].([]any)
	var stored map[string]any
	for _, x := range got {
		if m := x.(map[string]any); m["id"] == task {
			stored = m["data"].(map[string]any)
		}
	}
	stored["rank"] = 1000
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+task, map[string]any{"collection": "task", "data": stored}, amy.Access)
	set("Rank backlog", false)
	stored["rank"] = 2000
	if code, _ := e.Do("PUT", "/projects/"+conv+"/items/"+task, map[string]any{"collection": "task", "data": stored}, amy.Access); code != 403 {
		t.Fatalf("ranked without the right: %d", code)
	}
	if code, _ := e.Do("PATCH", "/projects/"+conv, map[string]any{"method": "waterfall"}, amy.Access); code == 200 {
		set("Change project type", false)
		if code, _ := e.Do("PATCH", "/projects/"+conv, map[string]any{"method": "agile"}, amy.Access); code != 403 {
			t.Fatalf("changed type without the right: %d", code)
		}
	}
}

// A project is managed on its own: nothing in it can point into, or take over an item of, another project.
func TestProjectsStaySeparate(t *testing.T) {
	e := testkit.New(t)
	ownerU, _ := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	o.Register(e)
	mk := func(name string) string {
		key, _ := e2ee.NewConversationKey()
		conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": name, "members": []string{}, "keys": o.WrapFor(e, key, 1, o.Person)}, ownerU.Access)["id"].(string)
		e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-01-01", "due": "2027-01-01"}, ownerU.Access)
		return conv
	}
	a, b := mk("A"), mk("B")
	id := func() string { return uuid.Must(uuid.NewV7()).String() }
	ms, sp := id(), id()
	put := func(conv, item, coll string, d map[string]any) int {
		code, _ := e.Do("PUT", "/projects/"+conv+"/items/"+item, map[string]any{"collection": coll, "data": d}, ownerU.Access)
		return code
	}
	if c := put(a, ms, "milestone", map[string]any{"name": "M", "iso": "2026-06-01"}); c != 200 {
		t.Fatal(c)
	}
	if c := put(a, sp, "sprint", map[string]any{"name": "S", "start": "2026-02-01", "end": "2026-02-14", "milestone": ms}); c != 200 {
		t.Fatal(c)
	}
	if c := put(b, id(), "sprint", map[string]any{"name": "S", "milestone": ms}); c != 400 {
		t.Fatalf("a sprint joined another project's milestone: %d", c)
	}
	if c := put(b, id(), "task", map[string]any{"title": "T", "sprint": sp}); c != 400 {
		t.Fatalf("a task joined another project's sprint: %d", c)
	}
	if c := put(b, ms, "milestone", map[string]any{"name": "Taken", "iso": "2026-06-01"}); c != 409 {
		t.Fatalf("took another project's item: %d", c)
	}
	if n := len(e.MustDo(200, "GET", "/projects/"+b+"/plan", nil, ownerU.Access)["items"].([]any)); n != 0 {
		t.Fatalf("project B holds %d items", n)
	}
}

func TestDeletingFromAProject(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f2", map[string]any{"levels": 0}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "D", "members": []string{amy.Person}, "keys": o.WrapFor(e, key, 1, o.Person, amy.Person)}, ownerU.Access)["id"].(string)
	e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-01-01", "due": "2027-01-01"}, ownerU.Access)
	id := func() string { return uuid.Must(uuid.NewV7()).String() }
	put := func(item, coll string, d map[string]any) {
		t.Helper()
		e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+item, map[string]any{"collection": coll, "data": d}, ownerU.Access)
	}
	del := func(item string) int {
		c, _ := e.Do("DELETE", "/projects/"+conv+"/items/"+item, nil, ownerU.Access)
		return c
	}
	ms, sp, task := id(), id(), id()
	put(ms, "milestone", map[string]any{"name": "M", "iso": "2026-06-01"})
	put(sp, "sprint", map[string]any{"name": "S", "start": "2026-02-01", "end": "2026-02-14", "milestone": ms})
	put(task, "task", map[string]any{"title": "T", "sprint": sp})
	if c := del(ms); c != 409 {
		t.Fatalf("deleted a milestone in use: %d", c)
	}
	if c := del(sp); c != 204 {
		t.Fatalf("sprint: %d", c)
	}
	for _, x := range e.MustDo(200, "GET", "/projects/"+conv+"/plan", nil, ownerU.Access)["items"].([]any) {
		if m := x.(map[string]any); m["id"] == task && m["data"].(map[string]any)["sprint"] != "" {
			t.Fatalf("the task did not return to the backlog: %v", m)
		}
	}
	if c := del(ms); c != 204 {
		t.Fatalf("milestone: %d", c)
	}
	if c := del(task); c != 204 {
		t.Fatalf("task: %d", c)
	}
	// every deletion answers to its own right
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	for right, coll := range map[string]string{"Delete milestones": "milestone", "Delete sprints": "sprint", "Delete tasks": "task", "Delete pages": "page"} {
		item := id()
		put(item, coll, map[string]any{"title": "X", "name": "X", "iso": "2026-06-01", "start": "2026-02-01", "end": "2026-02-14"})
		e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "projects", "right": right, "allowed": false}, ownerU.Access)
		if c, _ := e.Do("DELETE", "/projects/"+conv+"/items/"+item, nil, amy.Access); c != 403 {
			t.Fatalf("deleted a %s without %q: %d", coll, right, c)
		}
		e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "projects", "right": right, "allowed": true}, ownerU.Access)
		if c, _ := e.Do("DELETE", "/projects/"+conv+"/items/"+item, nil, amy.Access); c != 204 {
			t.Fatalf("could not delete a %s with %q: %d", coll, right, c)
		}
	}
	off := func(right string) {
		e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "projects", "right": right, "allowed": false}, ownerU.Access)
	}
	// a risk, an attachment and a workflow status each answer to their own right
	risk := id()
	put(risk, "risk", map[string]any{"text": "R", "level": "Low"})
	off("Delete risks")
	if c, _ := e.Do("DELETE", "/projects/"+conv+"/items/"+risk, nil, amy.Access); c != 403 {
		t.Fatalf("deleted a risk: %d", c)
	}
	act := id()
	e.MustDo(200, "PUT", "/projects/"+conv+"/items/"+act, map[string]any{"collection": "activity", "data": map[string]any{"text": "A"}}, amy.Access)
	if c, _ := e.Do("DELETE", "/projects/"+conv+"/items/"+act, nil, amy.Access); c != 403 {
		t.Fatalf("deleted activity without the right: %d", c)
	}
	pg := id()
	file := map[string]any{"name": "f.pdf", "ref": "x"}
	put(pg, "page", map[string]any{"title": "Files", "body": "", "files": []any{file}})
	off("Remove attachments")
	if c, _ := e.Do("PUT", "/projects/"+conv+"/items/"+pg, map[string]any{"collection": "page", "data": map[string]any{"title": "Files", "body": "", "files": []any{}}}, amy.Access); c != 403 {
		t.Fatalf("removed an attachment: %d", c)
	}
	if c, _ := e.Do("PUT", "/projects/"+conv+"/items/"+pg, map[string]any{"collection": "page", "data": map[string]any{"title": "Edited", "body": "new", "files": []any{}}}, amy.Access); c != 403 {
		t.Fatalf("removed an attachment inside an edit: %d", c)
	}
	off("Remove workflow statuses")
	short := []map[string]any{{"key": "todo", "name": "To do", "cat": 0}, {"key": "doing", "name": "In progress", "cat": 1}, {"key": "done", "name": "Done", "cat": 3}}
	if c, _ := e.Do("PATCH", "/projects/"+conv, map[string]any{"workflow": short}, amy.Access); c != 403 {
		t.Fatalf("removed a status: %d", c)
	}
	if c, _ := e.Do("DELETE", "/projects/"+conv, nil, amy.Access); c != 403 {
		t.Fatalf("deleted a project without the right: %d", c)
	}
	e.MustDo(204, "DELETE", "/projects/"+conv, nil, ownerU.Access)
	if c, _ := e.Do("GET", "/projects/"+conv+"/plan", nil, ownerU.Access); c != 404 {
		t.Fatalf("still there: %d", c)
	}
}
