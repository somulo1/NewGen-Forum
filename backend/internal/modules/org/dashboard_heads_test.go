package org_test

import (
	"testing"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// The head of a place resolves what its dashboard received, without the Solve right, but not what it sent.
func TestHeadsResolveWhatTheirPlaceReceived(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	fin := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	it := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "IT"}, ownerU.Access)["id"].(string)
	hana := e.Join(ownerU, slug, "Hana", "staff") // staff: no Solve right
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, hana, amy} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+hana.Person, map[string]any{"department_id": it}, ownerU.Access)
	e.MustDo(204, "PUT", "/org/units/"+it+"/head", map[string]any{"role": "staff"}, ownerU.Access)
	raise := func(to string) string {
		key, _ := e2ee.NewConversationKey()
		conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{hana.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, hana.Person)}, amy.Access)["id"].(string)
		e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": to}, amy.Access)
		return conv
	}
	got := raise(it)
	if tk := e.MustDo(200, "GET", "/tickets/"+got, nil, hana.Access); tk["can_work"] != true {
		t.Fatalf("head cannot work on what IT received: %v", tk["can_work"])
	}
	e.MustDo(200, "POST", "/tickets/"+got+"/act", map[string]any{"do": "step", "step": 3, "comment": "fixed"}, hana.Access)
	other := raise(fin)
	if c, _ := e.Do("POST", "/tickets/"+other+"/act", map[string]any{"do": "step", "step": 3, "comment": "not mine"}, hana.Access); c != 403 {
		t.Fatalf("IT's head resolved Finance's ticket: %d", c)
	}
	// a head resolves even what they raised to their own place themselves
	key, _ := e2ee.NewConversationKey()
	own := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{amy.Person}, "keys": hana.WrapFor(e, key, 1, hana.Person, amy.Person)}, hana.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": own, "category": "Hardware", "to_type": "department", "to_ref": it}, hana.Access)
	if tk := e.MustDo(200, "GET", "/tickets/"+own, nil, hana.Access); tk["can_work"] != true {
		t.Fatalf("head cannot work on what they sent to their own place: %v", tk["can_work"])
	}
	e.MustDo(200, "POST", "/tickets/"+own+"/act", map[string]any{"do": "step", "step": 3, "comment": "done myself"}, hana.Access)
	// someone who is not a head still cannot resolve what they raised
	if c, _ := e.Do("POST", "/tickets/"+other+"/act", map[string]any{"do": "step", "step": 3, "comment": "mine"}, amy.Access); c != 403 {
		t.Fatalf("the raiser resolved their own ticket: %d", c)
	}
	// and escalates what IT received: forwards it on, or reports it to the heads; not what is with Finance
	fwd := raise(it)
	if tk := e.MustDo(200, "GET", "/tickets/"+fwd, nil, hana.Access); tk["can_escalate"] != true {
		t.Fatalf("head cannot escalate what IT received: %v", tk["can_escalate"])
	}
	e.MustDo(200, "POST", "/tickets/"+fwd+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": fin, "category": "Hardware", "why": "Finance owns these"}, hana.Access)
	rep := raise(it)
	e.MustDo(200, "POST", "/tickets/"+rep+"/act", map[string]any{"do": "escalate", "why": "running late"}, hana.Access)
	if c, _ := e.Do("POST", "/tickets/"+other+"/act", map[string]any{"do": "escalate", "why": "needs attention", "to_type": "department", "to_ref": it, "category": "Hardware"}, hana.Access); c != 403 {
		t.Fatalf("IT's head escalated Finance's ticket: %d", c)
	}
}

// Who handles a ticket comes from the place's categories (set on its dashboard): a ticket raised before the category had people is
// given to them once they are on it, and a business request goes to the head of the department.
func TestCategoriesDecideWhoHandles(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	it := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "IT"}, ownerU.Access)["id"].(string)
	cat := e.MustDo(201, "POST", "/tickets/categories", map[string]any{"unit_type": "department", "unit_ref": it, "name": "Laptops", "tat_hours": 24}, ownerU.Access)
	sal := e.Join(ownerU, slug, "Salmon", "staff")
	hana := e.Join(ownerU, slug, "Hana", "staff")
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, sal, hana, amy} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+hana.Person, map[string]any{"department_id": it}, ownerU.Access)
	e.MustDo(204, "PUT", "/org/units/"+it+"/head", map[string]any{"role": "staff"}, ownerU.Access)
	newConv := func(members ...*testkit.Member) string {
		ids, people := []string{amy.Person}, []string{}
		for _, m := range members {
			ids, people = append(ids, m.Person), append(people, m.Person)
		}
		key, _ := e2ee.NewConversationKey()
		return e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": people, "keys": amy.WrapFor(e, key, 1, ids...)}, amy.Access)["id"].(string)
	}
	// raised while Laptops had nobody: unassigned (Salmon is on the ticket, as if added later, but not yet named on the category)
	tk := newConv(sal)
	if r := e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": tk, "category": "Laptops", "to_type": "department", "to_ref": it}, amy.Access); r["handler"] != nil {
		t.Fatalf("assigned before the category named anyone: %v", r["handler"])
	}
	// the head of IT names Salmon on Laptops; the next look at the ticket hands it to him
	e.MustDo(200, "PUT", "/tickets/categories/"+cat["id"].(string), map[string]any{"name": "Laptops", "tat_hours": 24, "members": []string{sal.Person}}, ownerU.Access)
	if r := e.MustDo(200, "POST", "/tickets/"+tk+"/route", nil, amy.Access); r["handler"] != sal.Person {
		t.Fatalf("not handed to the category's person: %v", r["handler"])
	}
	// a business request to IT goes to its head, who is on it
	e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	rq := e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": newConv(hana), "to_type": "department", "to_ref": it, "kind": "request", "title": "Budget"}, amy.Access)
	if rq["handler"] != hana.Person {
		t.Fatalf("business request not with IT's head: %v", rq["handler"])
	}
}

// A resolved, or even closed, ticket can be forwarded to another place: it opens there, tracked as received, on that place's TAT.
func TestClosedTicketsCanBeForwarded(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f5", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	fin := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	it := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "IT"}, ownerU.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets/categories", map[string]any{"unit_type": "department", "unit_ref": it, "name": "Laptops", "tat_hours": 6}, ownerU.Access)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": fin}, amy.Access)
	act := func(body map[string]any) map[string]any {
		return e.MustDo(200, "POST", "/tickets/"+conv+"/act", body, ownerU.Access)
	}
	act(map[string]any{"do": "step", "step": 3, "comment": "done"})
	act(map[string]any{"do": "step", "step": 4, "comment": "closing", "rating": 4})
	if c, _ := e.Do("POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "why": "late"}, ownerU.Access); c != 409 {
		t.Fatalf("reported a closed ticket: %d", c)
	}
	tk := act(map[string]any{"do": "escalate", "to_type": "department", "to_ref": it, "category": "Laptops", "why": "IT must check"})
	if tk["step"].(float64) >= 3 || tk["closed_at"] != nil || tk["tat_hours"].(float64) != 6 || tk["priority"] != "High" {
		t.Fatalf("forwarded closed ticket: step %v closed %v tat %v prio %v", tk["step"], tk["closed_at"], tk["tat_hours"], tk["priority"])
	}
	rows := e.MustDo(200, "GET", "/org/units/"+it+"/dashboard", nil, ownerU.Access)["tickets"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["dir"] != "Received" || rows[0].(map[string]any)["open"] != true {
		t.Fatalf("IT: %v", rows)
	}
	f := e.MustDo(200, "GET", "/org/units/"+fin+"/dashboard", nil, ownerU.Access)["tickets"].([]any)
	if len(f) != 1 || f[0].(map[string]any)["status"] != "Solved on time" {
		t.Fatalf("Finance keeps its own result: %v", f)
	}
}
