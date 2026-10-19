package chat

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/e2ee"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Message struct {
	ID            uuid.UUID  `json:"id"`
	Conversation  uuid.UUID  `json:"conversation"`
	SenderAccount uuid.UUID  `json:"sender_account"`
	SenderPerson  *uuid.UUID `json:"sender_person,omitempty"`
	SenderDevice  *uuid.UUID `json:"sender_device,omitempty"`
	SenderName    string     `json:"sender_name,omitempty"`
	Epoch         int        `json:"epoch"`
	Envelope      string     `json:"envelope"` // base64; ciphertext + signature, opened on the device
	Type          string     `json:"type"`
	ReplyTo       *uuid.UUID `json:"reply_to,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	Deleted       bool       `json:"deleted,omitempty"`
}

type SendIn struct {
	ID       uuid.UUID  `json:"id"` // chosen by the sender: it is part of what the message is cryptographically bound to
	Epoch    int        `json:"epoch"`
	Device   uuid.UUID  `json:"device"`
	Envelope string     `json:"envelope"`
	Type     string     `json:"type"`
	ReplyTo  *uuid.UUID `json:"reply_to"`
}

var typeRe = regexp.MustCompile(`^[a-z][a-z_]{1,19}$`)

// Send stores one encrypted message and tells everyone in the conversation. The server checks who may send and that the
// bytes look like an envelope; it cannot check what is inside.
func (s *Service) Send(ctx context.Context, p *tenant.Principal, conv uuid.UUID, in SendIn) (*Message, error) {
	env, err := base64.StdEncoding.DecodeString(in.Envelope)
	if err != nil || e2ee.ValidateEnvelope(env) != nil {
		return nil, httpx.Bad("That message is not a valid encrypted message.")
	}
	if in.Type == "" {
		in.Type = "text"
	}
	if !typeRe.MatchString(in.Type) {
		return nil, httpx.Bad("Unknown message type.")
	}
	if in.ID == uuid.Nil || in.ID.Version() != 7 {
		return nil, httpx.Bad("Messages need a time-ordered id (UUID v7) made by the sender.")
	}
	if d := time.Since(time.UnixMilli(int64(in.ID[0])<<40 | int64(in.ID[1])<<32 | int64(in.ID[2])<<24 | int64(in.ID[3])<<16 | int64(in.ID[4])<<8 | int64(in.ID[5]))); d > 10*time.Minute || d < -2*time.Minute {
		return nil, httpx.Bad("The message id carries the wrong time. Check the device clock.")
	}
	var out *Message
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, _, epoch, err := s.lockConvNoLock(ctx, tx, p, conv)
		if err != nil {
			return err
		}
		if m, r := sendRight(kind, in.Type); m != "" {
			if err := authz.Require(ctx, tx, p.Person, m, r); err != nil {
				return err
			}
		}
		if kind == "forum" && in.Type != "reaction" {
			if err := mayPost(ctx, tx, conv, p.Person); err != nil {
				return err
			}
		}
		if err := s.myDevice(ctx, tx, p.Account, in.Device); err != nil {
			return err
		}
		if in.Epoch != epoch {
			return httpx.Err(409, "epoch_changed", "The conversation's key changed. Fetch the new key and send again.")
		}
		if in.ReplyTo != nil {
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.messages where org_id=$1 and conversation_id=$2 and id=$3)`, p.Org, conv, *in.ReplyTo).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return httpx.Bad("The message you are replying to is not in this conversation.")
			}
		}
		if err := PostMessage(ctx, tx, s.A, Posted{Org: p.Org, Conversation: conv, ID: in.ID, Account: p.Account, Person: p.Person, Device: in.Device, Epoch: epoch, Envelope: env, Type: in.Type, ReplyTo: in.ReplyTo}); err != nil {
			return err
		}
		pid := p.Person
		out = &Message{ID: in.ID, Conversation: conv, SenderAccount: p.Account, SenderPerson: &pid, SenderDevice: &in.Device, SenderName: p.Name, Epoch: epoch, Envelope: in.Envelope, Type: in.Type, ReplyTo: in.ReplyTo, CreatedAt: time.Now().UTC()}
		return nil
	})
	if err != nil {
		if isDup(err) {
			return nil, httpx.Err(409, "duplicate", "That message was already sent.")
		}
		return nil, err
	}
	d, _ := json.Marshal(out)
	s.A.Publish(ctx, realtime.Event{Type: "message.new", Channel: realtime.ConvChannel(conv), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	return out, nil
}

func isDup(err error) bool {
	var pe *pgconn.PgError
	return errors.As(err, &pe) && pe.Code == "23505"
}

// History returns messages newest first, a page at a time, starting before `before` when given.
func (s *Service) History(ctx context.Context, p *tenant.Principal, conv uuid.UUID, before *uuid.UUID, limit int) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	out := []Message{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, conv); err != nil {
			return err
		}
		var first int
		if err := tx.QueryRow(ctx, `select first_epoch from hub.conversation_members where conversation_id=$1 and account_id=$2`, conv, p.Account).Scan(&first); err != nil {
			return err
		}
		q := `select m.id, m.sender_account, m.sender_person, m.sender_device, coalesce(pe.first_name || ' ' || pe.last_name, ''), m.epoch, m.ciphertext, m.content_type, m.reply_to, m.created_at, m.deleted_at is not null
			from hub.messages m left join hub.people pe on pe.id = m.sender_person
			where m.org_id=$1 and m.conversation_id=$2 and m.epoch >= $3`
		args := []any{p.Org, conv, first}
		if before != nil {
			q += ` and m.id < $4`
			args = append(args, *before)
		}
		rs, err := tx.Query(ctx, q+` order by m.id desc limit `+itoa(limit), args...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			m := Message{Conversation: conv}
			var ct []byte
			if err := rs.Scan(&m.ID, &m.SenderAccount, &m.SenderPerson, &m.SenderDevice, &m.SenderName, &m.Epoch, &ct, &m.Type, &m.ReplyTo, &m.CreatedAt, &m.Deleted); err != nil {
				return err
			}
			if !m.Deleted {
				m.Envelope = base64.StdEncoding.EncodeToString(ct)
			}
			out = append(out, m)
		}
		return rs.Err()
	})
	return out, err
}

func itoa(n int) string { return strconv.Itoa(n) }

// SyncPage is one page of Changes. When More is false the client stores Since (and an empty After) and asks with it next
// time; when true it asks again at once with Since and After to read the rest.
type SyncPage struct {
	Messages []Message `json:"messages"`
	More     bool      `json:"more"`
	Since    time.Time `json:"next_since"`
	After    string    `json:"next_after,omitempty"`
}

// syncOverlap is how far a finished catch-up steps back from "now". A message's created_at is set when its insert runs,
// so a slow transaction can become visible a moment after a read that was already past that time. Reading the last
// minute again (the client drops ids it already has) means such a message is picked up next time instead of lost.
const syncOverlap = "60 seconds"

// Changes lists what happened in a conversation from `since` on, oldest change first: messages stored since then and
// messages removed since then (marked deleted, without their ciphertext). (since, after) is a position in that order;
// a page holds at most `limit` rows. Clients use it to keep an open conversation current without re-reading its history.
func (s *Service) Changes(ctx context.Context, p *tenant.Principal, conv uuid.UUID, since time.Time, after *uuid.UUID, limit int) (*SyncPage, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	page := &SyncPage{Messages: []Message{}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if _, _, _, _, err := s.lockConvNoLock(ctx, tx, p, conv); err != nil {
			return err
		}
		var first int
		if err := tx.QueryRow(ctx, `select first_epoch from hub.conversation_members where conversation_id=$1 and account_id=$2`, conv, p.Account).Scan(&first); err != nil {
			return err
		}
		// the database clock decides "now", so app servers with drifting clocks agree
		var floor time.Time
		if err := tx.QueryRow(ctx, `select clock_timestamp() - interval '`+syncOverlap+`'`).Scan(&floor); err != nil {
			return err
		}
		afterID := uuid.Nil
		if after != nil {
			afterID = *after
		}
		// two index-backed halves (stored since / removed since), merged and ordered by when each last changed
		rs, err := tx.Query(ctx, `
			with changed as (
				select m.id, coalesce(m.deleted_at, m.created_at) as at
				  from hub.messages m where m.org_id=$1 and m.conversation_id=$2 and m.created_at >= $3 and m.epoch >= $4
				union
				select m.id, m.deleted_at as at
				  from hub.messages m where m.org_id=$1 and m.conversation_id=$2 and m.deleted_at >= $3 and m.epoch >= $4
			), ranked as (
				select id, max(at) as at from changed group by id
			)
			select m.id, m.sender_account, m.sender_person, m.sender_device, coalesce(pe.first_name || ' ' || pe.last_name, ''), m.epoch, m.ciphertext, m.content_type, m.reply_to, m.created_at, m.deleted_at is not null, r.at
			  from ranked r join hub.messages m on m.org_id=$1 and m.id=r.id
			  left join hub.people pe on pe.id = m.sender_person
			 where (r.at, r.id) > ($3::timestamptz, $5::uuid)
			 order by r.at, r.id
			 limit `+itoa(limit+1), p.Org, conv, since, first, afterID)
		if err != nil {
			return err
		}
		defer rs.Close()
		var lastAt time.Time
		for rs.Next() {
			m := Message{Conversation: conv}
			var ct []byte
			var at time.Time
			if err := rs.Scan(&m.ID, &m.SenderAccount, &m.SenderPerson, &m.SenderDevice, &m.SenderName, &m.Epoch, &ct, &m.Type, &m.ReplyTo, &m.CreatedAt, &m.Deleted, &at); err != nil {
				return err
			}
			if len(page.Messages) == limit {
				page.More = true
				break
			}
			if !m.Deleted {
				m.Envelope = base64.StdEncoding.EncodeToString(ct)
			}
			page.Messages = append(page.Messages, m)
			lastAt = at
		}
		if err := rs.Err(); err != nil {
			return err
		}
		switch {
		case page.More:
			page.Since, page.After = lastAt, page.Messages[len(page.Messages)-1].ID.String()
		case floor.After(since):
			page.Since = floor
		default:
			page.Since = since
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return page, nil
}

// Delete removes a message for everyone. The author can always do it; moderators need the Moderate right. The ciphertext
// is erased, and the removal is recorded in the audit log.
func (s *Service) Delete(ctx context.Context, p *tenant.Principal, conv, id uuid.UUID) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, _, _, err := s.lockConvNoLock(ctx, tx, p, conv)
		if err != nil {
			return err
		}
		var sender uuid.UUID
		if err := tx.QueryRow(ctx, `select sender_account from hub.messages where org_id=$1 and id=$2 and conversation_id=$3 and deleted_at is null`, p.Org, id, conv).Scan(&sender); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if sender != p.Account {
			if kind == "dm" {
				return httpx.ErrForbidden
			}
			if err := authz.Require(ctx, tx, p.Person, "forums", "Moderate"); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `update hub.messages set ciphertext='\x00'::bytea, deleted_at=now(), `+integrity.Bump+` where org_id=$1 and id=$2`, p.Org, id); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "messages", "id=$1 and org_id=$2", id, p.Org); err != nil {
			return err
		}
		if sender != p.Account {
			return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Removed a message", "Forums", p.IP, map[string]any{"conversation": conv, "message": id})
		}
		return nil
	})
	if err != nil {
		return err
	}
	d, _ := json.Marshal(map[string]any{"id": id, "conversation": conv})
	s.A.Publish(ctx, realtime.Event{Type: "message.deleted", Channel: realtime.ConvChannel(conv), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
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
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,coalesce($11, now()))`, m.Org, m.ID, m.Conversation, m.Account, m.Person, m.Device, m.Epoch, m.Envelope, m.Type, m.ReplyTo, m.CreatedAt); err != nil {
		return err
	}
	if err := a.Sign.Seal(ctx, tx, "messages", "id=$1 and org_id=$2", m.ID, m.Org); err != nil {
		return err
	}
	if m.Type != "reaction" && m.Type != "vote" {
		if _, err := tx.Exec(ctx, `update hub.conversation_stats set last_message=$3, last_at=greatest(coalesce(last_at, 'epoch'::timestamptz), coalesce($4, now())), message_count=message_count+1 where org_id=$1 and conversation_id=$2`, m.Org, m.Conversation, m.ID, m.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}
