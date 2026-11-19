// Package roles is roles and rights, reporting lines, and the signed-in person's own profile ("me").
package roles

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/people"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Service struct {
	A  *app.App
	Ap *approvals.Engine
}

func New(a *app.App, ap *approvals.Engine) *Service {
	s := &Service{A: a, Ap: ap}
	ap.Register("rights.role", s.applyRoleRight)
	ap.Register("rights.person", s.applyPersonRight)
	ap.Register("rights.group", func(context.Context, pgx.Tx, approvals.Applied) error { return errGroupsRetired })
	ap.Register("rights.batch", s.applyBatch)
	return s
}

type RoleInfo struct {
	Key     string `json:"id"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Scope   string `json:"scope"`
	Locked  bool   `json:"locked"`
	Needs   string `json:"needs,omitempty"`
	Members int    `json:"count"`
	Default bool   `json:"default"` // the role new people get when none is chosen
	// who this role reports to: a role, and optionally one named person in it
	ReportsToRole   string     `json:"reports_to_role,omitempty"`
	ReportsToPerson *uuid.UUID `json:"reports_to_person,omitempty"`
}

func (s *Service) Roles(ctx context.Context, p *tenant.Principal) ([]RoleInfo, error) {
	out := []RoleInfo{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "View roles"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select r.key, r.name, r.description, r.scope, r.locked, coalesce(r.needs_level,''), count(pp.person_id) filter (where x.status <> 'deleted'),
			coalesce(min(rt.key),''), (array_agg(rr.reports_to_person))[1], bool_or(dr.role_id is not null)
			from hub.roles r left join hub.person_roles pp on pp.role_id = r.id left join hub.people x on x.id = pp.person_id
			left join hub.role_reports rr on rr.role_id = r.id left join hub.roles rt on rt.id = rr.reports_to_role left join hub.default_role dr on dr.role_id = r.id
			group by r.id order by r.locked desc, min(r.key)`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var x RoleInfo
			if err := rs.Scan(&x.Key, &x.Name, &x.Desc, &x.Scope, &x.Locked, &x.Needs, &x.Members, &x.ReportsToRole, &x.ReportsToPerson, &x.Default); err != nil {
				return err
			}
			out = append(out, x)
		}
		return rs.Err()
	})
	return out, err
}

type RoleDetail struct {
	RoleInfo
	Rights  map[string]map[string]bool `json:"rights"`
	Pending map[string]bool            `json:"pending"`
}

// Role returns one role's complete rights matrix (module -> right -> allowed), plus changes still waiting for approval.
func (s *Service) Role(ctx context.Context, p *tenant.Principal, key string) (*RoleDetail, error) {
	var d RoleDetail
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "View roles"); err != nil {
			return err
		}
		var id uuid.UUID
		err := tx.QueryRow(ctx, `select id, key, name, description, scope, locked from hub.roles where key=$1`, key).Scan(&id, &d.Key, &d.Name, &d.Desc, &d.Scope, &d.Locked)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		d.Rights, d.Pending = map[string]map[string]bool{}, map[string]bool{}
		rs, err := tx.Query(ctx, `select module, right_name, allowed from hub.role_rights where role_id=$1`, id)
		if err != nil {
			return err
		}
		for rs.Next() {
			var m, r string
			var a bool
			if err := rs.Scan(&m, &r, &a); err != nil {
				rs.Close()
				return err
			}
			if d.Rights[m] == nil {
				d.Rights[m] = map[string]bool{}
			}
			d.Rights[m][r] = a
		}
		rs.Close()
		if d.Locked {
			// the owner role holds every right not switched off for it, including ones added to the catalog after it was made
			all, err := tx.Query(ctx, `select module, name from hub.rights`)
			if err != nil {
				return err
			}
			for all.Next() {
				var m, r string
				if err := all.Scan(&m, &r); err != nil {
					all.Close()
					return err
				}
				if d.Rights[m] == nil {
					d.Rights[m] = map[string]bool{}
				}
				if _, set := d.Rights[m][r]; !set {
					d.Rights[m][r] = true
				}
			}
			all.Close()
		}
		return tx.QueryRow(ctx, `select count(*) from hub.person_roles where role_id=$1`, id).Scan(&d.Members)
	})
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// errGroupsRetired answers an approval for a user group's right that was asked for before user groups were retired.
var errGroupsRetired = httpx.Err(409, "gone", "User groups were retired. Give the right to the roles on Roles & rights.")

type rightChange struct {
	Role    string    `json:"role,omitempty"`
	Person  uuid.UUID `json:"person,omitempty"`
	Module  string    `json:"module"`
	Right   string    `json:"right"`
	Allowed *bool     `json:"allowed"` // nil on a person removes their personal override
}

func (s *Service) validate(ctx context.Context, org uuid.UUID, c rightChange, person bool) error {
	if !person && c.Allowed == nil {
		return httpx.Bad("Say whether the right is allowed.")
	}
	var ok bool
	err := s.A.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
		if err := s.ensureCatalog(ctx, tx, org); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select exists (select 1 from hub.rights where module=$1 and name=$2)`, c.Module, c.Right).Scan(&ok)
	})
	if err != nil {
		return err
	}
	if !ok {
		return httpx.Bad("That right does not exist.")
	}
	return nil
}

// keeps refuses to take from the owner role a right it always keeps (see access.OwnerKeeps).
func keeps(locked bool, name string, c rightChange) error {
	if locked && c.Allowed != nil && !*c.Allowed && access.OwnerKeeps(c.Module, c.Right) {
		return httpx.Err(409, "kept_right", name+" always keeps \""+c.Right+"\", so someone can always manage roles and rights.")
	}
	return nil
}

// SetRoleRight changes one right of a role. Flow f8 may need approval first.
func (s *Service) SetRoleRight(ctx context.Context, p *tenant.Principal, role string, c rightChange) (*approvals.Request, bool, error) {
	c.Role, c.Person = role, uuid.Nil
	if err := s.validate(ctx, p.Org, c, false); err != nil {
		return nil, false, err
	}
	var req *approvals.Request
	var applied bool
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var locked bool
		var name string
		if err := tx.QueryRow(ctx, `select locked, name from hub.roles where key=$1`, role).Scan(&locked, &name); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if err := keeps(locked, name, c); err != nil {
			return err
		}
		verb := "allow"
		if !*c.Allowed {
			verb = "remove"
		}
		var err error
		req, applied, err = s.Ap.Submit(ctx, tx, p, "f8", "rights.role", fmt.Sprintf("Change rights: %s · %s · %s (%s)", name, c.Module, c.Right, verb), "", c)
		return err
	})
	return req, applied, err
}

func (s *Service) applyRoleRight(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var c rightChange
	if err := json.Unmarshal(a.Payload, &c); err != nil {
		return err
	}
	var locked bool
	var name string
	if err := tx.QueryRow(ctx, `select locked, name from hub.roles where key=$1`, c.Role).Scan(&locked, &name); err == nil {
		if err := keeps(locked, name, c); err != nil {
			return err
		}
	}
	tag, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed)
		select $5::uuid, r.id, $2::text, $3::text, $4::boolean from hub.roles r where r.key=$1
		on conflict (org_id, role_id, module, right_name) do update set allowed=excluded.allowed, `+integrity.BumpOn("role_rights"), c.Role, c.Module, c.Right, *c.Allowed, a.Org)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.Err(409, "gone", "That role or right no longer exists.")
	}
	if err := s.A.Sign.Seal(ctx, tx, "role_rights", "role_id=(select id from hub.roles where key=$1) and module=$2 and right_name=$3", c.Role, c.Module, c.Right); err != nil {
		return err
	}
	s.rightsChanged(a.Org)
	return nil
}

// SetPersonRight gives or takes away a single right for one person (nil removes the personal override).
func (s *Service) SetPersonRight(ctx context.Context, p *tenant.Principal, person uuid.UUID, c rightChange) (*approvals.Request, bool, error) {
	c.Person, c.Role = person, ""
	if err := s.validate(ctx, p.Org, c, true); err != nil {
		return nil, false, err
	}
	var req *approvals.Request
	var applied bool
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var name string
		var locked bool
		if err := tx.QueryRow(ctx, `select p.first_name || ' ' || p.last_name, coalesce(r.locked,false) from hub.people p left join hub.person_roles pr on pr.person_id=p.id left join hub.roles r on r.id=pr.role_id where p.id=$1 and p.status<>'deleted'`, person).Scan(&name, &locked); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if locked {
			return httpx.Err(409, "locked_role", "A super admin cannot be limited.")
		}
		what := "reset to the role"
		if c.Allowed != nil {
			what = map[bool]string{true: "allow", false: "remove"}[*c.Allowed]
		}
		var err error
		req, applied, err = s.Ap.Submit(ctx, tx, p, "f8", "rights.person", fmt.Sprintf("Change rights for %s: %s · %s (%s)", name, c.Module, c.Right, what), "", c)
		return err
	})
	return req, applied, err
}

func (s *Service) applyPersonRight(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var c rightChange
	if err := json.Unmarshal(a.Payload, &c); err != nil {
		return err
	}
	if c.Allowed == nil {
		_, err := tx.Exec(ctx, `delete from hub.person_rights where person_id=$1 and module=$2 and right_name=$3`, c.Person, c.Module, c.Right)
		s.rightsChanged(a.Org)
		return err
	}
	if _, err := tx.Exec(ctx, `insert into hub.person_rights (org_id, person_id, module, right_name, allowed) values ($1,$2,$3,$4,$5)
		on conflict (org_id, person_id, module, right_name) do update set allowed=excluded.allowed, `+integrity.BumpOn("person_rights"), a.Org, c.Person, c.Module, c.Right, *c.Allowed); err != nil {
		return err
	}
	if err := s.A.Sign.Seal(ctx, tx, "person_rights", "person_id=$1 and module=$2 and right_name=$3", c.Person, c.Module, c.Right); err != nil {
		return err
	}
	s.rightsChanged(a.Org)
	return nil
}

func (s *Service) rightsChanged(org uuid.UUID) {
	go s.A.Publish(context.Background(), realtime.Event{Type: "rights.changed", Channel: realtime.OrgChannel(org), Org: org.String()})
}

// ---- reporting lines ----

type Reporting struct {
	Role     string     `json:"role"`
	ToRole   *string    `json:"reports_to_role"`
	ToPerson *uuid.UUID `json:"reports_to_person"`
}

func (s *Service) SetRoleReport(ctx context.Context, p *tenant.Principal, r Reporting) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		if r.ToRole != nil && *r.ToRole == r.Role {
			return httpx.Bad("A role cannot report to itself.")
		}
		if _, err := tx.Exec(ctx, `insert into hub.role_reports (org_id, role_id, reports_to_role, reports_to_person)
			select $1, r.id, (select id from hub.roles where key=$3), $4 from hub.roles r where r.key=$2
			on conflict (org_id, role_id) do update set reports_to_role=excluded.reports_to_role, reports_to_person=excluded.reports_to_person, `+integrity.BumpOn("role_reports"),
			p.Org, r.Role, r.ToRole, r.ToPerson); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "role_reports", "role_id=(select id from hub.roles where key=$1)", r.Role); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed who "+r.Role+" reports to", "Roles", p.IP, r)
	})
}

// ---- me ----

type Me struct {
	Account  uuid.UUID           `json:"account"`
	Person   uuid.UUID           `json:"person"`
	Name     string              `json:"name"`
	Role     string              `json:"role"`
	RoleName string              `json:"role_name"`
	Status   string              `json:"status"`
	Rights   map[string][]string `json:"rights"`
	Profile  Profile             `json:"profile"`
}

type Profile struct {
	Availability string   `json:"status"`
	Cover        string   `json:"cover"`
	Photo        string   `json:"photo"`
	Phone        string   `json:"phone"`
	Bio          string   `json:"bio"`
	Skills       []string `json:"skills"`
	Headline     string   `json:"headline"`
}

func pctx(org, person uuid.UUID, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "person_profiles", Column: col, RowID: person.String()}
}

func (s *Service) Me(ctx context.Context, p *tenant.Principal) (*Me, error) {
	m := &Me{Account: p.Account, Person: p.Person, Name: p.Name, Rights: map[string][]string{}, Profile: Profile{Skills: []string{}}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := access.EnsureProjectRights(ctx, tx, p.Org, func(table, where string, args ...any) error { return s.A.Sign.Seal(ctx, tx, table, where, args...) }); err != nil {
			return err
		}
		var phone []byte
		err := tx.QueryRow(ctx, `select p.status, p.availability, p.phone_enc, p.first_name || ' ' || p.last_name, coalesce(r.key,''), coalesce(r.name,'')
			from hub.people p left join hub.person_roles pr on pr.person_id=p.id left join hub.roles r on r.id=pr.role_id where p.id=$1`, p.Person).Scan(&m.Status, &m.Profile.Availability, &phone, &m.Name, &m.Role, &m.RoleName)
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		if len(phone) > 0 {
			if b, err := crypto.FieldOpen(dek, crypto.Context{Org: p.Org.String(), Table: "people", Column: "phone_enc", RowID: p.Person.String()}, phone); err == nil {
				m.Profile.Phone = people.MaskPhone(string(b))
			}
		}
		var bio, skills, photo []byte
		var ptype *string
		m.Profile.Cover = "#1F3A2B"
		err = tx.QueryRow(ctx, `select bio_enc, skills_enc, photo_enc, photo_type, cover from hub.person_profiles where person_id=$1`, p.Person).Scan(&bio, &skills, &photo, &ptype, &m.Profile.Cover)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		open := func(col string, ct []byte) []byte {
			if len(ct) == 0 {
				return nil
			}
			b, _ := crypto.FieldOpen(dek, pctx(p.Org, p.Person, col), ct)
			return b
		}
		m.Profile.Bio = string(open("bio_enc", bio))
		if b := open("skills_enc", skills); b != nil {
			_ = json.Unmarshal(b, &m.Profile.Skills)
		}
		if b := open("photo_enc", photo); b != nil && ptype != nil {
			m.Profile.Photo = "data:" + *ptype + ";base64," + base64.StdEncoding.EncodeToString(b)
		}
		_ = tx.QueryRow(ctx, `select headline from hub.person_headlines where person_id=$1`, p.Person).Scan(&m.Profile.Headline)
		all := map[string][]string{}
		for _, mod := range access.Default().Modules {
			all[mod.Key] = mod.Rights
		}
		m.Rights, err = authz.Effective(ctx, tx, p.Person, all)
		return err
	})
	return m, err
}

type ProfileEdit struct {
	Availability *string   `json:"status"`
	Cover        *string   `json:"cover"`
	Photo        *string   `json:"photo"` // data URL; empty string removes it
	Phone        *string   `json:"phone"`
	Bio          *string   `json:"bio"`
	Skills       *[]string `json:"skills"`
	Headline     *string   `json:"headline"` // empty removes it
}

var availabilities = map[string]bool{"Available": true, "In a meeting": true, "Focusing": true, "Away": true}

const maxPhoto = 512 << 10
