// Package notify writes the bell notifications and pushes them to people's open devices.
package notify

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/mail"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Note struct {
	ID        uuid.UUID  `json:"id"`
	Kind      string     `json:"kind"`
	Text      string     `json:"text"`
	Color     string     `json:"color"`
	Ref       string     `json:"ref"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

type Pending struct {
	acct uuid.UUID
	org  uuid.UUID
	note Note
}

type queuedMail struct {
	org, person uuid.UUID
	subject     string
	card        mail.Card
}

type Batch struct {
	items []Pending
	mails []queuedMail
}

// Mail queues an email to one person; it goes out when Send is called (after the caller's transaction commits). The card's
// action URL may be a path such as "/tickets": the address of the site is put in front of it. A person with no email, or an
// email server that is down, never stops the action that caused the message.
func (b *Batch) Mail(org, person uuid.UUID, subject string, c mail.Card) {
	b.mails = append(b.mails, queuedMail{org, person, subject, c})
}

// Add records a notification inside the caller's transaction and queues its push; call Send after the transaction commits.
func (b *Batch) Add(ctx context.Context, tx pgx.Tx, org, person uuid.UUID, kind, text, color, ref string) error {
	n := Note{ID: app.NewID(), Kind: kind, Text: text, Color: color, Ref: ref, CreatedAt: app.Now()}
	var acct *uuid.UUID
	if err := tx.QueryRow(ctx, `with ins as (insert into hub.notifications (org_id, id, person_id, kind, text, color, ref) values ($1,$2,$3,$4,$5,$6,$7))
		select account_id from hub.people where id=$3`, org, n.ID, person, kind, text, color, ref).Scan(&acct); err != nil {
		return err
	}
	if acct != nil {
		b.items = append(b.items, Pending{*acct, org, n})
	}
	return nil
}

func (b *Batch) Send(ctx context.Context, a *app.App) {
	for _, p := range b.items {
		d, _ := json.Marshal(p.note)
		a.Publish(ctx, realtime.Event{Type: "notification", Channel: realtime.AcctChannel(p.acct), Org: p.org.String(), Data: d})
	}
	if len(b.mails) > 0 {
		go sendMails(a, b.mails)
	}
}

func sendMails(a *app.App, list []queuedMail) {
	if a.Mail == nil {
		return
	}
	ctx := context.Background()
	work := make(chan queuedMail)
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range work {
				if err := deliver(ctx, a, m); err != nil {
					slog.Warn("email not sent", "subject", m.subject, "err", err)
				}
			}
		}()
	}
	for _, m := range list {
		work <- m
	}
	close(work)
	wg.Wait()
}

// deliver looks the person up (name, address, organization) and sends the card.
func deliver(ctx context.Context, a *app.App, m queuedMail) error {
	var first, orgName string
	var acct uuid.UUID
	var enc []byte
	err := a.DB.WithOrg(ctx, m.org, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select p.first_name, o.name, ac.id, ac.email_enc from hub.people p join hub.orgs o on o.id = p.org_id join hub.accounts ac on ac.id = p.account_id where p.id=$1 and p.status='active'`, m.person).Scan(&first, &orgName, &acct, &enc)
	})
	if err != nil {
		return err
	}
	addr, err := crypto.FieldOpen(a.PlatformKey(), crypto.Context{Org: "platform", Table: "accounts", Column: "email_enc", RowID: acct.String()}, enc)
	if err != nil {
		return err
	}
	c := m.card
	c.Org = orgName
	if c.Greeting == "" {
		c.Greeting = "Hello " + first + ","
	}
	if u := c.Action[1]; strings.HasPrefix(u, "/") && len(a.Cfg.AllowedOrigins) > 0 {
		c.Action[1] = strings.TrimRight(a.Cfg.AllowedOrigins[0], "/") + u
	}
	return mail.SendCard(ctx, a.Mail, string(addr), m.subject, c)
}

type Module struct{ A *app.App }

func New(a *app.App) *Module { return &Module{A: a} }

func (m *Module) Public(chi.Router) {}

func (m *Module) Private(r chi.Router) {
	r.Get("/notifications", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if n <= 0 || n > 100 {
			n = 30
		}
		out := []Note{}
		unread := 0
		err = m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			rs, err := tx.Query(r.Context(), `select id, kind, text, color, ref, created_at, read_at from hub.notifications where person_id=$1 order by created_at desc, id desc limit `+strconv.Itoa(n), p.Person)
			if err != nil {
				return err
			}
			for rs.Next() {
				var x Note
				if err := rs.Scan(&x.ID, &x.Kind, &x.Text, &x.Color, &x.Ref, &x.CreatedAt, &x.ReadAt); err != nil {
					rs.Close()
					return err
				}
				out = append(out, x)
			}
			rs.Close()
			return tx.QueryRow(r.Context(), `select count(*) from hub.notifications where person_id=$1 and read_at is null`, p.Person).Scan(&unread)
		})
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"notifications": out, "unread": unread})
		return nil
	}))
	// Someone tagged people in a ticket comment, a wall post or a forum message. The message itself is end-to-end encrypted, so
	// the server is told only who was tagged and where; it sends each of them a bell notification and an email.
	r.Post("/mentions", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in struct {
			Where  string      `json:"where"` // ticket | wall | forum | project
			Ref    string      `json:"ref"`   // ticket number, forum or project name
			People []uuid.UUID `json:"people"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		place, path := map[string][2]string{"ticket": {"ticket ", "/tickets"}, "wall": {"a post on the community wall", "/wall"}, "forum": {"a forum message in ", "/"}, "project": {"a task in ", "/projects"}}[in.Where][0], map[string][2]string{"ticket": {"", "/tickets"}, "wall": {"", "/wall"}, "forum": {"", "/"}, "project": {"", "/projects"}}[in.Where][1]
		if place == "" || len(in.People) == 0 || len(in.People) > 20 || len(in.Ref) > 80 {
			return httpx.Bad("Say where people were tagged and who (up to 20).")
		}
		where := place + in.Ref
		var nb Batch
		err = m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			seen := map[uuid.UUID]bool{}
			for _, id := range in.People {
				if id == p.Person || seen[id] {
					continue
				}
				seen[id] = true
				var ok bool
				if err := tx.QueryRow(r.Context(), `select exists (select 1 from hub.people where id=$1 and status='active')`, id).Scan(&ok); err != nil || !ok {
					continue
				}
				if err := nb.Add(r.Context(), tx, p.Org, id, "mention", p.Name+" tagged you in "+where+".", "#C9B6FF", in.Ref); err != nil {
					return err
				}
				nb.Mail(p.Org, id, p.Name+" tagged you", mail.Card{Accent: mail.Violet, Title: "You were tagged", Preheader: p.Name + " tagged you in " + where,
					Lead: p.Name + " tagged you in " + where + ". Open Hub to see what they said.", Rows: [][2]string{{"Tagged by", p.Name}, {"Where", where}}, Action: [2]string{"Open in Hub", path}})
			}
			return nil
		})
		if err != nil {
			return err
		}
		nb.Send(r.Context(), m.A)
		w.WriteHeader(204)
		return nil
	}))
	r.Delete("/notifications", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		return m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			_, err := tx.Exec(r.Context(), `delete from hub.notifications where person_id=$1`, p.Person)
			if err == nil {
				w.WriteHeader(204)
			}
			return err
		})
	}))
	r.Post("/notifications/read", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		return m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			_, err := tx.Exec(r.Context(), `update hub.notifications set read_at=now() where person_id=$1 and read_at is null`, p.Person)
			if err == nil {
				w.WriteHeader(204)
			}
			return err
		})
	}))
}
