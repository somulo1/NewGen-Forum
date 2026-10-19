package seeder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"hub/internal/modules/approvals"
	"hub/internal/modules/chat"
	"hub/internal/modules/projects"
	"hub/internal/modules/tickets"
)

type mockTicket struct {
	ID, Title, Cat, Prio string
	Step                 int
	Hold                 bool
	Late                 bool
	RaisedBy             string
	Assignees            []string
	Handler              string
	Elapsed              float64
	Sync, Desc, Dept     string
	CreatedAt            any
	Escalated            bool
	Activity             []struct {
		Who, Kind, Text, Time string
	}
}

func (s *Seeder) seedTickets(ctx context.Context) {
	var list []mockTicket
	s.st("tickets", &list)
	for i := len(list) - 1; i >= 0; i-- { // oldest first
		t := list[i]
		raiser := s.actor(t.RaisedBy)
		created := s.ago(time.Duration(t.Elapsed*float64(time.Hour)) + time.Hour)
		// the ticket's conversation holds the requester and everyone who works on it
		var team []*Actor
		for _, n := range append([]string{t.Handler}, t.Assignees...) {
			if n != "" {
				team = append(team, s.actor(n))
			}
		}
		c, _ := s.newConversation(ctx, raiser, chat.CreateIn{Kind: "ticket"}, dedupeActors(team, raiser))
		first := s.say(ctx, c.ID, raiser, "text", map[string]any{"title": t.Title, "desc": t.Desc, "category": t.Cat}, created, nil)
		_ = first
		for k, a := range t.Activity {
			if a.Kind == "comment" {
				s.say(ctx, c.ID, s.actor(a.Who), "text", map[string]any{"text": a.Text}, created.Add(time.Duration(k+1)*20*time.Minute), nil)
			}
		}
		dept := strings.TrimSpace(t.Dept)
		in := tickets.CreateIn{Conversation: c.ID, Category: t.Cat, Priority: t.Prio, ToType: "desk", Sync: t.Sync, At: &created, Ref: t.ID}
		_ = dept
		tk, err := s.Tick.Create(ctx, s.principal(raiser), in)
		s.must(err, "ticket "+t.ID)
		owner := s.Owner.P
		at := func(d time.Duration) *time.Time { x := created.Add(d); return &x }
		if t.Handler != "" {
			h := s.actor(t.Handler)
			if tk.Handler == nil || *tk.Handler != h.Person {
				_, err = s.Tick.Act(ctx, owner, tk.ID, tickets.Action{Do: "assign", Handler: &h.Person, At: at(30 * time.Minute)})
				s.must(err, "assign "+t.ID)
			}
		}
		for st := 2; st <= t.Step && st <= 4; st++ {
			if st == 4 && t.Prio == "High" {
				// closing a High ticket normally needs approval; the mock shows it closed, so it was approved
				continue
			}
			_, err = s.Tick.Act(ctx, owner, tk.ID, tickets.Action{Do: "step", Step: &st, At: at(time.Duration(st) * time.Hour)})
			s.must(err, fmt.Sprintf("step %d %s", st, t.ID))
		}
		if t.Hold && t.Step == 2 {
			_, err = s.Tick.Act(ctx, owner, tk.ID, tickets.Action{Do: "hold", At: at(3 * time.Hour)})
			s.must(err, "hold "+t.ID)
		}
		if t.Escalated {
			_, err = s.Tick.Act(ctx, owner, tk.ID, tickets.Action{Do: "escalate", At: at(4 * time.Hour)})
			s.must(err, "escalate "+t.ID)
		}
	}
	// overdue tickets are found by the same sweep that runs every minute in production
	s.must(s.Tick.SweepAll(ctx), "ticket sweep")
}

func (s *Seeder) seedProjects(ctx context.Context) {
	var list []struct {
		ID, Name, Code, Color, Owner string
		Team                         []string
		Status, Goal, Start, Due     string
		Budget                       int
		Sync                         string
		Tickets                      []string
		Milestones                   []struct {
			ID, Name, Iso, Desc string
			Closed              bool
			ClosedOn, ClosedBy  string
			ClosedComment       string
			History             []any
		}
		Risks []struct{ Text, Level string }
		Tasks []struct {
			Title              string
			Col                int
			Primary, Secondary string
			Start, Due, Desc   string
			Links              []any
			Objective, Sprint  string
			Done               bool
		}
		Activity []struct{ Text, Time string }
		Sprints  []struct {
			ID, Name, Start, End, Milestone string
		}
	}
	s.st("projects", &list)
	adm := s.actor("David Otieno") // an administrator creates projects; the super admin approves them
	cols := []string{"todo", "doing", "review", "done"}
	for _, p := range list {
		owner := s.actor(p.Owner)
		var team []*Actor
		for _, n := range p.Team {
			team = append(team, s.actor(n))
		}
		c, _ := s.newConversation(ctx, adm, chat.CreateIn{Kind: "project", Name: p.Name, Color: p.Color}, dedupeActors(append(team, owner), adm))
		res, req, err := s.Proj.Create(ctx, adm.P, projects.CreateIn{Conversation: c.ID, Color: p.Color, Start: p.Start, Due: p.Due, Budget: &p.Budget, Sync: p.Sync, Links: p.Tickets})
		s.must(err, "project "+p.Name)
		if req != nil {
			s.approveAll(ctx, req)
		}
		st := p.Status
		_, err = s.Proj.Update(ctx, adm.P, c.ID, projects.Patch{Status: &st, Owner: &owner.Person})
		s.must(err, "project status and owner")
		_ = res
		// the plan itself is encrypted for the team
		msIDs := map[string]uuid.UUID{}
		for _, m := range p.Milestones {
			id := idAt(s.ago(30 * 24 * time.Hour))
			msIDs[m.ID] = id
			s.record(ctx, c.ID, adm, "milestone", id, map[string]any{"name": m.Name, "desc": m.Desc, "closedOn": m.ClosedOn, "closedBy": m.ClosedBy, "closedComment": m.ClosedComment, "history": m.History}, map[string]any{"due": m.Iso, "closed": m.Closed}, nil)
		}
		spIDs := map[string]uuid.UUID{}
		for _, sp := range p.Sprints {
			id := idAt(s.ago(29 * 24 * time.Hour))
			spIDs[sp.ID] = id
			meta := map[string]any{"start": sp.Start, "end": sp.End}
			if m, ok := msIDs[sp.Milestone]; ok {
				meta["milestone"] = m.String()
			}
			s.record(ctx, c.ID, adm, "sprint", id, map[string]any{"name": sp.Name}, meta, nil)
		}
		for _, t := range p.Tasks {
			id := idAt(s.ago(20 * 24 * time.Hour))
			meta := map[string]any{"column": cols[min(max(t.Col, 0), 3)]}
			if t.Due != "" {
				meta["due"] = t.Due
			}
			if t.Primary != "" {
				meta["assignee"] = s.actor(t.Primary).Person.String()
			}
			if _, ok := spIDs[t.Sprint]; ok {
				meta["sprint"] = t.Sprint
			}
			s.record(ctx, c.ID, adm, "task", id, map[string]any{"title": t.Title, "desc": t.Desc, "primary": t.Primary, "secondary": t.Secondary, "start": t.Start, "due": t.Due, "links": t.Links, "objective": t.Objective, "sprint": t.Sprint}, meta, nil)
		}
		for _, r := range p.Risks {
			s.record(ctx, c.ID, adm, "risk", idAt(s.ago(10*24*time.Hour)), map[string]any{"text": r.Text}, map[string]any{"level": r.Level}, nil)
		}
		for i := len(p.Activity) - 1; i >= 0; i-- {
			a := p.Activity[i]
			s.record(ctx, c.ID, adm, "activity", idAt(s.ago(parseAgoLoose(a.Time))), map[string]any{"text": a.Text, "time": a.Time}, nil, nil)
		}
	}
}

// approveAll walks a request through every level, standing in as the super admin at each one.
func (s *Seeder) approveAll(ctx context.Context, req *approvals.Request) {
	for i := 0; i < req.Levels; i++ {
		_, err := s.Ap.Decide(ctx, s.Owner.P, req.ID, true, "Approved for the demo.")
		s.must(err, "approve "+req.Ref)
	}
}

func (s *Seeder) seedCalendar(ctx context.Context) {
	var events []struct {
		ID, Title, Type, Date, Start, End string
		People                            []string
		Forum, Repeat, By                 string
	}
	s.st("events", &events)
	for _, e := range events {
		by := s.actor(e.By)
		var others []*Actor
		for _, n := range e.People {
			others = append(others, s.actor(n))
		}
		c, _ := s.newConversation(ctx, by, chat.CreateIn{Kind: "event"}, dedupeActors(others, by))
		typ := e.Type
		repeat := e.Repeat
		if repeat == "" {
			repeat = "None"
		}
		s.record(ctx, c.ID, by, "entry", idAt(s.ago(24*time.Hour)), map[string]any{"title": e.Title, "forum": e.Forum, "people": e.People, "by": e.By},
			map[string]any{"date": e.Date, "start": e.Start, "end": e.End, "type": typ, "repeat": repeat}, nil)
	}
}
