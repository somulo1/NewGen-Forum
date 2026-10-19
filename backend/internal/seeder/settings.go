package seeder

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/crypto"
	"hub/internal/integrity"
	"hub/internal/modules/comms"
	"hub/internal/modules/settings"
	"hub/internal/modules/tickets"
	"hub/internal/orgsettings"
)

var fastPW = crypto.PasswordParams{Time: 1, MemoryKiB: 8 * 1024, Threads: 1}

func (s *Seeder) setFlow(ctx context.Context, key string, levels int) {
	s.must(s.Ap.SetFlow(ctx, s.Owner.P, key, levels), "set flow")
}

func (s *Seeder) seedSettings(ctx context.Context) {
	o := s.Owner.P
	// approval flows exactly as the mock defines them (levels and approvers per flow)
	var acc struct {
		Flows []struct {
			ID        string   `json:"id"`
			Mod       string   `json:"mod"`
			Levels    int      `json:"levels"`
			Approvers []string `json:"approvers"`
		}
	}
	b, _ := json.Marshal(map[string]json.RawMessage{"Flows": s.data.state["flows"]})
	s.must(json.Unmarshal(b, &acc), "flows")
	for _, f := range acc.Flows {
		s.must(s.Ap.SetFlow(ctx, o, f.ID, f.Levels), "flow "+f.ID)
		// the mock names who approves each level; the demo gives that level's approval right to the role of that name
		for lvl, who := range f.Approvers {
			if lvl >= f.Levels {
				break
			}
			s.must(s.A.DB.WithOrg(ctx, s.OrgID, func(tx pgx.Tx) error {
				tag, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed)
					select $1, id, $3, $4, true from hub.roles where lower(name)=lower($2) and not locked
					on conflict (org_id, role_id, module, right_name) do update set allowed=true, `+integrity.BumpOn("role_rights"), s.OrgID, who, f.Mod, access.ApproveRight(lvl+1))
				if err != nil || tag.RowsAffected() == 0 {
					return err
				}
				return s.A.Sign.Seal(ctx, tx, "role_rights", "module=$1 and right_name=$2", f.Mod, access.ApproveRight(lvl+1))
			}), "approver right")
		}
	}

	var prof settings.Profile
	s.st("orgProfile", &prof)
	prof.Name = "Demo Organization"
	prof.Short = "Demo"
	s.must(s.Set.SetProfile(ctx, o, prof), "profile")
	var sec orgsettings.Security
	s.st("orgSecurity", &sec)
	s.must(s.Set.SetSecurity(ctx, o, sec), "security")
	var sub orgsettings.Subscription
	s.st("orgSub", &sub)
	sub.Seats = 300
	_, err := s.Set.SetSubscription(ctx, o, sub)
	s.must(err, "subscription")
	var bill settings.Billing
	s.st("orgBilling", &bill)
	s.must(s.Set.SetBilling(ctx, o, bill), "billing")
	var invs []struct {
		No, Date, Status string
		Amount           int64
	}
	s.st("invoices", &invs)
	for i := len(invs) - 1; i >= 0; i-- { // oldest first so invoice numbers run in order
		_, err := s.Set.IssueInvoiceOn(ctx, s.OrgID, invs[i].Amount, prof.Currency, invs[i].Status, invs[i].Date)
		s.must(err, "invoice")
	}

	// ticket timing policy and the default categories with the mock's turn-around times
	var ts struct {
		Cats []struct {
			Name string
			TAT  int
			Unit string
		}
		Start, End string
		Pause      struct{ Weekends, Holidays, Offhours, Onhold, Awaiting bool }
		Notify     struct{ Solved, Overdue, Nearing, Assigned bool }
		NearPct    int
		Escalate   string
	}
	s.st("tset", &ts)
	cfg := tickets.Config{Start: ts.Start, End: ts.End, Timezone: prof.Timezone, NearPct: ts.NearPct, Escalate: ts.Escalate, Holidays: []string{"2026-12-25", "2026-12-12", "2027-01-01"}}
	cfg.Pause.Weekends, cfg.Pause.Holidays, cfg.Pause.Offhours, cfg.Pause.Onhold, cfg.Pause.Awaiting = ts.Pause.Weekends, ts.Pause.Holidays, ts.Pause.Offhours, ts.Pause.Onhold, ts.Pause.Awaiting
	cfg.Notify.Solved, cfg.Notify.Overdue, cfg.Notify.Nearing, cfg.Notify.Assigned = ts.Notify.Solved, ts.Notify.Overdue, ts.Notify.Nearing, ts.Notify.Assigned
	s.must(s.Tick.SetSettings(ctx, o, cfg), "ticket settings")
	existing, err := s.Tick.Categories(ctx, o, "desk", nil)
	s.must(err, "categories")
	for _, c := range ts.Cats {
		hours := c.TAT
		if c.Unit == "days" {
			hours = c.TAT * 24
		}
		found := false
		for _, e := range existing {
			if e.Name == c.Name {
				e.TAT = hours
				_, err := s.Tick.SaveCategory(ctx, o, &e.ID, e)
				s.must(err, "category")
				found = true
			}
		}
		if !found {
			_, err := s.Tick.SaveCategory(ctx, o, nil, tickets.Category{Name: c.Name, TAT: hours, UnitType: "desk"})
			s.must(err, "category")
		}
	}

	// communication rules per level
	var rules map[string]map[string]comms.Rule
	s.st("commRules", &rules)
	for level, byType := range rules {
		for typ, r := range byType {
			s.must(s.Com.SetRule(ctx, o, level, typ, r), fmt.Sprintf("rule %s %s", level, typ))
		}
	}
	_ = strings.TrimSpace
}
