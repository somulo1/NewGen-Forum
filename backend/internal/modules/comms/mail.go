package comms

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/emaildetails"
	"hub/internal/mail"
	"hub/internal/notify"
)

// mailAudience emails everyone a communication was published to (not its sender). The content is end-to-end encrypted, so the
// email says what kind of message it is, who sent it and where, and sends people to Hub to read it.
func mailAudience(ctx context.Context, a *app.App, tx pgx.Tx, nb *notify.Batch, org uuid.UUID, d *Details, conv, sender uuid.UUID, typ, ref string) {
	_ = conv
	var by, kind, aud string
	_ = tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people where id=$1`, sender).Scan(&by)
	_ = tx.QueryRow(ctx, `select kind, coalesce(audience, '') from hub.conversations where id=$1`, conv).Scan(&kind, &aud)
	where, path := "a forum you are in", "/wall"
	switch {
	case kind == "wall" && aud == "org":
		where = "the community wall"
	case kind == "wall":
		where = "your " + aud + " wall"
	}
	card := mail.Card{Action: [2]string{"Read it in Hub", "/communications"}, Rows: [][2]string{{"Type", typ}, {"From", by}, {"Posted to", where}, {"Reference", ref}}}
	var subject string
	switch typ {
	case "Announcement":
		card.Accent, card.Title, subject = mail.Green, "New announcement", "New announcement from "+by
		card.Lead = by + " posted an announcement for you."
	case "Memo":
		card.Accent, card.Title, subject = mail.Amber, "New memo", "New memo from "+by
		card.Lead = by + " sent you a memo. Some memos ask you to act by a date, so please read it."
	case "Newsletter":
		card.Accent, card.Title, subject = mail.Violet, "New newsletter", "New newsletter from "+by
		card.Lead = "A new issue of the newsletter was published."
	case "Job post":
		card.Accent, card.Title, subject = mail.Blue, "New job opening", "New job opening posted by "+by
		card.Lead = "A new job opening was posted. Open it to read the role, the closing date and how to apply."
	default:
		return
	}
	card.Preheader = card.Title + " from " + by
	_ = path
	// the text itself, when the organization allows details in emails
	if d != nil && emaildetails.Allowed(ctx, a, tx, org) {
		label := "Title"
		if typ == "Job post" {
			label = "Job title"
		}
		rows := [][2]string{{label, d.Title}, {"Summary", d.Summary}}
		for _, f := range d.Fields {
			rows = append(rows, [2]string{f[0], f[1]})
		}
		card.Rows = append(rows, card.Rows...)
		var keep [][2]string
		for _, r := range card.Rows {
			if r[1] != "" {
				keep = append(keep, r)
			}
		}
		card.Rows = keep
		if d.Body != "" {
			card.Quote = [2]string{map[bool]string{true: "Role description", false: "Message"}[typ == "Job post"], d.Body}
		}
		if len(d.Files) > 0 {
			card.Rows = append(card.Rows, [2]string{"Attachments", strings.Join(d.Files, ", ")})
			card.Note = "Attached files are not sent by email. Open it in Hub to download them."
		}
		card.Preheader = d.Title
		subject = typ + ": " + d.Title
		card.Action[0] = "Read it in Hub"
	}
	rs, err := tx.Query(ctx, `select p.id from hub.conversation_members m join hub.people p on p.account_id = m.account_id
		where m.conversation_id=$1 and m.left_at is null and p.status='active' and p.id <> $2 limit 5000`, conv, sender)
	if err != nil {
		return
	}
	defer rs.Close()
	for rs.Next() {
		var id uuid.UUID
		if rs.Scan(&id) == nil {
			nb.Mail(org, id, subject, card)
		}
	}
}

// mailAward tells a winner by email.
func mailAward(nb *notify.Batch, org, person uuid.UUID, typ string) {
	nb.Mail(org, person, "You received an award: "+typ, mail.Card{Accent: mail.Amber, Title: "You received an award", Preheader: typ,
		Lead: "Congratulations! Your work was recognised.", Rows: [][2]string{{"Award", typ}}, Action: [2]string{"See it on your profile", "/profile"}})
}

// Details is what an email may say about a communication: its text and the names of its attachments (never the files).
type Details struct {
	Title   string      `json:"title"`
	Summary string      `json:"summary"`
	Body    string      `json:"body"`
	Fields  [][2]string `json:"fields"` // label, value (closing date, location, ...)
	Files   []string    `json:"files"`
}

func (d *Details) clean() *Details {
	if d == nil {
		return nil
	}
	out := &Details{Title: emaildetails.Clip(d.Title, 200), Summary: emaildetails.Clip(d.Summary, 300), Body: emaildetails.Clip(d.Body, 6000)}
	for i, f := range d.Fields {
		if i < 12 {
			out.Fields = append(out.Fields, [2]string{emaildetails.Clip(f[0], 60), emaildetails.Clip(f[1], 600)})
		}
	}
	for i, f := range d.Files {
		if i < 20 {
			out.Files = append(out.Files, emaildetails.Clip(f, 120))
		}
	}
	return out
}
