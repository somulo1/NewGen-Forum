package comms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/notify"
	"hub/internal/orgsettings"
	"hub/internal/tenant"
)

func (s *Service) Rules(ctx context.Context, p *tenant.Principal) (map[string]map[string]Rule, error) {
	out := map[string]map[string]Rule{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select level, type, config from hub.comm_rules`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var l, t string
			var raw []byte
			if err := rs.Scan(&l, &t, &raw); err != nil {
				return err
			}
			var r Rule
			_ = json.Unmarshal(raw, &r)
			if out[l] == nil {
				out[l] = map[string]Rule{}
			}
			out[l][t] = r
		}
		return rs.Err()
	})
	return out, err
}

func (s *Service) SetRule(ctx context.Context, p *tenant.Principal, level, typ string, r Rule) error {
	if (level != "departments" && level != "branches" && level != "sectors" && level != "units") || (typ != "Memo" && typ != "Announcement") {
		return httpx.Bad("Unknown rule.")
	}
	if len(r.To) > 8 {
		return httpx.Bad("Too many audiences.")
	}
	raw, _ := json.Marshal(r)
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Configure approvals"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.comm_rules (org_id, level, type, config) values ($1,$2,$3,$4) on conflict (org_id, level, type) do update set config=excluded.config, `+integrity.BumpOn("comm_rules"), p.Org, level, typ, raw); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "comm_rules", "level=$1 and type=$2", level, typ); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Changed the %s rule for %s", typ, level), "Settings", p.IP, r)
	})
}

// ---- awards ----

type Award struct {
	ID       uuid.UUID   `json:"id"`
	Ref      string      `json:"ref"`
	Type     string      `json:"type"`
	People   []uuid.UUID `json:"people"`
	Names    []string    `json:"names"`
	Citation string      `json:"citation"`
	Date     string      `json:"date"`
	Status   string      `json:"status"`
	By       string      `json:"by"`
	ByPerson uuid.UUID   `json:"by_person"`
	// when it was approved, and whether that was in the current weekly cycle (the Home board shows only those)
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
	InCycle    bool       `json:"in_cycle"`
	// where it still has to be announced ("wall" or a forum's conversation id); only the nominator sees these
	Targets []string `json:"targets,omitempty"`
}

func awardCtx(org, id uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "awards", Column: "citation_enc", RowID: id.String()}
}

func (s *Service) Nominate(ctx context.Context, p *tenant.Principal, typ string, people []uuid.UUID, citation string, targets []string) (*Award, *approvals.Request, error) {
	return s.NominateOn(ctx, p, typ, people, citation, "", targets)
}

// NominateOn is Nominate with a given award date (history loading); an empty date means today.
func (s *Service) NominateOn(ctx context.Context, p *tenant.Principal, typ string, people []uuid.UUID, citation, onDate string, targets []string) (*Award, *approvals.Request, error) {
	if len(targets) > 20 {
		return nil, nil, httpx.Bad("Choose up to 20 places to announce it.")
	}
	if typ == "" || len(typ) > 60 || len(people) == 0 || len(people) > 20 || len(citation) > 600 {
		return nil, nil, httpx.Bad("Choose the award, who gets it (up to 20 people) and add a citation of up to 600 characters.")
	}
	var out *Award
	var req *approvals.Request
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "awards", "Nominate"); err != nil {
			return err
		}
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.people where id = any($1) and status='active'`, people).Scan(&n); err != nil || n != len(dedupe(people)) {
			return httpx.Bad("Everyone nominated must be an active member.")
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		id := app.NewID()
		enc, _ := crypto.Field(dek, awardCtx(p.Org, id), []byte(citation))
		var seq int
		if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'award',11) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org).Scan(&seq); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.awards (org_id, id, ref, type, people, citation_enc, nominated_by, on_date) values ($1,$2,$3,$4,$5,$6,$7,coalesce(nullif($8,'')::date, current_date))`, p.Org, id, fmt.Sprintf("AW-%d", seq), typ, dedupe(people), enc, p.Person, onDate); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "awards", "id=$1", id); err != nil {
			return err
		}
		for _, t := range targets {
			if t != "wall" {
				if _, err := uuid.Parse(t); err != nil {
					return httpx.Bad("That place to announce it is not valid.")
				}
			}
			if _, err := tx.Exec(ctx, `insert into hub.award_posts (org_id, award_id, target) values ($1,$2,$3) on conflict do nothing`, p.Org, id, t); err != nil {
				return err
			}
		}
		if len(targets) > 0 {
			if err := s.A.Sign.Seal(ctx, tx, "award_posts", "award_id=$1", id); err != nil {
				return err
			}
		}
		r, applied, err := s.Ap.Submit(ctx, tx, p, "f7", "award.grant", "Award: "+typ, "", map[string]any{"award": id})
		if err != nil {
			return err
		}
		if !applied {
			req = r
		}
		out, err = s.award(ctx, tx, p.Org, dek, id)
		return err
	})
	return out, req, err
}

func dedupe(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	out := []uuid.UUID{}
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func (s *Service) award(ctx context.Context, tx pgx.Tx, org uuid.UUID, dek []byte, id uuid.UUID) (*Award, error) {
	a := &Award{}
	var enc []byte
	err := tx.QueryRow(ctx, `select a.id, a.ref, a.type, a.people, a.citation_enc, a.on_date::text, a.status, n.first_name || ' ' || n.last_name, a.nominated_by,
		coalesce((select array_agg(p.first_name || ' ' || p.last_name order by p.first_name) from hub.people p where p.id = any(a.people)), '{}')
		from hub.awards a join hub.people n on n.id = a.nominated_by where a.id=$1`, id).Scan(&a.ID, &a.Ref, &a.Type, &a.People, &enc, &a.Date, &a.Status, &a.By, &a.ByPerson, &a.Names)
	if err == nil && a.Status == "Approved" {
		// an award that needed no approval was approved when it was given
		err = tx.QueryRow(ctx, `select coalesce((select approved_at from hub.award_approvals where award_id=$1), (select created_at from hub.awards where id=$1))`, id).Scan(&a.ApprovedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if b, err := crypto.FieldOpen(dek, awardCtx(org, id), enc); err == nil {
		a.Citation = string(b)
	}
	return a, nil
}

// ClaimPost marks one place of an award as announced, once: it succeeds for the first caller and answers 409 after. Only the
// nominator claims, and only after the award is approved.
func (s *Service) ClaimPost(ctx context.Context, p *tenant.Principal, id uuid.UUID, target string) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var by uuid.UUID
		var status string
		if err := tx.QueryRow(ctx, `select nominated_by, status from hub.awards where id=$1`, id).Scan(&by, &status); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if by != p.Person {
			return httpx.ErrForbidden
		}
		if status != "Approved" {
			return httpx.Err(409, "not_approved", "The award has not been approved yet.")
		}
		tag, err := tx.Exec(ctx, `update hub.award_posts set claimed=true, `+integrity.Bump+` where award_id=$1 and target=$2 and not claimed`, id, target)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.Err(409, "already_posted", "That is already announced.")
		}
		return s.A.Sign.Seal(ctx, tx, "award_posts", "award_id=$1 and target=$2", id, target)
	})
}

// unposted lists the places a nominator's award still has to be announced.
func (s *Service) unposted(ctx context.Context, tx pgx.Tx, id uuid.UUID) []string {
	out := []string{}
	rs, err := tx.Query(ctx, `select target from hub.award_posts where award_id=$1 and not claimed order by target`, id)
	if err != nil {
		return out
	}
	defer rs.Close()
	for rs.Next() {
		var t string
		if rs.Scan(&t) == nil {
			out = append(out, t)
		}
	}
	return out
}

func (s *Service) applyAward(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var pl struct {
		Award uuid.UUID `json:"award"`
	}
	if err := json.Unmarshal(a.Payload, &pl); err != nil {
		return err
	}
	var people []uuid.UUID
	var typ string
	if err := tx.QueryRow(ctx, `update hub.awards set status='Approved', `+integrity.Bump+` where id=$1 and status='Pending approval' returning people, type`, pl.Award).Scan(&people, &typ); err != nil {
		return err
	}
	if err := s.A.Sign.Seal(ctx, tx, "awards", "id=$1", pl.Award); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `insert into hub.award_approvals (org_id, award_id) values ($1,$2) on conflict do nothing`, a.Org, pl.Award); err != nil {
		return err
	}
	var nb notify.Batch
	for _, id := range people {
		mailAward(&nb, a.Org, id, typ)
		if err := nb.Add(ctx, tx, a.Org, id, "award", "You received an award: "+typ+".", "#FFD66B", ""); err != nil {
			return err
		}
	}
	go nb.Send(context.Background(), s.A)
	return nil
}

// Cycle is the Performance awards board's weekly cycle: it starts at Saturday 00:00 in the organization's time zone, shows the
// awards approved since, and starts afresh (empty, waiting for the next award) at the next Saturday 00:00.
type Cycle struct {
	Start time.Time `json:"start"`
	Next  time.Time `json:"next"`
}

func cycleAt(now time.Time, loc *time.Location) Cycle {
	n := now.In(loc)
	back := (int(n.Weekday()) - int(time.Saturday) + 7) % 7
	start := time.Date(n.Year(), n.Month(), n.Day()-back, 0, 0, 0, 0, loc)
	return Cycle{Start: start.UTC(), Next: time.Date(start.Year(), start.Month(), start.Day()+7, 0, 0, 0, 0, loc).UTC()}
}

func (s *Service) orgLoc(ctx context.Context, tx pgx.Tx, org uuid.UUID) *time.Location {
	var prof struct {
		Timezone string `json:"timezone"`
	}
	_ = orgsettings.Get(ctx, s.A, tx, org, "profile", &prof)
	if l, err := time.LoadLocation(prof.Timezone); err == nil && prof.Timezone != "" {
		return l
	}
	l, _ := time.LoadLocation("Africa/Nairobi")
	if l == nil {
		l = time.UTC
	}
	return l
}

func (s *Service) Awards(ctx context.Context, p *tenant.Principal) ([]*Award, error) {
	out, _, err := s.AwardsCycle(ctx, p)
	return out, err
}

// AwardsCycle lists the awards and the current cycle.
func (s *Service) AwardsCycle(ctx context.Context, p *tenant.Principal) ([]*Award, Cycle, error) {
	var cyc Cycle
	out, err := s.awards(ctx, p, &cyc)
	return out, cyc, err
}

func (s *Service) awards(ctx context.Context, p *tenant.Principal, cyc *Cycle) ([]*Award, error) {
	out := []*Award{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "awards", "View"); err != nil {
			return err
		}
		approver, err := authz.Can(ctx, tx, p.Person, "awards", "Approve awards")
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		*cyc = cycleAt(time.Now(), s.orgLoc(ctx, tx, p.Org))
		rs, err := tx.Query(ctx, `select id from hub.awards where (status='Approved' or $2::boolean or nominated_by=$1) order by created_at desc limit 200`, p.Person, approver)
		if err != nil {
			return err
		}
		var ids []uuid.UUID
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				rs.Close()
				return err
			}
			ids = append(ids, id)
		}
		rs.Close()
		// the whole page in two queries (not two or three per award)
		if len(ids) == 0 {
			return nil
		}
		ar, err := tx.Query(ctx, `select a.id, a.ref, a.type, a.people, a.citation_enc, a.on_date::text, a.status, n.first_name || ' ' || n.last_name, a.nominated_by,
			coalesce((select array_agg(p.first_name || ' ' || p.last_name order by p.first_name) from hub.people p where p.id = any(a.people)), '{}'),
			case when a.status = 'Approved' then coalesce((select approved_at from hub.award_approvals x where x.award_id = a.id), a.created_at) end
			from hub.awards a join hub.people n on n.id = a.nominated_by where a.id = any($1)`, ids)
		if err != nil {
			return err
		}
		by := map[uuid.UUID]*Award{}
		for ar.Next() {
			a := &Award{}
			var enc []byte
			if err := ar.Scan(&a.ID, &a.Ref, &a.Type, &a.People, &enc, &a.Date, &a.Status, &a.By, &a.ByPerson, &a.Names, &a.ApprovedAt); err != nil {
				ar.Close()
				return err
			}
			if b, err := crypto.FieldOpen(dek, awardCtx(p.Org, a.ID), enc); err == nil {
				a.Citation = string(b)
			}
			a.InCycle = a.Status == "Approved" && a.ApprovedAt != nil && !a.ApprovedAt.Before(cyc.Start)
			by[a.ID] = a
		}
		ar.Close()
		if err := ar.Err(); err != nil {
			return err
		}
		tr, err := tx.Query(ctx, `select ap.award_id, ap.target from hub.award_posts ap join hub.awards a on a.id = ap.award_id
			where ap.award_id = any($1) and not ap.claimed and a.nominated_by = $2 and a.status = 'Approved' order by ap.target`, ids, p.Person)
		if err != nil {
			return err
		}
		for tr.Next() {
			var id uuid.UUID
			var t string
			if err := tr.Scan(&id, &t); err != nil {
				tr.Close()
				return err
			}
			if a := by[id]; a != nil {
				a.Targets = append(a.Targets, t)
			}
		}
		tr.Close()
		if err := tr.Err(); err != nil {
			return err
		}
		for _, id := range ids {
			if a := by[id]; a != nil {
				if a.ByPerson == p.Person && a.Status == "Approved" && a.Targets == nil {
					a.Targets = []string{}
				}
				out = append(out, a)
			}
		}
		return nil
	})
	return out, err
}

func (s *Service) Public(chi.Router) {}
