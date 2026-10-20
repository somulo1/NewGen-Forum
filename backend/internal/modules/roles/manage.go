package roles

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

// ensureCatalog installs the starting modules and rights for an organization that does not have them yet (older
// organizations were made before the catalog was stored per organization).
func (s *Service) ensureCatalog(ctx context.Context, tx pgx.Tx, org uuid.UUID) error {
	ok, err := access.Installed(ctx, tx)
	if err != nil || ok {
		return err
	}
	return access.Install(ctx, tx, org, func(table, where string, args ...any) error { return s.A.Sign.Seal(ctx, tx, table, where, args...) })
}

type CatalogRight struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Kind    string `json:"kind"`
	Builtin bool   `json:"builtin"`
}

type CatalogModule struct {
	Key     string         `json:"key"`
	Label   string         `json:"label"`
	Tint    string         `json:"tint"`
	Builtin bool           `json:"builtin"`
	Rights  []CatalogRight `json:"rights"`
}

type CatalogOut struct {
	Modules []CatalogModule `json:"modules"`
	Kinds   []string        `json:"kinds"`
}

// Catalog is the organization's list of modules and rights, as data.
func (s *Service) Catalog(ctx context.Context, p *tenant.Principal) (*CatalogOut, error) {
	out := &CatalogOut{Modules: []CatalogModule{}, Kinds: access.Kinds}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := s.ensureCatalog(ctx, tx, p.Org); err != nil {
			return err
		}
		ms, err := tx.Query(ctx, `select key, label, tint, builtin from hub.modules order by sort, label`)
		if err != nil {
			return err
		}
		idx := map[string]int{}
		for ms.Next() {
			var m CatalogModule
			if err := ms.Scan(&m.Key, &m.Label, &m.Tint, &m.Builtin); err != nil {
				ms.Close()
				return err
			}
			m.Rights = []CatalogRight{}
			idx[m.Key] = len(out.Modules)
			out.Modules = append(out.Modules, m)
		}
		ms.Close()
		rs, err := tx.Query(ctx, `select module, name, label, kind, builtin from hub.rights order by sort, name`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var mod string
			var r CatalogRight
			if err := rs.Scan(&mod, &r.Name, &r.Label, &r.Kind, &r.Builtin); err != nil {
				return err
			}
			if i, ok := idx[mod]; ok {
				out.Modules[i].Rights = append(out.Modules[i].Rights, r)
			}
		}
		return rs.Err()
	})
	return out, err
}

var keyRe = regexp.MustCompile(`[^a-z0-9]+`)

func slug(name string) string {
	k := strings.Trim(keyRe.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if len(k) > 40 {
		k = strings.Trim(k[:40], "-")
	}
	return k
}

type RoleIn struct {
	Name    string  `json:"name"`
	Desc    *string `json:"desc"`
	Scope   *string `json:"scope"`
	Default *bool   `json:"default"` // the role new people get when none is chosen
	Copy    string  `json:"copy_from"`
}

func clean(s string, max int) (string, bool) {
	s = strings.TrimSpace(s)
	return s, s != "" && len(s) <= max
}

// CreateRole makes a role with no rights (or a copy of another role's rights). The key is made from the name.
func (s *Service) CreateRole(ctx context.Context, p *tenant.Principal, in RoleIn) (string, error) {
	name, ok := clean(in.Name, 60)
	if !ok {
		return "", httpx.Bad("Give the role a name (up to 60 characters).")
	}
	desc, scope := "", "Whole organization"
	if in.Desc != nil {
		desc = strings.TrimSpace(*in.Desc)
	}
	if in.Scope != nil && strings.TrimSpace(*in.Scope) != "" {
		scope = strings.TrimSpace(*in.Scope)
	}
	if len(desc) > 300 || len(scope) > 60 {
		return "", httpx.Bad("That description is too long.")
	}
	base := slug(name)
	if base == "" {
		return "", httpx.Bad("The name needs letters or numbers.")
	}
	var key string
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		if err := s.ensureCatalog(ctx, tx, p.Org); err != nil {
			return err
		}
		var dup bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.roles where lower(name)=lower($1))`, name).Scan(&dup); err != nil {
			return err
		}
		if dup {
			return httpx.Err(409, "name_taken", "A role called "+name+" already exists.")
		}
		key = base
		for i := 2; ; i++ {
			var n int
			if err := tx.QueryRow(ctx, `select count(*) from hub.roles where key=$1`, key).Scan(&n); err != nil {
				return err
			}
			if n == 0 {
				break
			}
			key = base + "-" + string(rune('0'+i%10))
			if i > 9 {
				key = base + "-" + uuid.NewString()[:6]
			}
		}
		id := app.NewID()
		if _, err := tx.Exec(ctx, `insert into hub.roles (id, org_id, key, name, description, scope, locked) values ($1,$2,$3,$4,$5,$6,false)`, id, p.Org, key, name, desc, scope); err != nil {
			return err
		}
		if in.Copy != "" {
			if _, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed)
				select $1, $2, rr.module, rr.right_name, rr.allowed from hub.role_rights rr join hub.roles r on r.id = rr.role_id where r.key=$3 and not r.locked`, p.Org, id, in.Copy); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "roles", "id=$1", id); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "role_rights", "role_id=$1", id); err != nil {
			return err
		}
		if in.Default != nil && *in.Default {
			if err := s.setDefault(ctx, tx, p.Org, id); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Created the role "+name, "Roles", p.IP, map[string]any{"role": key})
	})
	if err != nil {
		return "", err
	}
	s.rightsChanged(p.Org)
	return key, nil
}

func (s *Service) setDefault(ctx context.Context, tx pgx.Tx, org, role uuid.UUID) error {
	if _, err := tx.Exec(ctx, `insert into hub.default_role (org_id, role_id) values ($1,$2) on conflict (org_id) do update set role_id=excluded.role_id, `+integrity.BumpOn("default_role"), org, role); err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "default_role", "org_id=$1", org)
}

// UpdateRole renames or describes a role, or makes it the default for new people.
func (s *Service) UpdateRole(ctx context.Context, p *tenant.Principal, key string, in RoleIn) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var id uuid.UUID
		var name string
		if err := tx.QueryRow(ctx, `select id, name from hub.roles where key=$1`, key).Scan(&id, &name); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		// every changed column goes in one update: a row changed twice before it is sealed again is refused
		newName, newDesc, newScope := (*string)(nil), (*string)(nil), (*string)(nil)
		if strings.TrimSpace(in.Name) != "" {
			n, ok := clean(in.Name, 60)
			if !ok {
				return httpx.Bad("A role name is up to 60 characters.")
			}
			var dup bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.roles where lower(name)=lower($1) and id <> $2)`, n, id).Scan(&dup); err != nil {
				return err
			}
			if dup {
				return httpx.Err(409, "name_taken", "A role called "+n+" already exists.")
			}
			newName = &n
		}
		if in.Desc != nil {
			if len(*in.Desc) > 300 {
				return httpx.Bad("That description is too long.")
			}
			d := strings.TrimSpace(*in.Desc)
			newDesc = &d
		}
		if in.Scope != nil {
			if len(*in.Scope) > 60 {
				return httpx.Bad("That scope is too long.")
			}
			sc := strings.TrimSpace(*in.Scope)
			newScope = &sc
		}
		if newName != nil || newDesc != nil || newScope != nil {
			if _, err := tx.Exec(ctx, `update hub.roles set name=coalesce($2, name), description=coalesce($3, description), scope=coalesce($4, scope), `+integrity.Bump+` where id=$1`, id, newName, newDesc, newScope); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "roles", "id=$1", id); err != nil {
			return err
		}
		if in.Default != nil {
			if *in.Default {
				if err := s.setDefault(ctx, tx, p.Org, id); err != nil {
					return err
				}
			} else if _, err := tx.Exec(ctx, `delete from hub.default_role where role_id=$1`, id); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the role "+name, "Roles", p.IP, map[string]any{"role": key})
	})
	if err == nil {
		s.rightsChanged(p.Org)
	}
	return err
}

// DeleteRole removes a role nobody holds. Locked (owner) roles stay.
func (s *Service) DeleteRole(ctx context.Context, p *tenant.Principal, key string) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var id uuid.UUID
		var name string
		var locked bool
		if err := tx.QueryRow(ctx, `select id, name, locked from hub.roles where key=$1`, key).Scan(&id, &name, &locked); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if locked {
			return httpx.Err(409, "locked_role", name+" cannot be deleted.")
		}
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.person_roles pr join hub.people p on p.id=pr.person_id where pr.role_id=$1 and p.status <> 'deleted'`, id).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return httpx.Err(409, "role_in_use", name+" is held by people. Move them to another role first.")
		}
		for _, q := range []string{
			`delete from hub.person_roles where role_id=$1`,
			`delete from hub.role_rights where role_id=$1`,
			`delete from hub.role_reports where role_id=$1 or reports_to_role=$1`,
			`delete from hub.default_role where role_id=$1`,
			`delete from hub.roles where id=$1`,
		} {
			if _, err := tx.Exec(ctx, q, id); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Deleted the role "+name, "Roles", p.IP, map[string]any{"role": key})
	})
	if err == nil {
		s.rightsChanged(p.Org)
	}
	return err
}

type RightIn struct {
	Module string `json:"module"`
	Name   string `json:"name"`
	Label  string `json:"label"`
	Kind   string `json:"kind"`
	Tint   string `json:"tint"`
}

var tintRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// AddModule adds a page of the organization's own to the catalog (its rights are then added to it).
func (s *Service) AddModule(ctx context.Context, p *tenant.Principal, in RightIn) error {
	key := slug(in.Module)
	label, ok := clean(in.Label, 60)
	if !ok {
		label, ok = clean(in.Module, 60)
	}
	if key == "" || !ok {
		return httpx.Bad("Give the module a name.")
	}
	if in.Tint != "" && !tintRe.MatchString(in.Tint) {
		return httpx.Bad("Colours look like #7CC4FF.")
	}
	if in.Tint == "" {
		in.Tint = "#9DB0A6"
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		if err := s.ensureCatalog(ctx, tx, p.Org); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `insert into hub.modules (org_id, key, label, tint, sort, builtin) select $1,$2,$3,$4, coalesce(max(sort),0)+1, false from hub.modules on conflict do nothing`, p.Org, key, label, in.Tint)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.Err(409, "exists", "That module already exists.")
		}
		if err := s.A.Sign.Seal(ctx, tx, "modules", "org_id=$1 and key=$2", p.Org, key); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Added the module "+label, "Roles", p.IP, map[string]any{"module": key})
	})
}

// AddRight adds a named permission to a module. Pages, buttons and approval levels can then be tied to it.
func (s *Service) AddRight(ctx context.Context, p *tenant.Principal, in RightIn) error {
	name, ok := clean(in.Name, 60)
	if !ok {
		return httpx.Bad("Give the right a name (up to 60 characters).")
	}
	label := name
	if l, ok := clean(in.Label, 80); ok {
		label = l
	}
	kind := in.Kind
	if kind == "" {
		kind = access.KindOf(name)
	}
	valid := false
	for _, k := range access.Kinds {
		valid = valid || k == kind
	}
	if !valid {
		return httpx.Bad("That kind of right does not exist.")
	}
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		if err := s.ensureCatalog(ctx, tx, p.Org); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `insert into hub.rights (org_id, module, name, label, kind, sort, builtin)
			select $1, m.key, $3, $4, $5, (select coalesce(max(sort),0)+1 from hub.rights where module=$2), false from hub.modules m where m.key=$2 on conflict do nothing`, p.Org, in.Module, name, label, kind)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.Err(409, "exists", "That module does not exist, or it already has that right.")
		}
		if err := s.A.Sign.Seal(ctx, tx, "rights", "org_id=$1 and module=$2 and name=$3", p.Org, in.Module, name); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Added the right "+name+" to "+in.Module, "Roles", p.IP, map[string]any{"module": in.Module, "right": name})
	})
}

// UpdateRight renames the label or regroups a right (the name itself is what the code and the grants refer to).
func (s *Service) UpdateRight(ctx context.Context, p *tenant.Principal, in RightIn) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		label, kind := strings.TrimSpace(in.Label), in.Kind
		if kind != "" {
			valid := false
			for _, k := range access.Kinds {
				valid = valid || k == kind
			}
			if !valid {
				return httpx.Bad("That kind of right does not exist.")
			}
		}
		tag, err := tx.Exec(ctx, `update hub.rights set label = case when $3 <> '' then $3 else label end, kind = case when $4 <> '' then $4 else kind end, `+integrity.Bump+` where module=$1 and name=$2`, in.Module, in.Name, label, kind)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		return s.A.Sign.Seal(ctx, tx, "rights", "module=$1 and name=$2", in.Module, in.Name)
	})
}

// RemoveRight deletes a right the organization added. Rights the server enforces stay.
func (s *Service) RemoveRight(ctx context.Context, p *tenant.Principal, module, name string) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var builtin bool
		if err := tx.QueryRow(ctx, `select builtin from hub.rights where module=$1 and name=$2`, module, name).Scan(&builtin); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if builtin {
			return httpx.Err(409, "builtin", "Hub itself uses that right, so it cannot be removed.")
		}
		for _, q := range []string{
			`delete from hub.role_rights where module=$1 and right_name=$2`,
			`delete from hub.person_rights where module=$1 and right_name=$2`,
			`delete from hub.rights where module=$1 and name=$2`,
		} {
			if _, err := tx.Exec(ctx, q, module, name); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Removed the right "+name+" from "+module, "Roles", p.IP, map[string]any{"module": module, "right": name})
	})
	if err == nil {
		s.rightsChanged(p.Org)
	}
	return err
}
