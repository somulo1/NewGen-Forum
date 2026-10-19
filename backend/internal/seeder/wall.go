package seeder

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"hub/internal/modules/chat"
)

func (s *Seeder) seedWall(ctx context.Context) {
	// one wall for the whole organization: every demo person is in its audience and holds its key
	c, _ := s.newConversation(ctx, s.Owner, chat.CreateIn{Kind: "wall", Name: "Community wall", Audience: "org", Color: "#F49AC1"}, s.All[1:])
	s.wall = c.ID

	var posts []struct {
		ID                  int
		Author, Meta        string
		Text                string
		Likes, Extra        int
		Shares              int
		PhotoLabel, PhotoBg string
		HasPoll             bool
		Comments            []struct{ Who, Text string }
	}
	s.st("posts", &posts)
	for i := len(posts) - 1; i >= 0; i-- { // oldest first
		p := posts[i]
		author := s.actor(p.Author)
		dept := strings.TrimSpace(strings.Split(p.Meta, "·")[0])
		when := s.ago(parseAgo(strings.TrimSpace(strings.Split(p.Meta, "·")[len(strings.Split(p.Meta, "·"))-1])))
		body := map[string]any{"text": p.Text, "department": dept}
		if p.PhotoLabel != "" {
			body["photo"] = map[string]string{"label": p.PhotoLabel, "bg": p.PhotoBg}
		}
		pid := s.say(ctx, c.ID, author, "post", body, when, nil)
		for k, cm := range p.Comments {
			s.say(ctx, c.ID, s.actor(cm.Who), "comment", map[string]any{"text": cm.Text}, when.Add(time.Duration(k+1)*4*time.Minute), &pid)
		}
		// reactions are separate encrypted messages from distinct people, so the counts are real
		for k, a := range s.pick("like:"+strings.ToLower(p.Author)+p.Text[:min(12, len(p.Text))], min(p.Likes, 60), author) {
			s.say(ctx, c.ID, a, "reaction", map[string]any{"kind": "like"}, when.Add(time.Duration(k+2)*time.Minute), &pid)
		}
		for k, a := range s.pick("love:"+strings.ToLower(p.Author)+p.Text[:min(12, len(p.Text))], min(p.Extra, 40), author) {
			s.say(ctx, c.ID, a, "reaction", map[string]any{"kind": "love"}, when.Add(time.Duration(k+3)*time.Minute), &pid)
		}
		for k, a := range s.pick("share:"+p.Text[:min(10, len(p.Text))], min(p.Shares, 6), author) {
			s.say(ctx, c.ID, a, "share", map[string]any{"note": ""}, when.Add(time.Duration(k+10)*time.Minute), &pid)
		}
	}

	var moments []struct {
		ID, Owner, Caption, Bg string
		Media                  struct {
			URL   string `json:"url"`
			IsImg bool   `json:"isImg"`
			Name  string `json:"name"`
		}
		Likes, Loves []string
	}
	s.st("moments", &moments)
	for i := len(moments) - 1; i >= 0; i-- {
		m := moments[i]
		owner := s.actor(m.Owner)
		when := s.ago(time.Duration(2+i*3) * time.Hour)
		mid := s.say(ctx, c.ID, owner, "moment", map[string]any{"caption": m.Caption, "bg": m.Bg, "media": m.Media}, when, nil)
		for k, n := range m.Likes {
			s.say(ctx, c.ID, s.actor(n), "reaction", map[string]any{"kind": "like"}, when.Add(time.Duration(k+1)*time.Minute), &mid)
		}
		for k, n := range m.Loves {
			s.say(ctx, c.ID, s.actor(n), "reaction", map[string]any{"kind": "love"}, when.Add(time.Duration(k+5)*time.Minute), &mid)
		}
	}
	_ = uuid.Nil
}
