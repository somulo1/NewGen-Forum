package roles_test

import (
	"testing"

	"hub/internal/testkit"
)

// Several rights marked together are ONE approval; an approver can approve some, reject some and leave the rest waiting.
func TestRightsBatchOneApprovalDecidedInParts(t *testing.T) {
	e := testkit.New(t)
	owner, slug := e.NewOrg("Olivia Owner")
	staff, _ := join(t, e, owner, slug, "Sam", "staff")
	bea, _ := join(t, e, owner, slug, "Bea", "staff")
	// Bea may approve rights changes (given while they need no approval); then they need one level, and the owner is the requester
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, owner.Access)
	e.MustDo(200, "PUT", "/people/"+beaID(t, e, bea)+"/rights", map[string]any{"module": "roles", "right": "Approve level 1", "allowed": true}, owner.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 1}, owner.Access)

	chg := func(mod, right string, allowed bool) map[string]any {
		return map[string]any{"module": mod, "right": right, "allowed": allowed}
	}
	res := e.MustDo(202, "PUT", "/roles/staff/rights/batch", map[string]any{"changes": []any{
		chg("wall", "Post", false), chg("wall", "Comment", false), chg("security", "View logs", true), chg("security", "View alerts", true), chg("roles", "View roles", true),
	}}, owner.Access)
	req := res["request"].(map[string]any)
	items := req["items"]
	_ = items
	// nothing changed yet, and there is exactly one waiting request
	me := func() map[string]any { return e.MustDo(200, "GET", "/me", nil, staff.Access) }
	if !has(me(), "wall", "Post") || has(me(), "security", "View logs") {
		t.Fatal("rights changed before approval")
	}
	pending := func() []any {
		return e.MustDo(200, "GET", "/approvals?status=Pending", nil, bea.Access)["requests"].([]any)
	}
	if n := len(pending()); n != 1 {
		t.Fatalf("one batch should be one approval, got %d", n)
	}
	first := pending()[0].(map[string]any)
	its := first["items"].([]any)
	if len(its) != 5 {
		t.Fatalf("items: %v", its)
	}
	id := func(i int) string { return its[i].(map[string]any)["id"].(string) }

	// approve two of the five (remove Post, allow View logs): the other three wait in a request of their own
	e.MustDo(200, "POST", "/approvals/"+first["id"].(string)+"/approve", map[string]any{"comment": "verified", "items": []string{id(0), id(2)}}, bea.Access)
	if has(me(), "wall", "Post") || !has(me(), "security", "View logs") {
		t.Fatalf("the approved two were not applied: %v", me()["rights"])
	}
	if has(me(), "security", "View alerts") || !has(me(), "wall", "Comment") || has(me(), "roles", "View roles") {
		t.Fatal("a right that was not chosen was applied")
	}
	left := pending()
	if len(left) != 1 || len(left[0].(map[string]any)["items"].([]any)) != 3 {
		t.Fatalf("the three others should wait as one request: %v", left)
	}
	rest := left[0].(map[string]any)
	rits := rest["items"].([]any)
	rid := func(i int) string { return rits[i].(map[string]any)["id"].(string) }

	// reject one of them: it is dropped; the last two still wait
	e.MustDo(200, "POST", "/approvals/"+rest["id"].(string)+"/reject", map[string]any{"items": []string{rid(0)}}, bea.Access)
	left = pending()
	if len(left) != 1 || len(left[0].(map[string]any)["items"].([]any)) != 2 {
		t.Fatalf("after rejecting one, two should wait: %v", left)
	}
	if has(me(), "wall", "Comment") != true {
		t.Fatal("a rejected right was applied")
	}
	// approving with nothing chosen approves everything that is left
	e.MustDo(200, "POST", "/approvals/"+left[0].(map[string]any)["id"].(string)+"/approve", map[string]any{}, bea.Access)
	if len(pending()) != 0 {
		t.Fatal("nothing should be waiting")
	}
	// an item that is not part of the request is refused
	res = e.MustDo(202, "PUT", "/roles/staff/rights/batch", map[string]any{"changes": []any{chg("wall", "Comment", false), chg("wall", "Post", true)}}, owner.Access)
	rq := res["request"].(map[string]any)["id"].(string)
	if code, _ := e.Do("POST", "/approvals/"+rq+"/approve", map[string]any{"items": []string{"not-an-item"}}, bea.Access); code != 400 {
		t.Fatalf("a foreign item was accepted: %d", code)
	}
	// staff cannot mark rights
	if code, _ := e.Do("PUT", "/roles/staff/rights/batch", map[string]any{"changes": []any{chg("wall", "Post", false)}}, staff.Access); code != 403 {
		t.Fatalf("staff marked rights: %d", code)
	}
	_ = testkit.Uniq
}

func beaID(t *testing.T, e *testkit.Env, u *testkit.User) string {
	t.Helper()
	return e.MustDo(200, "GET", "/me", nil, u.Access)["person"].(string)
}
