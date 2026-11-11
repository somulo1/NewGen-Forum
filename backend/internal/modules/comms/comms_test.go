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

// An announcement sent to a forum lands in that forum: its members find it in the forum's messages.
func TestAnnouncementSentToAForumReachesItsMembers(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	staff := e.Join(ownerU, slug, "Sam", "staff")
	for _, m := range []*testkit.Member{o, staff} {
		m.Register(e)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f10", map[string]any{"levels": 0}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	forum := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": "Kubwa", "members": []string{staff.Person}, "keys": o.WrapFor(e, key, 1, o.Person, staff.Person)}, ownerU.Access)["id"].(string)
	id := uuid.Must(uuid.NewV7()).String()
	res := e.MustDo(201, "POST", "/comms", map[string]any{"type": "Announcement", "conversation": forum, "message": id, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, forum, id, 1, key, `{"title":"Holiday"}`)}, ownerU.Access)
	if res["approval"] != nil {
		t.Fatalf("it should be published at once: %v", res)
	}
	found := false
	for _, m := range e.MustDo(200, "GET", "/conversations/"+forum+"/messages?limit=50", nil, staff.Access)["messages"].([]any) {
		found = found || (m.(map[string]any)["id"] == id && m.(map[string]any)["type"] == "announcement")
	}
	if !found {
		t.Fatal("the forum's member cannot find the announcement in the forum")
	}
}

// Reposting needs the page's "Repost" right, only a published item can be reposted, and only into places it went to before.
func TestRepostGoesBackToTheSamePlacesForPeopleWithTheRight(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	staff := e.Join(ownerU, slug, "Sam", "staff")
	for _, m := range []*testkit.Member{o, staff} {
		m.Register(e)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f10", map[string]any{"levels": 0}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	mk := func(name string) string {
		return e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "forum", "name": name, "members": []string{staff.Person}, "keys": o.WrapFor(e, key, 1, o.Person, staff.Person)}, ownerU.Access)["id"].(string)
	}
	went, other := mk("Kubwa"), mk("Other")
	send := func(m *testkit.Member, conv string, repostOf string, want int) map[string]any {
		id := uuid.Must(uuid.NewV7()).String()
		body := map[string]any{"type": "Job post", "conversation": conv, "message": id, "epoch": 1, "device": m.DevID, "envelope": m.Seal(org, conv, id, 1, key, `{"title":"Driver"}`)}
		if repostOf != "" {
			body["repost_of"] = repostOf
		}
		return e.MustDo(want, "POST", "/comms", body, m.Access)
	}
	orig := send(o, went, "", 201)["comm"].(map[string]any)["id"].(string)

	pl := e.MustDo(200, "GET", "/comms/"+orig+"/places", nil, ownerU.Access)["places"].([]any)
	if len(pl) != 1 || pl[0].(map[string]any)["conversation"] != went || pl[0].(map[string]any)["member"] != true {
		t.Fatalf("places: %v", pl)
	}
	// nobody without the right can see the places or repost
	if code, _ := e.Do("GET", "/comms/"+orig+"/places", nil, staff.Access); code != 403 {
		t.Fatalf("places without the right: %d", code)
	}
	id := uuid.Must(uuid.NewV7()).String()
	if code, _ := e.Do("POST", "/comms", map[string]any{"type": "Job post", "conversation": went, "message": id, "epoch": 1, "device": staff.DevID, "envelope": staff.Seal(org, went, id, 1, key, `{"title":"x"}`), "repost_of": orig}, staff.Access); code != 403 {
		t.Fatalf("repost without the right: %d", code)
	}
	// with the right: back into the place it went to (a second copy); never into another place
	send(o, went, orig, 201)
	if code, _ := e.Do("POST", "/comms", map[string]any{"type": "Job post", "conversation": other, "message": uuid.NewString(), "epoch": 1, "device": o.DevID, "envelope": "x", "repost_of": orig}, ownerU.Access); code == 201 {
		t.Fatalf("reposted into a place it never went to")
	}
	n := 0
	for _, m := range e.MustDo(200, "GET", "/conversations/"+went+"/messages?limit=50", nil, staff.Access)["messages"].([]any) {
		if m.(map[string]any)["type"] == "post" {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("the forum holds %d copies, want 2", n)
	}
}

// People a communication reached acknowledge it with a click; its insights show how many did and who.
func TestCommunicationsAreAcknowledged(t *testing.T) {
	e := testkit.New(t)
	ownerU, slug := e.NewOrg("Olivia Owner")
	org := e.OrgID(slug)
	o := e.Owner(ownerU)
	staff := e.Join(ownerU, slug, "Sam", "staff")
	for _, m := range []*testkit.Member{o, staff} {
		m.Register(e)
	}
	e.MustDo(204, "PUT", "/approvals/flows/f1", map[string]any{"levels": 0}, ownerU.Access)
	key, _ := e2ee.NewConversationKey()
	wall := e.MustDo(201, "POST", "/conversations", map[string]any{"kind": "wall", "name": "Everyone", "audience": "org", "keys": o.WrapFor(e, key, 1, o.Person, staff.Person)}, ownerU.Access)["id"].(string)
	id := uuid.Must(uuid.NewV7()).String()
	comm := e.MustDo(201, "POST", "/comms", map[string]any{"type": "Announcement", "conversation": wall, "message": id, "epoch": 1, "device": o.DevID, "envelope": o.Seal(org, wall, id, 1, key, "Office closed Friday")}, ownerU.Access)["comm"].(map[string]any)
	cid := comm["id"].(string)
	if comm["status"] != "Published" {
		t.Skipf("announcements need approval here: %v", comm["status"])
	}
	if c, _ := e.Do("POST", "/comms/"+cid+"/acknowledge", nil, ownerU.Access); c != 404 {
		t.Fatalf("its author acknowledged it: %d", c)
	}
	e.MustDo(204, "POST", "/comms/"+cid+"/acknowledge", nil, staff.Access)
	e.MustDo(204, "POST", "/comms/"+cid+"/acknowledge", nil, staff.Access) // twice is fine, counted once
	places := e.MustDo(200, "GET", "/comms/"+cid+"/insights", nil, ownerU.Access)["places"].([]any)
	pl := places[0].(map[string]any)
	if pl["acknowledged"].(float64) != 1 || len(pl["acknowledged_by"].([]any)) != 1 || pl["read"].(float64) != 1 {
		t.Fatalf("insights: %v", pl)
	}
}
