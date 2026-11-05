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
