package push

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/emaildetails"
	"hub/internal/realtime"
)

// chat messages that are worth a notification (not likes, votes, edits or read marks)
var pushTypes = map[string]bool{"text": true, "poll": true, "memo": true, "announcement": true, "post": true, "share": true}

func wanted(e realtime.Event) bool {
	switch e.Type {
	case "notification", "call.incoming", "call.started":
		return strings.HasPrefix(e.Channel, "acct:")
	case "message.new":
		return strings.HasPrefix(e.Channel, "conv:")
	}
	return false
}

func idAfter(channel, prefix string) uuid.UUID {
	id, _ := uuid.Parse(strings.TrimPrefix(channel, prefix))
	return id
}

func (s *Service) compose(ctx context.Context, e realtime.Event) (job, bool) {
	org, err := uuid.Parse(e.Org)
	if err != nil {
		return job{}, false
	}
	detail := s.details.allowed(ctx, s.a, org)
	switch e.Type {
	case "notification":
		var n struct {
			ID, Kind, Text, Ref string
		}
		if json.Unmarshal(e.Data, &n) != nil {
			return job{}, false
		}
		body := "You have a new notification."
		if detail && n.Text != "" {
			body = n.Text
		}
		return job{org, []uuid.UUID{idAfter(e.Channel, "acct:")}, Note{Title: "Hub", Body: body, Tag: "note:" + n.ID,
			Data: map[string]string{"kind": "notification", "id": n.ID, "ref": n.Ref, "about": n.Kind}}}, true
	case "call.incoming", "call.started":
		var c struct {
			Call         string `json:"call"`
			Kind         string `json:"kind"`
			Context      string `json:"context"`
			Conversation string `json:"conversation"`
			FromName     string `json:"from_name"`
		}
		if json.Unmarshal(e.Data, &c) != nil {
			return job{}, false
		}
		what := "voice call"
		if c.Kind == "video" {
			what = "video call"
		}
		title, body := "Hub", "Incoming "+what
		if e.Type == "call.started" {
			body = "A group " + what + " started"
		}
		if detail && c.FromName != "" {
			title = c.FromName
			if e.Type == "call.started" {
				body = c.FromName + " started a group " + what
			}
		}
		return job{org, []uuid.UUID{idAfter(e.Channel, "acct:")}, Note{Title: title, Body: body, Tag: "call:" + c.Call, Urgent: true,
			Data: map[string]string{"kind": "call", "call": c.Call, "conversation": c.Conversation, "context": c.Context}}}, true
	case "message.new":
		return s.chatNote(ctx, org, idAfter(e.Channel, "conv:"), e.Data, detail)
	}
	return job{}, false
}

// chatNote: everyone in a private chat or a forum except the sender (and those who muted it) hears of a new message.
func (s *Service) chatNote(ctx context.Context, org, conv uuid.UUID, data json.RawMessage, detail bool) (job, bool) {
	var m struct {
		Type          string    `json:"type"`
		SenderAccount uuid.UUID `json:"sender_account"`
		SenderName    string    `json:"sender_name"`
	}
	if json.Unmarshal(data, &m) != nil || !pushTypes[m.Type] {
		return job{}, false
	}
	var kind, name string
	var to []uuid.UUID
	err := s.a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `select kind, coalesce(name, '') from hub.conversations where id=$1`, conv).Scan(&kind, &name); err != nil {
			return err
		}
		if kind != "dm" && kind != "forum" {
			return nil
		}
		rs, err := tx.Query(ctx, `select distinct account_id from hub.conversation_members where conversation_id=$1 and left_at is null and not muted and account_id <> $2`, conv, m.SenderAccount)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var a uuid.UUID
			if err := rs.Scan(&a); err != nil {
				return err
			}
			to = append(to, a)
		}
		return rs.Err()
	})
	if err != nil || len(to) == 0 {
		return job{}, false
	}
	what := map[string]string{"poll": "New poll", "memo": "New memo", "announcement": "New announcement"}[m.Type]
	if what == "" {
		what = "New message"
	}
	title, body := "Hub", what
	if detail {
		if kind == "dm" && m.SenderName != "" {
			title = m.SenderName
		} else if kind == "forum" && name != "" {
			title = name
			if m.SenderName != "" {
				body = m.SenderName + ": " + strings.ToLower(what[:1]) + what[1:]
			}
		}
	}
	return job{org, to, Note{Title: title, Body: body, Tag: "conv:" + conv.String(),
		Data: map[string]string{"kind": kind, "conversation": conv.String()}}}, true
}

// detailCache remembers each organization's "details in emails" choice for a minute, so a busy chat does not read it per message.
type detailCache struct {
	mu   sync.Mutex
	seen map[uuid.UUID]detailEntry
}

type detailEntry struct {
	on bool
	at time.Time
}

func newDetailCache() *detailCache { return &detailCache{seen: map[uuid.UUID]detailEntry{}} }

func (c *detailCache) allowed(ctx context.Context, a *app.App, org uuid.UUID) bool {
	c.mu.Lock()
	if e, ok := c.seen[org]; ok && time.Since(e.at) < time.Minute {
		c.mu.Unlock()
		return e.on
	}
	c.mu.Unlock()
	on := false
	_ = a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error { on = emaildetails.Allowed(ctx, a, tx, org); return nil })
	c.mu.Lock()
	c.seen[org] = detailEntry{on, time.Now()}
	c.mu.Unlock()
	return on
}
