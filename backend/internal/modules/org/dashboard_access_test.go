package org_test

import (
	"testing"

	"hub/internal/testkit"
)

// A dashboard widget limited to a user group: only people whose role is in the group get it (GET /me/widgets) and may open the
// dashboard, even by typing its address; being placed in the department is not enough.
func TestUserGroupDecidesWhoOpensADashboard(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	cal := e.Join(ownerU, slug, "Cal", "hod")
	for _, m := range []*testkit.Member{amy, cal} {
		e.MustDo(200, "PATCH", "/people/"+m.Person, map[string]any{"department_id": dept}, ownerU.Access)
	}
	open := func(m *testkit.Member) int { c, _ := e.Do("GET", "/org/units/"+dept+"/dashboard", nil, m.Access); return c }
	// the departments widget is not switched on: nobody but an administrator opens department dashboards
	if c := open(amy); c != 403 {
		t.Fatalf("dashboard opened while its widget is off: %d", c)
	}
	group := e.MustDo(201, "POST", "/groups", map[string]any{"name": "Department heads", "roles": []string{"hod"}}, ownerU.Access)["id"].(string)
	e.MustDo(204, "PUT", "/settings/widgets", map[string]any{"departments": map[string]any{"enabled": true, "see": "Specific groups", "groups": []string{"ugroup:" + group}}}, ownerU.Access)
	allowed := func(m *testkit.Member) bool {
		for _, x := range e.MustDo(200, "GET", "/me/widgets", nil, m.Access)["allowed"].([]any) {
			if x == "departments" {
				return true
			}
		}
		return false
	}
	if !allowed(amy) || open(amy) != 200 {
		t.Fatalf("a member of the user group cannot use the dashboard: %v %d", allowed(amy), open(amy))
	}
	// Cal holds the same role but the group is changed to another role: he is placed in Finance, and still kept out
	e.MustDo(204, "PUT", "/settings/widgets", map[string]any{"departments": map[string]any{"enabled": true, "see": "Specific groups", "groups": []string{"ugroup:nobody"}}}, ownerU.Access)
	if allowed(cal) || open(cal) != 403 {
		t.Fatalf("someone outside the user group opened the dashboard: %v %d", allowed(cal), open(cal))
	}
	// for a dashboard, being placed in the department does not count: only user groups do
	e.MustDo(204, "PUT", "/settings/widgets", map[string]any{"departments": map[string]any{"enabled": true, "see": "Specific groups", "groups": []string{"departments:" + dept}}}, ownerU.Access)
	if allowed(amy) || open(amy) != 403 {
		t.Fatalf("a place opened a dashboard that only user groups may open: %v %d", allowed(amy), open(amy))
	}
	// switched on for everyone: anyone with the View dashboard right opens it again
	e.MustDo(204, "PUT", "/settings/widgets", map[string]any{"departments": map[string]any{"enabled": true, "see": "Everyone"}}, ownerU.Access)
	if open(cal) != 200 {
		t.Fatalf("dashboard for everyone: %d", open(cal))
	}
}
