package chat

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/orgsettings"
	"hub/internal/tenant"
)

// ownRules say which rights let someone change messages of each kind of conversation: the "own" rights for what they wrote
// themselves, and the moderation right for other people's. Kinds not listed here (notes, project and event threads) leave
// a person's own messages to them. All of these are rights the organization gives to roles.
type ownRule struct{ module, edit, del, moderate string }

var messageRules = map[string]ownRule{
	"forum":  {"forums", "Edit own messages", "Delete own messages", "Moderate"},
	"wall":   {"wall", "Edit own posts", "Delete own posts", "Delete others' posts"},
	"dm":     {"chat", "Edit own messages", "Delete own messages", ""},
	"ticket": {"tickets", "Edit own tickets", "Edit own tickets", ""},
}

// editable are the message types a person can edit (what they wrote, not reactions or votes).
var editable = map[string]bool{"text": true, "post": true, "comment": true}

// checkEdit: an edit is a new sealed message of type "edit" that points at the original. Only its author may edit it, and
// only with the "Edit own ..." right for that kind of conversation. Readers show the latest edit in place of the original.
func (s *Service) checkEdit(ctx context.Context, tx pgx.Tx, p *tenant.Principal, conv uuid.UUID, kind string, replyTo *uuid.UUID) error {
	if replyTo == nil {
		return httpx.Bad("Say which message you are editing.")
	}
	var sender uuid.UUID
	var typ string
	if err := tx.QueryRow(ctx, `select sender_account, content_type from hub.messages where org_id=$1 and id=$2 and conversation_id=$3 and deleted_at is null`, p.Org, *replyTo, conv).Scan(&sender, &typ); errors.Is(err, pgx.ErrNoRows) {
		return httpx.ErrNotFound
	} else if err != nil {
		return err
	}
	if sender != p.Account {
		return httpx.Err(403, "not_yours", "You can only edit what you wrote.")
	}
	if !editable[typ] {
		return httpx.Bad("That cannot be edited.")
	}
	if rule, ok := messageRules[kind]; ok {
		return authz.Require(ctx, tx, p.Person, rule.module, rule.edit)
	}
	return nil
}

// checkBlocked stops private chats or the wall when the organization has blocked them in its settings.
func (s *Service) checkBlocked(ctx context.Context, tx pgx.Tx, org uuid.UUID, kind string) error {
	if kind != "dm" && kind != "wall" && kind != "forum" {
		return nil
	}
	b, err := orgsettings.BlocksOf(ctx, s.A, tx, org)
	if err != nil {
		return err
	}
	if kind == "dm" && b.BlockPrivateChats {
		return httpx.Err(403, "blocked", "Private chats are switched off for this organization.")
	}
	if kind == "forum" && b.BlockForumChats {
		return httpx.Err(403, "blocked", "Forum chats are switched off for this organization.")
	}
	if kind == "wall" && b.BlockWall {
		return httpx.Err(403, "blocked", "The community wall is switched off for this organization.")
	}
	return nil
}

// sendRight is the right needed to put this kind of message into this kind of conversation (empty = membership is enough).
func sendRight(kind, typ string) (string, string) {
	switch kind {
	case "dm":
		return "chat", "Send private messages"
	case "wall":
		switch typ {
		case "post", "moment", "poll", "share", "memo", "announcement":
			return "wall", "Post"
		case "comment":
			return "wall", "Comment"
		}
		return "wall", "View"
	}
	return "", ""
}

// Posted is a message ready to store.
type Posted struct {
	Org, Conversation, ID, Account, Person, Device uuid.UUID
	Epoch                                          int
	Envelope                                       []byte
	Type                                           string
	ReplyTo                                        *uuid.UUID
	CreatedAt                                      *time.Time // history loading only (the seeder); live messages use the database clock
}

// PostMessage stores a message, seals it and updates the conversation's activity. Other modules (announcements that wait for
// approval, system notices) use it so every message in the database goes in the same way.
func PostMessage(ctx context.Context, tx pgx.Tx, a *app.App, m Posted) error {
	if _, err := tx.Exec(ctx, `insert into hub.messages (org_id, id, conversation_id, sender_account, sender_person, sender_device, epoch, ciphertext, content_type, reply_to, created_at)
		values ($1,$2,$3,$4,$5,nullif($6, '00000000-0000-0000-0000-000000000000'::uuid),$7,$8,$9,$10,coalesce($11, now()))`, m.Org, m.ID, m.Conversation, m.Account, m.Person, m.Device, m.Epoch, m.Envelope, m.Type, m.ReplyTo, m.CreatedAt); err != nil {
		return err
	}
	if err := a.Sign.Seal(ctx, tx, "messages", "id=$1 and org_id=$2", m.ID, m.Org); err != nil {
		return err
	}
	// changed_at moves for every kind of message (a like or an edit is a change a reconnecting device must catch up on);
	// the "latest message" fields only for real messages
	real := m.Type != "reaction" && m.Type != "vote"
	if _, err := tx.Exec(ctx, `update hub.conversation_stats set changed_at=now(),
		last_message = case when $5 then $3::uuid else last_message end,
		last_at = case when $5 then greatest(coalesce(last_at, 'epoch'::timestamptz), coalesce($4, now())) else last_at end,
		message_count = message_count + case when $5 then 1 else 0 end
		where org_id=$1 and conversation_id=$2`, m.Org, m.Conversation, m.ID, m.CreatedAt, real); err != nil {
		return err
	}
	return nil
}
