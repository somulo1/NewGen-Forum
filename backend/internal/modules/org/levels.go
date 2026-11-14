package org

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/orgsettings"
	"hub/internal/tenant"
)

// The structure levels an organization uses. Every organization decides for itself which of the four kinds of part it has
// (sector, branch, unit, department), what it calls each one (a "branch" may be a "region", a "unit" a "team") and which sits under
// which (a unit under a branch, a branch under a sector, a sector under the organization itself). Records always keep the fixed
// kind, so renaming a level never touches them; only the label shown changes.

var allKinds = []string{"sector", "branch", "unit", "department"}

type Level struct {
	Kind    string `json:"kind"`
	Enabled bool   `json:"enabled"`
	Label   string `json:"label"`
	Plural  string `json:"plural"`
	Parent  string `json:"parent"` // the kind it sits under, or "" for the organization itself
	// HeadRole: the role (its key) that heads every place of this level. Whoever holds it and is placed in a place heads that
	// place. Decided here for the whole level, and given to places added later too.
	HeadRole string `json:"head_role,omitempty"`
}

type Levels struct {
	Levels []Level `json:"levels"`
	Set    bool    `json:"set"` // false: the organization never configured its levels (these are worked out from what it has)
}

var defaultLabel = map[string][2]string{"sector": {"Sector", "Sectors"}, "branch": {"Branch", "Branches"}, "unit": {"Unit", "Units"}, "department": {"Department", "Departments"}}

// levels reads the organization's levels. Never configured: the kinds it already has are on, each under the nearest kind above it
// that is also on (the order sector > branch > unit > department), and the rest are off.
func (s *Service) levels(ctx context.Context, tx pgx.Tx, p *tenant.Principal) (Levels, error) {
	var saved Levels
	if err := orgsettings.Get(ctx, s.A, tx, p.Org, "structure", &saved); err != nil {
		return saved, err
	}
	if len(saved.Levels) > 0 {
		saved.Set = true
		return normalize(saved), nil
	}
	have := map[string]bool{}
	rs, err := tx.Query(ctx, `select distinct kind from hub.org_units`)
	if err != nil {
		return saved, err
	}
	for rs.Next() {
		var k string
		if rs.Scan(&k) == nil {
			have[k] = true
		}
	}
	rs.Close()
	out := Levels{}
	for i, k := range allKinds {
		l := Level{Kind: k, Enabled: have[k], Label: defaultLabel[k][0], Plural: defaultLabel[k][1]}
		for j := i - 1; j >= 0; j-- {
			if have[allKinds[j]] {
				l.Parent = allKinds[j]
				break
			}
		}
		out.Levels = append(out.Levels, l)
	}
	return out, nil
}

// normalize puts the four kinds in their fixed order, fills missing labels and drops anything unknown.
func normalize(in Levels) Levels {
	by := map[string]Level{}
	for _, l := range in.Levels {
		by[l.Kind] = l
	}
	out := Levels{Set: in.Set}
	for _, k := range allKinds {
		l, ok := by[k]
		if !ok {
			l = Level{Kind: k}
		}
		l.Kind = k
		if strings.TrimSpace(l.Label) == "" {
			l.Label = defaultLabel[k][0]
		}
		if strings.TrimSpace(l.Plural) == "" {
			l.Plural = defaultLabel[k][1]
		}
		out.Levels = append(out.Levels, l)
	}
	return out
}

func (l Levels) get(kind string) (Level, bool) {
	for _, x := range l.Levels {
		if x.Kind == kind {
			return x, true
		}
	}
	return Level{}, false
}

func (s *Service) Levels(ctx context.Context, p *tenant.Principal) (Levels, error) {
	var out Levels
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var err error
		out, err = s.levels(ctx, tx, p)
		return err
	})
	return out, err
}

// SetLevels saves the organization's levels. A level that still has entries cannot be switched off, and a level's parent cannot
// change while its entries sit under a different kind: the records stay straight.
func (s *Service) SetLevels(ctx context.Context, p *tenant.Principal, in Levels) (Levels, error) {
	in = normalize(in)
	in.Set = true
	on := map[string]bool{}
	for _, l := range in.Levels {
		if len(l.Label) > 40 || len(l.Plural) > 40 {
			return in, httpx.Bad("Keep each name under 40 characters.")
		}
		on[l.Kind] = l.Enabled
	}
	for _, l := range in.Levels {
		if !l.Enabled {
			continue
		}
		if l.Parent == l.Kind {
			return in, httpx.Bad(l.Label + " cannot sit under itself.")
		}
		if l.Parent != "" && !on[l.Parent] {
			return in, httpx.Bad(l.Label + " sits under a level that is switched off.")
		}
	}
	// no loops: follow each level's parent up to the organization
	for _, l := range in.Levels {
		seen := map[string]bool{}
		cur := l
		for cur.Enabled && cur.Parent != "" {
			if seen[cur.Kind] {
				return in, httpx.Bad("The levels go round in a circle. One of them must sit under the organization itself.")
			}
			seen[cur.Kind] = true
			cur, _ = in.get(cur.Parent)
		}
	}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		for _, l := range in.Levels {
			var n int
			if err := tx.QueryRow(ctx, `select count(*) from hub.org_units where kind=$1`, l.Kind).Scan(&n); err != nil {
				return err
			}
			if !l.Enabled && n > 0 {
				return httpx.Err(409, "level_in_use", "There are still "+itoa(n)+" "+strings.ToLower(l.Plural)+". Remove or merge them before switching "+strings.ToLower(l.Plural)+" off.")
			}
			if l.Enabled && n > 0 {
				var wrong int
				if err := tx.QueryRow(ctx, `select count(*) from hub.org_units u left join hub.org_units pu on pu.id = u.parent_id
					where u.kind=$1 and coalesce(pu.kind,'') <> $2`, l.Kind, l.Parent).Scan(&wrong); err != nil {
					return err
				}
				if wrong > 0 {
					return httpx.Err(409, "level_in_use", itoa(wrong)+" "+strings.ToLower(l.Plural)+" sit somewhere else now. Move them first, then change where "+strings.ToLower(l.Plural)+" sit.")
				}
			}
		}
		prev, _ := s.levels(ctx, tx, p)
		for i, l := range in.Levels {
			if !l.Enabled {
				in.Levels[i].HeadRole = ""
				continue
			}
			old, _ := prev.get(l.Kind)
			if l.HeadRole == "" && old.HeadRole == "" {
				continue // never decided at the level: places keep what they had
			}
			if err := applyLevelHead(ctx, tx, s, p, l.Kind, l.HeadRole); err != nil {
				return err
			}
		}
		if err := orgsettings.Set(ctx, s.A, tx, p.Org, "structure", in); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the organization's structure levels", "Settings", p.IP, in)
	})
	return in, err
}

// checkPlace says whether a part of this kind may be created or placed under this parent kind ("" = the organization itself).
func (l Levels) checkPlace(kind, parentKind string) error {
	lv, ok := l.get(kind)
	if !ok {
		return httpx.Bad("Unknown kind of part.")
	}
	if !lv.Enabled {
		return httpx.Bad("Your organization does not use " + strings.ToLower(lv.Plural) + ". Switch them on first.")
	}
	if lv.Parent != parentKind {
		if lv.Parent == "" {
			return httpx.Bad("Each " + strings.ToLower(lv.Label) + " sits directly under the organization.")
		}
		pl, _ := l.get(lv.Parent)
		return httpx.Bad("Choose the " + strings.ToLower(pl.Label) + " this " + strings.ToLower(lv.Label) + " sits under.")
	}
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// applyLevelHead gives every place of a level the level's head role (or takes it away when roleKey is empty).
func applyLevelHead(ctx context.Context, tx pgx.Tx, s *Service, p *tenant.Principal, kind, roleKey string) error {
	if roleKey == "" {
		if _, err := tx.Exec(ctx, `delete from hub.unit_heads where unit_id in (select id from hub.org_units where kind=$1)`, kind); err != nil {
			return err
		}
		return nil
	}
	var role uuid.UUID
	if err := tx.QueryRow(ctx, `select id from hub.roles where key=$1`, roleKey).Scan(&role); err != nil {
		return httpx.Bad("Choose a head role from the list.")
	}
	if _, err := tx.Exec(ctx, `delete from hub.unit_heads where unit_id in (select id from hub.org_units where kind=$1)`, kind); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `insert into hub.unit_heads (org_id, unit_id, role_id) select $1, id, $2 from hub.org_units where kind=$3`, p.Org, role, kind); err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "unit_heads", "unit_id in (select id from hub.org_units where kind=$1)", kind)
}
