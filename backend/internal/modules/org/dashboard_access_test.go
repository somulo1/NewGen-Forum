package org_test

import (
	"testing"

	"hub/internal/testkit"
)

// One rule decides who opens a place's dashboard, set in the dashboard's Widget settings: by roles (the role's own "View
// dashboard"), or by user groups only (a user group the role is in has "View dashboard"; the role's own setting does not count).
// The sidebar's list (GET /me/widgets) and the dashboard itself, even at a typed address, always agree.
func TestOneRuleDecidesWhoOpensADashboard(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	e.MustDo(200, "PATCH", "/people/"+amy.Person, map[string]any{"department_id": dept}, ownerU.Access)
	open := func() int {
		c, _ := e.Do("GET", "/org/units/"+dept+"/dashboard", nil, amy.Access)
		return c
	}
	listed := func() bool {
		for _, x := range e.MustDo(200, "GET", "/me/widgets", nil, amy.Access)["allowed"].([]any) {
			if x == "departments" {
				return true
			}
		}
		return false
	}
	agree := func(want bool, why string) {
		t.Helper()
		if l, c := listed(), open(); l != want || (c == 200) != want {
			t.Fatalf("%s: listed=%v dashboard=%d, want open=%v", why, l, c, want)
		}
	}
	widgets := func(cfg map[string]any) { e.MustDo(204, "PUT", "/settings/widgets", map[string]any{"departments": cfg}, ownerU.Access) }

	// switched off: nobody but an administrator
	agree(false, "widget off")

	// by roles: the role's own right opens it, and taking that right away closes it
	widgets(map[string]any{"enabled": true, "access": "roles"})
	e.MustDo(200, "PUT", "/roles/hod/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	agree(true, "role given View dashboard, by roles")
	e.MustDo(200, "PUT", "/roles/hod/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": false}, ownerU.Access)
	agree(false, "role's right removed, by roles")

	// by user groups only: a group's right opens it for every role in the group; the role's own right alone does not
	group := e.MustDo(201, "POST", "/groups", map[string]any{"name": "Department heads", "roles": []string{"hod"}}, ownerU.Access)["id"].(string)
	widgets(map[string]any{"enabled": true, "access": "groups"})
	e.MustDo(200, "PUT", "/roles/hod/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	agree(false, "only the role's own right, by user groups")
	e.MustDo(200, "PUT", "/groups/"+group+"/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	agree(true, "user group given View dashboard")
	// the role leaves the group: the dashboard closes with it
	e.MustDo(204, "PATCH", "/groups/"+group, map[string]any{"roles": []string{}}, ownerU.Access)
	agree(false, "role taken out of the group")

	// a dashboard saved before the choice existed ("Specific groups") counts as by user groups
	e.MustDo(204, "PATCH", "/groups/"+group, map[string]any{"roles": []string{"hod"}}, ownerU.Access)
	widgets(map[string]any{"enabled": true, "see": "Specific groups", "groups": []string{"ugroup:" + group}})
	agree(true, "older settings, by user groups")

	// a choice that is neither is refused
	if c, _ := e.Do("PUT", "/settings/widgets", map[string]any{"departments": map[string]any{"enabled": true, "access": "everyone"}}, ownerU.Access); c != 400 {
		t.Fatalf("an unknown access mode was saved: %d", c)
	}
}
