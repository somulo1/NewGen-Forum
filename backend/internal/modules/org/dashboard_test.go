package org_test

import (
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

func TestUnitDashboardIsBuiltFromRealRecords(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "hod")
	bob := e.Join(ownerU, slug, "Bob", "staff")
	for _, m := range []*testkit.Member{o, amy, bob} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+amy.Person, map[string]any{"department_id": dept, "title": "HR Manager"}, ownerU.Access)

	// before anything happens the page is empty and healthy, with the default threshold
	d := e.MustDo(200, "GET", "/org/units/"+dept+"/dashboard", nil, ownerU.Access)
	if d["perf"].(float64) != 100 || d["target"].(float64) != 95 || d["members"].(float64) != 1 {
		t.Fatalf("empty dashboard: %v", d)
	}
	// people without the right cannot open it
	if code, _ := e.Do("GET", "/org/units/"+dept+"/dashboard", nil, bob.Access); code != 403 {
		t.Fatalf("staff opened a dashboard: %d", code)
	}

	// a ticket Amy raised to her own department, with the owner on it
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": dept}, amy.Access)
	// ... and escalated by the owner to the desk with a reason
	esc := e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "desk", "why": "Needs the IT desk"}, ownerU.Access)
	if esc["to_type"] != "desk" || esc["escalated"] != true {
		t.Fatalf("escalation: %v", esc)
	}
	if code, _ := e.Do("POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": uuid.NewString()}, ownerU.Access); code != 400 {
		t.Fatalf("escalated to a department that does not exist: %d", code)
	}
	e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": dept, "person": o.Person, "why": "Back to Finance"}, ownerU.Access)

	// an approval request waiting for someone who holds the level-1 approval right (Amy, through a personal grant) shows up as received by Finance
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "people", "right": "Approve level 1", "allowed": true}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 1}, ownerU.Access)
	e.MustDo(202, "PATCH", "/people/"+bob.Person, map[string]any{"title": "Clerk"}, ownerU.Access)

	// a memo Amy sends to the whole organization
	k2, _ := e2ee.NewConversationKey()
	wall := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Everyone", "audience": "org", "keys": o.WrapFor(e, k2, 1, o.Person, amy.Person, bob.Person)}, ownerU.Access)["id"].(string)
	mid := uuid.Must(uuid.NewV7()).String()
	cm := e.MustDo(201, "POST", "/comms", map[string]any{"type": "Memo", "conversation": wall, "message": mid, "epoch": 1, "device": amy.DevID, "envelope": amy.Seal(org, wall, mid, 1, k2, "Budget dates")}, amy.Access)
	if cm["approval"] != nil {
		e.MustDo(200, "POST", "/approvals/"+cm["approval"].(map[string]any)["id"].(string)+"/approve", map[string]any{}, ownerU.Access)
	}

	d = e.MustDo(200, "GET", "/org/units/"+dept+"/dashboard", nil, ownerU.Access)
	// escalated to the desk and then back to Finance: two stays with Finance, the first passed on, the second still open
	dirs := map[string]int{}
	for _, x := range d["tickets"].([]any) {
		dirs[x.(map[string]any)["dir"].(string)]++
	}
	if len(d["tickets"].([]any)) != 2 || dirs["Escalated"] != 1 || dirs["Received"] != 1 {
		t.Fatalf("tickets: %v", d["tickets"])
	}
	// business requests come only from the request form now (see TestBusinessRequestsComeFromTheForm), never from approvals
	if reqs := d["requests"].([]any); len(reqs) != 0 {
		t.Fatalf("approvals listed as business requests: %v", reqs)
	}
	// organization-wide memos go out from the Communications page and are not a place's communications
	// (those are sent between places from the dashboards: see TestPlacesSendEachOtherCommunications)
	if comms := d["comms"].([]any); len(comms) != 0 {
		t.Fatalf("a memo listed among the place's communications: %v", comms)
	}
	_ = mid
	parts := d["parts"].(map[string]any)
	if parts["tickets"].(map[string]any)["received"].(float64) != 1 {
		t.Fatalf("parts: %v", parts)
	}

	// the threshold: only people who may set targets, and only 50-100
	e.MustDo(204, "PUT", "/org/units/"+dept+"/target", map[string]any{"target": 80}, ownerU.Access)
	if code, _ := e.Do("PUT", "/org/units/"+dept+"/target", map[string]any{"target": 20}, ownerU.Access); code != 400 {
		t.Fatalf("silly threshold: %d", code)
	}
	if code, _ := e.Do("PUT", "/org/units/"+dept+"/target", map[string]any{"target": 90}, bob.Access); code != 403 {
		t.Fatalf("staff set a threshold: %d", code)
	}
	if e.MustDo(200, "GET", "/org/units/"+dept+"/dashboard", nil, ownerU.Access)["target"].(float64) != 80 {
		t.Fatal("threshold not saved")
	}
	found := false
	for _, u := range e.MustDo(200, "GET", "/org/structure", nil, ownerU.Access)["units"].([]any) {
		if m := u.(map[string]any); m["id"] == dept && m["target"].(float64) == 80 {
			found = true
		}
	}
	if !found {
		t.Fatal("structure does not carry the threshold")
	}
}

// A sector's dashboard counts the tickets sent to the departments under it, and one still open within its time is not judged yet.
func TestSectorDashboardRollsUpAndWaitsForOpenTickets(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"sector": "", "department": "sector"})
	sector := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "sector", "name": "Central"}, ownerU.Access)["id"].(string)
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance", "parent_id": sector}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": dept}, amy.Access)
	for _, unit := range []string{dept, sector} {
		d := e.MustDo(200, "GET", "/org/units/"+unit+"/dashboard", nil, ownerU.Access)
		tk := d["parts"].(map[string]any)["tickets"].(map[string]any)
		if tk["waiting"].(float64) != 1 || tk["received"].(float64) != 0 || d["measured"].(float64) != 0 {
			t.Fatalf("unit %s: tickets %v measured %v", unit, tk, d["measured"])
		}
	}
}

// A business request is raised like a ticket, to a department only, and shows under the department's business requests.
func TestBusinessRequestsComeFromTheForm(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"sector": "", "department": "sector"})
	sector := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "sector", "name": "Central"}, ownerU.Access)["id"].(string)
	dept := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance", "parent_id": sector}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	conv := func() string {
		key, _ := e2ee.NewConversationKey()
		return e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	}
	if c, _ := e.Do("POST", "/tickets", map[string]any{"conversation": conv(), "category": "Hardware", "to_type": "sector", "to_ref": sector, "kind": "request"}, amy.Access); c != 400 {
		t.Fatalf("a business request to a sector: %d", c)
	}
	if c, _ := e.Do("POST", "/tickets", map[string]any{"conversation": conv(), "to_type": "department", "to_ref": dept, "kind": "request"}, amy.Access); c != 403 {
		t.Fatalf("a business request without access to departments: %d", c)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	tk := e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv(), "category": "Hardware", "to_type": "department", "to_ref": dept, "kind": "request", "priority": "High"}, amy.Access)
	if tk["kind"] != "request" || tk["priority"] != "Low" { // Hardware is 48 hours: beyond a day is low priority, whatever was asked
		t.Fatalf("request: kind %v priority %v", tk["kind"], tk["priority"])
	}
	d := e.MustDo(200, "GET", "/org/units/"+dept+"/dashboard", nil, ownerU.Access)
	if len(d["requests"].([]any)) != 1 || len(d["tickets"].([]any)) != 0 {
		t.Fatalf("department: %d requests, %d tickets", len(d["requests"].([]any)), len(d["tickets"].([]any)))
	}
	if s := e.MustDo(200, "GET", "/org/units/"+sector+"/dashboard", nil, ownerU.Access); len(s["requests"].([]any)) != 0 {
		t.Fatal("a sector lists business requests it received")
	}
	// escalated to another department: a fresh clock on that department's own business request turnaround
	ops := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Ops", "parent_id": sector}, ownerU.Access)["id"].(string)
	e.MustDo(204, "PUT", "/tickets/request-tat", map[string]any{"department": ops, "hours": 6}, ownerU.Access)
	moved := e.MustDo(200, "POST", "/tickets/"+tk["id"].(string)+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": ops, "why": "Ops handles it"}, ownerU.Access)
	if moved["tat_hours"].(float64) != 6 || moved["priority"] != "High" {
		t.Fatalf("escalated request: tat %v priority %v", moved["tat_hours"], moved["priority"])
	}
}

// A department sets how long business requests take, up to 3 months; a ticket it escalates stays on its dashboard as escalated.
func TestRequestTurnaroundAndEscalatedTickets(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	fin := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	if c, _ := e.Do("PUT", "/tickets/request-tat", map[string]any{"department": fin, "hours": 2161}, ownerU.Access); c != 400 {
		t.Fatalf("over 3 months: %d", c)
	}
	if c, _ := e.Do("PUT", "/tickets/request-tat", map[string]any{"department": fin, "hours": 6}, amy.Access); c != 403 {
		t.Fatalf("staff set the turnaround: %d", c)
	}
	e.MustDo(204, "PUT", "/tickets/request-tat", map[string]any{"department": fin, "hours": 6}, ownerU.Access)
	conv := func() string {
		key, _ := e2ee.NewConversationKey()
		return e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	rq := e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv(), "to_type": "department", "to_ref": fin, "kind": "request"}, amy.Access)
	if rq["tat_hours"].(float64) != 6 || rq["priority"] != "High" {
		t.Fatalf("request: tat %v priority %v", rq["tat_hours"], rq["priority"])
	}
	// a ticket to Finance, escalated by the owner to the desk: Finance still sees it, as escalated
	tk := conv()
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": tk, "category": "Hardware", "to_type": "department", "to_ref": fin}, amy.Access)
	e.MustDo(200, "POST", "/tickets/"+tk+"/act", map[string]any{"do": "escalate", "to_type": "desk", "why": "Needs the IT desk"}, ownerU.Access)
	d := e.MustDo(200, "GET", "/org/units/"+fin+"/dashboard", nil, ownerU.Access)
	var dirs []string
	for _, x := range d["tickets"].([]any) {
		dirs = append(dirs, x.(map[string]any)["dir"].(string))
	}
	if len(dirs) != 1 || dirs[0] != "Escalated" {
		t.Fatalf("finance tickets: %v", dirs)
	}
}

// Each place is judged on its own stay: its clock starts when it receives the ticket, and escalating within its own time counts
// for it. The place it moves to starts a new clock on the category it falls under there.
func TestEachPlaceIsJudgedOnItsOwnStay(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	fin := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	it := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "IT"}, ownerU.Access)["id"].(string)
	hr := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "HR"}, ownerU.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets/categories", map[string]any{"unit_type": "department", "unit_ref": it, "name": "Laptop repair", "tat_hours": 4}, ownerU.Access)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": fin}, amy.Access)
	act := func(body map[string]any) (int, map[string]any) {
		return e.Do("POST", "/tickets/"+conv+"/act", body, ownerU.Access)
	}
	if c, _ := act(map[string]any{"do": "escalate", "to_type": "department", "to_ref": it, "why": "IT"}); c != 400 {
		t.Fatalf("moved without a category: %d", c)
	}
	if c, _ := act(map[string]any{"do": "escalate", "category": "Laptop repair", "why": "longer"}); c != 400 {
		t.Fatalf("changed the category (and clock) without moving: %d", c)
	}
	c, tk := act(map[string]any{"do": "escalate", "to_type": "department", "to_ref": it, "category": "Laptop repair", "why": "IT fixes laptops"})
	if c != 200 || tk["tat_hours"].(float64) != 4 || tk["priority"] != "High" {
		t.Fatalf("escalate to IT: %d %v %v", c, tk["tat_hours"], tk["priority"])
	}
	tickets := func(unit string) []map[string]any {
		var out []map[string]any
		for _, x := range e.MustDo(200, "GET", "/org/units/"+unit+"/dashboard", nil, ownerU.Access)["tickets"].([]any) {
			out = append(out, x.(map[string]any))
		}
		return out
	}
	if f := tickets(fin); len(f) != 1 || f[0]["dir"] != "Escalated" || f[0]["ok"] != true {
		t.Fatalf("finance escalated in its own time: %v", f)
	}
	if i := tickets(it); len(i) != 1 || i[0]["dir"] != "Received" || i[0]["pending"] != true {
		t.Fatalf("IT has it, its clock running: %v", i)
	}
	// IT runs out of time, then passes it on: late for IT, a fresh clock for HR
	e.RootExec(`update hub.tickets set due_at = now() - interval '1 hour' where conversation_id = $1`, conv)
	if i := tickets(it); i[0]["status"] != "Overdue" {
		t.Fatalf("IT overdue: %v", i)
	}
	e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": hr, "category": "HR request", "why": "policy"}, ownerU.Access)
	if i := tickets(it); len(i) != 1 || i[0]["dir"] != "Escalated" || i[0]["ok"] != false {
		t.Fatalf("IT escalated late: %v", i)
	}
	if h := tickets(hr); len(h) != 1 || h[0]["pending"] != true {
		t.Fatalf("HR starts fresh: %v", h)
	}
	if f := tickets(fin); f[0]["ok"] != true {
		t.Fatalf("finance's result changed by what happened later: %v", f)
	}
	// HR resolves it in time
	e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "step", "step": 3, "comment": "done"}, ownerU.Access)
	if h := tickets(hr); h[0]["dir"] != "Received" || h[0]["ok"] != true || h[0]["status"] != "Solved on time" {
		t.Fatalf("HR solved: %v", h)
	}
}

// The dashboard shows who works on each ticket, and each person's own results; members see theirs whatever their rights.
func TestPeopleSeeTheirOwnPerformance(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	fin := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	it := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "IT"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	wes := e.Join(ownerU, slug, "Wes", "staff")
	for _, m := range []*testkit.Member{o, amy, wes} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+wes.Person, map[string]any{"department_id": fin}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person, wes.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person, wes.Person)}, amy.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": fin}, amy.Access)
	e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "assign", "handler": wes.Person}, ownerU.Access)
	d := e.MustDo(200, "GET", "/org/units/"+fin+"/dashboard?period=month", nil, ownerU.Access)
	if tk := d["tickets"].([]any)[0].(map[string]any); tk["assignee"] != "Wes Member" && tk["assignee_id"] != wes.Person {
		t.Fatalf("assignee: %v", tk)
	}
	// Wes forwards it to IT within Finance's time: on time for Wes
	e.MustDo(200, "POST", "/tickets/"+conv+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": it, "category": "Hardware", "why": "IT"}, ownerU.Access)
	mine := e.MustDo(200, "GET", "/org/my-performance", nil, wes.Access)["places"].([]any)
	if len(mine) != 1 {
		t.Fatalf("places: %v", mine)
	}
	me := mine[0].(map[string]any)["me"].(map[string]any)
	tk := me["tickets"].(map[string]any)
	if tk["received"].(float64) != 1 || tk["ok"].(float64) != 1 || me["id"] != wes.Person {
		t.Fatalf("Wes: %v", me)
	}
	if c, _ := e.Do("GET", "/org/units/"+fin+"/dashboard", nil, wes.Access); c != 403 {
		t.Fatalf("staff opened the whole dashboard: %d", c)
	}
}

// Places send each other communications from their dashboards: kept like tickets, on the receiving place's own turnaround.
func TestPlacesSendEachOtherCommunications(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"branch": "", "department": "branch"})
	br := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "branch", "name": "Kisumu"}, ownerU.Access)["id"].(string)
	fin := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance", "parent_id": br}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+amy.Person, map[string]any{"department_id": fin}, ownerU.Access)
	e.MustDo(204, "PUT", "/tickets/request-tat", map[string]any{"department": br, "hours": 12, "kind": "communication"}, ownerU.Access)
	if c, _ := e.Do("PUT", "/tickets/request-tat", map[string]any{"department": br, "hours": 12}, ownerU.Access); c != 404 {
		t.Fatalf("business request turnaround on a branch: %d", c)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	if c, _ := e.Do("POST", "/tickets", map[string]any{"conversation": conv, "to_type": "desk", "kind": "communication"}, amy.Access); c != 400 {
		t.Fatalf("a communication to the ticket desk: %d", c)
	}
	// a communication is sent on a place's behalf, by someone who belongs to it and may open its dashboard
	if c, _ := e.Do("POST", "/tickets", map[string]any{"conversation": conv, "to_type": "branch", "to_ref": br, "kind": "communication"}, amy.Access); c != 400 {
		t.Fatalf("a communication from nobody: %d", c)
	}
	if c, _ := e.Do("POST", "/tickets", map[string]any{"conversation": conv, "to_type": "branch", "to_ref": br, "kind": "communication", "from_ref": fin}, amy.Access); c != 403 {
		t.Fatalf("sent for Finance without its dashboard: %d", c)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	if c, _ := e.Do("POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "branch", "to_ref": br, "from_ref": fin}, amy.Access); c != 400 {
		t.Fatalf("a ticket on a place's behalf: %d", c)
	}
	cm := e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "to_type": "branch", "to_ref": br, "kind": "communication", "title": "Stock count dates", "from_ref": fin}, amy.Access)
	if cm["kind"] != "communication" || cm["tat_hours"].(float64) != 12 || cm["priority"] != "Medium" {
		t.Fatalf("communication: %v %v %v", cm["kind"], cm["tat_hours"], cm["priority"])
	}
	dirs := func(unit string) []string {
		var out []string
		d := e.MustDo(200, "GET", "/org/units/"+unit+"/dashboard", nil, ownerU.Access)
		for _, x := range d["comms"].([]any) {
			out = append(out, x.(map[string]any)["dir"].(string))
		}
		if len(d["tickets"].([]any)) != 0 {
			t.Fatalf("a communication listed as a ticket: %v", d["tickets"])
		}
		return out
	}
	if b := dirs(br); len(b) != 1 || b[0] != "Received" {
		t.Fatalf("branch: %v", b)
	}
	if c := e.MustDo(200, "GET", "/org/units/"+br+"/dashboard", nil, ownerU.Access)["comms"].([]any)[0].(map[string]any); c["from"] != "Finance" || c["from_unit"] != true {
		t.Fatalf("from: %v %v", c["from"], c["from_unit"])
	}
	// the people of the branch acknowledge it; the sender cannot
	if c, _ := e.Do("POST", "/tickets/"+conv+"/acknowledge", nil, amy.Access); c != 403 {
		t.Fatalf("the sender acknowledged it: %d", c)
	}
	bea := e.Join(ownerU, slug, "Bea", "staff")
	bea.Register(e)
	e.MustDo(200, "PATCH", "/people/"+bea.Person, map[string]any{"branch_id": br}, ownerU.Access)
	ack := e.MustDo(200, "POST", "/tickets/"+conv+"/acknowledge", nil, bea.Access)
	if acks := ack["acknowledgements"].([]any); len(acks) != 1 {
		t.Fatalf("acknowledgements: %v", acks)
	}
	if f := dirs(fin); len(f) != 1 || f[0] != "Sent" {
		t.Fatalf("finance: %v", f)
	}
}

// A department's business requests are tracked like its communications: received, passed on, and sent (from its dashboard too).
func TestDepartmentBusinessRequestsAreTrackedBothWays(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	fin := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "Finance"}, ownerU.Access)["id"].(string)
	it := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "IT"}, ownerU.Access)["id"].(string)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	for _, m := range []*testkit.Member{o, amy} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+amy.Person, map[string]any{"department_id": fin}, ownerU.Access)
	e.MustDo(200, "PUT", "/people/"+amy.Person+"/rights", map[string]any{"module": "departments", "right": "View dashboard", "allowed": true}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{o.Person}, "keys": amy.WrapFor(e, key, 1, amy.Person, o.Person)}, amy.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "to_type": "department", "to_ref": it, "kind": "request", "from_ref": fin, "title": "Budget lines"}, amy.Access)
	dirs := func(unit string) []string {
		var out []string
		for _, x := range e.MustDo(200, "GET", "/org/units/"+unit+"/dashboard", nil, ownerU.Access)["requests"].([]any) {
			out = append(out, x.(map[string]any)["dir"].(string))
		}
		return out
	}
	if f := dirs(fin); len(f) != 1 || f[0] != "Sent" {
		t.Fatalf("finance sent: %v", f)
	}
	if i := dirs(it); len(i) != 1 || i[0] != "Received" {
		t.Fatalf("IT received: %v", i)
	}
}

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
	if c, _ := e.Do("POST", "/tickets/"+other+"/act", map[string]any{"do": "escalate", "to_type": "department", "to_ref": it, "category": "Hardware"}, hana.Access); c != 403 {
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
