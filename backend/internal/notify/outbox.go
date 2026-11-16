package notify

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/mail"
)

// QueueAudience puts one email for every recipient into the outbox, in one statement: recipients is a query that returns
// person ids (with its own arguments after the job id, as $2, $3 ...). Nothing is held in memory and nobody is capped: the worker
// sends them in small batches (DrainOutbox).
func QueueAudience(ctx context.Context, tx pgx.Tx, org uuid.UUID, subject string, card mail.Card, recipients string, args ...any) error {
	b, err := json.Marshal(card)
	if err != nil {
		return err
	}
	var job uuid.UUID
	if err := tx.QueryRow(ctx, `insert into hub.mail_jobs (org_id, subject, card) values ($1,$2,$3) returning id`, org, subject, b).Scan(&job); err != nil {
		return err
	}
	all := append([]any{job, org}, args...)
	_, err = tx.Exec(ctx, `insert into hub.mail_outbox (job_id, org_id, person_id) select $1, $2, r.id from (`+recipients+`) r on conflict do nothing`, all...)
	return err
}

// DrainOutbox sends what is due, at most batch emails per organization per call, and retries failures with a growing wait
// (giving up after 5 tries). It returns how many were sent.
func DrainOutbox(ctx context.Context, a *app.App, batch int) int {
	if a.Mail == nil {
		return 0
	}
	var orgs []uuid.UUID
	_ = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select distinct org_id from hub.mail_outbox where next_at <= now() limit 100`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var o uuid.UUID
			if rs.Scan(&o) == nil {
				orgs = append(orgs, o)
			}
		}
		return rs.Err()
	})
	sent := 0
	for _, org := range orgs {
		type item struct {
			job, person uuid.UUID
			attempts    int
			subject     string
			card        mail.Card
			first, name string
			acct        uuid.UUID
			enc         []byte
		}
		var items []item
		err := a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
			// claim a batch: taken out of the outbox now, put back below if sending fails
			rs, err := tx.Query(ctx, `delete from hub.mail_outbox o using (
					select job_id, person_id from hub.mail_outbox where org_id=$1 and next_at <= now() order by next_at limit $2 for update skip locked) x
				where o.job_id = x.job_id and o.person_id = x.person_id returning o.job_id, o.person_id, o.attempts`, org, batch)
			if err != nil {
				return err
			}
			var ids []uuid.UUID
			jobs := map[uuid.UUID]bool{}
			for rs.Next() {
				var it item
				if err := rs.Scan(&it.job, &it.person, &it.attempts); err != nil {
					rs.Close()
					return err
				}
				items = append(items, it)
				ids = append(ids, it.person)
				jobs[it.job] = true
			}
			rs.Close()
			if len(items) == 0 {
				return nil
			}
			cards := map[uuid.UUID]item{}
			var jl []uuid.UUID
			for j := range jobs {
				jl = append(jl, j)
			}
			jr, err := tx.Query(ctx, `select id, subject, card from hub.mail_jobs where id = any($1)`, jl)
			if err != nil {
				return err
			}
			for jr.Next() {
				var id uuid.UUID
				var it item
				var b []byte
				if err := jr.Scan(&id, &it.subject, &b); err != nil {
					jr.Close()
					return err
				}
				_ = json.Unmarshal(b, &it.card)
				cards[id] = it
			}
			jr.Close()
			who := map[uuid.UUID]item{}
			pr, err := tx.Query(ctx, `select p.id, p.first_name, o.name, ac.id, ac.email_enc from hub.people p join hub.orgs o on o.id = p.org_id
				join hub.accounts ac on ac.id = p.account_id where p.id = any($1) and p.status='active'`, ids)
			if err != nil {
				return err
			}
			for pr.Next() {
				var id uuid.UUID
				var it item
				if err := pr.Scan(&id, &it.first, &it.name, &it.acct, &it.enc); err != nil {
					pr.Close()
					return err
				}
				who[id] = it
			}
			pr.Close()
			for i := range items {
				c, w := cards[items[i].job], who[items[i].person]
				items[i].subject, items[i].card = c.subject, c.card
				items[i].first, items[i].name, items[i].acct, items[i].enc = w.first, w.name, w.acct, w.enc
			}
			// jobs with nobody left to send to are finished
			_, err = tx.Exec(ctx, `delete from hub.mail_jobs j where j.id = any($1) and not exists (select 1 from hub.mail_outbox o where o.job_id = j.id)`, jl)
			return err
		})
		if err != nil {
			slog.Warn("mail outbox", "err", err)
			continue
		}
		var retry []item
		for _, it := range items {
			if it.enc == nil {
				continue // no active account any more: nothing to send
			}
			if err := deliver(ctx, a, queuedMail{org: org, person: it.person, subject: it.subject, card: it.card}, it.first, it.name, it.acct, it.enc); err != nil {
				if it.attempts < 4 {
					retry = append(retry, it)
				}
				continue
			}
			sent++
		}
		if len(retry) > 0 {
			_ = a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
				for _, it := range retry {
					wait := time.Duration(1<<it.attempts) * time.Minute
					// the job may have been tidied away when its last row was claimed: bring it back first
					if _, err := tx.Exec(ctx, `insert into hub.mail_jobs (id, org_id, subject, card) select $1, $2, $3, $4 where not exists (select 1 from hub.mail_jobs where id=$1)`,
						it.job, org, it.subject, mustJSON(it.card)); err != nil {
						return err
					}
					if _, err := tx.Exec(ctx, `insert into hub.mail_outbox (job_id, org_id, person_id, attempts, next_at) values ($1,$2,$3,$4, now() + make_interval(mins => $5)) on conflict do nothing`,
						it.job, org, it.person, it.attempts+1, int(wait/time.Minute)); err != nil {
						return err
					}
				}
				return nil
			})
		}
	}
	return sent
}

func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
