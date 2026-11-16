// Package comms is memos, announcements and newsletters, and performance awards. A communication's content is end-to-end
// encrypted for its audience; it is held, sealed, until the people who must approve it have, then posted.
package comms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/e2ee"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/modules/chat"
	"hub/internal/notify"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Service struct {
	A  *app.App
	Ap *approvals.Engine
}

func New(a *app.App, ap *approvals.Engine) *Service {
	s := &Service{A: a, Ap: ap}
	ap.Register("comm.publish", s.applyPublish)
	ap.Register("award.grant", s.applyAward)
	return s
}

type Comm struct {
	ID           uuid.UUID  `json:"id"`
	Ref          string     `json:"ref"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	By           uuid.UUID  `json:"by"`
	ByName       string     `json:"by_name"`
	Conversation uuid.UUID  `json:"conversation"`
	Message      uuid.UUID  `json:"message"`
	CreatedAt    time.Time  `json:"created_at"`
	PublishedAt  *time.Time `json:"published_at,omitempty"`
	Read         int        `json:"read"`
	Audience     int        `json:"audience"` // distinct people it reached across every place it went to, except the author (read/acknowledged likewise)
	IRead        bool       `json:"i_read"`
	IAcked       bool       `json:"i_acked"`
	Acked        int        `json:"acknowledged"`
	Batch        *uuid.UUID `json:"batch,omitempty"` // the sends made together, one per place
}

type SendIn struct {
	Type         string    `json:"type"` // Memo | Announcement | Newsletter
	Conversation uuid.UUID `json:"conversation"`
	Message      uuid.UUID `json:"message"` // chosen by the sender: the content is bound to it
	Epoch        int       `json:"epoch"`
	Device       uuid.UUID `json:"device"`
	Envelope     string    `json:"envelope"`
	// Body: for the community wall (server-readable), the content itself (JSON) instead of an envelope
	Body    json.RawMessage `json:"body"`
	Details *Details        `json:"details"` // what emails may say about it (written into emails, never stored on its own)
	Batch   uuid.UUID       `json:"batch"`   // optional: the sends made together (one per place) share one
	// RepostOf is the published communication this one posts again to a place it went to before. It needs the page's "Repost"
	// right instead of the right to send a new one.
	RepostOf uuid.UUID `json:"repost_of"`
}

func (s *Service) Send(ctx context.Context, p *tenant.Principal, in SendIn) (*Comm, *approvals.Request, error) {
	right := map[string][2]string{"Memo": {"memos", "Send memos"}, "Announcement": {"announcements", "Post"}, "Newsletter": {"newsletter", "Publish newsletters"}, "Job post": {"jobposts", "Post job posts"}}[in.Type]
	if right[0] == "" {
		return nil, nil, httpx.Bad("Choose memo, announcement, newsletter or job post.")
	}
	plain := len(in.Body) > 0
	var env []byte
	var err error
	if plain {
		if len(in.Body) > 256<<10 || !json.Valid(in.Body) {
			return nil, nil, httpx.Bad("That is too long or not valid.")
		}
	} else {
		env, err = base64.StdEncoding.DecodeString(in.Envelope)
		if err != nil || e2ee.ValidateEnvelope(env) != nil {
			return nil, nil, httpx.Bad("That is not a valid encrypted message.")
		}
	}
	if in.Message.Version() != 7 {
		return nil, nil, httpx.Bad("Messages need a time-ordered id (UUID v7).")
	}
	var out *Comm
	var req *approvals.Request
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if in.RepostOf != uuid.Nil {
			// posting an item again: its page's Repost right, only for a published item of the same kind, and only into a place it went to
			if err := authz.Require(ctx, tx, p.Person, right[0], "Repost"); err != nil {
				return err
			}
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.comms o where o.id=$1 and o.status='Published' and o.type=$2
				and (o.conversation_id=$3 or (o.batch is not null and exists (select 1 from hub.comms b where b.batch=o.batch and b.conversation_id=$3))))`, in.RepostOf, in.Type, in.Conversation).Scan(&ok); err != nil {
				return err
			}
			if !ok {
				return httpx.Err(409, "not_a_place", "That item was not posted there, so it cannot be reposted there.")
			}
		} else if err := authz.Require(ctx, tx, p.Person, right[0], right[1]); err != nil {
			return err
		}
		var kind string
		var epoch int
		var audience *string
		if err := tx.QueryRow(ctx, `select c.kind, c.epoch, c.audience from hub.conversations c join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null where c.id=$1`, in.Conversation, p.Account).Scan(&kind, &epoch, &audience); err != nil {
			return httpx.ErrNotFound
		}
		if kind != "wall" && kind != "forum" {
			return httpx.Bad("Send it to a wall audience or a forum.")
		}
		if plain {
			// the community wall is server-readable: its content is kept by the server, no key or device involved
			if kind != "wall" {
				return httpx.Bad("Forums are end-to-end encrypted: send an encrypted message.")
			}
			if env, err = chat.SealPlain(ctx, tx, s.A, p.Org, in.Message, in.Body); err != nil {
				return httpx.Bad("That is not valid.")
			}
			in.Epoch = epoch
		} else {
			if in.Epoch != epoch {
				return httpx.Err(409, "epoch_changed", "The key changed. Fetch the new key and send again.")
			}
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.devices where id=$1 and account_id=$2 and revoked_at is null)`, in.Device, p.Account).Scan(&ok); err != nil || !ok {
				return httpx.Bad("That device is not yours or was removed.")
			}
		}
		var n int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'comm',209) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org).Scan(&n); err != nil {
			return err
		}
		id := app.NewID()
		ref := fmt.Sprintf("C-%d", n)
		if _, err := tx.Exec(ctx, `insert into hub.comms (org_id, id, ref, type, by_person, conversation_id, message_id, epoch, device, envelope, batch) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,nullif($11, '00000000-0000-0000-0000-000000000000'::uuid))`,
			p.Org, id, ref, in.Type, p.Person, in.Conversation, in.Message, epoch, in.Device, env, in.Batch); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "comms", "id=$1", id); err != nil {
			return err
		}
		// announcements to the whole organization follow flow f1; everything else follows f10 (communication rules)
		flow := "f10"
		if in.Type == "Announcement" && audience != nil && *audience == "org" {
			flow = "f1"
		}
		r, applied, err := s.Ap.Submit(ctx, tx, p, flow, "comm.publish", in.Type+" "+ref, "", map[string]any{"comm": id, "details": in.Details.clean()})
		if err != nil {
			return err
		}
		if !applied {
			req = r
		}
		out, err = s.get(ctx, tx, p, id)
		return err
	})
	return out, req, err
}

type pub struct {
	Comm    uuid.UUID `json:"comm"`
	Details *Details  `json:"details,omitempty"` // for the emails only; it waits in the approval request, which is already stored encrypted
}

func (s *Service) applyPublish(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var pl pub
	if err := json.Unmarshal(a.Payload, &pl); err != nil {
		return err
	}
	var conv, msg, device, by uuid.UUID
	var epoch int
	var typ string
	var env []byte
	var acct *uuid.UUID
	if err := tx.QueryRow(ctx, `select c.conversation_id, c.message_id, c.device, c.by_person, c.epoch, c.type, c.envelope, p.account_id from hub.comms c join hub.people p on p.id=c.by_person where c.id=$1 and c.status='Pending approval' for update of c`, pl.Comm).
		Scan(&conv, &msg, &device, &by, &epoch, &typ, &env, &acct); err != nil {
		return err
	}
	var cur int
	if err := tx.QueryRow(ctx, `select epoch from hub.conversations where id=$1`, conv).Scan(&cur); err != nil {
		return err
	}
	s.bumpCounts(ctx, a.Org)
	status := "Published"
	if (cur != epoch && !chat.IsPlain(env)) || acct == nil {
		status = "Stale" // the audience's key changed while it waited: the sender must send it again
	} else if err := chat.PostMessage(ctx, tx, s.A, chat.Posted{Org: a.Org, Conversation: conv, ID: msg, Account: *acct, Person: by, Device: device, Epoch: epoch, Envelope: env, Type: map[string]string{"Memo": "memo", "Announcement": "announcement", "Newsletter": "post", "Job post": "post"}[typ]}); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `update hub.comms set status=$2, published_at = case when $2='Published' then now() end, `+integrity.Bump+` where id=$1`, pl.Comm, status); err != nil {
		return err
	}
	if err := s.A.Sign.Seal(ctx, tx, "comms", "id=$1", pl.Comm); err != nil {
		return err
	}
	var nb notify.Batch
	text := typ + " published."
	if status == "Stale" {
		text = typ + " could not be posted because the audience's key changed. Please send it again."
	}
	if err := nb.Add(ctx, tx, a.Org, by, "comm", text, "#B8F36B", ""); err != nil {
		return err
	}
	if status == "Published" {
		var cref string
		_ = tx.QueryRow(ctx, `select ref from hub.comms where id=$1`, pl.Comm).Scan(&cref)
		mailAudience(ctx, s.A, tx, &nb, a.Org, pl.Details, conv, by, typ, cref)
	}
	go func() {
		nb.Send(context.Background(), s.A)
		if status == "Published" {
			d, _ := json.Marshal(map[string]any{"comm": pl.Comm, "type": typ})
			s.A.Publish(context.Background(), realtime.Event{Type: "comm.published", Channel: realtime.ConvChannel(conv), Org: a.Org.String(), Data: d})
		}
	}()
	return nil
}

// the read and acknowledged counts follow the same audience rule as the insights (see inAudience), so every screen agrees
// onePlace: the communication went to one place only (no other published copy in its batch). Then its audience is that place's
// kept member count (minus the author, when a member), and its reads and acknowledgements are counted from the readers, never by
// listing the whole audience. Several places need the distinct count across them, done exactly as before.
const onePlace = `(c.batch is null or not exists (select 1 from hub.comms c3 where c3.batch = c.batch and c3.id <> c.id and c3.status = 'Published'))`

const authorMember = `(case when exists (select 1 from hub.conversation_members am join hub.people ap on ap.account_id = am.account_id and ap.org_id = am.org_id
	where am.conversation_id = c.conversation_id and am.left_at is null and ap.id = c.by_person) then 1 else 0 end)`

var sel = `select c.id, c.ref, c.type, c.status, c.by_person, p.first_name || ' ' || p.last_name, c.conversation_id, c.message_id, c.created_at, c.published_at, c.batch,
	case when ` + onePlace + ` then (select count(*) from hub.comm_reads r where r.comm_id = c.id and ` + inAudience("r.person_id") + `)
	else (select count(distinct r.person_id) from hub.comm_reads r where r.comm_id in (` + grpComms + `) and r.person_id in (` + grpAudience + `)) end,
	exists (select 1 from hub.comm_reads r where r.comm_id = c.id and r.person_id = $1),
	exists (select 1 from hub.comm_acks a where a.comm_id = c.id and a.person_id = $1),
	case when ` + onePlace + ` then (select count(*) from hub.comm_acks a where a.comm_id = c.id and ` + inAudience("a.person_id") + `)
	else (select count(distinct a.person_id) from hub.comm_acks a where a.comm_id in (` + grpComms + `) and a.person_id in (` + grpAudience + `)) end,
	case when ` + onePlace + ` then greatest(coalesce((select cc.members from hub.conversation_counts cc where cc.conversation_id = c.conversation_id), 0) - ` + authorMember + `, 0)
	else (select count(*) from (` + grpAudience + `) ga) end
	from hub.comms c join hub.people p on p.id = c.by_person`

func scan(r pgx.Row) (*Comm, error) {
	x := &Comm{}
	err := r.Scan(&x.ID, &x.Ref, &x.Type, &x.Status, &x.By, &x.ByName, &x.Conversation, &x.Message, &x.CreatedAt, &x.PublishedAt, &x.Batch, &x.Read, &x.IRead, &x.IAcked, &x.Acked, &x.Audience)
	return x, err
}

func (s *Service) get(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID) (*Comm, error) {
	c, err := scan(tx.QueryRow(ctx, sel+` where c.id=$2`, p.Person, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	return c, err
}

// List shows communications the person may see: published ones addressed to a conversation they are in, plus their own and,
// for people who run communications, everything.
func (s *Service) List(ctx context.Context, p *tenant.Principal, typ string) ([]*Comm, error) {
	out := []*Comm{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		all, err := authz.Can(ctx, tx, p.Person, "announcements", "See who read")
		if err != nil {
			return err
		}
		q := sel + ` where ($2::text = '' or c.type = $2) and ($3::boolean or c.by_person = $1 or (c.status = 'Published' and exists (select 1 from hub.conversation_members m where m.conversation_id = c.conversation_id and m.account_id = $4 and m.left_at is null)))
			order by c.created_at desc limit 200`
		rs, err := tx.Query(ctx, q, p.Person, typ, all, p.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			c, err := scan(rs)
			if err != nil {
				return err
			}
			if !all && c.By != p.Person {
				c.Read = 0
			}
			out = append(out, c)
		}
		return rs.Err()
	})
	return out, err
}

// Content returns the sealed envelope of a communication (same visibility rule as List), plus who sealed it.
func (s *Service) Content(ctx context.Context, p *tenant.Principal, id uuid.UUID) (map[string]any, error) {
	var out map[string]any
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		all, err := authz.Can(ctx, tx, p.Person, "announcements", "See who read")
		if err != nil {
			return err
		}
		var conv, msg, device uuid.UUID
		var epoch int
		var env []byte
		var acct *uuid.UUID
		var by uuid.UUID
		err = tx.QueryRow(ctx, `select c.conversation_id, c.message_id, c.device, c.epoch, c.envelope, c.by_person, pe.account_id
			from hub.comms c join hub.people pe on pe.id = c.by_person
			where c.id=$2 and ($3::boolean or c.by_person = $1 or (c.status = 'Published' and exists (select 1 from hub.conversation_members m where m.conversation_id = c.conversation_id and m.account_id = $4 and m.left_at is null)))`,
			p.Person, id, all, p.Account).Scan(&conv, &msg, &device, &epoch, &env, &by, &acct)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		out = map[string]any{"id": id, "conversation": conv, "message": msg, "epoch": epoch, "device": device,
			"envelope": base64.StdEncoding.EncodeToString(env), "sender_person": by, "sender_account": acct}
		return nil
	})
	return out, err
}

// Contents is Content for many communications at once (the list screen opens them all), skipping the ones the caller may not see.
func (s *Service) Contents(ctx context.Context, p *tenant.Principal, ids []uuid.UUID) ([]map[string]any, error) {
	out := []map[string]any{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		all, err := authz.Can(ctx, tx, p.Person, "announcements", "See who read")
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org) // for wall items, kept readable by the server (fetched before reading rows)
		if err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select c.id, c.conversation_id, c.message_id, c.device, c.epoch, c.envelope, c.by_person, pe.account_id
			from hub.comms c join hub.people pe on pe.id = c.by_person
			where c.id = any($2) and ($3::boolean or c.by_person = $1 or (c.status = 'Published' and exists (select 1 from hub.conversation_members m where m.conversation_id = c.conversation_id and m.account_id = $4 and m.left_at is null)))`,
			p.Person, ids, all, p.Account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id, conv, msg, device, by uuid.UUID
			var epoch int
			var env []byte
			var acct *uuid.UUID
			if err := rs.Scan(&id, &conv, &msg, &device, &epoch, &env, &by, &acct); err != nil {
				return err
			}
			row := map[string]any{"id": id, "conversation": conv, "message": msg, "epoch": epoch, "device": device, "sender_person": by, "sender_account": acct}
			if chat.IsPlain(env) {
				if b, ok := chat.OpenPlain(dek, p.Org, msg, env); ok {
					row["body"] = b
				}
			} else {
				row["envelope"] = base64.StdEncoding.EncodeToString(env)
			}
			out = append(out, row)
		}
		return rs.Err()
	})
	return out, err
}

// MarkSeen records communications the person really saw on screen in a wall or forum (the browser reports a post once it
// was at least 60% visible for 2 seconds). Only published communications in that conversation, that the person is a
// current member of and did not write themselves, are counted.
func (s *Service) MarkSeen(ctx context.Context, p *tenant.Principal, conv uuid.UUID, msgs []uuid.UUID) (int, error) {
	if len(msgs) == 0 {
		return 0, nil
	}
	if len(msgs) > 200 {
		msgs = msgs[:200]
	}
	n := 0
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		s.bumpCounts(ctx, p.Org)
		tag, err := tx.Exec(ctx, `insert into hub.comm_reads (org_id, comm_id, person_id)
			select c.org_id, c.id, $3 from hub.comms c
			join hub.conversation_members m on m.conversation_id = c.conversation_id and m.account_id = $4 and m.left_at is null
			where c.conversation_id = $1 and c.message_id = any($2) and c.status = 'Published' and c.by_person <> $3
			on conflict do nothing`, conv, msgs, p.Person, p.Account)
		n = int(tag.RowsAffected())
		return err
	})
	return n, err
}

// MarkRead records that the person read a communication (once).
func (s *Service) MarkRead(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		s.bumpCounts(ctx, p.Org)
		tag, err := tx.Exec(ctx, `insert into hub.comm_reads (org_id, comm_id, person_id) select $1, c.id, $3 from hub.comms c
			join hub.conversation_members m on m.conversation_id = c.conversation_id and m.account_id = $4 and m.left_at is null
			where c.id=$2 and c.status='Published' on conflict do nothing`, p.Org, id, p.Person, p.Account)
		if err != nil {
			return err
		}
		_ = tag
		return nil
	})
}

// Acknowledge records that the caller received and noted a published communication sent to a place they are in (it also counts as
// read). Once is enough; clicking again changes nothing.
func (s *Service) Acknowledge(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		s.bumpCounts(ctx, p.Org)
		tag, err := tx.Exec(ctx, `insert into hub.comm_acks (org_id, comm_id, person_id) select $1, c.id, $3 from hub.comms c
			join hub.conversation_members m on m.conversation_id = c.conversation_id and m.account_id = $4 and m.left_at is null
			where c.id=$2 and c.status='Published' and c.by_person <> $3 on conflict do nothing`, p.Org, id, p.Person, p.Account)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.comm_acks where comm_id=$1 and person_id=$2)`, id, p.Person).Scan(&ok); err != nil || ok {
				return err
			}
			return httpx.ErrNotFound
		}
		s.bumpCounts(ctx, p.Org)
		_, err = tx.Exec(ctx, `insert into hub.comm_reads (org_id, comm_id, person_id) values ($1,$2,$3) on conflict do nothing`, p.Org, id, p.Person)
		return err
	})
}

// ---- communication rules ----

type Rule struct {
	On       bool     `json:"on"`
	To       []string `json:"to"`
	Sender   string   `json:"sender"`
	Approver string   `json:"approver"`
}
