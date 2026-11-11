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
	ID           uuid.UUID  `json:"id"`
	Ref          string     `json:"ref"`
	Type         string     `json:"type"`
	Date         string     `json:"date"`
	Dir          string     `json:"dir"` // Received | Sent
	Status       string     `json:"status"`
	OK           bool       `json:"ok"`
	Pending      bool       `json:"pending,omitempty"` // still open and not yet due: not counted either way
	Counterparty string     `json:"cp"`
	Conversation uuid.UUID  `json:"conversation"` // the communication's own thread (it is kept like a ticket)
	Message      uuid.UUID  `json:"message,omitempty"`
	Assignee     string     `json:"assignee"`
	AssigneeID   *uuid.UUID `json:"assignee_id,omitempty"`
	Open         bool       `json:"open"`
	From         string     `json:"from"`
	FromUnit     bool       `json:"from_unit"`
}

type ReqStep struct {
	Level    int    `json:"level"`
	Who      string `json:"who"`
	Decision string `json:"decision"`
	At       string `json:"at"`
}

type ReqRow struct {
	ID           uuid.UUID  `json:"id"`
	Ref          string     `json:"ref"`
	Date         string     `json:"date"`
	Dir          string     `json:"dir"`
	Subject      string     `json:"subject"`
	Status       string     `json:"status"`
	OK           bool       `json:"ok"`
	Pending      bool       `json:"pending,omitempty"` // still open and not yet due: not counted either way
	Counterparty string     `json:"cp"`
	By           string     `json:"by"`
	Assignee     string     `json:"assignee"`
	AssigneeID   *uuid.UUID `json:"assignee_id,omitempty"`
	Open         bool       `json:"open"`      // still with this place, not resolved or passed on
	From         string     `json:"from"`      // the place it was sent on behalf of, or the person who raised it
	FromUnit     bool       `json:"from_unit"` // sent from a place's dashboard, not by a person for themselves
	Level        int        `json:"level"`
	Levels       int        `json:"levels"`
	Approvers    []string   `json:"approvers"`
	History      []ReqStep  `json:"history"`
	Closed       bool       `json:"closed"`
}

type TicketRow struct {
	ID           uuid.UUID  `json:"id"`
	Ref          string     `json:"ref"`
	Date         string     `json:"date"`
	Dir          string     `json:"dir"` // Received | Sent | Escalated
	Category     string     `json:"category"`
	Priority     string     `json:"priority"`
	Step         int        `json:"step"`
	Status       string     `json:"status"`
	OK           bool       `json:"ok"`
	Pending      bool       `json:"pending,omitempty"` // still open and not yet due: not counted either way
	Counterparty string     `json:"cp"`
	Assignee     string     `json:"assignee"`
	AssigneeID   *uuid.UUID `json:"assignee_id,omitempty"`
	Open         bool       `json:"open"`      // still with this place, not resolved or passed on
	From         string     `json:"from"`      // the place it was sent on behalf of, or the person who raised it
	FromUnit     bool       `json:"from_unit"` // sent from a place's dashboard, not by a person for themselves
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

// MemberPerf is how one person in the unit did in the period: the tickets, business requests and communications (the ones places
// send each other from their dashboards, not organization-wide memos) assigned to them, resolved or escalated on within the
// place's turnaround, out of those decided. Only people who were assigned something are listed.
type MemberPerf struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Title    string    `json:"title"`
	Tickets  Part      `json:"tickets"`
	Requests Part      `json:"requests"`
	Comms    Part      `json:"comms"`
	Perf     int       `json:"perf"`
	Measured int       `json:"measured"`
}

type Dashboard struct {
	Unit     Unit            `json:"unit"`
	Target   int             `json:"target"`
	Perf     int             `json:"perf"`
	Measured int             `json:"measured"` // how many items the performance is based on (0: nothing to judge yet)
	People   []MemberPerf    `json:"people"`
	Since    string          `json:"since"` // the start of the period it covers
	Parts    map[string]Part `json:"parts"`
	Members  int             `json:"members"`
	Comms    []CommRow       `json:"comms"`
	Requests []ReqRow        `json:"requests"`
	Tickets  []TicketRow     `json:"tickets"`
	Projects []ProjectRow    `json:"projects"`
}

type selfView struct{}

// MyPerf is one person's own results in each place they belong to (department, unit, branch, sector), whatever their rights.
type MyPerf struct {
	Unit Unit       `json:"unit"`
	Me   MemberPerf `json:"me"`
}

// MyPerformance lets anyone see how they did in the period in each place they are placed in. Only their own figures are returned.
func (s *Service) MyPerformance(ctx context.Context, p *tenant.Principal, period string) ([]MyPerf, error) {
	var places []uuid.UUID
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var a, b, c, d *uuid.UUID
		if err := tx.QueryRow(ctx, `select department_id, unit_id, branch_id, sector_id from hub.people where id=$1`, p.Person).Scan(&a, &b, &c, &d); err != nil {
			return err
		}
		for _, x := range []*uuid.UUID{a, b, c, d} {
			if x != nil {
				places = append(places, *x)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	out := []MyPerf{}
	for _, u := range places {
		dash, err := s.Dashboard(context.WithValue(ctx, selfView{}, true), p, u, 30, period)
		if err != nil {
			return nil, err
		}
		for _, m := range dash.People {
			if m.ID == p.Person {
				out = append(out, MyPerf{Unit: dash.Unit, Me: m})
			}
		}
	}
	return out, nil
}

func ptrID(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
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
	d := &Dashboard{Parts: map[string]Part{}, Comms: []CommRow{}, Requests: []ReqRow{}, Tickets: []TicketRow{}, Projects: []ProjectRow{}, People: []MemberPerf{}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `select id, kind, parent_id, name, coalesce(code,'') from hub.org_units where id=$1`, id).Scan(&d.Unit.ID, &d.Unit.Kind, &d.Unit.Parent, &d.Unit.Name, &d.Unit.Code); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		col := memberCol[d.Unit.Kind]
		if ctx.Value(selfView{}) == nil {
			if err := authz.Require(ctx, tx, p.Person, rightsMod[d.Unit.Kind], "View dashboard"); err != nil {
				return err
			}
		}
		d.Target = DefaultTarget
		_ = tx.QueryRow(ctx, `select target_pct from hub.unit_targets where unit_id=$1`, id).Scan(&d.Target)
		if err := tx.QueryRow(ctx, `select count(*) from hub.people where `+col+`=$1 and status='active'`, id).Scan(&d.Members); err != nil {
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

		// communications here are the ones places send each other from their dashboards (not the organization-wide memos and
		// announcements): they are read with tickets below.

		// ---- tickets and business requests ----
		// Each stay with this unit or a unit under it (a sector's branches, a branch's departments) is judged on its own clock: from
		// when it arrived here to when it was resolved or escalated on (see tickets/hops.go). Its people's tickets to anyone else
		// are listed as sent and not judged here. Business requests are tickets marked as requests.
		rs, err := tx.Query(ctx, `
			with recursive below(id) as (select $1::uuid union select u.id from hub.org_units u join below on u.parent_id = below.id),
			stays as (
			  select h.ticket_id, h.unit_ref, h.received_at, h.left_at, h.outcome, h.on_time, h.due_at, h.handler from hub.ticket_hops h where h.unit_ref in (select id from below)
			  union all
			  -- tickets from before stays were kept: one stay, at the place they are now
			  select t.conversation_id, t.to_ref, t.created_at, case when t.step >= 3 then t.solved_at end, case when t.step >= 3 then 'resolved' end,
			    case when t.step >= 3 then (t.due_at is null or t.solved_at <= t.due_at) end, case when t.step >= 3 then t.due_at end, t.handler
			  from hub.tickets t where t.to_ref in (select id from below) and not exists (select 1 from hub.ticket_hops x where x.ticket_id = t.conversation_id)
			),
			rows as (
			  select true as here, s.ticket_id, s.received_at as at, s.left_at, s.outcome, s.on_time, s.due_at, s.handler from stays s where s.received_at > $2
			  union all
			  select false, t.conversation_id, t.created_at, null, null, null, null, null::uuid from hub.tickets t
			  -- sent by its people for themselves, or by anyone on this place's behalf (from its dashboard)
			  where ((t.raised_by in `+um+` and t.from_ref is null) or t.from_ref in (select id from below)) and t.created_at > $2
			    -- where it first went (the first stay; for older tickets, where they are now) is not this unit
			    and not coalesce(coalesce((select h.unit_ref from hub.ticket_hops h where h.ticket_id = t.conversation_id order by h.received_at limit 1), t.to_ref) in (select id from below), false)
			)
			select x.here, t.conversation_id, t.ref, x.at, t.category, t.priority, t.step, t.due_at, t.solved_at, x.left_at, x.outcome, x.on_time, x.due_at,
			  coalesce(t.kind, 'ticket'), coalesce(rp.first_name || ' ' || rp.last_name, ''), coalesce(rdx.name, ''), t.to_type, coalesce(tu.name, ''),
			  -- who works on it: for a stay that ended, whoever had it then; otherwise whoever has it now
			  case when x.left_at is not null then x.handler else t.handler end, coalesce(hp.first_name || ' ' || hp.last_name, ''),
			  coalesce(fu.name, '')
			from rows x join hub.tickets t on t.conversation_id = x.ticket_id
			left join hub.org_units fu on fu.id = t.from_ref
			left join hub.people hp on hp.id = (case when x.left_at is not null then x.handler else t.handler end)
			join hub.people rp on rp.id = t.raised_by left join hub.org_units rdx on rdx.id = rp.department_id
			left join hub.org_units tu on tu.id = t.to_ref
			order by x.at desc limit 300`, id, since)
		if err != nil {
			return err
		}
		stepNames := []string{"Raised", "Assigned", "In progress", "Resolved", "Closed"}
		now := time.Now()
		for rs.Next() {
			var r TicketRow
			var at time.Time
			var here bool
			var due, solved, left, stayDue *time.Time
			var outcome *string
			var onTime *bool
			var kind string
			var raiser, rdept, toType, toName, fromUnit string
			if err := rs.Scan(&here, &r.ID, &r.Ref, &at, &r.Category, &r.Priority, &r.Step, &due, &solved, &left, &outcome, &onTime, &stayDue, &kind, &raiser, &rdept, &toType, &toName, &r.AssigneeID, &r.Assignee, &fromUnit); err != nil {
				rs.Close()
				return err
			}
			r.Date = at.UTC().Format("2006-01-02")
			// who it is from: the place it was sent on behalf of, or the person who raised it
			r.From, r.FromUnit = raiser, fromUnit != ""
			if r.FromUnit {
				r.From, rdept = fromUnit, fromUnit
			}
			where := toName
			if where == "" {
				where = "Ticket desk"
			}
			switch {
			case !here:
				// raised by its people to someone else: shown, not judged here
				r.Dir, r.Counterparty = "Sent", where
				r.Status = stepNames[min(max(r.Step, 0), 4)]
				if r.Step < 3 && due != nil && due.Before(now) {
					r.Status = "Overdue"
				}
			case outcome != nil && *outcome == "escalated":
				r.Dir, r.Counterparty = "Escalated", where
				r.OK = onTime != nil && *onTime
				r.Status = map[bool]string{true: "Escalated on time", false: "Escalated late"}[r.OK]
			case outcome != nil && *outcome == "resolved":
				r.Dir, r.Counterparty = "Received", rdept
				r.OK = onTime != nil && *onTime
				r.Status = map[bool]string{true: "Solved on time", false: "Solved late"}[r.OK]
			default:
				// still here: judged once it is resolved or escalated, or once its time here runs out
				r.Dir, r.Counterparty = "Received", rdept
				r.Status = stepNames[min(max(r.Step, 0), 4)]
				if due != nil && due.Before(now) {
					r.Status = "Overdue"
				} else {
					r.Pending = true
				}
			}
			_ = left
			_ = stayDue
			_ = solved
			if kind == "communication" {
				d.Comms = append(d.Comms, CommRow{ID: r.ID, Ref: r.Ref, Type: "Communication", Date: r.Date, Dir: r.Dir, Status: r.Status, OK: r.OK, Pending: r.Pending, From: r.From, FromUnit: r.FromUnit,
					Counterparty: r.Counterparty, Conversation: r.ID, Assignee: r.Assignee, AssigneeID: r.AssigneeID, Open: r.Dir != "Sent" && (r.Pending || r.Status == "Overdue")})
				continue
			}
			if kind == "request" {
				// only a department receives business requests; elsewhere they are what its people sent
				if d.Unit.Kind != "department" && r.Dir != "Sent" {
					continue
				}
				d.Requests = append(d.Requests, ReqRow{ID: r.ID, Ref: r.Ref, Date: r.Date, Dir: r.Dir, Subject: r.Category, Status: r.Status, OK: r.OK, Pending: r.Pending,
					Counterparty: r.Counterparty, By: raiser, From: r.From, FromUnit: r.FromUnit, Assignee: r.Assignee, AssigneeID: r.AssigneeID, Open: r.Pending || r.Status == "Overdue", Approvers: []string{}, History: []ReqStep{}, Closed: r.Step >= 3})
				continue
			}
			r.Open = r.Dir != "Sent" && (r.Pending || r.Status == "Overdue")
			d.Tickets = append(d.Tickets, r)
		}
		rs.Close()
		if err := rs.Err(); err != nil {
			return err
		}

		// ---- its people (their results are filled from what they worked on, below) ----
		ps, err := tx.Query(ctx, `select pe.id, pe.first_name || ' ' || pe.last_name, coalesce(pe.title, '')
			from hub.people pe where pe.`+col+` = $1 and pe.status = 'active' order by pe.first_name, pe.last_name`, id)
		if err != nil {
			return err
		}
		for ps.Next() {
			var m MemberPerf
			if err := ps.Scan(&m.ID, &m.Name, &m.Title); err != nil {
				ps.Close()
				return err
			}
			d.People = append(d.People, m)
		}
		ps.Close()
		if err := ps.Err(); err != nil {
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
		if r.Dir != "Sent" && r.Pending {
			waiting++
		} else if r.Dir != "Sent" {
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
	// each person: what they worked on, counted the same way
	byID := map[uuid.UUID]*MemberPerf{}
	for i := range d.People {
		byID[d.People[i].ID] = &d.People[i]
	}
	add := func(part *Part, id *uuid.UUID, dir string, pending, ok bool) {
		if id == nil || dir == "Sent" {
			return
		}
		if pending {
			part.Waiting++
			return
		}
		part.Received++
		if ok {
			part.OK++
		}
	}
	for _, r := range d.Tickets {
		if m := byID[ptrID(r.AssigneeID)]; m != nil {
			add(&m.Tickets, r.AssigneeID, r.Dir, r.Pending, r.OK)
		}
	}
	for _, r := range d.Comms {
		if m := byID[ptrID(r.AssigneeID)]; m != nil {
			add(&m.Comms, r.AssigneeID, r.Dir, r.Pending, r.OK)
		}
	}
	for _, r := range d.Requests {
		if m := byID[ptrID(r.AssigneeID)]; m != nil {
			add(&m.Requests, r.AssigneeID, r.Dir, r.Pending, r.OK)
		}
	}
	for i := range d.People {
		m := &d.People[i]
		for _, part := range []*Part{&m.Tickets, &m.Requests, &m.Comms} {
			part.Pct = pct(part.OK, part.Received)
		}
		n, ok := m.Tickets.Received+m.Requests.Received+m.Comms.Received, m.Tickets.OK+m.Requests.OK+m.Comms.OK
		m.Perf, m.Measured = pct(ok, n), n
	}
	// only people who were given something to work on: without any, there is nothing to measure them by
	assigned := d.People[:0]
	for _, m := range d.People {
		if m.Tickets.Received+m.Tickets.Waiting+m.Requests.Received+m.Requests.Waiting+m.Comms.Received+m.Comms.Waiting > 0 {
			assigned = append(assigned, m)
		}
	}
	d.People = assigned
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
		if len(period) == 0 || period[0] == "" {
			period = []string{"month"} // this month unless asked otherwise
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
