// Package comms is memos, announcements and newsletters, and performance awards. A communication's content is end-to-end
// encrypted for its audience; it is held, sealed, until the people who must approve it have, then posted.
package comms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
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
	IRead        bool       `json:"i_read"`
}

type SendIn struct {
	Type         string    `json:"type"` // Memo | Announcement | Newsletter
	Conversation uuid.UUID `json:"conversation"`
	Message      uuid.UUID `json:"message"` // chosen by the sender: the content is bound to it
	Epoch        int       `json:"epoch"`
	Device       uuid.UUID `json:"device"`
	Envelope     string    `json:"envelope"`
}

func (s *Service) Send(ctx context.Context, p *tenant.Principal, in SendIn) (*Comm, *approvals.Request, error) {
	right := map[string][2]string{"Memo": {"memos", "Send memos"}, "Announcement": {"announcements", "Post"}, "Newsletter": {"newsletter", "Publish newsletters"}}[in.Type]
	if right[0] == "" {
		return nil, nil, httpx.Bad("Choose memo, announcement or newsletter.")
	}
	env, err := base64.StdEncoding.DecodeString(in.Envelope)
	if err != nil || e2ee.ValidateEnvelope(env) != nil {
		return nil, nil, httpx.Bad("That is not a valid encrypted message.")
	}
	if in.Message.Version() != 7 {
		return nil, nil, httpx.Bad("Messages need a time-ordered id (UUID v7).")
	}
	var out *Comm
	var req *approvals.Request
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, right[0], right[1]); err != nil {
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
		if in.Epoch != epoch {
			return httpx.Err(409, "epoch_changed", "The key changed. Fetch the new key and send again.")
		}
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.devices where id=$1 and account_id=$2 and revoked_at is null)`, in.Device, p.Account).Scan(&ok); err != nil || !ok {
			return httpx.Bad("That device is not yours or was removed.")
		}
		var n int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'comm',209) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org).Scan(&n); err != nil {
			return err
		}
		id := app.NewID()
		ref := fmt.Sprintf("C-%d", n)
		if _, err := tx.Exec(ctx, `insert into hub.comms (org_id, id, ref, type, by_person, conversation_id, message_id, epoch, device, envelope) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			p.Org, id, ref, in.Type, p.Person, in.Conversation, in.Message, epoch, in.Device, env); err != nil {
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
		r, applied, err := s.Ap.Submit(ctx, tx, p, flow, "comm.publish", in.Type+" "+ref, "", map[string]any{"comm": id})
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
	Comm uuid.UUID `json:"comm"`
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
	status := "Published"
	if cur != epoch || acct == nil {
		status = "Stale" // the audience's key changed while it waited: the sender must send it again
	} else if err := chat.PostMessage(ctx, tx, s.A, chat.Posted{Org: a.Org, Conversation: conv, ID: msg, Account: *acct, Person: by, Device: device, Epoch: epoch, Envelope: env, Type: map[string]string{"Memo": "memo", "Announcement": "announcement", "Newsletter": "post"}[typ]}); err != nil {
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
	go func() {
		nb.Send(context.Background(), s.A)
		if status == "Published" {
			d, _ := json.Marshal(map[string]any{"comm": pl.Comm, "type": typ})
			s.A.Publish(context.Background(), realtime.Event{Type: "comm.published", Channel: realtime.ConvChannel(conv), Org: a.Org.String(), Data: d})
		}
	}()
	return nil
}

const sel = `select c.id, c.ref, c.type, c.status, c.by_person, p.first_name || ' ' || p.last_name, c.conversation_id, c.message_id, c.created_at, c.published_at,
	(select count(*) from hub.comm_reads r where r.comm_id = c.id), exists (select 1 from hub.comm_reads r where r.comm_id = c.id and r.person_id = $1)
	from hub.comms c join hub.people p on p.id = c.by_person`

func scan(r pgx.Row) (*Comm, error) {
	x := &Comm{}
	err := r.Scan(&x.ID, &x.Ref, &x.Type, &x.Status, &x.By, &x.ByName, &x.Conversation, &x.Message, &x.CreatedAt, &x.PublishedAt, &x.Read, &x.IRead)
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

// MarkRead records that the person read a communication (once).
func (s *Service) MarkRead(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
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

// ---- communication rules ----

type Rule struct {
	On       bool     `json:"on"`
	To       []string `json:"to"`
	Sender   string   `json:"sender"`
	Approver string   `json:"approver"`
}

func (s *Service) Rules(ctx context.Context, p *tenant.Principal) (map[string]map[string]Rule, error) {
	out := map[string]map[string]Rule{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select level, type, config from hub.comm_rules`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var l, t string
			var raw []byte
			if err := rs.Scan(&l, &t, &raw); err != nil {
				return err
			}
			var r Rule
			_ = json.Unmarshal(raw, &r)
			if out[l] == nil {
				out[l] = map[string]Rule{}
			}
			out[l][t] = r
		}
		return rs.Err()
	})
	return out, err
}

func (s *Service) SetRule(ctx context.Context, p *tenant.Principal, level, typ string, r Rule) error {
	if (level != "departments" && level != "branches" && level != "sectors" && level != "units") || (typ != "Memo" && typ != "Announcement") {
		return httpx.Bad("Unknown rule.")
	}
	if len(r.To) > 8 {
		return httpx.Bad("Too many audiences.")
	}
	raw, _ := json.Marshal(r)
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Configure approvals"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.comm_rules (org_id, level, type, config) values ($1,$2,$3,$4) on conflict (org_id, level, type) do update set config=excluded.config, `+integrity.BumpOn("comm_rules"), p.Org, level, typ, raw); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "comm_rules", "level=$1 and type=$2", level, typ); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Changed the %s rule for %s", typ, level), "Settings", p.IP, r)
	})
}

// ---- awards ----

type Award struct {
	ID       uuid.UUID   `json:"id"`
	Ref      string      `json:"ref"`
	Type     string      `json:"type"`
	People   []uuid.UUID `json:"people"`
	Names    []string    `json:"names"`
	Citation string      `json:"citation"`
	Date     string      `json:"date"`
	Status   string      `json:"status"`
	By       string      `json:"by"`
}

func awardCtx(org, id uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "awards", Column: "citation_enc", RowID: id.String()}
}

func (s *Service) Nominate(ctx context.Context, p *tenant.Principal, typ string, people []uuid.UUID, citation string) (*Award, *approvals.Request, error) {
	return s.NominateOn(ctx, p, typ, people, citation, "")
}

// NominateOn is Nominate with a given award date (history loading); an empty date means today.
func (s *Service) NominateOn(ctx context.Context, p *tenant.Principal, typ string, people []uuid.UUID, citation, onDate string) (*Award, *approvals.Request, error) {
	if typ == "" || len(typ) > 60 || len(people) == 0 || len(people) > 20 || len(citation) > 600 {
		return nil, nil, httpx.Bad("Choose the award, who gets it (up to 20 people) and add a citation of up to 600 characters.")
	}
	var out *Award
	var req *approvals.Request
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "awards", "Nominate"); err != nil {
			return err
		}
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.people where id = any($1) and status='active'`, people).Scan(&n); err != nil || n != len(dedupe(people)) {
			return httpx.Bad("Everyone nominated must be an active member.")
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		id := app.NewID()
		enc, _ := crypto.Field(dek, awardCtx(p.Org, id), []byte(citation))
		var seq int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'award',11) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org).Scan(&seq); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.awards (org_id, id, ref, type, people, citation_enc, nominated_by, on_date) values ($1,$2,$3,$4,$5,$6,$7,coalesce(nullif($8,'')::date, current_date))`, p.Org, id, fmt.Sprintf("AW-%d", seq), typ, dedupe(people), enc, p.Person, onDate); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "awards", "id=$1", id); err != nil {
			return err
		}
		r, applied, err := s.Ap.Submit(ctx, tx, p, "f7", "award.grant", "Award: "+typ, "", map[string]any{"award": id})
		if err != nil {
			return err
		}
		if !applied {
			req = r
		}
		out, err = s.award(ctx, tx, p.Org, dek, id)
		return err
	})
	return out, req, err
}

func dedupe(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	out := []uuid.UUID{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func (s *Service) award(ctx context.Context, tx pgx.Tx, org uuid.UUID, dek []byte, id uuid.UUID) (*Award, error) {
	a := &Award{}
	var enc []byte
	err := tx.QueryRow(ctx, `select a.id, a.ref, a.type, a.people, a.citation_enc, a.on_date::text, a.status, n.first_name || ' ' || n.last_name,
		coalesce((select array_agg(p.first_name || ' ' || p.last_name order by p.first_name) from hub.people p where p.id = any(a.people)), '{}')
		from hub.awards a join hub.people n on n.id = a.nominated_by where a.id=$1`, id).Scan(&a.ID, &a.Ref, &a.Type, &a.People, &enc, &a.Date, &a.Status, &a.By, &a.Names)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if b, err := crypto.FieldOpen(dek, awardCtx(org, id), enc); err == nil {
		a.Citation = string(b)
	}
	return a, nil
}

func (s *Service) applyAward(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var pl struct {
		Award uuid.UUID `json:"award"`
	}
	if err := json.Unmarshal(a.Payload, &pl); err != nil {
		return err
	}
	var people []uuid.UUID
	var typ string
	if err := tx.QueryRow(ctx, `update hub.awards set status='Approved', `+integrity.Bump+` where id=$1 and status='Pending approval' returning people, type`, pl.Award).Scan(&people, &typ); err != nil {
		return err
	}
	if err := s.A.Sign.Seal(ctx, tx, "awards", "id=$1", pl.Award); err != nil {
		return err
	}
	var nb notify.Batch
	for _, id := range people {
		if err := nb.Add(ctx, tx, a.Org, id, "award", "You received an award: "+typ+".", "#FFD66B", ""); err != nil {
			return err
		}
	}
	go nb.Send(context.Background(), s.A)
	return nil
}

func (s *Service) Awards(ctx context.Context, p *tenant.Principal) ([]*Award, error) {
	out := []*Award{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "awards", "View"); err != nil {
			return err
		}
		approver, err := authz.Can(ctx, tx, p.Person, "awards", "Approve awards")
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select id from hub.awards where (status='Approved' or $2::boolean or nominated_by=$1) order by created_at desc limit 200`, p.Person, approver)
		if err != nil {
			return err
		}
		var ids []uuid.UUID
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				rs.Close()
				return err
			}
			ids = append(ids, id)
		}
		rs.Close()
		for _, id := range ids {
			a, err := s.award(ctx, tx, p.Org, dek, id)
			if err != nil {
				return err
			}
			out = append(out, a)
		}
		return nil
	})
	return out, err
}

func (s *Service) Public(chi.Router) {}

func (s *Service) Private(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	r.Post("/comms", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in SendIn
		if err := httpx.Decode(r, &in, 600<<10); err != nil {
			return err
		}
		c, req, err := s.Send(r.Context(), p, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"comm": c, "approval": req})
		return nil
	}))
	r.Get("/comms", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := s.List(r.Context(), p, r.URL.Query().Get("type"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"comms": l})
		return nil
	}))
	r.Post("/comms/{id}/read", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		if err := s.MarkRead(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/comms/rules", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		m, err := s.Rules(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"rules": m})
		return nil
	}))
	r.Put("/comms/rules/{level}/{type}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in Rule
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if err := s.SetRule(r.Context(), p, chi.URLParam(r, "level"), chi.URLParam(r, "type"), in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/awards", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in struct {
			Type     string      `json:"type"`
			People   []uuid.UUID `json:"people"`
			Citation string      `json:"citation"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		a, req, err := s.Nominate(r.Context(), p, in.Type, in.People, in.Citation)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"award": a, "approval": req})
		return nil
	}))
	r.Get("/awards", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := s.Awards(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"awards": l})
		return nil
	}))
}
