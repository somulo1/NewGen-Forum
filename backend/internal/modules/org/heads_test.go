package org_test

import (
	"testing"

	"hub/internal/testkit"
)

// A unit's head is a role: everyone who holds that role and is placed in the unit heads it.
func TestUnitHeadIsARole(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": "", "branch": ""})
	add := func(kind, name string) string {
		return e.MustDo(201, "POST", "/org/units", map[string]any{"kind": kind, "name": name}, ownerU.Access)["id"].(string)
	}
	hr, fin, nairobi := add("department", "HR"), add("department", "Finance"), add("branch", "Nairobi")
	ann := e.Join(ownerU, slug, "Ann", "hod") // holds the head role, placed in HR
	bea := e.Join(ownerU, slug, "Bea", "hod") // holds the head role, placed in Finance
	cy := e.Join(ownerU, slug, "Cy", "staff") // placed in HR, a different role
	dee := e.Join(ownerU, slug, "Dee", "hod") // holds the head role but is placed nowhere
	e.MustDo(200, "PATCH", "/people/"+ann.Person, map[string]any{"department_id": hr, "branch_id": nairobi}, ownerU.Access)
	e.MustDo(200, "PATCH", "/people/"+bea.Person, map[string]any{"department_id": fin}, ownerU.Access)
	e.MustDo(200, "PATCH", "/people/"+cy.Person, map[string]any{"department_id": hr}, ownerU.Access)
	_ = dee

	heads := func(unit string) []string {
		var out []string
		for _, u := range e.MustDo(200, "GET", "/org/structure", nil, ownerU.Access)["units"].([]any) {
			m := u.(map[string]any)
			if m["id"] != unit {
				continue
			}
			hs, _ := m["heads"].([]any)
			for _, h := range hs {
				out = append(out, h.(map[string]any)["name"].(string))
			}
		}
		return out
	}
	if got := heads(hr); len(got) != 0 {
		t.Fatalf("a unit with no head role has heads: %v", got)
	}
	// only someone who may edit the organization can set it, and only to a real role
	if code, _ := e.Do("PUT", "/org/units/"+hr+"/head", map[string]any{"role": "hod"}, cy.Access); code != 403 {
		t.Fatalf("staff set a head role: %d", code)
	}
	if code, _ := e.Do("PUT", "/org/units/"+hr+"/head", map[string]any{"role": "nope"}, ownerU.Access); code != 400 {
		t.Fatalf("an invented role was accepted: %d", code)
	}
	e.MustDo(204, "PUT", "/org/units/"+hr+"/head", map[string]any{"role": "hod"}, ownerU.Access)
	// the head of HR is Ann (holds the role, placed in HR): not Cy (other role), not Bea (placed in Finance), not Dee (placed nowhere)
	if got := heads(hr); len(got) != 1 || got[0] != "Ann T" {
		t.Fatalf("heads of HR: %v", got)
	}
	// the same role heads another unit too, with that unit's own people
	e.MustDo(204, "PUT", "/org/units/"+fin+"/head", map[string]any{"role": "hod"}, ownerU.Access)
	if got := heads(fin); len(got) != 1 || got[0] != "Bea T" {
		t.Fatalf("heads of Finance: %v", got)
	}
	// changing it and clearing it
	e.MustDo(204, "PUT", "/org/units/"+hr+"/head", map[string]any{"role": "staff"}, ownerU.Access)
	if got := heads(hr); len(got) != 1 || got[0] != "Cy T" {
		t.Fatalf("heads of HR after the change: %v", got)
	}
	e.MustDo(204, "PUT", "/org/units/"+hr+"/head", map[string]any{"role": ""}, ownerU.Access)
	if got := heads(hr); len(got) != 0 {
		t.Fatalf("heads after clearing: %v", got)
	}
	// removing a unit takes its head role with it
	legal := add("department", "Legal")
	e.MustDo(204, "PUT", "/org/units/"+legal+"/head", map[string]any{"role": "hod"}, ownerU.Access)
	e.MustDo(204, "DELETE", "/org/units/"+legal, nil, ownerU.Access)
}
