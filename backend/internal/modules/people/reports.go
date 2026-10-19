package people

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ReportsTo is who a person answers to: the manager set for them, else the person their role reports to, else the role.
type ReportsTo struct {
	Kind    string `json:"kind"` // person | role | none
	Person  *Ref   `json:"person,omitempty"`
	Role    string `json:"role,omitempty"`
	RoleKey string `json:"role_key,omitempty"`
}

// defaultReports is who each role reports to until the organization chooses otherwise.
var defaultReports = map[string]string{"staff": "hod", "it": "hod", "finance": "hod", "hod": "branch", "unit": "branch", "branch": "sector", "sector": "admin"}

type roleLine struct {
	name     string
	holders  int
	locked   bool
	set      bool // the organization has a row for this role (even one that says "no one")
	toRole   string
	toPerson *Ref
}

func (s *Service) reportsTo(ctx context.Context, tx pgx.Tx, person uuid.UUID, manager *Ref) (*ReportsTo, error) {
	if manager != nil {
		return &ReportsTo{Kind: "person", Person: manager}, nil
	}
	var mine string
	if err := tx.QueryRow(ctx, `select coalesce((select r.key from hub.person_roles pr join hub.roles r on r.id = pr.role_id where pr.person_id=$1 limit 1), '')`, person).Scan(&mine); err != nil {
		return nil, err
	}
	rs, err := tx.Query(ctx, `select r.key, r.name, r.locked,
			(select count(*) from hub.person_roles pr join hub.people x on x.id = pr.person_id where pr.role_id = r.id and x.status <> 'deleted'),
			rr.role_id is not null, coalesce(rt.key,''), mg.id, coalesce(mg.first_name || ' ' || mg.last_name, '')
		from hub.roles r left join hub.role_reports rr on rr.role_id = r.id left join hub.roles rt on rt.id = rr.reports_to_role
		left join hub.people mg on mg.id = rr.reports_to_person`)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	lines := map[string]*roleLine{}
	for rs.Next() {
		var key string
		var l roleLine
		var pid *uuid.UUID
		var pname string
		if err := rs.Scan(&key, &l.name, &l.locked, &l.holders, &l.set, &l.toRole, &pid, &pname); err != nil {
			return nil, err
		}
		if pid != nil {
			l.toPerson = &Ref{*pid, strings.TrimSpace(pname)}
		}
		lines[key] = &l
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	cur := mine
	for hop := 0; hop < 8 && cur != ""; hop++ {
		l := lines[cur]
		if l == nil {
			break
		}
		if l.toPerson != nil && l.toPerson.ID != person {
			return &ReportsTo{Kind: "person", Person: l.toPerson, Role: l.name, RoleKey: cur}, nil
		}
		next := l.toRole
		if !l.set {
			next = defaultReports[cur]
		}
		if next == "" || next == cur {
			break
		}
		// a role nobody holds is skipped: the person reports to whatever sits above it
		if t := lines[next]; t != nil && (t.holders > 0 || t.locked) {
			return &ReportsTo{Kind: "role", Role: t.name, RoleKey: next}, nil
		}
		cur = next
	}
	return &ReportsTo{Kind: "none"}, nil
}
