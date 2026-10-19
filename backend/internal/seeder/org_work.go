package seeder

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"hub/internal/e2ee"
	"hub/internal/modules/approvals"
	"hub/internal/modules/chat"
	"hub/internal/modules/comms"
	"hub/internal/modules/files"
	"hub/internal/modules/projects"
	"hub/internal/notify"
)

func (s *Seeder) seedComms(ctx context.Context) {
	var cs []struct {
		ID, Type, Title, Body, By, Date string
		To                              []string
		Status                          string
		Read                            int
	}
	s.st("comms", &cs)
	// department walls so memos can go to a department
	walls := map[string]uuid.UUID{"Community wall": s.wall}
	deptWall := func(name string) (uuid.UUID, bool) {
		if id, ok := walls[name]; ok {
			return id, true
		}
		d, ok := s.Units["department:"+name]
		if !ok {
			return uuid.Nil, false
		}
		members := s.byDept(name)
		c, _ := s.newConversation(ctx, s.Owner, chat.CreateIn{Kind: "wall", Name: name, Audience: "department", AudienceRef: &d, Color: "#FFD66B"}, dedupeActors(members, s.Owner))
		walls[name] = c.ID
		return c.ID, true
	}
	for i := len(cs) - 1; i >= 0; i-- {
		c := cs[i]
		by := s.actor(c.By)
		conv := s.wall
		for _, to := range c.To {
			if f, ok := s.forumByName(to); ok {
				conv = f.Conv
				break
			}
			if id, ok := deptWall(to); ok {
				conv = id
				break
			}
		}
		id := idAt(s.ago(time.Duration(i+1) * 26 * time.Hour))
		typ := c.Type
		var res *comms.Comm
		var req *approvals.Request
		// the stated sender sends it; if their role may not send that kind, an administrator sends it for them
		candidates := []*Actor{by, s.actor("David Otieno")}
		if by == s.Owner { // the super admin approves everything here, so someone else must be the requester
			candidates = candidates[1:]
			by = candidates[0]
		}
		for _, sender := range candidates {
			if !s.inConversation(conv, sender) {
				s.addToFolder(ctx, conv, sender)
			}
			env := s.sealFor(conv, sender, id, map[string]any{"title": c.Title, "body": c.Body, "to": c.To, "type": typ, "from": c.By})
			var err error
			res, req, err = s.Com.Send(ctx, s.principal(sender), comms.SendIn{Type: typ, Conversation: conv, Message: id, Epoch: s.ConvEp[conv], Device: sender.DevID, Envelope: env})
			if err == nil {
				break
			}
			if sender != by || !strings.Contains(err.Error(), "forbidden") {
				s.must(err, "comm "+c.Title)
			}
		}
		if c.Status == "Published" && req != nil {
			s.approveAll(ctx, req)
		} else if req != nil && c.Status != "Published" {
			// "awaiting approval" in the mock: one level has said yes and the next is waiting
			if req.Levels > 1 {
				_, err := s.Ap.Decide(ctx, s.Owner.P, req.ID, true, "Looks fine to me.")
				s.must(err, "first level")
			}
		}
		// the read count: that many people opened it
		if c.Status == "Published" {
			for _, r := range s.pick("read:"+c.ID, min(c.Read, len(s.All)-2)) {
				if s.inConversation(conv, r) {
					_ = s.Com.MarkRead(ctx, s.principal(r), res.ID)
				}
			}
		}
	}

	var aw []struct {
		ID, Type, Citation, Date, Status string
		People                           []string
	}
	s.st("awardList", &aw)
	for i := len(aw) - 1; i >= 0; i-- {
		a := aw[i]
		var people []uuid.UUID
		for _, n := range a.People {
			people = append(people, s.actor(n).Person)
		}
		nominator := s.byRole("hod")[0]
		_, req, err := s.Com.NominateOn(ctx, s.principal(nominator), a.Type, people, a.Citation, a.Date)
		s.must(err, "award "+a.Type)
		if a.Status == "Approved" && req != nil {
			s.approveAll(ctx, req)
		}
	}

	// the files-and-projects items the mock approval inbox shows (a file shared outside, a project awaiting its last approval)
	fin := s.folder["finance"]
	by := s.actor("Amina Wanjiru")
	if !s.inConversation(fin, by) {
		s.addToFolder(ctx, fin, by)
	}
	fid := s.upload(ctx, fin, by, mockFile{Name: "Q3 budget tracker.xlsx", Kind: "XLS", Size: "86 KB", By: by.Name, When: "Yesterday"})
	keyEnv := s.sealFor(fin, by, idAt(s.Now), map[string]any{"name": "Q3 budget tracker.xlsx", "link_key": "demo"})
	_, req, err := s.File.CreateShare(ctx, s.principal(by), fid, files.ShareIn{ExpiresInHours: 72, MaxDownloads: 3, KeyEnvelope: keyEnv})
	s.must(err, "share request")
	_ = req

	kev := s.actor("Kevin Mutua")
	pc, _ := s.newConversation(ctx, s.actor("David Otieno"), chat.CreateIn{Kind: "project", Name: "Customer self-service portal", Color: "#C9B6FF"}, dedupeActors([]*Actor{kev, s.actor("Brian Kiprop"), s.actor("Amina Wanjiru")}, s.actor("David Otieno")))
	_, preq, err := s.Proj.Create(ctx, s.actor("David Otieno").P, projects.CreateIn{Conversation: pc.ID, Color: "#C9B6FF", Start: "2026-11-01", Due: "2027-05-01", Sync: "Not synced"})
	s.must(err, "pending project")
	for i := 0; i < preq.Levels-1; i++ {
		_, err := s.Ap.Decide(ctx, s.Owner.P, preq.ID, true, "Approved.")
		s.must(err, "project level")
	}
	_ = bytes.NewReader
	_ = e2ee.Version
}

func (s *Seeder) seedCalls(ctx context.Context) {
	var log []struct {
		ID, Ctx, Ref, Kind, Dir string
		At                      int64
		Dur                     int
	}
	s.st("callLog", &log)
	for i := len(log) - 1; i >= 0; i-- {
		l := log[i]
		other := s.actorSafe(l.Ref)
		if other == nil {
			continue
		}
		host, others := s.Owner, []uuid.UUID{other.Person}
		if l.Dir == "in" || l.Dir == "missed" || l.Dir == "blocked" {
			host, others = other, []uuid.UUID{s.Owner.Person}
		}
		status := map[string]string{"in": "ended", "out": "ended", "missed": "missed", "blocked": "declined"}[l.Dir]
		ctxName := "dm"
		var conv *uuid.UUID
		if l.Ctx == "forum" {
			ctxName = "forum"
		}
		at := time.UnixMilli(l.At).UTC()
		if at.After(s.Now) {
			at = s.ago(time.Duration(i+1) * time.Hour)
		}
		s.must(s.Call.RecordPast(ctx, s.OrgID, host.Person, l.Kind, ctxName, conv, others, status, at, l.Dur), fmt.Sprintf("call %s", l.ID))
	}
}

// actorSafe finds the other side of a call: a person, or (for forum calls) a stable member of the named forum.
func (s *Seeder) actorSafe(ref string) *Actor {
	for _, f := range s.forums {
		if f.Mock == ref && len(f.Members) > 1 {
			return f.Members[1]
		}
	}
	if a, ok := s.ByName[strings.TrimSpace(ref)]; ok {
		return a
	}
	return nil
}

func (s *Seeder) seedNotifications(ctx context.Context) {
	var notifs []struct {
		Text, Color, Time, Ticket string
	}
	s.st("notifs", &notifs)
	var nb notify.Batch
	s.must(s.A.DB.WithOrg(ctx, s.OrgID, func(tx pgxTx) error {
		for _, n := range notifs {
			if err := nb.Add(ctx, tx, s.OrgID, s.Owner.Person, "ticket", n.Text, n.Color, n.Ticket); err != nil {
				return err
			}
		}
		return nil
	}), "notifications")
	nb.Send(ctx, s.A)

	var widgets []string
	var theme string
	s.st("widgets", &widgets)
	s.st("theme", &theme)
	prefs := fmt.Sprintf(`{"widgets":%s,"theme":%q,"hiddenWidgets":[]}`, mustJSON(widgets), theme)
	s.must(s.Rol.SetPrefs(ctx, s.Owner.P, []byte(prefs)), "preferences")
}
