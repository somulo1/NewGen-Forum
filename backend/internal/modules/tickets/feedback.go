package tickets

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/tenant"
)

// Feedback is how closed tickets were rated, for the people who worked on them (the handler and everyone it was assigned to)
// and for the heads of the place each ticket was with (or of a place above it).
type Feedback struct {
	ID         uuid.UUID  `json:"id"`
	Ref        string     `json:"ref"`
	Title      string     `json:"title"`
	Category   string     `json:"category"`
	Rating     int        `json:"rating"`
	Remark     string     `json:"remark,omitempty"`
	AutoClosed bool       `json:"auto_closed"`
	ClosedAt   *time.Time `json:"closed_at,omitempty"`
	RaisedBy   string     `json:"raised_by_name"`
	Handler    string     `json:"handler_name"`
	Mine       bool       `json:"mine"` // I worked on it (otherwise I see it as a head)
}

type FeedbackPage struct {
	Items   []Feedback `json:"items"`
	Total   int        `json:"total"`
	Average float64    `json:"average"`
	Page    int        `json:"page"`
	Size    int        `json:"size"`
}

// feedbackScope: closed, rated tickets I worked on, or that are with a place I head (or below it).
const feedbackScope = `with recursive headed(id) as (
		select x from hub.people pe cross join unnest(array[pe.sector_id, pe.branch_id, pe.unit_id, pe.department_id]) x
		where pe.id = $1 and x is not null and exists (select 1 from hub.person_roles pr join hub.roles r on r.id = pr.role_id where pr.person_id = $1 and r.scope like 'Each %')
		union select u.id from hub.org_units u join headed h on u.parent_id = h.id)
	`

const feedbackWhere = ` where t.step = 4 and t.rating is not null and (t.handler = $1
		or exists (select 1 from hub.ticket_assignees a where a.ticket_id = t.conversation_id and a.person_id = $1)
		or t.to_ref in (select id from headed))`

// TicketFeedback lists the ratings the caller may see, newest first, a page at a time.
func (s *Service) TicketFeedback(ctx context.Context, p *tenant.Principal, page, size int) (*FeedbackPage, error) {
	if size <= 0 || size > 50 {
		size = 20
	}
	if page < 1 {
		page = 1
	}
	out := &FeedbackPage{Items: []Feedback{}, Page: page, Size: size}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, feedbackScope+`select count(*), coalesce(avg(t.rating), 0)::float8 from hub.tickets t`+feedbackWhere, p.Person).Scan(&out.Total, &out.Average); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, feedbackScope+ticketSelect+feedbackWhere+` order by t.closed_at desc nulls last, t.conversation_id limit $2 offset $3`, p.Person, size, (page-1)*size)
		if err != nil {
			return err
		}
		var list []*Ticket
		for rs.Next() {
			t, err := scanTicket(rs)
			if err != nil {
				rs.Close()
				return err
			}
			list = append(list, t)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}
		if err := s.fillText(ctx, tx, p.Org, list, true); err != nil {
			return err
		}
		for _, t := range list {
			f := Feedback{ID: t.ID, Ref: t.Ref, Title: t.Title, Category: t.Category, AutoClosed: t.AutoClosed, ClosedAt: t.ClosedAt, RaisedBy: t.RaisedByName, Handler: t.HandlerName}
			if t.Rating != nil {
				f.Rating = *t.Rating
			}
			f.Mine = t.Handler != nil && *t.Handler == p.Person
			for _, a := range t.Assignees {
				f.Mine = f.Mine || a == p.Person
			}
			// the remark is the closing comment: the raiser's "Closed: ..." or Hub's note on an automatic close
			for _, c := range t.Comments {
				if strings.HasPrefix(c.Text, "Closed: ") {
					f.Remark = strings.TrimPrefix(c.Text, "Closed: ")
				} else if c.Person == nil {
					f.Remark = c.Text
				}
			}
			out.Items = append(out.Items, f)
		}
		return nil
	})
	return out, err
}
