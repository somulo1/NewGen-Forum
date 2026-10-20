package security

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// LogDetail is everything known about one audit log entry: the entry, the alert it raised (if any), what else the same
// person and address did around it, and the proof that it is still the entry that was written.
type LogDetail struct {
	Log
	Seq      int64     `json:"seq,omitempty"`
	Sealed   *bool     `json:"sealed,omitempty"` // true: the entry matches its signature in the tamper-proof log
	Hash     string    `json:"hash,omitempty"`
	Prev     string    `json:"prev_hash,omitempty"`
	Alert    *AlertRef `json:"alert,omitempty"`
	Related  []Log     `json:"related"`      // the same person, newest first
	SameIP   int       `json:"same_ip"`      // entries from this address in the last 24 hours
	SameUser int       `json:"same_user"`    // entries by this person in the last 24 hours
	Failed   int       `json:"failed_today"` // failed or blocked entries by this person in the last 24 hours
}

type AlertRef struct {
	Ref      string `json:"ref"`
	Title    string `json:"title"`
	Severity string `json:"severity"`
	Status   string `json:"status"`
	Detail   string `json:"detail"` // what exactly happened, as the alert says it
}

func (m *Module) LogDetail(ctx context.Context, p *tenant.Principal, id uuid.UUID) (*LogDetail, error) {
	d := &LogDetail{Related: []Log{}}
	err := m.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "security", "View logs"); err != nil {
			return err
		}
		var seq *int64
		err := tx.QueryRow(ctx, `select id, ts, actor_name, action, module, ip, location, result, rule, ledger_seq from hub.audit_logs where id=$1`, id).
			Scan(&d.ID, &d.At, &d.User, &d.Action, &d.Module, &d.IP, &d.Location, &d.Result, &d.Rule, &seq)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		if seq != nil {
			d.Seq = *seq
			var ts time.Time
			var actor *uuid.UUID
			var action, subject string
			var ph, prev, hash, sig []byte
			if err := tx.QueryRow(ctx, `select ts, actor, action, subject, payload_hash, prev_hash, hash, sig from hub.ledger where seq=$1`, *seq).Scan(&ts, &actor, &action, &subject, &ph, &prev, &hash, &sig); err == nil {
				ok := m.A.Ledger.CheckEntry(p.Org, *seq, ts, actor, action, subject, ph, prev, hash, sig)
				d.Sealed = &ok
				d.Hash, d.Prev = hex.EncodeToString(hash), hex.EncodeToString(prev)
				var a AlertRef
				if err := tx.QueryRow(ctx, `select ref, title, severity, status, detail from hub.alerts where ledger_seq=$1 order by created_at limit 1`, *seq).Scan(&a.Ref, &a.Title, &a.Severity, &a.Status, &a.Detail); err == nil {
					d.Alert = &a
				}
			}
		}
		if err := tx.QueryRow(ctx, `select (select count(*) from hub.audit_logs where ip=$2::text and $2::text <> '' and ts > $1::timestamptz - interval '24 hours' and ts <= $1::timestamptz),
			(select count(*) from hub.audit_logs where actor_name=$3::text and ts > $1::timestamptz - interval '24 hours' and ts <= $1::timestamptz),
			(select count(*) from hub.audit_logs where actor_name=$3::text and result <> 'Success' and ts > $1::timestamptz - interval '24 hours' and ts <= $1::timestamptz)`, d.At, d.IP, d.User).Scan(&d.SameIP, &d.SameUser, &d.Failed); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select id, ts, actor_name, action, module, ip, location, result, rule from hub.audit_logs where actor_name=$1 and id <> $2 and ts <= $3 order by ts desc, id desc limit 6`, d.User, id, d.At)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var l Log
			if err := rs.Scan(&l.ID, &l.At, &l.User, &l.Action, &l.Module, &l.IP, &l.Location, &l.Result, &l.Rule); err != nil {
				return err
			}
			d.Related = append(d.Related, l)
		}
		return rs.Err()
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}
