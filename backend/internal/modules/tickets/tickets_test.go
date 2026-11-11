package tickets_test

import (
	"strings"
	"testing"
	"time"

	"hub/internal/e2ee"
	"hub/internal/modules/tickets"
	"hub/internal/testkit"
)

func TestTATRespectsWorkingHours(t *testing.T) {
	c := tickets.Config{Start: "08:00", End: "17:00", Timezone: "Africa/Nairobi"}
	c.Pause.Weekends, c.Pause.Offhours = true, true
	loc, _ := time.LoadLocation("Africa/Nairobi")
	// Friday 2026-10-16 16:00 local + 2 working hours: 1h today, then Monday 08:00-09:00
	from := time.Date(2026, 10, 16, 16, 0, 0, 0, loc)
	got := c.Due(from, 2).In(loc)
	if want := time.Date(2026, 10, 19, 9, 0, 0, 0, loc); !got.Equal(want) {
		t.Fatalf("due %v want %v", got, want)
	}
	// holidays are skipped
	c.Pause.Holidays, c.Holidays = true, []string{"2026-10-19"}
	got = c.Due(from, 2).In(loc)
	if want := time.Date(2026, 10, 20, 9, 0, 0, 0, loc); !got.Equal(want) {
		t.Fatalf("due with holiday %v want %v", got, want)
	}
	// with nothing paused, the clock simply runs
	free := tickets.Config{Start: "08:00", End: "17:00", Timezone: "UTC"}
	if got := free.Due(from, 5); !got.Equal(from.Add(5 * time.Hour)) {
		t.Fatalf("free clock %v", got)
	}
}

func TestTicketRoutingHandOverAndEscalation(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	it := e.Join(ownerU, slug, "Ian", "it")
	raiser := e.Join(ownerU, slug, "Faith", "staff")
	for _, m := range []*testkit.Member{o, it, raiser} {
		m.Register(e)
	}
	// the IT handler takes every "Network & Wi-Fi" ticket
	cats := e.MustDo(200, "GET", "/tickets/categories", nil, ownerU.Access)["categories"].([]any)
	if len(cats) != 5 {
		t.Fatalf("default categories: %d", len(cats))
	}
	var net map[string]any
	for _, c := range cats {
		if c.(map[string]any)["name"] == "Network & Wi-Fi" {
			net = c.(map[string]any)
		}
	}
	e.MustDo(200, "PUT", "/tickets/categories/"+net["id"].(string), map[string]any{"name": "Network & Wi-Fi", "tat_hours": 4, "priority": "High", "members": []string{it.Person}}, ownerU.Access)
	// staff cannot reconfigure categories or settings
	if code, _ := e.Do("PUT", "/tickets/categories/"+net["id"].(string), map[string]any{"name": "x", "tat_hours": 1}, raiser.Access); code != 403 {
		t.Fatalf("staff edited categories: %d", code)
	}
	if code, _ := e.Do("PUT", "/tickets/settings", map[string]any{"start": "08:00", "end": "17:00", "timezone": "UTC", "nearPct": 80}, raiser.Access); code != 403 {
		t.Fatalf("staff edited settings: %d", code)
	}

	// Faith raises a ticket: her device makes the conversation with a key for her and the IT handler
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{it.Person}, "keys": raiser.WrapFor(e, key, 1, raiser.Person, it.Person)}, raiser.Access)["id"].(string)
	raiser.Send(e, org, conv, 1, key, "Laptop will not connect to the office Wi-Fi", 201)
	tk := e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Network & Wi-Fi", "to_type": "desk"}, raiser.Access)
	if tk["handler"] != it.Person || tk["step"].(float64) != 1 || tk["priority"] != "High" || tk["due_at"] == nil || tk["ref"] == "" {
		t.Fatalf("routing: %v", tk)
	}
	// the handler can read the thread; a bystander cannot even list the conversation's messages
	if k := it.ReadKey(e, conv, 1); string(k) != string(key) {
		t.Fatal("handler has no key")
	}
	other := e.Join(ownerU, slug, "Nosy", "staff")
	other.Register(e)
	if code, _ := e.Do("GET", "/conversations/"+conv+"/messages", nil, other.Access); code != 404 {
		t.Fatalf("bystander read the ticket thread: %d", code)
	}
	// the handler works it; the raiser may not close it
	id := conv
	st := 2
	e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": st}, it.Access)
	if code, _ := e.Do("POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 4, "comment": "closing", "rating": 4}, raiser.Access); code != 403 {
		t.Fatalf("requester closed own ticket: %d", code)
	}
	e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "hold"}, it.Access)
	got := e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "resume"}, it.Access)
	if got["hold"] == true {
		t.Fatal("still on hold")
	}
	// resolving needs a comment on what was done
	if code, _ := e.Do("POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 3, "comment": "  "}, it.Access); code != 400 {
		t.Fatalf("resolved without a comment: %d", code)
	}
	e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 3, "comment": "Restarted the router"}, it.Access)
	// resolved but not closed: it may still be reopened
	e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 2}, it.Access)
	e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 3, "comment": "Replaced the router"}, it.Access)
	// the raiser was told it was resolved
	notes := e.MustDo(200, "GET", "/notifications", nil, raiser.Access)
	if notes["unread"].(float64) < 1 {
		t.Fatalf("no notification: %v", notes)
	}
	full := e.MustDo(200, "GET", "/tickets/"+id, nil, ownerU.Access)
	if ev := full["events"].([]any); len(ev) < 5 {
		t.Fatalf("history too short: %v", ev)
	}
	// a resolved ticket nobody closes is closed by the sweep 30 minutes later: 5 stars and a comment from Hub
	ts0 := tickets.New(e.A)
	if err := ts0.SweepAll(t.Context()); err != nil {
		t.Fatal(err)
	}
	if got := e.MustDo(200, "GET", "/tickets/"+id, nil, ownerU.Access); got["step"].(float64) != 3 {
		t.Fatalf("closed before 30 minutes: %v", got["step"])
	}
	e.AgeResolved(id, 31*time.Minute)
	if err := ts0.SweepAll(t.Context()); err != nil {
		t.Fatal(err)
	}
	shut := e.MustDo(200, "GET", "/tickets/"+id, nil, raiser.Access)
	if shut["step"].(float64) != 4 || shut["rating"].(float64) != 5 || shut["auto_closed"] != true || shut["closed_at"] == nil {
		t.Fatalf("not closed automatically: %v", shut)
	}
	var byHub, resolved bool
	for _, c := range shut["comments"].([]any) {
		m := c.(map[string]any)
		byHub = byHub || (m["person"] == nil && m["who"] == "Hub" && strings.Contains(m["text"].(string), "Closed automatically"))
		resolved = resolved || m["text"] == "Resolved: Replaced the router"
	}
	if !byHub || !resolved {
		t.Fatalf("comments: %v", shut["comments"])
	}
	// closed is final: no reopening, no comments; the resolver sees the rating, the raiser has none to see
	if code, _ := e.Do("POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 2}, it.Access); code != 409 {
		t.Fatalf("reopened a closed ticket: %d", code)
	}
	if code, _ := e.Do("POST", "/tickets/"+id+"/comments", map[string]any{"text": "hello"}, raiser.Access); code != 409 {
		t.Fatalf("commented on a closed ticket: %d", code)
	}
	fb := e.MustDo(200, "GET", "/tickets/feedback", nil, it.Access)
	if fb["total"].(float64) != 1 || fb["average"].(float64) != 5 || fb["items"].([]any)[0].(map[string]any)["mine"] != true {
		t.Fatalf("resolver feedback: %v", fb)
	}
	if fb := e.MustDo(200, "GET", "/tickets/feedback", nil, raiser.Access); fb["total"].(float64) != 0 {
		t.Fatalf("raiser sees feedback: %v", fb)
	}
	// a second sweep leaves it alone
	if err := ts0.SweepAll(t.Context()); err != nil {
		t.Fatal(err)
	}
	if again := e.MustDo(200, "GET", "/tickets/"+id, nil, raiser.Access); len(again["comments"].([]any)) != len(shut["comments"].([]any)) {
		t.Fatal("closed twice")
	}
	// a ticket past its TAT is escalated by the sweep
	key2, _ := e2ee.NewConversationKey()
	c2 := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{it.Person}, "keys": raiser.WrapFor(e, key2, 1, raiser.Person, it.Person)}, raiser.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": c2, "category": "Network & Wi-Fi"}, raiser.Access)
	if _, err := e.A.DB.Pool.Exec(t.Context(), `select 1`); err != nil {
		t.Fatal(err)
	}
	e.MakeOverdue(c2)
	ts := tickets.New(e.A)
	if err := ts.SweepAll(t.Context()); err != nil {
		t.Fatal(err)
	}
	late := e.MustDo(200, "GET", "/tickets/"+c2, nil, ownerU.Access)
	if late["late"] != true || late["escalated"] != true {
		t.Fatalf("overdue ticket not escalated: %v", late)
	}
}

// Nobody works on a ticket they raised (they may close it), and nobody escalates a ticket into their own place at the
// level it already sits at.
func TestOwnTicketsAndEscalationScope(t *testing.T) {
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
	e.Levels(ownerU, map[string]string{"sector": "", "branch": "sector", "unit": "branch"})
	sector := add("sector", "Central", "")
	branch := add("branch", "Nairobi", sector)
	other := add("branch", "Kisumu", sector)
	unit := add("unit", "Support", branch)
	staff := e.Join(ownerU, slug, "Sam", "staff")
	for _, m := range []*testkit.Member{o, staff} {
		m.Register(e)
	}
	e.MustDo(200, "PATCH", "/people/"+o.Person, map[string]any{"sector_id": sector, "branch_id": branch, "unit_id": unit}, ownerU.Access)
	raise := func(by *testkit.Member, toType, toRef string) string {
		key, _ := e2ee.NewConversationKey()
		peer := o
		if by == o {
			peer = staff
		}
		conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{peer.Person}, "keys": by.WrapFor(e, key, 1, by.Person, peer.Person)}, by.Access)["id"].(string)
		e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": toType, "to_ref": toRef}, by.Access)
		return conv
	}
	act := func(conv string, body map[string]any) int {
		code, _ := e.Do("POST", "/tickets/"+conv+"/act", body, ownerU.Access)
		return code
	}

	// the owner raised it: they cannot work on it or take it, but may hand it to someone else and close it
	mine := raise(o, "unit", unit)
	if c := act(mine, map[string]any{"do": "assign", "handler": o.Person}); c != 403 {
		t.Fatalf("raiser assigned to self: %d", c)
	}
	if c := act(mine, map[string]any{"do": "step", "step": 2}); c != 403 {
		t.Fatalf("raiser started work: %d", c)
	}
	if c := act(mine, map[string]any{"do": "assign", "handler": staff.Person}); c != 403 {
		t.Fatalf("the raiser (not a head of the unit) handed it over: %d", c)
	}
	if c := act(mine, map[string]any{"do": "step", "step": 4, "comment": "closing", "rating": 6}); c != 400 {
		t.Fatalf("six stars: %d", c)
	}
	if c := act(mine, map[string]any{"do": "step", "step": 4, "rating": 4, "comment": "Quick, thanks"}); c != 200 {
		t.Fatalf("raiser closing: %d", c)
	}
	if got := e.MustDo(200, "GET", "/tickets/"+mine, nil, o.Access); got["rating"].(float64) != 4 || got["auto_closed"] != false || len(got["comments"].([]any)) != 1 {
		t.Fatalf("rating or remark not kept: %v", got)
	}

	// a unit ticket: not into the owner's own unit again, but up into their own branch is fine
	tk := raise(staff, "unit", unit)
	if c := act(tk, map[string]any{"do": "escalate", "why": "needs attention", "to_type": "unit", "to_ref": unit}); c != 400 {
		t.Fatalf("escalated to where it is: %d", c)
	}
	if c := act(tk, map[string]any{"do": "escalate", "why": "needs attention", "to_type": "branch", "to_ref": branch}); c != 200 {
		t.Fatalf("unit into own branch: %d", c)
	}
	// now a branch ticket: not into their own branch, but another branch or their sector is fine
	if c := act(tk, map[string]any{"do": "escalate", "why": "needs attention", "to_type": "branch", "to_ref": branch}); c != 400 {
		t.Fatalf("branch into own branch: %d", c)
	}
	if c := act(tk, map[string]any{"do": "escalate", "why": "needs attention", "to_type": "branch", "to_ref": other}); c != 200 {
		t.Fatalf("into another branch: %d", c)
	}
	if c := act(tk, map[string]any{"do": "escalate", "why": "needs attention", "to_type": "sector", "to_ref": sector}); c != 200 {
		t.Fatalf("up into own sector: %d", c)
	}
	sec := raise(staff, "sector", sector)
	if c := act(sec, map[string]any{"do": "escalate", "why": "needs attention", "to_type": "sector", "to_ref": sector}); c != 400 {
		t.Fatalf("sector into own sector: %d", c)
	}
}

// A ticket's subject, description and comments are kept by the server (encrypted at rest, not end to end): its participants
// read them, others with the View right see only the facts, and the text never goes out on the organization-wide channel.
func TestTicketTextIsServerReadableForParticipants(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	it := e.Join(ownerU, slug, "Ian", "it")
	raiser := e.Join(ownerU, slug, "Faith", "staff")
	for _, m := range []*testkit.Member{o, it, raiser} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{it.Person}, "keys": raiser.WrapFor(e, key, 1, raiser.Person, it.Person)}, raiser.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "title": "Printer jammed", "description": "Third floor printer"}, raiser.Access)
	e.MustDo(201, "POST", "/tickets/"+conv+"/comments", map[string]any{"text": "On my way"}, it.Access)
	if code, _ := e.Do("POST", "/tickets/"+conv+"/comments", map[string]any{"text": "nosy"}, ownerU.Access); code != 403 {
		t.Fatalf("a non-participant commented: %d", code)
	}
	if code, _ := e.Do("POST", "/tickets/"+conv+"/comments", map[string]any{"text": "  "}, raiser.Access); code != 400 {
		t.Fatalf("empty comment: %d", code)
	}

	got := e.MustDo(200, "GET", "/tickets/"+conv, nil, raiser.Access)
	if got["title"] != "Printer jammed" || got["description"] != "Third floor printer" {
		t.Fatalf("text: %v", got)
	}
	cs := got["comments"].([]any)
	if len(cs) != 1 || cs[0].(map[string]any)["text"] != "On my way" || cs[0].(map[string]any)["person"] != it.Person {
		t.Fatalf("comments: %v", cs)
	}
	// the owner may view tickets but is not on this one: facts only
	other := e.MustDo(200, "GET", "/tickets/"+conv, nil, ownerU.Access)
	if other["title"] != nil || other["comments"] != nil || other["participant"] != false {
		t.Fatalf("non-participant saw the text: %v", other)
	}
	for _, row := range e.MustDo(200, "GET", "/tickets", nil, ownerU.Access)["tickets"].([]any) {
		if r := row.(map[string]any); r["id"] == conv && r["title"] != nil {
			t.Fatalf("list leaked the subject: %v", r)
		}
	}
	listed := false
	for _, row := range e.MustDo(200, "GET", "/tickets", nil, it.Access)["tickets"].([]any) {
		if r := row.(map[string]any); r["id"] == conv && r["title"] == "Printer jammed" {
			listed = true
		}
	}
	if !listed {
		t.Fatal("a participant's list is missing the subject")
	}
	// stored encrypted, not in clear
	var raw []byte
	if err := e.A.DB.Pool.QueryRow(t.Context(), `select body_enc from hub.ticket_comments limit 1`).Scan(&raw); err == nil && strings.Contains(string(raw), "On my way") {
		t.Fatal("comment stored in clear")
	}
}

// The head of the receiving department assigns, reassigns and escalates; the person working on it may escalate but not hand it on;
// the person who raised it may do neither.
func TestTicketRolesHeadWorkerAndRaiser(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	e.MustDo(204, "PUT", "/approvals/flows/f11", map[string]any{"levels": 0}, ownerU.Access)
	e.MustDo(204, "PUT", "/approvals/flows/f8", map[string]any{"levels": 0}, ownerU.Access)
	e.Levels(ownerU, map[string]string{"department": ""})
	it := e.MustDo(201, "POST", "/org/units", map[string]any{"kind": "department", "name": "IT"}, ownerU.Access)["id"].(string)
	hana := e.Join(ownerU, slug, "Hana", "staff")
	wes := e.Join(ownerU, slug, "Wes", "staff")
	rae := e.Join(ownerU, slug, "Rae", "staff")
	for _, m := range []*testkit.Member{hana, wes, rae} {
		m.Register(e)
	}
	for _, m := range []*testkit.Member{hana, wes} {
		e.MustDo(200, "PATCH", "/people/"+m.Person, map[string]any{"department_id": it}, ownerU.Access)
	}
	e.MustDo(200, "PUT", "/people/"+hana.Person+"/rights", map[string]any{"module": "tickets", "right": "Assign & hand over", "allowed": true}, ownerU.Access)
	e.MustDo(200, "PUT", "/people/"+rae.Person+"/rights", map[string]any{"module": "tickets", "right": "Assign & hand over", "allowed": true}, ownerU.Access) // the right alone is not enough outside IT

	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{hana.Person, wes.Person}, "keys": rae.WrapFor(e, key, 1, rae.Person, hana.Person, wes.Person)}, rae.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv, "category": "Hardware", "to_type": "department", "to_ref": it}, rae.Access)
	act := func(who *testkit.Member, body map[string]any) int {
		code, _ := e.Do("POST", "/tickets/"+conv+"/act", body, who.Access)
		return code
	}
	if c := act(rae, map[string]any{"do": "assign", "handler": wes.Person}); c != 403 {
		t.Fatalf("the raiser (not head of IT) assigned: %d", c)
	}
	if c := act(hana, map[string]any{"do": "assign", "handler": wes.Person}); c != 200 {
		t.Fatalf("the head of IT assigned: %d", c)
	}
	if c := act(wes, map[string]any{"do": "assign", "handler": hana.Person}); c != 403 {
		t.Fatalf("the worker reassigned: %d", c)
	}
	can := func(who *testkit.Member) (bool, bool) {
		d := e.MustDo(200, "GET", "/tickets/"+conv, nil, who.Access)
		return d["can_assign"] == true, d["can_escalate"] == true
	}
	if a, x := can(wes); a || !x {
		t.Fatalf("worker should escalate but not assign: assign=%v escalate=%v", a, x)
	}
	if a, x := can(hana); !a || !x {
		t.Fatalf("head should assign and escalate: assign=%v escalate=%v", a, x)
	}
	if a, x := can(rae); a || x {
		t.Fatalf("raiser should do neither: assign=%v escalate=%v", a, x)
	}
	if c := act(rae, map[string]any{"do": "escalate", "why": "needs attention"}); c != 403 {
		t.Fatalf("the raiser escalated: %d", c)
	}
	part := func(period string) map[string]any {
		q := ""
		if period != "" {
			q = "?period=" + period
		}
		return e.MustDo(200, "GET", "/org/units/"+it+"/dashboard"+q, nil, ownerU.Access)["parts"].(map[string]any)["tickets"].(map[string]any)
	}
	// performance: received and still open within its time -> waiting, not judged either way yet
	if p := part(""); p["received"].(float64) != 0 || p["waiting"].(float64) != 1 {
		t.Fatalf("before escalating: %v", p)
	}
	if c := act(wes, map[string]any{"do": "escalate", "why": "needs attention"}); c != 200 {
		t.Fatalf("the worker escalated: %d", c)
	}
	// raising the alarm without handing it on keeps it here: still this department's, on its own clock, not judged yet
	// (handing it to another place within the time is what counts as escalated in time: see org's TestEachPlaceIsJudgedOnItsOwnStay)
	for _, period := range []string{"", "month", "quarter", "year"} {
		if p := part(period); p["received"].(float64) != 0 || p["waiting"].(float64) != 1 {
			t.Fatalf("after raising the alarm (%q): %v", period, p)
		}
	}
	// a head of department may assign a ticket they raised themselves to their own department
	hal := e.Join(ownerU, slug, "Hal", "hod")
	hal.Register(e)
	e.MustDo(200, "PATCH", "/people/"+hal.Person, map[string]any{"department_id": it}, ownerU.Access)
	key2, _ := e2ee.NewConversationKey()
	conv2 := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "ticket", "members": []string{wes.Person}, "keys": hal.WrapFor(e, key2, 1, hal.Person, wes.Person)}, hal.Access)["id"].(string)
	e.MustDo(201, "POST", "/tickets", map[string]any{"conversation": conv2, "category": "Hardware", "to_type": "department", "to_ref": it}, hal.Access)
	if code, _ := e.Do("POST", "/tickets/"+conv2+"/act", map[string]any{"do": "assign", "handler": wes.Person}, hal.Access); code != 200 {
		t.Fatalf("a head assigning their own ticket in their department: %d", code)
	}
}
