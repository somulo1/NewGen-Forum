package org

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

const DefaultTarget = 95

var (
	memberCol  = map[string]string{"sector": "sector_id", "branch": "branch_id", "unit": "unit_id", "department": "department_id"}
	rightsMod  = map[string]string{"sector": "sectors", "branch": "branches", "unit": "units", "department": "departments"}
	ontimeWait = 48 * time.Hour
)

type CommRow struct {
	ID           uuid.UUID `json:"id"`
	Ref          string    `json:"ref"`
	Type         string    `json:"type"`
	Date         string    `json:"date"`
	Dir          string    `json:"dir"` // Received | Sent
	Status       string    `json:"status"`
	OK           bool      `json:"ok"`
	Pending      bool      `json:"pending,omitempty"` // still open and not yet due: not counted either way
	Counterparty string    `json:"cp"`
	Conversation uuid.UUID `json:"conversation"`
	Message      uuid.UUID `json:"message"` // the subject is sealed: open this message to read it
}

type ReqStep struct {
	Level    int    `json:"level"`
	Who      string `json:"who"`
	Decision string `json:"decision"`
	At       string `json:"at"`
}

type ReqRow struct {
	ID           uuid.UUID `json:"id"`
	Ref          string    `json:"ref"`
	Date         string    `json:"date"`
	Dir          string    `json:"dir"`
	Subject      string    `json:"subject"`
	Status       string    `json:"status"`
	OK           bool      `json:"ok"`
	Pending      bool      `json:"pending,omitempty"` // still open and not yet due: not counted either way
	Counterparty string    `json:"cp"`
	By           string    `json:"by"`
	Level        int       `json:"level"`
	Levels       int       `json:"levels"`
	Approvers    []string  `json:"approvers"`
	History      []ReqStep `json:"history"`
	Closed       bool      `json:"closed"`
}

type TicketRow struct {
	ID           uuid.UUID `json:"id"`
	Ref          string    `json:"ref"`
	Date         string    `json:"date"`
	Dir          string    `json:"dir"` // Received | Sent | Escalated
	Category     string    `json:"category"`
	Priority     string    `json:"priority"`
	Step         int       `json:"step"`
	Status       string    `json:"status"`
	OK           bool      `json:"ok"`
	Pending      bool      `json:"pending,omitempty"` // still open and not yet due: not counted either way
	Counterparty string    `json:"cp"`
}

type ProjectRow struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Status string    `json:"status"`
}

type Part struct {
	Pct      int `json:"pct"`
	Received int `json:"received"` // what was counted: decided items (done, or past their time)
	OK       int `json:"ok"`
	Waiting  int `json:"waiting"` // received but still open and within time
}

type Dashboard struct {
	Unit     Unit            `json:"unit"`
	Target   int             `json:"target"`
	Perf     int             `json:"perf"`
	Measured int             `json:"measured"` // how many items the performance is based on (0: nothing to judge yet)
	Since    string          `json:"since"`    // the start of the period it covers
	Parts    map[string]Part `json:"parts"`
	Members  int             `json:"members"`
	Comms    []CommRow       `json:"comms"`
	Requests []ReqRow        `json:"requests"`
	Tickets  []TicketRow     `json:"tickets"`
	Projects []ProjectRow    `json:"projects"`
}

func pct(ok, n int) int {
	if n == 0 {
		return 100
	}
	return int(float64(ok)/float64(n)*100 + 0.5)
}

// Targets returns every unit's threshold (units with none use the default).
func (s *Service) Targets(ctx context.Context, p *tenant.Principal) (map[uuid.UUID]int, error) {
	out := map[uuid.UUID]int{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select unit_id, target_pct from hub.unit_targets`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			var v int
			if err := rs.Scan(&id, &v); err != nil {
				return err
			}
			out[id] = v
		}
		return rs.Err()
	})
	return out, err
}

func (s *Service) SetTarget(ctx context.Context, p *tenant.Principal, id uuid.UUID, v int) error {
	if v < 50 || v > 100 {
		return httpx.Bad("Choose a threshold between 50 and 100%.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "performance", "Set targets"); err != nil {
			return err
		}
		var name string
		if err := tx.QueryRow(ctx, `select name from hub.org_units where id=$1`, id).Scan(&name); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.unit_targets (org_id, unit_id, target_pct) values ($1,$2,$3) on conflict (org_id, unit_id) do update set target_pct=excluded.target_pct, updated_at=now(), `+integrity.BumpOn("unit_targets"), p.Org, id, v); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "unit_targets", "unit_id=$1", id); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Set the performance threshold for %s to %d%%", name, v), "Settings", p.IP, map[string]any{"unit": id, "target": v})
	})
}

// Dashboard builds a unit's page from real records over the last `days` days: what it received and sent (memos and
// announcements, approval requests, tickets), how much of what it received was handled on time, and its projects.
// Dashboard builds a unit's dashboard from what it received in a period: period "month", "quarter" or "year" is the current calendar
// month, quarter or year; otherwise the last `days` days (30 by default).
func (s *Service) Dashboard(ctx context.Context, p *tenant.Principal, id uuid.UUID, days int, period ...string) (*Dashboard, error) {
	if days <= 0 || days > 3660 {
		days = 30
	}
	d := &Dashboard{Parts: map[string]Part{}, Comms: []CommRow{}, Requests: []ReqRow{}, Tickets: []TicketRow{}, Projects: []ProjectRow{}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `select id, kind, parent_id, name, coalesce(code,'') from hub.org_units where id=$1`, id).Scan(&d.Unit.ID, &d.Unit.Kind, &d.Unit.Parent, &d.Unit.Name, &d.Unit.Code); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		col := memberCol[d.Unit.Kind]
		if err := authz.Require(ctx, tx, p.Person, rightsMod[d.Unit.Kind], "View dashboard"); err != nil {
			return err
		}
		d.Target = DefaultTarget
		_ = tx.QueryRow(ctx, `select target_pct from hub.unit_targets where unit_id=$1`, id).Scan(&d.Target)
		if err := tx.QueryRow(ctx, `select count(*) from hub.people where `+col+`=$1 and status='active'`, id).Scan(&d.Members); err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		um := `(select id from hub.people where ` + col + `=$1 and status <> 'deleted')`
		since := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
		if len(period) > 0 {
			if ps, ok := periodStart(period[0], time.Now()); ok {
				since = ps
			}
		}
		d.Since = since.UTC().Format(time.RFC3339)

		// ---- memos, announcements, newsletters ----
		rs, err := tx.Query(ctx, `
			select c.id, c.ref, c.type, c.published_at, c.conversation_id, c.message_id,
			  c.by_person in `+um+`,
			  (select count(*) from hub.conversation_members m join hub.people pe on pe.account_id = m.account_id and pe.org_id = m.org_id
			     where m.conversation_id = c.conversation_id and m.left_at is null and pe.id in `+um+`),
			  (select count(*) from hub.conversation_members m where m.conversation_id = c.conversation_id and m.left_at is null),
			  (select count(*) from hub.comm_reads r where r.comm_id = c.id and r.person_id in `+um+`),
			  (select count(*) from hub.comm_reads r where r.comm_id = c.id and r.person_id in `+um+` and r.at <= c.published_at + $3::interval),
			  (select count(*) from hub.comm_reads r where r.comm_id = c.id),
			  coalesce(dx.name, ''), conv.name_enc, conv.id
			from hub.comms c
			join hub.conversations conv on conv.id = c.conversation_id
			join hub.people sp on sp.id = c.by_person left join hub.org_units dx on dx.id = sp.department_id
			where c.status = 'Published' and c.published_at > $2
			  and (c.by_person in `+um+` or exists (select 1 from hub.conversation_members m join hub.people pe on pe.account_id = m.account_id and pe.org_id = m.org_id
			        where m.conversation_id = c.conversation_id and m.left_at is null and pe.id in `+um+`))
			order by c.published_at desc limit 200`, id, since, fmt.Sprintf("%d seconds", int(ontimeWait.Seconds())))
		if err != nil {
			return err
		}
		for rs.Next() {
			var r CommRow
			var pub time.Time
			var sent bool
			var inUnit, inConv, reads, onTime, allReads int
			var dept string
			var nameEnc []byte
			var convID uuid.UUID
			if err := rs.Scan(&r.ID, &r.Ref, &r.Type, &pub, &r.Conversation, &r.Message, &sent, &inUnit, &inConv, &reads, &onTime, &allReads, &dept, &nameEnc, &convID); err != nil {
				rs.Close()
				return err
			}
			r.Date = pub.UTC().Format("2006-01-02")
			cp := "Community wall"
			if len(nameEnc) > 0 {
				if b, err := crypto.FieldOpen(dek, crypto.Context{Org: p.Org.String(), Table: "conversations", Column: "name_enc", RowID: convID.String()}, nameEnc); err == nil && len(b) > 0 {
					cp = string(b)
				}
			}
			if sent {
				r.Dir, r.Counterparty = "Sent", cp
				switch {
				case inConv > 0 && allReads >= inConv-1:
					r.Status, r.OK = "Read by all", true
				case allReads > 0:
					r.Status, r.OK = "Partly read", true
				default:
					r.Status, r.OK = "Delivered", true
				}
			} else {
				r.Dir, r.Counterparty = "Received", dept
				// a unit has acknowledged a broadcast when a quarter of its people (at least one) have opened it
				need := max(1, (inUnit+3)/4)
				switch {
				case onTime >= need && inUnit > 0:
					r.Status, r.OK = "Acknowledged on time", true
				case reads >= need && inUnit > 0:
					r.Status = "Acknowledged late"
				case time.Since(pub) < ontimeWait:
					// still within its time to be read: not judged yet
					r.Status, r.Pending = "Waiting to be read", true
				default:
					r.Status = "Not acknowledged"
				}
			}
			d.Comms = append(d.Comms, r)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}

		// ---- business requests: raised from the ticket form to a department (tickets marked as requests, read with them below) ----

		// ---- tickets ----
		// a ticket sent to this unit or to any unit under it (a sector's branches, a branch's departments) is work this unit received
		rs, err = tx.Query(ctx, `
			with recursive below(id) as (select $1::uuid union select u.id from hub.org_units u join below on u.parent_id = below.id)
			select t.conversation_id, t.ref, t.created_at, t.category, t.priority, t.step, t.late, t.escalated, t.due_at, t.solved_at, t.kind = 'request',
			  coalesce(rp.first_name || ' ' || rp.last_name, ''),
			  (select min(e.at) from hub.ticket_events e where e.ticket_id = t.conversation_id and e.kind = 'event' and e.text like 'Escalated%'),
			  coalesce($2::text <> '' and t.to_ref in (select id from below), false), coalesce(t.escalated_from && array(select id from below), false), coalesce(rdx.name, ''), t.to_type, coalesce(tu.name, '')
			from hub.tickets t join hub.people rp on rp.id = t.raised_by left join hub.org_units rdx on rdx.id = rp.department_id
			left join hub.org_units tu on tu.id = t.to_ref
			where t.created_at > $3 and (t.to_ref in (select id from below) or t.raised_by in `+um+` or t.escalated_from && array(select id from below))
			order by t.created_at desc limit 300`, id, d.Unit.Kind, since)
		if err != nil {
			return err
		}
		stepNames := []string{"Raised", "Assigned", "In progress", "Resolved", "Closed"}
		for rs.Next() {
			var r TicketRow
			var created time.Time
			var late, esc, here, passedOn bool
			var due, solved, escAt *time.Time
			var rdept, toType, toName, raiser string
			var isReq *bool
			if err := rs.Scan(&r.ID, &r.Ref, &created, &r.Category, &r.Priority, &r.Step, &late, &esc, &due, &solved, &isReq, &raiser, &escAt, &here, &passedOn, &rdept, &toType, &toName); err != nil {
				rs.Close()
				return err
			}
			r.Date = created.UTC().Format("2006-01-02")
			overdue := late || (r.Step < 3 && due != nil && due.Before(time.Now())) || (r.Step >= 3 && due != nil && solved != nil && solved.After(*due))
			// Performance counts what was received and then resolved in time, or escalated in time (by a person, before its time ran
			// out; the automatic escalation of an overdue ticket does not count). Anything still open counts against it until then.
			solvedOnTime := r.Step >= 3 && solved != nil && (due == nil || !solved.After(*due))
			escOnTime := escAt != nil && (due == nil || !escAt.After(*due))
			r.OK = solvedOnTime || escOnTime
			r.Pending = !r.OK && !overdue && r.Step < 3
			r.Status = stepNames[min(max(r.Step, 0), 4)]
			if overdue {
				r.Status = "Overdue"
				if r.Step >= 3 {
					r.Status = "Solved late"
				}
			} else if r.Step >= 3 {
				r.Status = "Solved on time"
			} else if escOnTime {
				r.Status = "Escalated on time"
			}
			switch {
			// Received: with this unit now (raised by its own people too, and escalated to it). Escalated: this unit passed it on.
			// Sent: its people raised it to someone else.
			case here:
				r.Dir, r.Counterparty = "Received", rdept
			case passedOn:
				r.Dir, r.Counterparty = "Escalated", toName
				if toName == "" {
					r.Counterparty = "Ticket desk"
				}
			default:
				r.Dir = "Sent"
				if toName != "" {
					r.Counterparty = toName
				} else {
					r.Counterparty = "Ticket desk"
				}
			}
			if isReq != nil && *isReq {
				// only a department receives business requests; elsewhere they are what its people sent
				if d.Unit.Kind != "department" && r.Dir != "Sent" {
					continue
				}
				d.Requests = append(d.Requests, ReqRow{ID: r.ID, Ref: r.Ref, Date: r.Date, Dir: r.Dir, Subject: r.Category, Status: r.Status, OK: r.OK, Pending: r.Pending,
					Counterparty: r.Counterparty, By: raiser, Approvers: []string{}, History: []ReqStep{}, Closed: r.Step >= 3})
				continue
			}
			d.Tickets = append(d.Tickets, r)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}

		// ---- projects the unit's people are part of ----
		rs, err = tx.Query(ctx, `select p.conversation_id, p.code, p.status from hub.projects p
			where p.status <> 'Pending approval' and exists (select 1 from hub.conversation_members m join hub.people pe on pe.account_id = m.account_id and pe.org_id = m.org_id
			   where m.conversation_id = p.conversation_id and m.left_at is null and pe.id in `+um+`) order by p.code`, id)
		if err != nil {
			return err
		}
		for rs.Next() {
			var r ProjectRow
			if err := rs.Scan(&r.ID, &r.Code, &r.Status); err != nil {
				rs.Close()
				return err
			}
			d.Projects = append(d.Projects, r)
		}
		rs.Close()
		return rs.Err()
	})
	if err != nil {
		return nil, err
	}
	var okAll, nAll, waiting int
	count := func(name string, n, ok int) {
		d.Parts[name] = Part{Pct: pct(ok, n), Received: n, OK: ok, Waiting: waiting}
		okAll, nAll, waiting = okAll+ok, nAll+n, 0
	}
	n, ok := 0, 0
	for _, r := range d.Comms {
		if r.Dir == "Received" && r.Pending {
			waiting++
		} else if r.Dir == "Received" {
			n++
			if r.OK {
				ok++
			}
		}
	}
	count("comms", n, ok)
	n, ok = 0, 0
	for _, r := range d.Requests {
		if r.Dir != "Sent" && r.Pending {
			waiting++
		} else if r.Dir != "Sent" {
			n++
			if r.OK {
				ok++
			}
		}
	}
	count("requests", n, ok)
	n, ok = 0, 0
	for _, r := range d.Tickets {
		if r.Dir != "Sent" && r.Pending {
			waiting++
		} else if r.Dir != "Sent" {
			n++
			if r.OK {
				ok++
			}
		}
	}
	count("tickets", n, ok)
	d.Perf, d.Measured = pct(okAll, nAll), nAll
	_ = strings.TrimSpace
	return d, nil
}

type UnitPerf struct {
	ID       uuid.UUID       `json:"id"`
	Kind     string          `json:"kind"`
	Perf     int             `json:"perf"`
	Measured int             `json:"measured"`
	Target   int             `json:"target"`
	Parts    map[string]Part `json:"parts"`
}

// Performance gives every unit the caller may open a one-line summary (for the org overview cards and widgets). Building
// it reads each unit's records (capped at 200 units), so callers should cache the answer for a minute or so.
func (s *Service) Performance(ctx context.Context, p *tenant.Principal, period ...string) ([]UnitPerf, error) {
	units, err := s.Structure(ctx, p)
	if err != nil {
		return nil, err
	}
	out := []UnitPerf{}
	for i, u := range units {
		if i >= 200 {
			break
		}
		d, err := s.Dashboard(ctx, p, u.ID, 30, period...)
		var he *httpx.Error
		if errors.As(err, &he) && he.Status == 403 {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, UnitPerf{ID: u.ID, Kind: u.Kind, Perf: d.Perf, Measured: d.Measured, Target: d.Target, Parts: d.Parts})
	}
	return out, nil
}

// periodStart is the first moment of the current calendar month, quarter or year.
func periodStart(period string, now time.Time) (time.Time, bool) {
	y, m, _ := now.Date()
	switch period {
	case "month":
		return time.Date(y, m, 1, 0, 0, 0, 0, now.Location()), true
	case "quarter":
		return time.Date(y, ((m-1)/3)*3+1, 1, 0, 0, 0, 0, now.Location()), true
	case "year":
		return time.Date(y, 1, 1, 0, 0, 0, 0, now.Location()), true
	}
	return time.Time{}, false
}
