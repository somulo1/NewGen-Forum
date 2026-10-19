package people_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/testkit"
)

var tokenRe = regexp.MustCompile(`/join\?t=(\S+)`)

func person(first, email, title string) map[string]any {
	return map[string]any{"first_name": first, "last_name": "Tester", "email": email, "phone": "+254700000001", "date_of_birth": "1990-04-05", "gender": "Female", "title": title}
}

// join onboards a person as owner and makes them sign up and accept the emailed invitation.
func join(t *testing.T, e *testkit.Env, owner *testkit.User, slug, first, title string) (*testkit.User, string) {
	t.Helper()
	email := testkit.Uniq()
	out := e.MustDo(201, "POST", "/people", map[string]any{"people": []any{person(first, email, title)}}, owner.Access)
	res := out["results"].([]any)[0].(map[string]any)
	if res["status"] != "invited" {
		t.Fatalf("not invited: %v", res)
	}
	msgs := e.Dev.To(email)
	if len(msgs) != 1 {
		t.Fatalf("invitation email missing")
	}
	tok := tokenRe.FindStringSubmatch(msgs[0].Body)[1]
	e.Signup(first, email)
	u := e.Login(email, "")
	e.MustDo(200, "POST", "/invites/accept", map[string]any{"token": tok}, u.Access)
	return e.Login(email, slug), res["id"].(string)
}

func TestOnboardJoinStatusApprovalsAndReset(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	a, _ := join(t, e, owner, slug, "Alice", "Analyst")
	_, bID := join(t, e, owner, slug, "Beth", "HR Manager")

	// a staff member may not onboard, edit or act on anyone
	if code, _ := e.Do("POST", "/people", map[string]any{"people": []any{person("X", testkit.Uniq(), "")}}, a.Access); code != 403 {
		t.Fatalf("staff onboarded someone: %d", code)
	}
	if code, _ := e.Do("POST", "/people/"+bID+"/status", map[string]any{"action": "block"}, a.Access); code != 403 {
		t.Fatalf("staff blocked someone: %d", code)
	}
	// an invitation works once and only for the account that owns the invited email
	email := testkit.Uniq()
	e.MustDo(201, "POST", "/people", map[string]any{"people": []any{person("Carl", email, "")}}, owner.Access)
	tok := tokenRe.FindStringSubmatch(e.Dev.To(email)[0].Body)[1]
	other := testkit.Uniq()
	e.Signup("Mallory", other)
	if code, _ := e.Do("POST", "/invites/accept", map[string]any{"token": tok}, e.Login(other, "").Access); code != 400 {
		t.Fatalf("a stranger used someone else's invitation: %d", code)
	}
	e.Signup("Carl", email)
	carl := e.Login(email, "")
	e.MustDo(200, "POST", "/invites/accept", map[string]any{"token": tok}, carl.Access)
	if code, _ := e.Do("POST", "/invites/accept", map[string]any{"token": tok}, carl.Access); code != 400 {
		t.Fatalf("invitation used twice: %d", code)
	}
}

func TestEditsFollowTheApprovalFlow(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	a, aID := join(t, e, owner, slug, "Alice", "Analyst")
	b, _ := join(t, e, owner, slug, "Beth", "HR Manager")

	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 1, "approvers": []string{"HR Manager", "x", "y"}}, owner.Access)
	res := e.MustDo(202, "PATCH", "/people/"+aID, map[string]any{"title": "Senior Analyst"}, owner.Access)
	req := res["request"].(map[string]any)
	id := req["id"].(string)
	// nothing changed yet
	if got := e.MustDo(200, "GET", "/people/"+aID, nil, owner.Access)["title"]; got != "Analyst" {
		t.Fatalf("title changed before approval: %v", got)
	}
	// the requester cannot approve their own request; a plain staff member cannot either
	if code, _ := e.Do("POST", "/approvals/"+id+"/approve", map[string]any{}, owner.Access); code != 403 {
		t.Fatalf("requester approved own request: %d", code)
	}
	if code, _ := e.Do("POST", "/approvals/"+id+"/approve", map[string]any{}, a.Access); code != 403 {
		t.Fatalf("wrong person approved: %d", code)
	}
	done := e.MustDo(200, "POST", "/approvals/"+id+"/approve", map[string]any{"comment": "fine"}, b.Access)
	if done["status"] != "Approved" {
		t.Fatalf("not approved: %v", done)
	}
	if got := e.MustDo(200, "GET", "/people/"+aID, nil, owner.Access)["title"]; got != "Senior Analyst" {
		t.Fatalf("approved change not applied: %v", got)
	}
	if code, _ := e.Do("POST", "/approvals/"+id+"/approve", map[string]any{}, b.Access); code != 409 {
		t.Fatalf("double decision allowed: %d", code)
	}
	// a rejected request changes nothing
	r2 := e.MustDo(202, "PATCH", "/people/"+aID, map[string]any{"title": "Director"}, owner.Access)["request"].(map[string]any)["id"].(string)
	e.MustDo(200, "POST", "/approvals/"+r2+"/reject", map[string]any{"comment": "no"}, b.Access)
	if got := e.MustDo(200, "GET", "/people/"+aID, nil, owner.Access)["title"]; got != "Senior Analyst" {
		t.Fatalf("rejected change applied: %v", got)
	}
	// with zero levels the change is direct
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0, "approvers": []string{"a", "b", "c"}}, owner.Access)
	e.MustDo(200, "PATCH", "/people/"+aID, map[string]any{"title": "Lead"}, owner.Access)
	if got := e.MustDo(200, "GET", "/people/"+aID, nil, owner.Access)["title"]; got != "Lead" {
		t.Fatalf("direct change not applied: %v", got)
	}
	// non-admins cannot change flows
	if code, _ := e.Do("PUT", "/approvals/flows/f11", map[string]any{"levels": 0, "approvers": []string{"a"}}, a.Access); code != 403 {
		t.Fatalf("staff changed a flow: %d", code)
	}
}

func TestBlockSignsOutAndResetForcesChange(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	a, aID := join(t, e, owner, slug, "Alice", "Analyst")
	e.MustDo(200, "GET", "/people", nil, a.Access)

	e.MustDo(204, "POST", "/people/"+aID+"/status", map[string]any{"action": "block"}, owner.Access)
	if code, _ := e.Do("GET", "/people", nil, a.Access); code != 401 {
		t.Fatalf("blocked user still has access: %d", code)
	}
	if code, body := e.Do("POST", "/auth/login", map[string]any{"email": a.Email, "password": testkit.Password, "org": slug}, ""); code != 403 {
		t.Fatalf("blocked user signed in: %d %v", code, body)
	}
	if code, _ := e.Do("POST", "/people/"+aID+"/status", map[string]any{"action": "block"}, owner.Access); code != 409 {
		t.Fatalf("blocking twice: %d", code)
	}
	e.MustDo(204, "POST", "/people/"+aID+"/status", map[string]any{"action": "unblock"}, owner.Access)
	a = e.Login(a.Email, slug)

	// the owner cannot lock themselves out
	var oid string
	me := e.MustDo(200, "GET", "/people?q=Olivia", nil, owner.Access)["people"].([]any)
	oid = me[0].(map[string]any)["id"].(string)
	if code, _ := e.Do("POST", "/people/"+oid+"/status", map[string]any{"action": "block"}, owner.Access); code != 409 {
		t.Fatalf("self block: %d", code)
	}

	temp := e.MustDo(200, "POST", "/people/"+aID+"/reset-password", nil, owner.Access)["temporary_password"].(string)
	if code, _ := e.Do("GET", "/people", nil, a.Access); code != 401 {
		t.Fatalf("old session survived a reset: %d", code)
	}
	out := e.MustDo(200, "POST", "/auth/login", map[string]any{"email": a.Email, "password": temp, "org": slug}, "")
	if out["must_change_password"] != true {
		t.Fatalf("temporary password did not force a change: %v", out)
	}
	tmpTok := out["access_token"].(string)
	if code, body := e.Do("GET", "/people", nil, tmpTok); code != 403 || fmt.Sprint(body["error"]) == "" {
		t.Fatalf("a temporary password reached data: %d", code)
	}
	e.MustDo(200, "POST", "/auth/password", map[string]any{"current_password": temp, "new_password": "Brand5NewPass"}, tmpTok)
	out = e.MustDo(200, "POST", "/auth/login", map[string]any{"email": a.Email, "password": "Brand5NewPass", "org": slug}, "")
	if out["must_change_password"] == true {
		t.Fatal("still forced to change after changing")
	}
}

func TestDirectoryPagingPrivacyAndEncryption(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	var list []any
	for i := 0; i < 5; i++ {
		list = append(list, person(fmt.Sprintf("P%02d", i), testkit.Uniq(), "Clerk"))
	}
	e.MustDo(201, "POST", "/people", map[string]any{"people": list}, owner.Access)
	// bad rows are reported without stopping the good ones
	res := e.MustDo(201, "POST", "/people", map[string]any{"people": []any{person("Good", testkit.Uniq(), ""), map[string]any{"first_name": "Bad"}}}, owner.Access)["results"].([]any)
	if res[0].(map[string]any)["status"] != "invited" || res[1].(map[string]any)["status"] != "rejected" {
		t.Fatalf("partial import: %v", res)
	}
	seen, cursor := map[string]bool{}, ""
	for pages := 0; pages < 10; pages++ {
		out := e.MustDo(200, "GET", "/people?limit=3&cursor="+cursor, nil, owner.Access)
		for _, p := range out["people"].([]any) {
			id := p.(map[string]any)["id"].(string)
			if seen[id] {
				t.Fatalf("person listed twice across pages")
			}
			seen[id] = true
		}
		if out["total"].(float64) != 7 {
			t.Fatalf("total %v", out["total"])
		}
		next, _ := out["next"].(string)
		if next == "" {
			break
		}
		cursor = next
	}
	if len(seen) != 7 {
		t.Fatalf("paging saw %d of 7", len(seen))
	}
	// a plain staff member sees contact details but not other people's date of birth or gender
	st, _ := join(t, e, owner, slug, "Sam", "Clerk")
	for _, p := range e.MustDo(200, "GET", "/people?limit=50", nil, st.Access)["people"].([]any) {
		m := p.(map[string]any)
		if m["name"] != "Sam Tester" && (m["date_of_birth"] != nil || m["gender"] != nil) {
			t.Fatalf("staff saw someone's date of birth or gender: %v", m)
		}
	}
	// the database never holds the phone number in clear
	var raw []byte
	err := e.A.DB.Tx(context.Background(), func(tx pgx.Tx) error {
		var org uuid.UUID
		if err := tx.QueryRow(context.Background(), `select id from hub.orgs where slug=$1`, slug).Scan(&org); err != nil {
			return err
		}
		return e.A.DB.WithOrg(context.Background(), org, func(tx2 pgx.Tx) error {
			return tx2.QueryRow(context.Background(), `select phone_enc from hub.people where phone_enc is not null limit 1`).Scan(&raw)
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) == 0 || bytes.Contains(raw, []byte("254700")) {
		t.Fatalf("phone is stored readable: %q", raw)
	}
}

// A person's "reports to" is read from the organization's reporting lines, and the Roles list shows the same lines.
func TestReportsToFollowsReportingLines(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	_, aID := join(t, e, owner, slug, "Alice", "Analyst")
	ownerID := e.MustDo(200, "GET", "/me", nil, owner.Access)["person"].(string)

	reports := func() map[string]any {
		return e.MustDo(200, "GET", "/people/"+aID, nil, owner.Access)["reports_to"].(map[string]any)
	}
	// nobody holds the roles above staff yet, so there is no one to report to
	if r := reports(); r["kind"] != "none" {
		t.Fatalf("expected no one, got %v", r)
	}
	// the organization says staff report to the Super admin role
	e.MustDo(204, "PUT", "/roles/reporting", map[string]any{"role": "staff", "reports_to_role": "superadmin"}, owner.Access)
	if r := reports(); r["kind"] != "role" || r["role_key"] != "superadmin" {
		t.Fatalf("expected the Super admin role, got %v", r)
	}
	// naming a person in that role makes it that person
	e.MustDo(204, "PUT", "/roles/reporting", map[string]any{"role": "staff", "reports_to_role": "superadmin", "reports_to_person": ownerID}, owner.Access)
	if r := reports(); r["kind"] != "person" || r["person"].(map[string]any)["id"] != ownerID {
		t.Fatalf("expected Olivia, got %v", r)
	}
	// and the roles list carries the line
	var staff map[string]any
	for _, x := range e.MustDo(200, "GET", "/roles", nil, owner.Access)["roles"].([]any) {
		if x.(map[string]any)["id"] == "staff" {
			staff = x.(map[string]any)
		}
	}
	if staff["reports_to_role"] != "superadmin" || staff["reports_to_person"] != ownerID {
		t.Fatalf("roles list lost the reporting line: %v", staff)
	}
}
