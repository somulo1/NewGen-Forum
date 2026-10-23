// Package roles is roles and rights, reporting lines, and the signed-in person's own profile ("me").
package roles

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
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
	ap.Register("rights.group", s.applyGroupRight)
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
			d.Rights[m][r] = a || d.Locked
		}
		rs.Close()
		if d.Locked {
			// a locked role (the owner) holds every right in the catalog, including ones added after it was made
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
				d.Rights[m][r] = true
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

// SetRoleRight changes one right of a role. Flow f8 may need approval first; locked roles can never be limited.
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
		if locked {
			return httpx.Err(409, "locked_role", name+" cannot be limited.")
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
	tag, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed)
		select $5::uuid, r.id, $2::text, $3::text, $4::boolean from hub.roles r where r.key=$1 and not r.locked
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
}

func pctx(org, person uuid.UUID, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "person_profiles", Column: col, RowID: person.String()}
}

func (s *Service) Me(ctx context.Context, p *tenant.Principal) (*Me, error) {
	m := &Me{Account: p.Account, Person: p.Person, Name: p.Name, Rights: map[string][]string{}, Profile: Profile{Skills: []string{}}}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
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
				m.Profile.Phone = string(b)
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
}

var availabilities = map[string]bool{"Available": true, "In a meeting": true, "Focusing": true, "Away": true}

const maxPhoto = 512 << 10

// EditMe saves the caller's own profile. This never needs approval: it is their own page.
func (s *Service) EditMe(ctx context.Context, p *tenant.Principal, e ProfileEdit) error {
	if e.Availability != nil && !availabilities[*e.Availability] {
		return httpx.Bad("Choose a status from the list.")
	}
	if e.Bio != nil && len(*e.Bio) > 2000 {
		return httpx.Bad("Keep your bio under 2,000 characters.")
	}
	if e.Phone != nil && *e.Phone != "" && (len(*e.Phone) < 5 || len(*e.Phone) > 24) {
		return httpx.Bad("Enter a valid phone number.")
	}
	if e.Cover != nil && (len(*e.Cover) != 7 || (*e.Cover)[0] != '#') {
		return httpx.Bad("Choose a colour like #1F3A2B.")
	}
	var skills []string
	if e.Skills != nil {
		for _, k := range *e.Skills {
			if k = strings.TrimSpace(k); k != "" && len(k) <= 40 {
				skills = append(skills, k)
			}
		}
		if len(skills) > 30 {
			return httpx.Bad("Add up to 30 skills.")
		}
	}
	var photoBytes []byte
	var photoType string
	if e.Photo != nil && *e.Photo != "" {
		head, data, ok := strings.Cut(*e.Photo, ",")
		if !ok || !strings.HasPrefix(head, "data:image/") || !strings.HasSuffix(head, ";base64") {
			return httpx.Bad("Choose an image file for your profile photo.")
		}
		photoType = strings.TrimSuffix(strings.TrimPrefix(head, "data:"), ";base64")
		if photoType != "image/png" && photoType != "image/jpeg" && photoType != "image/webp" && photoType != "image/gif" {
			return httpx.Bad("Use a PNG, JPEG, WebP or GIF photo.")
		}
		var err error
		if photoBytes, err = base64.StdEncoding.DecodeString(data); err != nil {
			return httpx.Bad("The photo could not be read.")
		}
		if len(photoBytes) > maxPhoto {
			return httpx.Bad("Profile photos must be 512 KB or smaller.")
		}
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		if e.Availability != nil || e.Phone != nil {
			var phone any
			if e.Phone != nil {
				b, _ := crypto.Field(dek, crypto.Context{Org: p.Org.String(), Table: "people", Column: "phone_enc", RowID: p.Person.String()}, []byte(*e.Phone))
				if *e.Phone == "" {
					b = nil
				}
				phone = b
			}
			if _, err := tx.Exec(ctx, `update hub.people set availability=coalesce($2, availability), phone_enc = case when $3::boolean then $4 else phone_enc end, updated_at=now(), `+integrity.Bump+` where id=$1`,
				p.Person, e.Availability, e.Phone != nil, phone); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "people", "id=$1", p.Person); err != nil {
				return err
			}
		}
		if e.Cover != nil || e.Photo != nil || e.Bio != nil || e.Skills != nil {
			if _, err := tx.Exec(ctx, `insert into hub.person_profiles (org_id, person_id) values ($1,$2) on conflict do nothing`, p.Org, p.Person); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "person_profiles", "person_id=$1", p.Person); err != nil {
				return err
			}
			sealIf := func(col string, plain []byte, set bool) (any, error) {
				if !set {
					return nil, nil
				}
				if len(plain) == 0 {
					return []byte{}, nil
				}
				return crypto.Field(dek, pctx(p.Org, p.Person, col), plain)
			}
			bio, err := sealIf("bio_enc", []byte(deref(e.Bio)), e.Bio != nil)
			if err != nil {
				return err
			}
			var sk any
			if e.Skills != nil {
				raw, _ := json.Marshal(skills)
				if sk, err = sealIf("skills_enc", raw, true); err != nil {
					return err
				}
			}
			ph, err := sealIf("photo_enc", photoBytes, e.Photo != nil)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `update hub.person_profiles set
				bio_enc = case when $2::bytea is null then bio_enc when length($2)=0 then null else $2 end,
				skills_enc = case when $3::bytea is null then skills_enc else $3 end,
				photo_enc = case when $4::bytea is null then photo_enc when length($4)=0 then null else $4 end,
				photo_type = case when $4::bytea is null then photo_type when length($4)=0 then null else $5 end,
				cover = coalesce($6, cover), updated_at=now(), `+integrity.Bump+` where person_id=$1`, p.Person, bio, sk, ph, photoType, e.Cover); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "person_profiles", "person_id=$1", p.Person); err != nil {
				return err
			}
		}
		return nil
	})
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ---- the person's app settings (home widgets, theme ...) ----

func prefsCtx(org, person uuid.UUID) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "person_prefs", Column: "data_enc", RowID: person.String()}
}

func (s *Service) Prefs(ctx context.Context, p *tenant.Principal) (json.RawMessage, error) {
	out := json.RawMessage(`{}`)
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var enc []byte
		err := tx.QueryRow(ctx, `select data_enc from hub.person_prefs where person_id=$1`, p.Person).Scan(&enc)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		raw, err := crypto.FieldOpen(dek, prefsCtx(p.Org, p.Person), enc)
		if err != nil {
			return err
		}
		out = raw
		return nil
	})
	return out, err
}

func (s *Service) SetPrefs(ctx context.Context, p *tenant.Principal, raw json.RawMessage) error {
	var probe map[string]any
	if len(raw) > 16<<10 || json.Unmarshal(raw, &probe) != nil || probe == nil {
		return httpx.Bad("Settings must be a JSON object under 16 KB.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		dek, _, err := s.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		enc, err := crypto.Field(dek, prefsCtx(p.Org, p.Person), raw)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into hub.person_prefs (org_id, person_id, data_enc) values ($1,$2,$3) on conflict (org_id, person_id) do update set data_enc=excluded.data_enc, updated_at=now(), `+integrity.BumpOn("person_prefs"), p.Org, p.Person, enc); err != nil {
			return err
		}
		return s.A.Sign.Seal(ctx, tx, "person_prefs", "person_id=$1", p.Person)
	})
}

// PersonRights lists the personal overrides one person has on top of their role.
func (s *Service) PersonRights(ctx context.Context, p *tenant.Principal, person uuid.UUID) ([]map[string]any, error) {
	out := []map[string]any{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select module, right_name, allowed from hub.person_rights where person_id=$1 order by module, right_name`, person)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var m, r string
			var a bool
			if err := rs.Scan(&m, &r, &a); err != nil {
				return err
			}
			out = append(out, map[string]any{"module": m, "right": r, "allowed": a})
		}
		return rs.Err()
	})
	return out, err
}
