package org_test

import (
	"testing"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// Moving and merging parts of the structure keeps everyone's placement whole, and tickets follow merged places.
func TestStructureMoveAndMerge(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	add := func(kind, name, parent string) string {
		body := map[string]any{"kind": kind, "name": name}
		if parent != "" {
			body["parent_id"] = parent
		}
		return e.MustDo(201, "POST", "/org/units", body, ownerU.Access)["id"].(string)
	}
	north, south := add("sector", "North", ""), add("sector", "South", "")
	nairobi := add("branch", "Nairobi", north)
	ops := add("unit", "Operations", nairobi)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	staff := e.Join(ownerU, slug, "Sam", "staff")
	for _, m := range []*testkit.Member{o, amy, staff} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+amy.Person, map[string]any{"sector_id": north, "branch_id": nairobi, "unit_id": ops}, ownerU.Access)
	placed := func() map[string]any {
		p := e.MustDo(200, "GET", "/people/"+amy.Person, nil, ownerU.Access)
		id := func(k string) any {
			if m, ok := p[k].(map[string]any); ok {
				return m["id"]
			}
			return nil
		}
		return map[string]any{"sector_id": id("sector"), "branch_id": id("branch"), "unit_id": id("unit")}
	}

	// staff cannot reshape the organization
	if code, _ := e.Do("POST", "/org/units/"+nairobi+"/move", map[string]any{"parent_id": south}, staff.Access); code != 403 {
		t.Fatalf("staff moved a branch: %d", code)
	}
	// a branch cannot sit under a unit, nor a unit under itself
	if code, _ := e.Do("POST", "/org/units/"+nairobi+"/move", map[string]any{"parent_id": ops}, ownerU.Access); code != 400 {
		t.Fatalf("branch under unit: %d", code)
	}
	if code, _ := e.Do("POST", "/org/units/"+ops+"/move", map[string]any{"parent_id": ops}, ownerU.Access); code != 400 {
		t.Fatalf("unit under itself: %d", code)
	}

	// the branch moves to the South sector: Amy, in a unit of that branch, is now in South
	e.MustDo(204, "POST", "/org/units/"+nairobi+"/move", map[string]any{"parent_id": south}, ownerU.Access)
	if p := placed(); p["sector_id"] != south || p["branch_id"] != nairobi || p["unit_id"] != ops {
		t.Fatalf("after moving the branch: %v", p)
	}

	// a ticket sent to the Operations unit
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": staff.WrapFor(e, key, 1, staff.Person, o.Person)}, staff.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "unit", "to_ref": ops}, staff.Access)

	// Operations merges into a new Support unit under another branch: Amy, her branch and the ticket follow
	mombasa := add("branch", "Mombasa", north)
	support := add("unit", "Support", mombasa)
	if code, _ := e.Do("POST", "/org/units/merge", map[string]any{"into": support, "from": []string{nairobi}}, ownerU.Access); code != 400 {
		t.Fatalf("merged a branch into a unit: %d", code)
	}
	e.MustDo(204, "POST", "/org/units/merge", map[string]any{"into": support, "from": []string{ops}}, ownerU.Access)
	if p := placed(); p["sector_id"] != north || p["branch_id"] != mombasa || p["unit_id"] != support {
		t.Fatalf("after merging the unit: %v", p)
	}
	if tk := e.MustDo(200, "GET", "/tickets/"+conv, nil, ownerU.Access); tk["to_ref"] != support {
		t.Fatalf("ticket did not follow the merge: %v", tk["to_ref"])
	}

	// the sectors merge: the branches under South (Nairobi) now sit under North
	e.MustDo(204, "POST", "/org/units/merge", map[string]any{"into": north, "from": []string{south}}, ownerU.Access)
	for _, u := range e.MustDo(200, "GET", "/org/structure", nil, ownerU.Access)["units"].([]any) {
		um := u.(map[string]any)
		if um["id"] == south {
			t.Fatal("the merged sector is still there")
		}
		if um["id"] == nairobi && um["parent_id"] != north {
			t.Fatalf("Nairobi did not move to North: %v", um)
		}
	}
}
