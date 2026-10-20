package comms_test

import (
	"testing"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/testkit"
)

func TestAnnouncementWaitsForApprovalThenReachesTheAudience(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	hod := e.Join(ownerU, slug, "Hana", "hod")
	staff := e.Join(ownerU, slug, "Sam", "staff")
	for _, m := range []*testkit.Member{o, hod, staff} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	wall := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Everyone", "audience": "org", "keys": o.WrapFor(e, key, 1, o.Person, hod.Person, staff.Person)}, ownerU.Access)["id"].(string)

	send := func(m *testkit.Member, typ string, want int) map[string]any {
		id := uuid.Must(uuid.NewV7()).String()
		return e.MustDo(want, "POST", "/comms", map[string]any{"type": typ, "conversation": wall, "message": id, "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, wall, id, 1, key, "Office closed Friday")}, m.Access)
	}
	send(staff, "Announcement", 403) // staff cannot announce
	res := send(hod, "Announcement", 201)
	comm := res["comm"].(map[string]any)
	if comm["status"] != "Pending approval" || res["approval"] == nil {
		t.Fatalf("expected to wait: %v", res)
	}
	// nothing reached the wall yet
	if msgs := e.MustDo(200, "GET", "/conversations/"+wall+"/messages", nil, staff.Access)["messages"].([]any); len(msgs) != 0 {
		t.Fatalf("posted before approval: %v", msgs)
	}
	ap := res["approval"].(map[string]any)["id"].(string)
	e.MustDo(200, "POST", "/approvals/"+ap+"/approve", map[string]any{}, ownerU.Access)
	e.MustDo(200, "POST", "/approvals/"+ap+"/approve", map[string]any{}, ownerU.Access)
	msgs := e.MustDo(200, "GET", "/conversations/"+wall+"/messages", nil, staff.Access)["messages"].([]any)
	if len(msgs) != 1 || msgs[0].(map[string]any)["type"] != "announcement" || msgs[0].(map[string]any)["id"] != comm["message"] {
		t.Fatalf("not posted: %v", msgs)
	}
	// reads are counted once per person
	cid := comm["id"].(string)
	e.MustDo(204, "POST", "/comms/"+cid+"/read", nil, staff.Access)
	e.MustDo(204, "POST", "/comms/"+cid+"/read", nil, staff.Access)
	e.MustDo(204, "POST", "/comms/"+cid+"/read", nil, o.Access)
	l := e.MustDo(200, "GET", "/comms", nil, hod.Access)["comms"].([]any)
	if len(l) != 1 || l[0].(map[string]any)["status"] != "Published" || int(l[0].(map[string]any)["read"].(float64)) != 2 {
		t.Fatalf("list: %v", l)
	}
	// staff see it in their list but not the read count
	sl := e.MustDo(200, "GET", "/comms", nil, staff.Access)["comms"].([]any)
	if len(sl) != 1 || int(sl[0].(map[string]any)["read"].(float64)) != 0 {
		t.Fatalf("staff list: %v", sl)
	}
}

func TestAwardsNeedApprovalAndNotifyWinners(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	hod := e.Join(ownerU, slug, "Hana", "hod")
	win := e.Join(ownerU, slug, "Wanjiru", "staff")
	res := e.MustDo(201, "POST", "/awards", map[string]any{"type": "Employee of the month", "people": []string{win.Person}, "citation": "99.9% uptime"}, hod.Access)
	if res["award"].(map[string]any)["status"] != "Pending approval" {
		t.Fatalf("award: %v", res)
	}
	if l := e.MustDo(200, "GET", "/awards", nil, win.Access)["awards"].([]any); len(l) != 0 {
		t.Fatalf("an unapproved award is visible: %v", l)
	}
	ap := res["approval"].(map[string]any)["id"].(string)
	e.MustDo(200, "POST", "/approvals/"+ap+"/approve", map[string]any{}, ownerU.Access)
	e.MustDo(200, "POST", "/approvals/"+ap+"/approve", map[string]any{}, ownerU.Access)
	l := e.MustDo(200, "GET", "/awards", nil, win.Access)["awards"].([]any)
	if len(l) != 1 || l[0].(map[string]any)["citation"] != "99.9% uptime" {
		t.Fatalf("award not visible: %v", l)
	}
	n := e.MustDo(200, "GET", "/notifications", nil, win.Access)
	if n["unread"].(float64) != 1 {
		t.Fatalf("winner not told: %v", n)
	}
	// staff cannot nominate people who are not members, or nominate at all without the right
	if code, _ := e.Do("POST", "/awards", map[string]any{"type": "X", "people": []string{uuid.NewString()}, "citation": "y"}, hod.Access); code != 400 {
		t.Fatalf("stranger nominated: %d", code)
	}
}

// Where an award should be announced is remembered, and each place is claimed once by the nominator after approval.
func TestAwardAnnouncementTargetsAreClaimedOnce(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	hod := e.Join(ownerU, slug, "Hana", "hod")
	win := e.Join(ownerU, slug, "Wanjiru", "staff")
	forum := uuid.NewString()
	res := e.MustDo(201, "POST", "/awards", map[string]any{"type": "Team player", "people": []string{win.Person}, "citation": "helped", "targets": []string{"wall", forum}}, hod.Access)
	id := res["award"].(map[string]any)["id"].(string)
	// not approved: nothing to claim yet, and the targets are hidden
	if code, _ := e.Do("POST", "/awards/"+id+"/claim", map[string]any{"target": "wall"}, hod.Access); code != 409 {
		t.Fatalf("claimed before approval: %d", code)
	}
	ap := res["approval"].(map[string]any)["id"].(string)
	e.MustDo(200, "POST", "/approvals/"+ap+"/approve", map[string]any{}, ownerU.Access)
	e.MustDo(200, "POST", "/approvals/"+ap+"/approve", map[string]any{}, ownerU.Access)
	find := func(tok string) map[string]any {
		for _, a := range e.MustDo(200, "GET", "/awards", nil, tok)["awards"].([]any) {
			if a.(map[string]any)["id"] == id {
				return a.(map[string]any)
			}
		}
		t.Fatal("award not listed")
		return nil
	}
	if tg, _ := find(hod.Access)["targets"].([]any); len(tg) != 2 {
		t.Fatalf("the nominator should see both places: %v", find(hod.Access))
	}
	if find(win.Access)["targets"] != nil {
		t.Fatal("someone else saw where the award is to be announced")
	}
	if code, _ := e.Do("POST", "/awards/"+id+"/claim", map[string]any{"target": "wall"}, win.Access); code != 403 {
		t.Fatalf("a stranger claimed the post: %d", code)
	}
	e.MustDo(204, "POST", "/awards/"+id+"/claim", map[string]any{"target": "wall"}, hod.Access)
	if code, _ := e.Do("POST", "/awards/"+id+"/claim", map[string]any{"target": "wall"}, hod.Access); code != 409 {
		t.Fatalf("posted twice: %d", code)
	}
	if tg, _ := find(hod.Access)["targets"].([]any); len(tg) != 1 || tg[0] != forum {
		t.Fatalf("only the forum should be left: %v", tg)
	}
}

func TestJobPostInsightsCountReadsLikesAndRepliesForAllowedPeopleOnly(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	adm := e.Join(ownerU, slug, "Ada", "admin")
	staff := e.Join(ownerU, slug, "Sam", "staff")
	for _, m := range []*testkit.Member{o, adm, staff} {
		m.Register(e)
	}
	key, _ := e2ee.NewConversationKey()
	wall := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Everyone", "audience": "org", "keys": o.WrapFor(e, key, 1, o.Person, adm.Person, staff.Person)}, ownerU.Access)["id"].(string)
	id := uuid.Must(uuid.NewV7()).String()
	batch := uuid.Must(uuid.NewV7()).String()
	res := e.MustDo(201, "POST", "/comms", map[string]any{"type": "Job post", "batch": batch, "conversation": wall, "message": id, "epoch": 1, "device": adm.DevID, "envelope": adm.Seal(org, wall, id, 1, key, `{"title":"Driver"}`)}, adm.Access)
	cid := res["comm"].(map[string]any)["id"].(string)
	if ap, ok := res["approval"].(map[string]any); ok {
		for i := 0; i < 3; i++ {
			e.Do("POST", "/approvals/"+ap["id"].(string)+"/approve", map[string]any{}, ownerU.Access)
		}
	}
	e.MustDo(204, "POST", "/comms/"+cid+"/read", nil, staff.Access)
	react := func(typ string) {
		rid := uuid.Must(uuid.NewV7()).String()
		e.MustDo(201, "POST", "/conversations/"+wall+"/messages", map[string]any{"id": rid, "epoch": 1, "device": staff.DevID, "envelope": staff.Seal(org, wall, rid, 1, key, `{"text":"x"}`), "type": typ, "reply_to": id}, staff.Access)
	}
	react("reaction")
	react("text")
	react("text")
	e.MustDo(403, "GET", "/comms/"+cid+"/insights", nil, staff.Access)
	places := e.MustDo(200, "GET", "/comms/"+cid+"/insights", nil, ownerU.Access)["places"].([]any)
	p := places[0].(map[string]any)
	if len(places) != 1 || int(p["read"].(float64)) != 1 || int(p["likes"].(float64)) != 1 || int(p["replies"].(float64)) != 2 || int(p["audience"].(float64)) != 3 {
		t.Fatalf("insights: %v", places)
	}
}
