package tickets

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// A ticket's text (subject, description, comments) is read by the server and encrypted at rest with the organization's data
// key; every value is bound to its table, column and row. Only the ticket's participants (the people in its conversation) see it.

type Comment struct {
	ID     uuid.UUID `json:"id"`
	At     time.Time `json:"at"`
	Person uuid.UUID `json:"person"`
	Name   string    `json:"who"`
	Text   string    `json:"text"`
}

const (
	maxTitle   = 200
	maxDesc    = 10000
	maxComment = 5000
)

func textCtx(org, id uuid.UUID, table, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: table, Column: col, RowID: id.String()}
}

func cleanText(v string, max int, what string) (string, error) {
	v = strings.TrimSpace(v)
	if utf8.RuneCountInString(v) > max {
		return "", httpx.Bad(what + " is too long.")
	}
	return v, nil
}

func (s *Service) saveText(ctx context.Context, tx pgx.Tx, org, ticket uuid.UUID, title, desc string) error {
	dek, _, err := s.A.OrgDEK(ctx, tx, org)
	if err != nil {
		return err
	}
	te, err := crypto.Field(dek, textCtx(org, ticket, "ticket_text", "title_enc"), []byte(title))
	if err != nil {
		return err
	}
	var de []byte
	if desc != "" {
		if de, err = crypto.Field(dek, textCtx(org, ticket, "ticket_text", "description_enc"), []byte(desc)); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `insert into hub.ticket_text (org_id, ticket_id, title_enc, description_enc) values ($1,$2,$3,$4)`, org, ticket, te, de); err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "ticket_text", "ticket_id=$1", ticket)
}

func (s *Service) participant(ctx context.Context, tx pgx.Tx, p *tenant.Principal, ticket uuid.UUID) (bool, error) {
	var ok bool
	err := tx.QueryRow(ctx, `select exists (select 1 from hub.conversation_members where conversation_id=$1 and account_id=$2 and left_at is null)`, ticket, p.Account).Scan(&ok)
	return ok, err
}

// fillText adds the subject and description to the tickets given; comments too when withComments.
func (s *Service) fillText(ctx context.Context, tx pgx.Tx, org uuid.UUID, list []*Ticket, withComments bool) error {
	if len(list) == 0 {
		return nil
	}
	dek, _, err := s.A.OrgDEK(ctx, tx, org)
	if err != nil {
		return err
	}
	by := map[uuid.UUID]*Ticket{}
	ids := make([]uuid.UUID, 0, len(list))
	for _, t := range list {
		by[t.ID] = t
		ids = append(ids, t.ID)
	}
	rs, err := tx.Query(ctx, `select ticket_id, title_enc, description_enc from hub.ticket_text where ticket_id = any($1)`, ids)
	if err != nil {
		return err
	}
	for rs.Next() {
		var id uuid.UUID
		var te, de []byte
		if err := rs.Scan(&id, &te, &de); err != nil {
			rs.Close()
			return err
		}
		t := by[id]
		if b, err := crypto.FieldOpen(dek, textCtx(org, id, "ticket_text", "title_enc"), te); err == nil {
			t.Title = string(b)
		}
		if len(de) > 0 {
			if b, err := crypto.FieldOpen(dek, textCtx(org, id, "ticket_text", "description_enc"), de); err == nil {
				t.Description = string(b)
			}
		}
	}
	rs.Close()
	if err := rs.Err(); err != nil || !withComments {
		return err
	}
	cs, err := tx.Query(ctx, `select c.id, c.ticket_id, c.created_at, c.author, coalesce(p.first_name || ' ' || p.last_name,''), c.body_enc
		from hub.ticket_comments c left join hub.people p on p.id = c.author where c.ticket_id = any($1) order by c.created_at, c.id`, ids)
	if err != nil {
		return err
	}
	defer cs.Close()
	for cs.Next() {
		var c Comment
		var tid uuid.UUID
		var be []byte
		if err := cs.Scan(&c.ID, &tid, &c.At, &c.Person, &c.Name, &be); err != nil {
			return err
		}
		b, err := crypto.FieldOpen(dek, textCtx(org, c.ID, "ticket_comments", "body_enc"), be)
		if err != nil {
			continue
		}
		c.Text = string(b)
		by[tid].Comments = append(by[tid].Comments, c)
	}
	return cs.Err()
}

// AddComment posts a comment on a ticket. Only its participants may.
func (s *Service) AddComment(ctx context.Context, p *tenant.Principal, ticket uuid.UUID, text string) (*Comment, error) {
	text, err := cleanText(text, maxComment, "The comment")
	if err != nil {
		return nil, err
	}
	if text == "" {
		return nil, httpx.Bad("Write a comment first.")
	}
	var out *Comment
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.tickets where conversation_id=$1`, ticket).Scan(&n); err != nil {
			return err
		}
		if n == 0 {
			return httpx.ErrNotFound
		}
		if ok, err := s.participant(ctx, tx, p, ticket); err != nil {
			return err
		} else if !ok {
			return httpx.ErrForbidden
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		id := app.NewID()
		be, err := crypto.Field(dek, textCtx(p.Org, id, "ticket_comments", "body_enc"), []byte(text))
		if err != nil {
			return err
		}
		c := &Comment{ID: id, Person: p.Person, Name: p.Name, Text: text}
		if err := tx.QueryRow(ctx, `insert into hub.ticket_comments (org_id, id, ticket_id, author, body_enc) values ($1,$2,$3,$4,$5) returning created_at`, p.Org, id, ticket, p.Person, be).Scan(&c.At); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "ticket_comments", "id=$1", id); err != nil {
			return err
		}
		out = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	if t, err := s.Get(ctx, p, ticket); err == nil {
		s.changed(ctx, p, t)
	}
	return out, nil
}
