package roles_test

import (
	"encoding/base64"
	"regexp"
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
	if code, _ := e.Do("PUT", "/roles/superadmin/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, owner.Access); code != 409 {
		t.Fatalf("super admin was limited: %d", code)
	}
	// f8 needs 3 levels by default: the change waits, then nothing happens until all approve
	res := e.MustDo(202, "PUT", "/roles/staff/rights", map[string]any{"module": "wall", "right": "Post", "allowed": false}, owner.Access)
	if !has(e.MustDo(200, "GET", "/me", nil, staff.Access), "wall", "Post") {
		t.Fatal("rights changed before approval")
	}
	_ = res
	// make it direct and see the change take effect, with the audit trail written
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0, "approvers": []string{"a", "b", "c"}}, owner.Access)
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
	if len(roles) != 10 {
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
