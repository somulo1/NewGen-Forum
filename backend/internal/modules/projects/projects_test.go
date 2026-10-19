package projects_test

import (
	"testing"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

func TestProjectNeedsApprovalThenGoesLive(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	hod := e.Join(ownerU, slug, "Hana", "hod")
	o.Register(e)
	hod.Register(e)

	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "project", "name": "Mobile App 2.0", "keys": hod.WrapFor(e, key, 1, hod.Person)}, hod.Access)["id"].(string)
	res := e.MustDo(201, "POST", "/projects", map[string]any{"conversation": conv, "start": "2026-09-01", "due": "2026-12-15", "budget_pct": 58}, hod.Access)
	pr := res["project"].(map[string]any)
	if pr["status"] != "Pending approval" || pr["code"] != "PRJ-01" || res["approval"] == nil {
		t.Fatalf("created: %v", res)
	}
	// until approved nobody can move it along
	if code, _ := e.Do("PATCH", "/projects/"+conv, map[string]any{"status": "At risk"}, hod.Access); code != 409 {
		t.Fatalf("pending project edited: %d", code)
	}
	// f2 has 3 levels; the super admin stands in at every level
	ap := res["approval"].(map[string]any)["id"].(string)
	for i := 0; i < 3; i++ {
		e.MustDo(200, "POST", "/approvals/"+ap+"/approve", map[string]any{}, ownerU.Access)
	}
	list := e.MustDo(200, "GET", "/projects", nil, hod.Access)["projects"].([]any)
	if len(list) != 1 || list[0].(map[string]any)["status"] != "On track" {
		t.Fatalf("not live: %v", list)
	}
	e.MustDo(200, "PATCH", "/projects/"+conv, map[string]any{"status": "At risk", "budget_pct": 70}, hod.Access)
	if code, _ := e.Do("PATCH", "/projects/"+conv, map[string]any{"status": "Fixed"}, hod.Access); code != 400 {
		t.Fatalf("bad status: %d", code)
	}
}
