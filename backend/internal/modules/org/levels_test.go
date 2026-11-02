package org_test

import (
	"testing"

	"hub/internal/testkit"
)

// An organization uses only the levels it switched on, under the parents it chose, with its own names for them.
func TestStructureLevelsFollowTheOrganization(t *testing.T) {
	e := testkit.New(t)
	ownerU, _ := e.NewOrg("Olivia Owner")
	add := func(kind, name, parent string) (int, map[string]any) {
		body := map[string]any{"kind": kind, "name": name}
		if parent != "" {
			body["parent_id"] = parent
		}
		return e.Do("POST", "/org/units", body, ownerU.Access)
	}
	// nothing switched on: nothing can be made
	if code, _ := add("sector", "North", ""); code != 400 {
		t.Fatalf("a sector before sectors were switched on: %d", code)
	}
	// units under branches, branches under the organization; branches are called regions here
	levels := []any{
		map[string]any{"kind": "branch", "enabled": true, "label": "Region", "plural": "Regions", "parent": ""},
		map[string]any{"kind": "unit", "enabled": true, "parent": "branch"},
	}
	got := e.MustDo(200, "PUT", "/org/levels", map[string]any{"levels": levels}, ownerU.Access)
	if lv := got["levels"].([]any)[1].(map[string]any); lv["label"] != "Region" || lv["enabled"] != true {
		t.Fatalf("levels: %v", got)
	}
	code, br := add("branch", "Coast", "")
	if code != 201 {
		t.Fatalf("a region: %d", code)
	}
	if code, _ := add("unit", "Lonely", ""); code != 400 {
		t.Fatalf("a unit without its region: %d", code)
	}
	if code, _ := add("unit", "Retail", br["id"].(string)); code != 201 {
		t.Fatalf("a unit in a region: %d", code)
	}
	if code, _ := add("department", "IT", ""); code != 400 {
		t.Fatalf("a department when departments are off: %d", code)
	}
	// a level with entries cannot be switched off, nor moved under another level while its entries sit elsewhere
	off := []any{map[string]any{"kind": "branch", "enabled": true, "parent": ""}, map[string]any{"kind": "unit", "enabled": false}}
	if code, _ := e.Do("PUT", "/org/levels", map[string]any{"levels": off}, ownerU.Access); code != 409 {
		t.Fatalf("switching off a level that has entries: %d", code)
	}
	moved := []any{map[string]any{"kind": "sector", "enabled": true, "parent": ""}, map[string]any{"kind": "branch", "enabled": true, "parent": "sector"}, map[string]any{"kind": "unit", "enabled": true, "parent": "branch"}}
	if code, _ := e.Do("PUT", "/org/levels", map[string]any{"levels": moved}, ownerU.Access); code != 409 {
		t.Fatalf("putting regions under sectors while they sit under the organization: %d", code)
	}
	// no circles
	circle := []any{map[string]any{"kind": "sector", "enabled": true, "parent": "department"}, map[string]any{"kind": "department", "enabled": true, "parent": "sector"}, map[string]any{"kind": "branch", "enabled": true, "parent": ""}, map[string]any{"kind": "unit", "enabled": true, "parent": "branch"}}
	if code, _ := e.Do("PUT", "/org/levels", map[string]any{"levels": circle}, ownerU.Access); code != 400 {
		t.Fatalf("levels in a circle: %d", code)
	}
	// renaming a level keeps every record as it is
	levels[0] = map[string]any{"kind": "branch", "enabled": true, "label": "Area", "plural": "Areas", "parent": ""}
	e.MustDo(200, "PUT", "/org/levels", map[string]any{"levels": levels}, ownerU.Access)
	if n := len(e.MustDo(200, "GET", "/org/structure", nil, ownerU.Access)["units"].([]any)); n != 2 {
		t.Fatalf("renaming a level changed the records: %d", n)
	}
}
