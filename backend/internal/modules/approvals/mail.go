package approvals

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/mail"
	"hub/internal/notify"
)

// approversOf lists who may decide a level: the person chosen for it, or, when nobody was chosen, everyone who holds the right
// "Approve level N" in the flow's module (owners included). The requester never decides their own request.
func (e *Engine) approversOf(ctx context.Context, tx pgx.Tx, f *Flow, level int, requester uuid.UUID) []uuid.UUID {
	var chosen uuid.UUID
	if tx.QueryRow(ctx, `select person_id from hub.approval_designees where flow_key=$1 and level=$2`, f.Key, level).Scan(&chosen) == nil {
		if chosen == requester {
			return nil
		}
		return []uuid.UUID{chosen}
	}
	rs, err := tx.Query(ctx, `select p.id from hub.people p
		left join hub.person_roles pr on pr.person_id = p.id left join hub.roles r on r.id = pr.role_id
		where p.status='active' and p.id <> $3 and (coalesce(r.locked, false) or coalesce(
			(select x.allowed from hub.person_rights x where x.person_id = p.id and x.module=$1 and x.right_name=$2),
			(select rr.allowed from hub.role_rights rr where rr.role_id = pr.role_id and rr.module=$1 and rr.right_name=$2), false))`,
		f.Module, access.ApproveRight(level), requester)
	if err != nil {
		return nil
	}
	defer rs.Close()
	var out []uuid.UUID
	for rs.Next() {
		var id uuid.UUID
		if rs.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out
}

// mailWaiting tells the people who must decide the request's current level that it is waiting for them (bell and email).
func (e *Engine) mailWaiting(ctx context.Context, tx pgx.Tx, org uuid.UUID, f *Flow, r *Request) {
	var nb notify.Batch
	for _, id := range e.approversOf(ctx, tx, f, r.Level, r.Requester) {
		c := mail.Card{Accent: mail.Amber, Title: "A request is waiting for your approval", Preheader: r.Ref + ": " + r.Title,
			Lead:   r.By + " sent a request that needs your decision.",
			Rows:   [][2]string{{"Request", r.Ref}, {"Type", f.Action}, {"What", r.Title}, {"Requested by", r.By}, {"Step", fmt.Sprintf("Approval %d of %d", r.Level, r.Levels)}, {"Sent", r.CreatedAt.UTC().Format("2 Jan 2006, 15:04 UTC")}},
			Action: [2]string{"Review and decide", "/approvals"}, Note: "You can approve or reject it from the Approvals page."}
		if r.Note != "" {
			c.Quote = [2]string{"Note from " + r.By, r.Note}
		}
		nb.Mail(org, id, "Approval needed: "+r.Title, c)
		_ = nb.Add(ctx, tx, org, id, "approval", r.Ref+" is waiting for your approval: "+r.Title, "#FFB86B", r.Ref)
	}
	nb.Send(context.Background(), e.A)
}

// mailDecided tells the requester how their request ended.
func (e *Engine) mailDecided(ctx context.Context, tx pgx.Tx, org uuid.UUID, r *Request, by, comment string) {
	var nb notify.Batch
	approved := r.Status == "Approved"
	c := mail.Card{Accent: mail.Green, Title: "Your request was approved", Preheader: r.Ref + ": " + r.Title,
		Lead: "Everyone who had to approve it has done so, and it is now going ahead.",
		Rows: [][2]string{{"Request", r.Ref}, {"What", r.Title}, {"Decided by", by}}, Action: [2]string{"Open request", "/approvals"}}
	subject, text, color := "Approved: "+r.Title, r.Ref+" was approved: "+r.Title, "#B8F36B"
	if !approved {
		c.Accent, c.Title = mail.Red, "Your request was rejected"
		c.Lead = "Your request was not approved. You can read the reason below and send a new request if needed."
		subject, text, color = "Rejected: "+r.Title, r.Ref+" was rejected: "+r.Title, "#FF9B8A"
	}
	if comment != "" {
		c.Quote = [2]string{"Comment from " + by, comment}
	}
	nb.Mail(org, r.Requester, subject, c)
	_ = nb.Add(ctx, tx, org, r.Requester, "approval", text, color, r.Ref)
	nb.Send(context.Background(), e.A)
}
