package tickets

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/emaildetails"
	"hub/internal/mail"
	"hub/internal/notify"
)

// Details is what an email may say about a ticket's content. The content is end-to-end encrypted; the raiser's app gives the
// server this copy when the organization allows details in emails (see package emaildetails). Files are never included.
type Details struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Files       int    `json:"files"` // how many files are attached (the names are not sent)
}

func (d *Details) clean() *Details {
	if d == nil {
		return nil
	}
	return &Details{Title: emaildetails.Clip(d.Title, 200), Description: emaildetails.Clip(d.Description, 3000), Files: d.Files}
}

// mailTicket queues the email for one ticket event, with everything the person needs to know about the ticket.
func (s *Service) mailTicket(ctx context.Context, tx pgx.Tx, nb *notify.Batch, org, to uuid.UUID, kind string, t *Ticket, d *Details, by, why string) {
	ref := t.Ref
	c := mail.Card{Action: [2]string{"Open ticket " + ref, "/tickets"}}
	var subject string
	switch kind {
	case "assigned":
		subject = ref + " was assigned to you"
		c.Title, c.Preheader = "A ticket was assigned to you", "Ticket "+ref+" needs your attention."
		c.Lead = "A ticket was raised in a category you handle and it has come to you."
	case "handover":
		subject = ref + " was handed over to you"
		c.Title, c.Preheader = "A ticket was handed over to you", "Ticket "+ref+" is now yours to handle."
		c.Lead = by + " handed this ticket over to you."
	case "escalated":
		subject = ref + " was escalated to you"
		c.Title, c.Preheader = "A ticket was escalated to you", "Ticket "+ref+" needs a decision."
		c.Lead = "This ticket needs someone with more authority to take it forward."
	case "resolved":
		subject = ref + " was resolved"
		c.Title, c.Preheader = "Your ticket was resolved", "Ticket "+ref+" is marked as resolved."
		c.Lead = "Good news: the ticket you raised has been resolved. Open it to check the outcome, and tell us if something is still wrong."
		c.Action[0] = "Review ticket " + ref
	case "overdue":
		subject = ref + " is overdue"
		c.Title, c.Preheader = "A ticket is overdue", "Ticket "+ref+" passed its target time."
		c.Lead = "This ticket has run past its target time and has been escalated. Please deal with it now."
	case "near":
		subject = ref + " is nearing its deadline"
		c.Title, c.Preheader = "A ticket is nearly due", "Ticket "+ref+" is close to its target time."
		c.Lead = "This ticket is close to its target time. Please give it priority."
	default:
		return
	}
	have := d != nil && emaildetails.Allowed(ctx, s.A, tx, org)
	row := func(k, v string) {
		if v != "" {
			c.Rows = append(c.Rows, [2]string{k, v})
		}
	}
	row("Ticket", ref)
	if have {
		row("Subject", d.Title)
	}
	row("Category", t.Category)
	row("Priority", t.Priority)
	row("Raised by", t.RaisedByName)
	row("Handler", t.HandlerName)
	row("Status", t.StepName)
	if t.DueAt != nil {
		row("Due", t.DueAt.UTC().Format("2 Jan 2006, 15:04 UTC"))
	}
	if have && d.Files > 0 {
		row("Attachments", map[bool]string{true: "1 file", false: ""}[d.Files == 1]+map[bool]string{true: "", false: itoa(d.Files) + " files"}[d.Files != 1])
	}
	switch {
	case why != "":
		c.Quote = [2]string{"Why it was escalated", why}
		if have && d.Description != "" {
			c.Note = "Ticket description: " + emaildetails.Clip(d.Description, 600)
		}
	case have && d.Description != "":
		c.Quote = [2]string{"Description", d.Description}
	}
	if have && d.Files > 0 {
		c.Note = joinNote(c.Note, "Attached files are not sent by email. Open the ticket in Hub to see them.")
	} else if !have {
		c.Note = joinNote(c.Note, "Open the ticket in Hub to read its details.")
	}
	nb.Mail(org, to, subject, c)
}

func joinNote(a, b string) string {
	if a == "" {
		return b
	}
	return a + "\n\n" + b
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
