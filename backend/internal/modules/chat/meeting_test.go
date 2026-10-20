package chat_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// The Home page's "Meetings & standups" end to end on the server: an organizer schedules a meeting for people, only those
// people see it on their calendar, an invitee answers RSVP (and only for themselves), and a repeating standup is listed.
func TestMeetingIsSeenByInviteesAndRsvpIsPerPerson(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	amy := e.Join(ownerU, slug, "Amy", "staff")
	bob := e.Join(ownerU, slug, "Bob", "staff")
	for _, m := range []*testkit.Member{o, amy, bob} {
		m.Register(e)
	}
	today := time.Now().UTC().Format("2006-01-02")
	from, to := time.Now().AddDate(0, 0, -1).UTC().Format("2006-01-02"), time.Now().AddDate(0, 0, 8).UTC().Format("2006-01-02")

	// the organizer makes the meeting's container (members = invitees) and puts the entry in it
	key, _ := e2ee.NewConversationKey()
	conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "event", "members": []string{amy.Person}, "keys": o.WrapFor(e, key, 1, o.Person, amy.Person)}, ownerU.Access)["id"].(string)
	seal := func(m *testkit.Member, id, text string) string { return m.Seal(org, conv, id, 1, key, text) }
	put := func(m *testkit.Member, collection string, parent *string, text string, meta map[string]any, want int) string {
		t.Helper()
		id := uuid.Must(uuid.NewV7()).String()
		body := map[string]any{"id": id, "collection": collection, "epoch": 1, "device": m.DevID, "envelope": seal(m, id, text), "meta": meta}
		if parent != nil {
			body["parent"] = *parent
		}
		if code, out := e.Do("POST", "/containers/"+conv+"/records", body, m.Access); code != want {
			t.Fatalf("%s: want %d got %d %v", collection, want, code, out)
		}
		return id
	}
	meeting := put(o, "entry", nil, `{"title":"Sprint review"}`, map[string]any{"date": today, "start": "10:00", "end": "11:00", "type": "Meeting", "repeat": "None"}, 201)
	standup := put(o, "entry", nil, `{"title":"Daily standup"}`, map[string]any{"date": today, "start": "09:15", "end": "09:30", "type": "Standup", "repeat": "Weekdays"}, 201)

	listed := func(m *testkit.Member) map[string]string {
		out := map[string]string{}
		for _, x := range e.MustDo(200, "GET", "/calendar?from="+from+"&to="+to, nil, m.Access)["entries"].([]any) {
			r := x.(map[string]any)
			meta := r["meta"].(map[string]any)
			out[r["id"].(string)] = meta["type"].(string) + "/" + meta["repeat"].(string)
		}
		return out
	}
	// Amy was invited: she sees both. Bob was not: he sees neither.
	if got := listed(amy); got[meeting] != "Meeting/None" || got[standup] != "Standup/Weekdays" {
		t.Fatalf("Amy's calendar: %v", got)
	}
	if got := listed(bob); len(got) != 0 {
		t.Fatalf("Bob sees a meeting he was not invited to: %v", got)
	}
	// an uninvited person cannot answer for it either
	id := uuid.Must(uuid.NewV7()).String()
	if code, _ := e.Do("POST", "/containers/"+conv+"/records", map[string]any{"id": id, "collection": "rsvp", "parent": meeting, "epoch": 1, "device": bob.DevID, "envelope": seal(bob, id, `{"going":true}`), "meta": map[string]any{"going": true}}, bob.Access); code != 404 {
		t.Fatalf("uninvited RSVP: %d", code)
	}
	// Amy answers Going, then changes her mind; the organizer reads her answer, and it is only hers
	rsvp := put(amy, "rsvp", &meeting, `{"going":true}`, map[string]any{"going": true}, 201)
	mine := func(m *testkit.Member) (bool, bool) {
		for _, x := range e.MustDo(200, "GET", "/containers/"+conv+"/records?collection=rsvp", nil, m.Access)["records"].([]any) {
			r := x.(map[string]any)
			if r["id"] == rsvp {
				return r["meta"].(map[string]any)["going"] == true, true
			}
		}
		return false, false
	}
	if going, ok := mine(amy); !ok || !going {
		t.Fatal("Amy's RSVP is missing")
	}
	putBack := map[string]any{"collection": "rsvp", "parent": meeting, "epoch": 1, "device": amy.DevID, "envelope": seal(amy, rsvp, `{"going":false}`), "meta": map[string]any{"going": false}}
	e.MustDo(200, "PUT", "/containers/"+conv+"/records/"+rsvp, putBack, amy.Access)
	if going, ok := mine(o); !ok || going {
		t.Fatalf("the organizer does not see Amy's changed answer: %v %v", going, ok)
	}
	byOrganizer := map[string]any{"collection": "rsvp", "parent": meeting, "epoch": 1, "device": o.DevID, "envelope": seal(o, rsvp, `{"going":true}`), "meta": map[string]any{"going": true}}
	if code, _ := e.Do("PUT", "/containers/"+conv+"/records/"+rsvp, byOrganizer, o.Access); code != 403 && code != 404 {
		t.Fatalf("the organizer changed Amy's answer: %d", code)
	}
	// the organizer cancels it for everyone
	e.MustDo(204, "DELETE", "/containers/"+conv+"/records/"+meeting, nil, o.Access)
	if got := listed(amy); got[meeting] != "" || got[standup] == "" {
		t.Fatalf("after cancelling: %v", got)
	}
}
