package tickets_test

import (
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
	if code, _ := e.Do("POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 4}, raiser.Access); code != 403 {
		t.Fatalf("requester closed own ticket: %d", code)
	}
	e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "hold"}, it.Access)
	got := e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "resume"}, it.Access)
	if got["hold"] == true {
		t.Fatal("still on hold")
	}
	e.MustDo(200, "POST", "/tickets/"+id+"/act", map[string]any{"do": "step", "step": 3}, it.Access)
	// the raiser was told it was resolved
	notes := e.MustDo(200, "GET", "/notifications", nil, raiser.Access)
	if notes["unread"].(float64) < 1 {
		t.Fatalf("no notification: %v", notes)
	}
	full := e.MustDo(200, "GET", "/tickets/"+id, nil, ownerU.Access)
	if ev := full["events"].([]any); len(ev) < 5 {
		t.Fatalf("history too short: %v", ev)
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
