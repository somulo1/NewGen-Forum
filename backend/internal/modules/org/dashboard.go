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
	Counterparty string    `json:"cp"`
}

type ProjectRow struct {
	ID     uuid.UUID `json:"id"`
	Code   string    `json:"code"`
	Status string    `json:"status"`
}

type Part struct {
	Pct      int `json:"pct"`
	Received int `json:"received"`
	OK       int `json:"ok"`
}

type Dashboard struct {
	Unit     Unit            `json:"unit"`
	Target   int             `json:"target"`
	Perf     int             `json:"perf"`
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
func (s *Service) Dashboard(ctx context.Context, p *tenant.Principal, id uuid.UUID, days int) (*Dashboard, error) {
	if days <= 0 || days > 365 {
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

		// ---- approval requests ----
		rs, err = tx.Query(ctx, `
			select r.id, r.seq, r.title_enc, r.level, r.status, r.created_at, r.decided_at, f.levels, array[]::text[],
			  r.requester in `+um+`,
			  exists (select 1 from hub.approval_steps s where s.request_id = r.id and s.approver in `+um+`),
			  exists (select 1 from hub.people pe left join hub.person_roles pr on pr.person_id = pe.id left join hub.roles ro on ro.id = pr.role_id
			     where pe.id in `+um+` and r.status = 'Pending' and pe.status = 'active' and (coalesce(ro.locked,false) or coalesce(
			        (select x.allowed from hub.person_rights x where x.person_id = pe.id and x.module = f.module and x.right_name = 'Approve level ' || r.level),
			        case when pr.role_id is null then null else hub.role_right(pr.role_id, f.module, 'Approve level ' || r.level) end, false))),
			  coalesce(dx.name, ''), rq.first_name || ' ' || rq.last_name
			from hub.approval_requests r join hub.approval_flows f on f.key = r.flow_key
			join hub.people rq on rq.id = r.requester left join hub.org_units dx on dx.id = rq.department_id
			where r.created_at > $2 and r.status <> 'Cancelled'
			  and (r.requester in `+um+` or exists (select 1 from hub.approval_steps s where s.request_id = r.id and s.approver in `+um+`)
			       or exists (select 1 from hub.people pe left join hub.person_roles pr on pr.person_id = pe.id left join hub.roles ro on ro.id = pr.role_id
			            where pe.id in `+um+` and r.status = 'Pending' and pe.status = 'active' and (coalesce(ro.locked,false) or coalesce(
			        (select x.allowed from hub.person_rights x where x.person_id = pe.id and x.module = f.module and x.right_name = 'Approve level ' || r.level),
			        case when pr.role_id is null then null else hub.role_right(pr.role_id, f.module, 'Approve level ' || r.level) end, false))))
			order by r.created_at desc limit 200`, id, since)
		if err != nil {
			return err
		}
		var reqIDs []uuid.UUID
		for rs.Next() {
			var r ReqRow
			var seq int
			var enc []byte
			var created time.Time
			var decided *time.Time
			var sent, decidedHere, waitingHere bool
			var dept, status string
			if err := rs.Scan(&r.ID, &seq, &enc, &r.Level, &status, &created, &decided, &r.Levels, &r.Approvers, &sent, &decidedHere, &waitingHere, &dept, &r.By); err != nil {
				rs.Close()
				return err
			}
			r.Ref, r.Date = fmt.Sprintf("AP-%d", seq), created.UTC().Format("2006-01-02")
			r.Levels = max(1, r.Levels)
			if b, err := crypto.FieldOpen(dek, crypto.Context{Org: p.Org.String(), Table: "approval_requests", Column: "title_enc", RowID: r.ID.String()}, enc); err == nil {
				r.Subject = string(b)
			}
			r.Closed = status != "Pending"
			if sent {
				r.Dir, r.Counterparty = "Sent", "Approvers"
			} else {
				r.Dir, r.Counterparty = "Received", dept
			}
			age := time.Since(created)
			if decided != nil {
				age = decided.Sub(created)
			}
			switch {
			case status == "Pending" && age > ontimeWait:
				r.Status, r.OK = "Overdue", false
			case status == "Pending":
				r.Status, r.OK = "In progress", true
			case age > ontimeWait:
				r.Status, r.OK = "Resolved late", false
			default:
				r.Status, r.OK = "Resolved on time", true
			}
			if status == "Rejected" {
				r.Status = "Rejected"
			}
			r.History = []ReqStep{}
			d.Requests = append(d.Requests, r)
			reqIDs = append(reqIDs, r.ID)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}
		if len(reqIDs) > 0 {
			at := make(map[uuid.UUID]int, len(reqIDs))
			for i, id := range reqIDs {
				at[id] = i
			}
			ss, err := tx.Query(ctx, `select s.request_id, s.level, pe.first_name || ' ' || pe.last_name, s.decision, s.created_at from hub.approval_steps s join hub.people pe on pe.id = s.approver where s.request_id = any($1) order by s.request_id, s.level`, reqIDs)
			if err != nil {
				return err
			}
			for ss.Next() {
				var rid uuid.UUID
				var st ReqStep
				var when time.Time
				if err := ss.Scan(&rid, &st.Level, &st.Who, &st.Decision, &when); err != nil {
					ss.Close()
					return err
				}
				st.At = when.UTC().Format(time.RFC3339)
				if i, ok := at[rid]; ok {
					d.Requests[i].History = append(d.Requests[i].History, st)
				}
			}
			ss.Close()
		}

		// ---- tickets ----
		rs, err = tx.Query(ctx, `
			select t.conversation_id, t.ref, t.created_at, t.category, t.priority, t.step, t.late, t.escalated, t.due_at, t.solved_at,
			  (t.to_type = $2 and t.to_ref = $1), coalesce(rdx.name, ''), t.to_type, coalesce(tu.name, '')
			from hub.tickets t join hub.people rp on rp.id = t.raised_by left join hub.org_units rdx on rdx.id = rp.department_id
			left join hub.org_units tu on tu.id = t.to_ref
			where t.created_at > $3 and ((t.to_type = $2 and t.to_ref = $1) or t.raised_by in `+um+`)
			order by t.created_at desc limit 300`, id, d.Unit.Kind, since)
		if err != nil {
			return err
		}
		stepNames := []string{"Raised", "Assigned", "In progress", "Resolved", "Closed"}
		for rs.Next() {
			var r TicketRow
			var created time.Time
			var late, esc, here bool
			var due, solved *time.Time
			var rdept, toType, toName string
			if err := rs.Scan(&r.ID, &r.Ref, &created, &r.Category, &r.Priority, &r.Step, &late, &esc, &due, &solved, &here, &rdept, &toType, &toName); err != nil {
				rs.Close()
				return err
			}
			r.Date = created.UTC().Format("2006-01-02")
			overdue := late || (r.Step < 3 && due != nil && due.Before(time.Now())) || (r.Step >= 3 && due != nil && solved != nil && solved.After(*due))
			r.OK = !overdue
			r.Status = stepNames[min(max(r.Step, 0), 4)]
			if overdue {
				r.Status = "Overdue"
				if r.Step >= 3 {
					r.Status = "Solved late"
				}
			} else if r.Step >= 3 {
				r.Status = "Solved on time"
			}
			switch {
			case here && esc:
				r.Dir, r.Counterparty = "Escalated", rdept
			case here:
				r.Dir, r.Counterparty = "Received", rdept
			default:
				r.Dir = "Sent"
				if toName != "" {
					r.Counterparty = toName
				} else {
					r.Counterparty = "Ticket desk"
				}
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
	var okAll, nAll int
	count := func(name string, n, ok int) {
		d.Parts[name] = Part{Pct: pct(ok, n), Received: n, OK: ok}
		okAll, nAll = okAll+ok, nAll+n
	}
	n, ok := 0, 0
	for _, r := range d.Comms {
		if r.Dir == "Received" {
			n++
			if r.OK {
				ok++
			}
		}
	}
	count("comms", n, ok)
	n, ok = 0, 0
	for _, r := range d.Requests {
		if r.Dir == "Received" {
			n++
			if r.OK {
				ok++
			}
		}
	}
	count("requests", n, ok)
	n, ok = 0, 0
	for _, r := range d.Tickets {
		if r.Dir != "Sent" {
			n++
			if r.OK {
				ok++
			}
		}
	}
	count("tickets", n, ok)
	d.Perf = pct(okAll, nAll)
	_ = strings.TrimSpace
	return d, nil
}

type UnitPerf struct {
	ID     uuid.UUID       `json:"id"`
	Kind   string          `json:"kind"`
	Perf   int             `json:"perf"`
	Target int             `json:"target"`
	Parts  map[string]Part `json:"parts"`
}

// Performance gives every unit the caller may open a one-line summary (for the org overview cards and widgets). Building
// it reads each unit's records (capped at 200 units), so callers should cache the answer for a minute or so.
func (s *Service) Performance(ctx context.Context, p *tenant.Principal) ([]UnitPerf, error) {
	units, err := s.Structure(ctx, p)
	if err != nil {
		return nil, err
	}
	out := []UnitPerf{}
	for i, u := range units {
		if i >= 200 {
			break
		}
		d, err := s.Dashboard(ctx, p, u.ID, 30)
		var he *httpx.Error
		if errors.As(err, &he) && he.Status == 403 {
			continue
		}
		if err != nil {
			return nil, err
		}
		out = append(out, UnitPerf{ID: u.ID, Kind: u.Kind, Perf: d.Perf, Target: d.Target, Parts: d.Parts})
	}
	return out, nil
}
