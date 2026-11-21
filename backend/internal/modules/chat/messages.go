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
	// Body: a server-readable message (the community wall) carries its content here instead of an envelope
	Body json.RawMessage `json:"body,omitempty"`
	// wall posts and comments: counts and my own reaction / vote, kept by the server (see wall.go)
	Counts     map[string]int      `json:"counts,omitempty"`
	MyReaction string              `json:"my_reaction,omitempty"`
	MyVote     string              `json:"my_vote,omitempty"`
	Names      map[string][]string `json:"names,omitempty"` // the first people per reaction (who liked, who loved)
	Type       string              `json:"type"`
	ReplyTo    *uuid.UUID          `json:"reply_to,omitempty"`
	CreatedAt  time.Time           `json:"created_at"`
	Deleted    bool                `json:"deleted,omitempty"`
}

type SendIn struct {
	ID       uuid.UUID   `json:"id"` // chosen by the sender: it is part of what the message is cryptographically bound to
	Epoch    int         `json:"epoch"`
	Device   uuid.UUID   `json:"device"`
	Envelope string      `json:"envelope"`
	Type     string      `json:"type"`
	ReplyTo  *uuid.UUID  `json:"reply_to"`
	Files    []uuid.UUID `json:"files"` // the uploaded files this message carries, so deleting it erases them too
	// Body: on the community wall, the content itself (JSON), read and kept by the server; no envelope, device or key epoch
	Body json.RawMessage `json:"body"`
}

var typeRe = regexp.MustCompile(`^[a-z][a-z_]{1,19}$`)

// A phone writes messages while offline and sends them when it reconnects; the id keeps the time it was written, so it
// shows in its place. Older than this, or ahead of the server clock by more than a little, the id is refused.
const (
	offlineGrace = 48 * time.Hour // the same limit as every change made offline (see server/lww.go)
	clockAhead   = 2 * time.Minute
)

// Send stores one encrypted message and tells everyone in the conversation. The server checks who may send and that the
// bytes look like an envelope; it cannot check what is inside.
func (s *Service) Send(ctx context.Context, p *tenant.Principal, conv uuid.UUID, in SendIn) (*Message, error) {
	plain := len(in.Body) > 0
	var env []byte
	var err error
	if plain {
		if len(in.Body) > 64<<10 || !json.Valid(in.Body) {
			return nil, httpx.Bad("That message is too long or not valid.")
		}
	} else {
		env, err = base64.StdEncoding.DecodeString(in.Envelope)
		if err != nil || e2ee.ValidateEnvelope(env) != nil {
			return nil, httpx.Bad("That message is not a valid encrypted message.")
		}
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
	age := time.Since(time.UnixMilli(int64(in.ID[0])<<40 | int64(in.ID[1])<<32 | int64(in.ID[2])<<24 | int64(in.ID[3])<<16 | int64(in.ID[4])<<8 | int64(in.ID[5])))
	if age > offlineGrace {
		return nil, httpx.Err(410, "offline_expired", "This message waited more than 48 hours without a connection, so it was not sent.")
	}
	if age < -clockAhead {
		return nil, httpx.Bad("The message id carries the wrong time. Check the device clock.")
	}
	var out *Message
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, _, epoch, err := s.lockConvNoLock(ctx, tx, p, conv)
		if err != nil {
			return err
		}
		if err := s.checkBlocked(ctx, tx, p.Org, kind); err != nil {
			return err
		}
		if m, r := sendRight(kind, in.Type); m != "" {
			if err := authz.Require(ctx, tx, p.Person, m, r); err != nil {
				return err
			}
		}
		if in.Type == "edit" {
			if err := s.checkEdit(ctx, tx, p, conv, kind, in.ReplyTo); err != nil {
				return err
			}
		}
		if kind == "ticket" {
			// a closed ticket's conversation is final: nothing more is said or attached
			var closed bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.tickets where conversation_id=$1 and step >= 4)`, conv).Scan(&closed); err != nil {
				return err
			}
			if closed {
				return httpx.Err(409, "ticket_closed", "This ticket is closed. Its conversation is read-only.")
			}
		}
		if kind == "forum" && in.Type != "reaction" {
			if err := mayPost(ctx, tx, conv, p.Person); err != nil {
				return err
			}
		}
		if plain {
			// only the community wall is server-readable; its likes and votes go through their own endpoints (counted by the server)
			if kind != "wall" {
				return httpx.Bad("This conversation is end-to-end encrypted: send an encrypted message.")
			}
			if in.Type == "reaction" || in.Type == "vote" {
				return httpx.Bad("Like and vote through the wall's own endpoints.")
			}
			if env, err = SealPlain(ctx, tx, s.A, p.Org, in.ID, in.Body); err != nil {
				return httpx.Bad("That message is not valid.")
			}
		} else {
			if err := s.myDevice(ctx, tx, p.Account, in.Device); err != nil {
				return err
			}
			if in.Epoch != epoch {
				return httpx.Err(409, "epoch_changed", "The conversation's key changed. Fetch the new key and send again.")
			}
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
		if err := s.checkEngagement(ctx, tx, p, conv, kind, in.Type, in.ReplyTo); err != nil {
			return err
		}
		dev := in.Device
		if plain {
			dev = uuid.Nil
		}
		if err := PostMessage(ctx, tx, s.A, Posted{Org: p.Org, Conversation: conv, ID: in.ID, Account: p.Account, Person: p.Person, Device: dev, Epoch: epoch, Envelope: env, Type: in.Type, ReplyTo: in.ReplyTo}); err != nil {
			return err
		}
		link := linkFiles
		if plain {
			link = linkPlainFiles
		}
		if err := link(ctx, tx, p.Org, in.ID, p.Person, in.Files); err != nil {
			return err
		}
		pid := p.Person
		// the time the database stored is the one everybody compares with (read ticks), so the echo carries that one
		stored := time.Now().UTC()
		if err := tx.QueryRow(ctx, `select created_at from hub.messages where org_id=$1 and id=$2`, p.Org, in.ID).Scan(&stored); err != nil {
			return err
		}
		out = &Message{ID: in.ID, Conversation: conv, SenderAccount: p.Account, SenderPerson: &pid, SenderName: p.Name, Epoch: epoch, Type: in.Type, ReplyTo: in.ReplyTo, CreatedAt: stored.UTC()}
		if plain {
			out.Body = in.Body
		} else {
			out.SenderDevice, out.Envelope = &in.Device, in.Envelope
		}
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
	var tampered []uuid.UUID
	defer func() { reportTampering(context.WithoutCancel(ctx), s.A, p.Org, conv, tampered) }()
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, _, _, err := s.lockConvNoLock(ctx, tx, p, conv)
		if err != nil {
			return err
		}
		var dek []byte // the organization's key, for server-readable (wall) messages; fetched before reading rows
		if kind == "wall" {
			dek, _, _ = s.A.OrgDEK(ctx, tx, p.Org)
		}
		// a blocked forum still opens, for its members and calls, but shows no messages
		if err := s.checkBlocked(ctx, tx, p.Org, kind); err != nil {
			if kind == "forum" {
				return nil
			}
			return err
		}
		var first int
		if err := tx.QueryRow(ctx, `select first_epoch from hub.conversation_members where conversation_id=$1 and account_id=$2`, conv, p.Account).Scan(&first); err != nil {
			return err
		}
		cleared := s.clearedAt(ctx, tx, conv, p.Account)
		q := `select m.id, m.sender_account, m.sender_person, m.sender_device, coalesce(pe.first_name || ' ' || pe.last_name, ''), m.epoch, m.ciphertext, m.content_type, m.reply_to, m.created_at, m.deleted_at is not null, m.ver, m.mac, hub.canon(m)
			from hub.messages m left join hub.people pe on pe.id = m.sender_person
			where m.org_id=$1 and m.conversation_id=$2 and m.epoch >= $3 and m.created_at > $4::timestamptz
			and not (m.content_type='moment' and m.created_at < now() - interval '24 hours')`
		args := []any{p.Org, conv, first, cleared}
		if before != nil {
			q += ` and m.id < $5`
			args = append(args, *before)
		}
		rs, err := tx.Query(ctx, q+` order by m.id desc limit `+itoa(limit), args...)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			m := Message{Conversation: conv}
			var ct, mac []byte
			var ver int
			var canon string
			if err := rs.Scan(&m.ID, &m.SenderAccount, &m.SenderPerson, &m.SenderDevice, &m.SenderName, &m.Epoch, &ct, &m.Type, &m.ReplyTo, &m.CreatedAt, &m.Deleted, &ver, &mac, &canon); err != nil {
				return err
			}
			// a row changed outside the system (a forged like or vote) is left out and reported
			if !s.A.Sign.Verify(mac, "messages", ver, canon) {
				tampered = append(tampered, m.ID)
				continue
			}
			if !m.Deleted && IsPlain(ct) {
				if b, ok := OpenPlain(dek, p.Org, m.ID, ct); ok {
					m.Body = b
				}
			} else if !m.Deleted {
				m.Envelope = base64.StdEncoding.EncodeToString(ct)
			}
			out = append(out, m)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		if kind == "wall" {
			return s.wallExtras(ctx, tx, p, out)
		}
		return nil
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
	var tampered []uuid.UUID
	defer func() { reportTampering(context.WithoutCancel(ctx), s.A, p.Org, conv, tampered) }()
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		kind, _, _, _, err := s.lockConvNoLock(ctx, tx, p, conv)
		if err != nil {
			return err
		}
		var dek []byte // the organization's key, for server-readable (wall) messages; fetched before reading rows
		if kind == "wall" {
			dek, _, _ = s.A.OrgDEK(ctx, tx, p.Org)
		}
		// a blocked forum still opens, for its members and calls, but shows no messages
		if err := s.checkBlocked(ctx, tx, p.Org, kind); err != nil {
			if kind == "forum" {
				return nil
			}
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
				  from hub.messages m where m.org_id=$1 and m.conversation_id=$2 and m.created_at >= $3 and m.epoch >= $4 and m.created_at > $6::timestamptz
				union
				select m.id, m.deleted_at as at
				  from hub.messages m where m.org_id=$1 and m.conversation_id=$2 and m.deleted_at >= $3 and m.epoch >= $4 and m.created_at > $6::timestamptz
			), ranked as (
				select id, max(at) as at from changed group by id
			)
			select m.id, m.sender_account, m.sender_person, m.sender_device, coalesce(pe.first_name || ' ' || pe.last_name, ''), m.epoch, m.ciphertext, m.content_type, m.reply_to, m.created_at, m.deleted_at is not null, r.at, m.ver, m.mac, hub.canon(m)
			  from ranked r join hub.messages m on m.org_id=$1 and m.id=r.id
			  left join hub.people pe on pe.id = m.sender_person
			 where (r.at, r.id) > ($3::timestamptz, $5::uuid) and not (m.content_type='moment' and m.created_at < now() - interval '24 hours')
			 order by r.at, r.id
			 limit `+itoa(limit+1), p.Org, conv, since, first, afterID, s.clearedAt(ctx, tx, conv, p.Account))
		if err != nil {
			return err
		}
		defer rs.Close()
		var lastAt time.Time
		var lastID uuid.UUID
		for rs.Next() {
			m := Message{Conversation: conv}
			var ct, mac []byte
			var at time.Time
			var ver int
			var canon string
			if err := rs.Scan(&m.ID, &m.SenderAccount, &m.SenderPerson, &m.SenderDevice, &m.SenderName, &m.Epoch, &ct, &m.Type, &m.ReplyTo, &m.CreatedAt, &m.Deleted, &at, &ver, &mac, &canon); err != nil {
				return err
			}
			if len(page.Messages) == limit {
				page.More = true
				break
			}
			lastAt, lastID = at, m.ID
			if !s.A.Sign.Verify(mac, "messages", ver, canon) {
				tampered = append(tampered, m.ID)
				continue
			}
			if !m.Deleted && IsPlain(ct) {
				if b, ok := OpenPlain(dek, p.Org, m.ID, ct); ok {
					m.Body = b
				}
			} else if !m.Deleted {
				m.Envelope = base64.StdEncoding.EncodeToString(ct)
			}
			page.Messages = append(page.Messages, m)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		if kind == "wall" {
			if err := s.wallExtras(ctx, tx, p, page.Messages); err != nil {
				return err
			}
		}
		switch {
		case page.More:
			page.Since, page.After = lastAt, lastID.String()
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
		if err := s.checkBlocked(ctx, tx, p.Org, kind); err != nil {
			return err
		}
		var sender uuid.UUID
		var ctype string
		if err := tx.QueryRow(ctx, `select sender_account, content_type from hub.messages where org_id=$1 and id=$2 and conversation_id=$3 and deleted_at is null`, p.Org, id, conv).Scan(&sender, &ctype); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		// votes and likes are protected: a vote is final for everyone (it is never removed, not even by its owner or a moderator), and a
		// like can be taken back only by the person who gave it
		if ctype == "vote" {
			return httpx.Err(409, "vote_final", "A vote is final and cannot be taken back.")
		}
		if ctype == "reaction" && sender != p.Account {
			return httpx.ErrForbidden
		}
		if rule, ok := messageRules[kind]; ok {
			if err := authz.RequireOwn(ctx, tx, p.Person, rule.module, rule.moderate, rule.del, sender == p.Account); err != nil {
				return err
			}
		} else if sender != p.Account {
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
		if err := touchChanged(ctx, tx, p.Org, conv); err != nil {
			return err
		}
		// what hangs on it goes too: comments, likes, and every picture or video it carried (erased from storage, not just hidden)
		if err := s.eraseAround(ctx, tx, p.Org, []uuid.UUID{id}); err != nil {
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
	s.purge(ctx)
	d, _ := json.Marshal(map[string]any{"id": id, "conversation": conv})
	s.A.Publish(ctx, realtime.Event{Type: "message.deleted", Channel: realtime.ConvChannel(conv), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	return nil
}
