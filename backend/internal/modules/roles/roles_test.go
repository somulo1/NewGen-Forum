package roles_test

import (
	"encoding/base64"
	"github.com/google/uuid"
	"regexp"
	"strings"
	"testing"

	"hub/internal/testkit"
)

var tokenRe = regexp.MustCompile(`/join\?t=(\S+)`)

func join(t *testing.T, e *testkit.Env, owner *testkit.User, slug, first, role string) (*testkit.User, string) {
	t.Helper()
	email := testkit.Uniq()
	out := e.MustDo(201, "POST", "/people", map[string]any{"people": []any{map[string]any{"first_name": first, "last_name": "T", "email": email, "phone": "+254700000001",
		"date_of_birth": "1990-04-05", "gender": "Male", "role": role}}}, owner.Access)
	id := out["results"].([]any)[0].(map[string]any)["id"].(string)
	tok := tokenRe.FindStringSubmatch(e.Dev.To(email)[0].Body)[1]
	e.Signup(first, email)
	e.MustDo(200, "POST", "/invites/accept", map[string]any{"token": tok}, e.Login(email, "").Access)
	return e.Login(email, slug), id
}

func has(m map[string]any, mod, right string) bool {
	list, _ := m["rights"].(map[string]any)[mod].([]any)
	for _, r := range list {
		if r == right {
			return true
		}
	}
	return false
}

func TestRolesRightsAndMe(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	staff, sid := join(t, e, owner, slug, "Sam", "staff")

	me := e.MustDo(200, "GET", "/me", nil, staff.Access)
	if has(me, "security", "View logs") || !has(me, "wall", "Post") {
		t.Fatalf("staff default rights wrong: %v", me["rights"])
	}
	om := e.MustDo(200, "GET", "/me", nil, owner.Access)
	if !has(om, "roles", "Manage roles & rights") {
		t.Fatal("super admin lacks rights")
	}
	// staff cannot change rights
	if code, _ := e.Do("PUT", "/roles/staff/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, staff.Access); code != 403 {
		t.Fatalf("staff changed rights: %d", code)
	}
	// the super admin role cannot be limited
	if code, _ := e.Do("PUT", "/roles/owner/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, owner.Access); code != 409 {
		t.Fatalf("super admin was limited: %d", code)
	}
	// with 3 approval levels on f8 the change waits, and nothing happens until all approve
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 3}, owner.Access)
	res := e.MustDo(202, "PUT", "/roles/staff/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, owner.Access)
	if !has(e.MustDo(200, "GET", "/me", nil, staff.Access), "wall", "Post") {
		t.Fatal("rights changed before approval")
	}
	_ = res
	// make it direct and see the change take effect, with the audit trail written
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, owner.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, owner.Access)
	e.MustDo(200, "PUT", "/roles/staff/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, owner.Access)
	if has(e.MustDo(200, "GET", "/me", nil, staff.Access), "wall", "Post") {
		t.Fatal("role change did not take effect")
	}
	// personal override beats the role; null removes it
	e.MustDo(200, "PUT", "/people/"+sid+"/rights", map[string]any{"module": "wall", "right": "Post", "allowed": true}, owner.Access)
	if !has(e.MustDo(200, "GET", "/me", nil, staff.Access), "wall", "Post") {
		t.Fatal("personal override ignored")
	}
	e.MustDo(200, "PUT", "/people/"+sid+"/rights", map[string]any{"module": "wall", "right": "Post"}, owner.Access)
	if has(e.MustDo(200, "GET", "/me", nil, staff.Access), "wall", "Post") {
		t.Fatal("override not removed")
	}
	if code, _ := e.Do("PUT", "/people/"+sid+"/rights", map[string]any{"module": "nope", "right": "x", "allowed": true}, owner.Access); code != 400 {
		t.Fatalf("invented right accepted: %d", code)
	}

	roles := e.MustDo(200, "GET", "/roles", nil, owner.Access)["roles"].([]any)
	if len(roles) != 10 { // the test sample set
		t.Fatalf("want 10 roles, got %d", len(roles))
	}
	d := e.MustDo(200, "GET", "/roles/hod", nil, owner.Access)
	if d["rights"].(map[string]any)["tickets"].(map[string]any)["Close"] != true {
		t.Fatalf("hod matrix wrong: %v", d["rights"])
	}

	// my profile
	png := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("\x89PNG....fake"))
	e.MustDo(204, "PATCH", "/me", map[string]any{"status": "Focusing", "bio": "Hello there", "skills": []string{"Go", " ", "SQL"}, "phone": "+254711111111", "photo": png}, staff.Access)
	pf := e.MustDo(200, "GET", "/me", nil, staff.Access)["profile"].(map[string]any)
	if pf["status"] != "Focusing" || pf["bio"] != "Hello there" || pf["photo"] != png || len(pf["skills"].([]any)) != 2 || pf["phone"] != "+254711111111" {
		t.Fatalf("profile not saved: %v", pf)
	}
	if code, _ := e.Do("PATCH", "/me", map[string]any{"status": "Hacking"}, staff.Access); code != 400 {
		t.Fatalf("invalid status: %d", code)
	}
	if code, _ := e.Do("PATCH", "/me", map[string]any{"photo": "data:text/html;base64,PGI+"}, staff.Access); code != 400 {
		t.Fatalf("non-image photo: %d", code)
	}
}

// Roles are made by the organization, not built in: a new organization has only its owner role, and everything else (roles,
// what they may do, who approves each level) is created and changed through the API.
func TestRolesAreCreatedAndRightsAssignedByTheOrganization(t *testing.T) {
	e := testkit.New(t)
	// an organization made without the test sample set
	email := testkit.Uniq()
	e.Signup("Olivia Owner", email)
	u := e.Login(email, "")
	slug := "t-" + strings.ToLower(strings.ReplaceAll(uuid.NewString(), "-", "")[:14])
	e.MustDo(201, "POST", "/orgs", map[string]any{"name": "Bare Co", "slug": slug}, u.Access)
	owner := e.Login(email, slug)

	list := func() []any { return e.MustDo(200, "GET", "/roles", nil, owner.Access)["roles"].([]any) }
	if l := list(); len(l) != 1 || l[0].(map[string]any)["id"] != "owner" {
		t.Fatalf("a new organization should start with only its owner role: %v", l)
	}
	// the catalog is data: pages, rights, kinds and approval levels
	cat := e.MustDo(200, "GET", "/access/catalog", nil, owner.Access)
	kinds := map[string]bool{}
	approve := false
	for _, m := range cat["modules"].([]any) {
		for _, r := range m.(map[string]any)["rights"].([]any) {
			rm := r.(map[string]any)
			kinds[rm["kind"].(string)] = true
			approve = approve || rm["name"] == "Approve level 2"
		}
	}
	for _, k := range []string{"view", "create", "edit", "delete", "export", "approve"} {
		if !kinds[k] {
			t.Fatalf("catalog has no %s rights", k)
		}
	}
	if !approve {
		t.Fatal("no approval-level rights in the catalog")
	}

	// make a role, give it rights, give it to a person
	made := e.MustDo(201, "POST", "/roles", map[string]any{"name": "Records clerk", "desc": "Looks after records"}, owner.Access)
	key := made["id"].(string)
	if key != "records-clerk" {
		t.Fatalf("key: %v", key)
	}
	if code, _ := e.Do("POST", "/roles", map[string]any{"name": "records clerk"}, owner.Access); code != 409 {
		t.Fatalf("duplicate name allowed: %d", code)
	}
	e.MustDo(200, "PUT", "/roles/"+key+"/rights", map[string]any{"module": "files", "right": "Export", "allowed": true}, owner.Access)
	e.MustDo(200, "PUT", "/roles/"+key+"/rights", map[string]any{"module": "files", "right": "Delete", "allowed": true}, owner.Access)
	e.MustDo(204, "PATCH", "/roles/"+key, map[string]any{"name": "Records officer", "default": true}, owner.Access)
	d := e.MustDo(200, "GET", "/roles/"+key, nil, owner.Access)
	if d["name"] != "Records officer" || d["rights"].(map[string]any)["files"].(map[string]any)["Export"] != true {
		t.Fatalf("role not saved: %v", d)
	}
	sam, _ := join(t, e, owner, slug, "Sam", "") // no role named: the default role is used
	me := e.MustDo(200, "GET", "/me", nil, sam.Access)
	if !has(me, "files", "Export") || has(me, "files", "Upload") {
		t.Fatalf("default role rights wrong: %v", me["rights"])
	}

	// your own right: added to a module, assigned, and gone again when removed
	e.MustDo(204, "POST", "/access/rights", map[string]any{"module": "files", "name": "Print", "kind": "export"}, owner.Access)
	e.MustDo(200, "PUT", "/roles/"+key+"/rights", map[string]any{"module": "files", "right": "Print", "allowed": true}, owner.Access)
	if code, _ := e.Do("DELETE", "/access/rights?module=files&name=Export", nil, owner.Access); code != 409 {
		t.Fatalf("a right the server uses was removed: %d", code)
	}
	e.MustDo(204, "DELETE", "/access/rights?module=files&name=Print", nil, owner.Access)
	if code, _ := e.Do("PUT", "/roles/"+key+"/rights", map[string]any{"module": "files", "right": "Print", "allowed": true}, owner.Access); code != 400 {
		t.Fatalf("a removed right was granted: %d", code)
	}

	// the owner role stays, and a role in use cannot be deleted
	if code, _ := e.Do("DELETE", "/roles/owner", nil, owner.Access); code != 409 {
		t.Fatalf("owner role deleted: %d", code)
	}
	if code, _ := e.Do("DELETE", "/roles/"+key, nil, owner.Access); code != 409 {
		t.Fatalf("a role people hold was deleted: %d", code)
	}
	spare := e.MustDo(201, "POST", "/roles", map[string]any{"name": "Spare"}, owner.Access)["id"].(string)
	e.MustDo(204, "DELETE", "/roles/"+spare, nil, owner.Access)
	// a person without the right cannot manage roles
	if code, _ := e.Do("POST", "/roles", map[string]any{"name": "Sneaky"}, sam.Access); code != 403 {
		t.Fatalf("a clerk made a role: %d", code)
	}
}

// Approval levels are rights: whoever holds "Approve level N" in the flow's module can decide level N.
func TestApprovalLevelsAreRights(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	amy, amyID := join(t, e, owner, slug, "Amy", "staff")
	bob, bobID := join(t, e, owner, slug, "Bob", "staff")
	cat, catID := join(t, e, owner, slug, "Cat", "staff")
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, owner.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, owner.Access)
	grant := func(role, mod, right string) {
		e.MustDo(200, "PUT", "/roles/"+role+"/rights", map[string]any{"module": mod, "right": right, "allowed": true}, owner.Access)
	}
	one := e.MustDo(201, "POST", "/roles", map[string]any{"name": "First reviewer"}, owner.Access)["id"].(string)
	two := e.MustDo(201, "POST", "/roles", map[string]any{"name": "Second reviewer"}, owner.Access)["id"].(string)
	grant(one, "people", "Approve level 1")
	grant(two, "people", "Approve level 2")
	e.MustDo(200, "PATCH", "/people/"+bobID, map[string]any{"role": one}, owner.Access)
	e.MustDo(200, "PATCH", "/people/"+catID, map[string]any{"role": two}, owner.Access)

	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 2}, owner.Access)
	req := e.MustDo(202, "PATCH", "/people/"+amyID, map[string]any{"title": "Lead"}, owner.Access)["request"].(map[string]any)
	id := req["id"].(string)
	// level 1: only the first reviewer; the second reviewer and plain staff must wait their turn
	for name, u := range map[string]*testkit.User{"second reviewer": cat, "staff": amy} {
		if code, _ := e.Do("POST", "/approvals/"+id+"/approve", map[string]any{}, u.Access); code != 403 {
			t.Fatalf("%s decided level 1: %d", name, code)
		}
	}
	if step := e.MustDo(200, "POST", "/approvals/"+id+"/approve", map[string]any{}, bob.Access); step["status"] != "Pending" || step["level"] != float64(2) {
		t.Fatalf("level 1 should move on to level 2: %v", step)
	}
	// level 2: now the first reviewer is not the right person
	if code, _ := e.Do("POST", "/approvals/"+id+"/approve", map[string]any{}, bob.Access); code != 403 {
		t.Fatalf("level 1 reviewer decided level 2: %d", code)
	}
	if done := e.MustDo(200, "POST", "/approvals/"+id+"/approve", map[string]any{}, cat.Access); done["status"] != "Approved" {
		t.Fatalf("not approved: %v", done)
	}
	if got := e.MustDo(200, "GET", "/people/"+amyID, nil, owner.Access)["title"]; got != "Lead" {
		t.Fatalf("approved change not applied: %v", got)
	}
}
