package calls_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"hub/internal/modules/calls"
	"hub/internal/testkit"
)

// Everyone who was in a call with someone else rates it once; missed calls ask for nothing; only the host reads the remarks.
func TestCallFeedbackFromEveryoneWhoWasIn(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	b := e.Join(ownerU, slug, "Ben", "staff")
	org := uuid.MustParse(e.OrgID(slug))
	m := calls.New(e.A)
	ago := time.Now().Add(-time.Hour)
	if err := m.RecordPast(t.Context(), org, uuid.MustParse(o.Person), "video", "meeting", nil, []uuid.UUID{uuid.MustParse(a.Person)}, "ended", ago, 600); err != nil {
		t.Fatal(err)
	}
	if err := m.RecordPast(t.Context(), org, uuid.MustParse(o.Person), "voice", "dm", nil, []uuid.UUID{uuid.MustParse(b.Person)}, "missed", ago.Add(time.Minute), 0); err != nil {
		t.Fatal(err)
	}
	due := func(m *testkit.Member) []any {
		return e.MustDo(200, "GET", "/calls/feedback", nil, m.Access)["due"].([]any)
	}
	if d := due(b); len(d) != 0 {
		t.Fatalf("a missed call asked for feedback: %v", d)
	}
	dO, dA := due(o), due(a)
	if len(dO) != 1 || len(dA) != 1 {
		t.Fatalf("due: %v / %v", dO, dA)
	}
	call := dA[0].(map[string]any)["call"].(string)
	if code, _ := e.Do("POST", "/calls/"+call+"/feedback", map[string]any{"stars": 0}, a.Access); code != 400 {
		t.Fatalf("no stars: %d", code)
	}
	if code, _ := e.Do("POST", "/calls/"+call+"/feedback", map[string]any{"stars": 4}, b.Access); code != 400 {
		t.Fatalf("someone not in the call rated it: %d", code)
	}
	e.MustDo(204, "POST", "/calls/"+call+"/feedback", map[string]any{"stars": 4, "remark": "Good agenda"}, a.Access)
	if code, _ := e.Do("POST", "/calls/"+call+"/feedback", map[string]any{"stars": 1}, a.Access); code != 409 {
		t.Fatalf("rated twice: %d", code)
	}
	e.MustDo(204, "POST", "/calls/"+call+"/feedback", map[string]any{"stars": 5}, o.Access)
	if len(due(a)) != 0 || len(due(o)) != 0 {
		t.Fatal("still asked after rating")
	}
	if code, _ := e.Do("GET", "/calls/"+call+"/feedback", nil, a.Access); code != 404 {
		t.Fatalf("a participant read everyone's feedback: %d", code)
	}
	fb := e.MustDo(200, "GET", "/calls/"+call+"/feedback", nil, o.Access)["feedback"].([]any)
	if len(fb) != 2 || fb[0].(map[string]any)["remark"] != "Good agenda" || fb[0].(map[string]any)["stars"].(float64) != 4 {
		t.Fatalf("feedback: %v", fb)
	}
	all := e.MustDo(200, "GET", "/calls/feedback/all?context=meeting", nil, ownerU.Access)
	if all["total"].(float64) != 1 || all["ratings"].(float64) != 2 || len(all["calls"].([]any)[0].(map[string]any)["ratings"].([]any)) != 2 {
		t.Fatalf("org feedback: %v", all)
	}
	if code, _ := e.Do("GET", "/calls/feedback/all", nil, a.Access); code != 403 {
		t.Fatalf("staff read the organization's feedback: %d", code)
	}
}

// Only calls where people really talked are rated, and closing the card without rating stops the question for good.
func TestFeedbackOnlyForTalksAndAskedOnce(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	o := e.Owner(ownerU)
	a := e.Join(ownerU, slug, "Amy", "staff")
	org := uuid.MustParse(e.OrgID(slug))
	m := calls.New(e.A)
	ago := time.Now().Add(-time.Hour)
	if err := m.RecordPast(t.Context(), org, uuid.MustParse(o.Person), "voice", "dm", nil, []uuid.UUID{uuid.MustParse(a.Person)}, "ended", ago, 30); err != nil {
		t.Fatal(err)
	}
	if err := m.RecordPast(t.Context(), org, uuid.MustParse(o.Person), "video", "meeting", nil, []uuid.UUID{uuid.MustParse(a.Person)}, "ended", ago.Add(time.Minute), 600); err != nil {
		t.Fatal(err)
	}
	due := func(mb *testkit.Member) []any {
		return e.MustDo(200, "GET", "/calls/feedback", nil, mb.Access)["due"].([]any)
	}
	if len(due(o)) != 2 {
		t.Fatalf("answered call and meeting: %v", due(o))
	}
	// the meeting: Amy came in only after Olivia had left, so they never talked
	var meeting string
	for _, d := range due(o) {
		if d.(map[string]any)["context"] == "meeting" {
			meeting = d.(map[string]any)["call"].(string)
		}
	}
	e.RootExec(`update hub.call_participants set joined_at = now() - interval '5 minutes', left_at = now() - interval '4 minutes' where call_id = $1 and person_id = $2`, meeting, a.Person)
	e.RootExec(`update hub.call_participants set joined_at = now() - interval '10 minutes', left_at = now() - interval '6 minutes' where call_id = $1 and person_id = $2`, meeting, o.Person)
	if d := due(o); len(d) != 1 || d[0].(map[string]any)["context"] != "dm" {
		t.Fatalf("a meeting where nobody overlapped is rated: %v", d)
	}
	call := due(o)[0].(map[string]any)["call"].(string)
	e.MustDo(204, "POST", "/calls/"+call+"/feedback/skip", nil, o.Access)
	e.MustDo(204, "POST", "/calls/"+call+"/feedback/skip", nil, o.Access)
	if len(due(o)) != 0 {
		t.Fatal("asked again after cancelling")
	}
	if len(due(a)) != 1 {
		t.Fatal("cancelling for one person hid it from the other")
	}
}
