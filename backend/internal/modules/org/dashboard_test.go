package org_test

import (
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

func TestUnitDashboardIsBuiltFromRealRecords(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	bob := e.Join(ownerU, slug, "Bob", "staff")
	for _, m := range []*testkit.Member{o, amy, bob} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+amy.Person, map[string]any{"department_id": dept, "title": "HR Manager"}, ownerU.Access)

	// before anything happens the page is empty and healthy, with the default threshold
	d := e.MustDo(200, "GET", "/org/units/"+dept+"/dashboard", nil, ownerU.Access)
	if d["perf"].(float64) != 100 || d["target"].(float64) != 95 || d["members"].(float64) != 1 {
		t.Fatalf("empty dashboard: %v", d)
	}
	// people without the right cannot open it
	if code, _ := e.Do("GET", "/org/units/"+dept+"/dashboard", nil, bob.Access); code != 403 {
		t.Fatalf("staff opened a dashboard: %d", code)
	}

	// a ticket Amy raised to her own department, with the owner on it
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": dept}, amy.Access)
	// ... and escalated by the owner to the desk with a reason
	esc := e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "desk", "why": "Needs the IT desk"}, ownerU.Access)
	if esc["to_type"] != "desk" || esc["escalated"] != true {
		t.Fatalf("escalation: %v", esc)
	}
	if code, _ := e.Do("POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": uuid.NewString()}, ownerU.Access); code != 400 {
		t.Fatalf("escalated to a department that does not exist: %d", code)
	}
	e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": dept, "person": amy.Person, "why": "Back to Finance"}, ownerU.Access)

	// an approval request waiting for someone who holds the level-1 approval right (Amy, through a personal grant) shows up as received by Finance
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "people", "right": "Approve level 1", "allowed": true}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 1}, ownerU.Access)
	e.MustDo(202, "PATCH", "/people/"+bob.Person, map[string]any{"title": "Clerk"}, ownerU.Access)

	// a memo Amy sends to the whole organization
	k2, _ := e2ee.NewConversationKey()
	wall := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Everyone", "audience": "org", "keys": o.WrapFor(e, k2, 1, o.Person, amy.Person, bob.Person)}, ownerU.Access)["id"].(string)
	mid := uuid.Must(uuid.NewV7()).String()
	cm := e.MustDo(201, "POST", "/comms", map[string]any{"type": "Memo", "conversation": wall, "message": mid, "epoch": 1, "device": amy.DevID, "envelope": amy.Seal(org, wall, mid, 1, k2, "Budget dates")}, amy.Access)
	if cm["approval"] != nil {
		e.MustDo(200, "POST", "/approvals/"+cm["approval"].(map[string]any)["id"].(string)+"/approve", map[string]any{}, ownerU.Access)
	}

	d = e.MustDo(200, "GET", "/org/units/"+dept+"/dashboard", nil, ownerU.Access)
	if n := len(d["tickets"].([]any)); n != 1 {
		t.Fatalf("tickets: %v", d["tickets"])
	}
	tk := d["tickets"].([]any)[0].(map[string]any)
	if tk["dir"] != "Escalated" {
		t.Fatalf("ticket direction: %v", tk)
	}
	reqs := d["requests"].([]any)
	dirs := map[string]bool{}
	for _, r := range reqs {
		dirs[r.(map[string]any)["dir"].(string)] = r.(map[string]any)["subject"] != ""
	}
	if len(reqs) != 2 || !dirs["Received"] || !dirs["Sent"] {
		t.Fatalf("want one request received (waiting for the HR Manager) and one sent (her memo): %v", reqs)
	}
	comms := d["comms"].([]any)
	if len(comms) != 1 || comms[0].(map[string]any)["dir"] != "Sent" || comms[0].(map[string]any)["message"] != mid {
		t.Fatalf("comms: %v", comms)
	}
	parts := d["parts"].(map[string]any)
	if parts["tickets"].(map[string]any)["received"].(float64) != 1 {
		t.Fatalf("parts: %v", parts)
	}

	// the threshold: only people who may set targets, and only 50-100
	e.MustDo(204, "PUT", "/org/units/"+dept+"/target", map[string]any{"target": 80}, ownerU.Access)
	if code, _ := e.Do("PUT", "/org/units/"+dept+"/target", map[string]any{"target": 20}, ownerU.Access); code != 400 {
		t.Fatalf("silly threshold: %d", code)
	}
	if code, _ := e.Do("PUT", "/org/units/"+dept+"/target", map[string]any{"target": 90}, bob.Access); code != 403 {
		t.Fatalf("staff set a threshold: %d", code)
	}
	if e.MustDo(200, "GET", "/org/units/"+dept+"/dashboard", nil, ownerU.Access)["target"].(float64) != 80 {
		t.Fatal("threshold not saved")
	}
	found := false
	for _, u := range e.MustDo(200, "GET", "/org/structure", nil, ownerU.Access)["units"].([]any) {
		if m := u.(map[string]any); m["id"] == dept && m["target"].(float64) == 80 {
			found = true
		}
	}
	if !found {
		t.Fatal("structure does not carry the threshold")
	}
}
