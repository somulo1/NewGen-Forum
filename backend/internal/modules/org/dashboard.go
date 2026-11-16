package org

import (
	"context"
	"errors"
	"fmt"
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
	At           string     `json:"at"`
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
	At           string     `json:"at"`        // when it reached this place
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
	At           string     `json:"at"`        // when it reached this place
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
