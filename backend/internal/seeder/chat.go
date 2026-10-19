package seeder

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"hub/internal/modules/chat"
)

type forumData struct {
	ID, Name, Initials, Color, Privacy, About string
	Unread, Members, Online                   int
	Updates                                   []struct {
		Who, Initials, Color, Time, Text string
		Replies                          int
	}
}

var agoRe = regexp.MustCompile(`(?i)^(\d+)\s*(min|h|d|day|days|hour|hours|week|weeks)\b`)

// parseAgo turns "14 min ago", "1 h ago", "3 days ago", "Yesterday" into how long ago it was.
func parseAgo(s string) time.Duration {
	s = strings.TrimSpace(s)
	if m := agoRe.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		switch strings.ToLower(m[2]) {
		case "min":
			return time.Duration(n) * time.Minute
		case "h", "hour", "hours":
			return time.Duration(n) * time.Hour
		case "d", "day", "days":
			return time.Duration(n) * 24 * time.Hour
		default:
			return time.Duration(n) * 7 * 24 * time.Hour
		}
	}
	switch strings.ToLower(s) {
	case "yesterday":
		return 26 * time.Hour
	case "just now", "now", "":
		return 30 * time.Second
	}
	return time.Hour
}

var nairobi = func() *time.Location { l, _ := time.LoadLocation("Africa/Nairobi"); return l }()

// parseClock turns "2:10 PM" into today at that time (Nairobi), or yesterday if that is still to come.
func (s *Seeder) parseClock(c string) time.Time {
	t, err := time.ParseInLocation("3:04 PM", strings.TrimSpace(c), nairobi)
	if err != nil {
		return s.ago(time.Hour)
	}
	n := s.Now.In(nairobi)
	at := time.Date(n.Year(), n.Month(), n.Day(), t.Hour(), t.Minute(), 0, 0, nairobi)
	if at.After(s.Now) {
		at = at.AddDate(0, 0, -1)
	}
	return at.UTC()
}

// pick chooses n distinct demo people in a repeatable order (so the demo looks the same every time it is seeded).
func (s *Seeder) pick(seed string, n int, skip ...*Actor) []*Actor {
	if n <= 0 {
		return nil
	}
	skipSet := map[*Actor]bool{}
	for _, a := range skip {
		skipSet[a] = true
	}
	type scored struct {
		a *Actor
		h uint32
	}
	var l []scored
	for _, a := range s.All {
		if skipSet[a] || a.Role == "guest" {
			continue
		}
		h := fnv.New32a()
		h.Write([]byte(seed + "|" + a.Email))
		l = append(l, scored{a, h.Sum32()})
	}
	for i := 1; i < len(l); i++ { // insertion sort by hash: small lists, no extra imports
		for j := i; j > 0 && l[j].h < l[j-1].h; j-- {
			l[j], l[j-1] = l[j-1], l[j]
		}
	}
	out := make([]*Actor, 0, n)
	for _, x := range l {
		out = append(out, x.a)
		if len(out) == n {
			break
		}
	}
	return out
}

func (s *Seeder) seedDMs(ctx context.Context) {
	var dms []struct {
		ID, With string
		Unread   int
		Msgs     []struct{ From, Text, Time string }
	}
	s.st("dms", &dms)
	for _, d := range dms {
		other := s.actor(d.With)
		c, _ := s.newConversation(ctx, s.Owner, chat.CreateIn{Kind: "dm"}, []*Actor{other})
		var ids []uuid.UUID
		var theirs []bool
		for _, m := range d.Msgs {
			from := s.Owner
			if m.From != "me" {
				from = s.actor(m.From)
			}
			id := s.say(ctx, c.ID, from, "text", map[string]any{"text": m.Text}, s.parseClock(m.Time), nil)
			ids, theirs = append(ids, id), append(theirs, from != s.Owner)
		}
		s.markUnread(ctx, c.ID, ids, theirs, d.Unread)
	}
}

var replyPool = []string{"Thanks for the update.", "On it.", "Noted, thank you.", "Agreed.", "I will check and come back to you.", "Good to know.", "Sounds good.", "Done on my side."}

func (s *Seeder) seedForums(ctx context.Context) {
	for _, f := range s.data.forums {
		creator := s.Owner // the administrator sets the forums up, as the mock's "Forum admin" does
		members := s.pick("forum:"+f.ID, f.Members-1, creator)
		// everyone who speaks in the forum is in it, and so is the owner
		names := map[string]bool{}
		var must []*Actor
		for _, u := range f.Updates {
			a := s.actor(u.Who)
			if !names[a.Email] {
				names[a.Email] = true
				must = append(must, a)
			}
		}
		set := map[*Actor]bool{creator: true}
		var all []*Actor
		for _, a := range append(must, members...) {
			if !set[a] && len(all) < f.Members-1 {
				set[a] = true
				all = append(all, a)
			}
		}
		priv := "invite"
		if strings.HasPrefix(strings.ToLower(f.Privacy), "open") {
			priv = "open"
		}
		c, _ := s.newConversation(ctx, creator, chat.CreateIn{Kind: "forum", Name: f.Name, About: f.About, Color: f.Color, Privacy: priv}, all)
		// oldest first
		var ids []uuid.UUID
		var theirs []bool
		for i := len(f.Updates) - 1; i >= 0; i-- {
			u := f.Updates[i]
			from := s.actor(u.Who)
			at := s.ago(parseAgo(u.Time))
			id := s.say(ctx, c.ID, from, "text", map[string]any{"text": u.Text}, at, nil)
			ids, theirs = append(ids, id), append(theirs, from != s.Owner)
			for r := 0; r < u.Replies; r++ {
				rep := all[(i*7+r*3)%len(all)]
				if rep == from {
					rep = creator
				}
				rid := s.say(ctx, c.ID, rep, "text", map[string]any{"text": replyPool[(i+r)%len(replyPool)]}, at.Add(time.Duration(r+1)*3*time.Minute), &id)
				ids, theirs = append(ids, rid), append(theirs, rep != s.Owner)
			}
		}
		s.markUnread(ctx, c.ID, ids, theirs, f.Unread)
		s.forums = append(s.forums, forumRef{f.ID, c.ID, creator, append([]*Actor{creator}, all...)})
	}
	s.seedPolls(ctx)
	s.seedForumRules(ctx)
}

type forumRef struct {
	Mock    string
	Conv    uuid.UUID
	Creator *Actor
	Members []*Actor
}

func (s *Seeder) forum(mock string) forumRef {
	for _, f := range s.forums {
		if f.Mock == mock {
			return f
		}
	}
	panic("seed: no forum " + mock)
}

func (s *Seeder) forumByName(name string) (forumRef, bool) {
	for i, f := range s.data.forums {
		if f.Name == name {
			return s.forums[i], true
		}
	}
	return forumRef{}, false
}

func (s *Seeder) seedPolls(ctx context.Context) {
	var polls []struct {
		ID, Forum, Question string
		Options             []struct {
			ID, Label string
			Votes     int
		}
		By      string
		At      int64
		Locked  bool
		Members int
	}
	s.st("polls", &polls)
	for _, p := range polls {
		f := s.forum(p.Forum)
		by := s.actor(p.By)
		at := time.UnixMilli(p.At).UTC()
		opts := make([]map[string]string, len(p.Options))
		for i, o := range p.Options {
			opts[i] = map[string]string{"id": o.ID, "label": o.Label}
		}
		pid := s.say(ctx, f.Conv, by, "poll", map[string]any{"question": p.Question, "options": opts}, at, nil)
		voter := 0
		for _, o := range p.Options {
			for v := 0; v < o.Votes && voter < len(f.Members); v++ {
				vt := f.Members[voter]
				voter++
				s.say(ctx, f.Conv, vt, "vote", map[string]any{"option": o.ID}, at.Add(time.Duration(voter+1)*time.Minute), &pid)
			}
		}
		if p.Locked {
			s.say(ctx, f.Conv, by, "poll_lock", map[string]any{"locked": true}, at.Add(48*time.Hour), &pid)
		}
	}
}

func (s *Seeder) seedForumRules(ctx context.Context) {
	var fc map[string]struct {
		Mode    string
		Roles   []string
		Blocked []string
		Allowed []string
	}
	s.st("forumChat", &fc)
	for mock, r := range fc {
		f := s.forum(mock)
		roleKeys := []string{}
		for _, name := range r.Roles {
			switch strings.ToLower(name) {
			case "admin":
				roleKeys = append(roleKeys, "admin", "superadmin")
			case "moderator":
				roleKeys = append(roleKeys, "hod")
			}
		}
		rules := chat.ForumRules{PostMode: r.Mode, Roles: roleKeys, Blocked: []uuid.UUID{}, Allowed: []uuid.UUID{}}
		if rules.PostMode == "" {
			rules.PostMode = "everyone"
		}
		s.must(s.Chat.SetRules(ctx, s.principal(f.Creator), f.Conv, rules), fmt.Sprintf("forum rules %s", mock))
	}
}
