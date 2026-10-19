package org_test

import (
	"testing"

	"hub/internal/testkit"
)

func TestStructureLevelsAreOptionalAndRemovable(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	staff := e.Join(ownerU, slug, "Sam", "staff")

	// an organization with only departments, and branches without sectors: nothing forces the levels above
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	br := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "branch", "name": "Nairobi"}, ownerU.Access)["id"].(string)
	un := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "unit", "name": "Retail Desk", "parent_id": br}, ownerU.Access)["id"].(string)
	// but what is given must fit
	if code, _ := e.Do("POST", "/org/units", map[string]any{"kind": "sector", "name": "Bad", "parent_id": br}, ownerU.Access); code != 400 {
		t.Fatalf("sector under a branch: %d", code)
	}
	if code, _ := e.Do("POST", "/org/units", map[string]any{"kind": "branch", "name": "Mombasa", "parent_id": un}, ownerU.Access); code != 400 {
		t.Fatalf("branch under a unit: %d", code)
	}
	// only people allowed to edit the organization may change it
	if code, _ := e.Do("DELETE", "/org/units/"+dept, nil, staff.Access); code != 403 {
		t.Fatalf("staff removed a department: %d", code)
	}
	// a unit with something under it, or people in it, cannot be removed
	if code, _ := e.Do("DELETE", "/org/units/"+br, nil, ownerU.Access); code != 409 {
		t.Fatalf("removed a branch that has a unit: %d", code)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(200, "PATCH", "/people/"+staff.Person, map[string]any{"department_id": dept}, ownerU.Access)
	if code, _ := e.Do("DELETE", "/org/units/"+dept, nil, ownerU.Access); code != 409 {
		t.Fatalf("removed a department that has people: %d", code)
	}
	// once empty it can go, and the level disappears from the structure
	e.MustDo(204, "DELETE", "/org/units/"+un, nil, ownerU.Access)
	e.MustDo(204, "DELETE", "/org/units/"+br, nil, ownerU.Access)
	n := 0
	for _, u := range e.MustDo(200, "GET", "/org/structure", nil, ownerU.Access)["units"].([]any) {
		if u.(map[string]any)["kind"] != "department" {
			t.Fatalf("a removed unit is still listed: %v", u)
		}
		n++
	}
	if n != 1 {
		t.Fatalf("want 1 unit left, got %d", n)
	}
}
