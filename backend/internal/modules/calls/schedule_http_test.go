package calls_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/callsvc"
	"hub/internal/e2ee"
	"hub/internal/testkit"
)

// A scheduled meeting's room is refused before it opens (nothing is created), and opens normally once it is on.
func TestScheduledMeetingRoomOnlyOpensWhenItIsOn(t *testing.T) {
	e := testkit.New(t)
	svc := httptest.NewServer(callsvc.New(e.A.Cfg.CallsSecret, e.A.Rdb, nil).Handler())
	defer svc.Close()
	e.A.Cfg.CallsURLs = []string{svc.URL}
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	o.Register(e)
	a.Register(e)
	loc, _ := time.LoadLocation(e.A.Cfg.ScheduleTZ)
	now := time.Now().In(loc)
	entry := func(start, end time.Time) string {
		key, _ := e2ee.NewConversationKey()
		conv := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "event", "members": []string{a.Person}, "keys": o.WrapFor(e, key, 1, o.Person, a.Person)}, ownerU.Access)["id"].(string)
		id := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/containers/"+conv+"/records", map[string]any{"id": id, "collection": "entry", "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, conv, id, 1, key, `{"title":"Sync"}`),
			"meta": map[string]any{"date": start.Format("2006-01-02"), "start": start.Format("15:04"), "end": end.Format("15:04"), "type": "Meeting", "repeat": "None"}}, ownerU.Access)
		return conv
	}
	// 20 minutes from now: too early. Nothing is created.
	soon := now.Add(20 * time.Minute)
	later := entry(soon, soon.Add(time.Hour))
	if code, out := e.Do("POST", "/calls", map[string]any{"kind": "video", "context": "meeting", "conversation": later}, ownerU.Access); code != 409 {
		t.Fatalf("early start: %d %v", code, out)
	}
	// started 3 minutes ago, 'til in an hour: on.
	on := entry(now.Add(-3*time.Minute), now.Add(time.Hour))
	e.MustDo(201, "POST", "/calls", map[string]any{"kind": "video", "context": "meeting", "conversation": on}, ownerU.Access)
	// ended 10 minutes ago: over.
	over := entry(now.Add(-70*time.Minute), now.Add(-10*time.Minute))
	if code, out := e.Do("POST", "/calls", map[string]any{"kind": "video", "context": "meeting", "conversation": over}, ownerU.Access); code != 409 {
		t.Fatalf("late start: %d %v", code, out)
	}
	// opens in 4 minutes: inside the 5-minute lead.
	lead := now.Add(4 * time.Minute)
	e.MustDo(201, "POST", "/calls", map[string]any{"kind": "video", "context": "meeting", "conversation": entry(lead, lead.Add(time.Hour))}, ownerU.Access)
}
