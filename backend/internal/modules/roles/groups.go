package roles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/tenant"
)

// User groups: a named set of roles. A right given to a group is held by everyone whose role is in it, on top of what the
// role itself allows (see hub.role_right). A person still has one role.

type GroupInfo struct {
	Key   string   `json:"id"`
	Name  string   `json:"name"`
	Desc  string   `json:"desc"`
	Roles []string `json:"roles"` // role keys
	Count int      `json:"count"` // people whose role is in the group
}

type GroupDetail struct {
	GroupInfo
	Rights map[string]map[string]bool `json:"rights"`
}

type GroupIn struct {
	Name  string   `json:"name"`
	Desc  *string  `json:"desc"`
	Roles []string `json:"roles"` // nil: leave as it is
}

func (s *Service) Groups(ctx context.Context, p *tenant.Principal) ([]GroupInfo, error) {
	out := []GroupInfo{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "View roles"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select g.key, g.name, g.description,
				coalesce(array_agg(r.key order by r.key) filter (where r.id is not null), '{}'),
				(select count(*) from hub.person_roles pr join hub.people x on x.id = pr.person_id
				  where x.status <> 'deleted' and pr.role_id in (select role_id from hub.role_group_roles where group_id = g.id))
			from hub.role_groups g left join hub.role_group_roles gr on gr.group_id = g.id left join hub.roles r on r.id = gr.role_id
			group by g.id order by lower(g.name)`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var g GroupInfo
			if err := rs.Scan(&g.Key, &g.Name, &g.Desc, &g.Roles, &g.Count); err != nil {
				return err
			}
			out = append(out, g)
		}
		return rs.Err()
	})
	return out, err
}

// Group returns one group with its complete rights (module -> right -> allowed).
func (s *Service) Group(ctx context.Context, p *tenant.Principal, key string) (*GroupDetail, error) {
	all, err := s.Groups(ctx, p)
	if err != nil {
		return nil, err
	}
	var d *GroupDetail
	for _, g := range all {
		if g.Key == key {
			d = &GroupDetail{GroupInfo: g, Rights: map[string]map[string]bool{}}
		}
	}
	if d == nil {
		return nil, httpx.ErrNotFound
	}
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select x.module, x.right_name, x.allowed from hub.role_group_rights x join hub.role_groups g on g.id = x.group_id where g.key=$1`, key)
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
			if d.Rights[m] == nil {
				d.Rights[m] = map[string]bool{}
			}
			d.Rights[m][r] = a
		}
		return rs.Err()
	})
	return d, err
}

// setGroupRoles replaces which roles a group holds. Roles that cannot be limited (super admin) are not worth grouping.
func (s *Service) setGroupRoles(ctx context.Context, tx pgx.Tx, org, group uuid.UUID, keys []string) error {
	seen := map[string]bool{}
	uniq := keys[:0:0]
	for _, k := range keys {
		if k = strings.TrimSpace(k); k != "" && !seen[k] {
			seen[k] = true
			uniq = append(uniq, k)
		}
	}
	rs, err := tx.Query(ctx, `select id, key, name, locked from hub.roles where key = any($1)`, uniq)
	if err != nil {
		return err
	}
	var ids []uuid.UUID
	found := map[string]bool{}
	for rs.Next() {
		var id uuid.UUID
		var k, n string
		var locked bool
		if err := rs.Scan(&id, &k, &n, &locked); err != nil {
			rs.Close()
			return err
		}
		if locked {
			rs.Close()
			return httpx.Err(409, "locked_role", n+" already has every right, so it does not need a group.")
		}
		found[k] = true
		ids = append(ids, id)
	}
	rs.Close()
	for _, k := range uniq {
		if !found[k] {
			return httpx.Bad("One of the roles does not exist.")
		}
	}
	if _, err := tx.Exec(ctx, `delete from hub.role_group_roles where group_id=$1`, group); err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := tx.Exec(ctx, `insert into hub.role_group_roles (org_id, group_id, role_id) values ($1,$2,$3)`, org, group, id); err != nil {
			return err
		}
	}
	return s.A.Sign.Seal(ctx, tx, "role_group_roles", "group_id=$1", group)
}

func (s *Service) CreateGroup(ctx context.Context, p *tenant.Principal, in GroupIn) (string, error) {
	name, ok := clean(in.Name, 60)
	if !ok {
		return "", httpx.Bad("Give the group a name (up to 60 characters).")
	}
	desc := ""
	if in.Desc != nil {
		desc = strings.TrimSpace(*in.Desc)
	}
	if len(desc) > 300 {
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
		var dup bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.role_groups where lower(name)=lower($1))`, name).Scan(&dup); err != nil {
			return err
		}
		if dup {
			return httpx.Err(409, "name_taken", "A group called "+name+" already exists.")
		}
		key = base
		for i := 2; ; i++ {
			var n int
			if err := tx.QueryRow(ctx, `select count(*) from hub.role_groups where key=$1`, key).Scan(&n); err != nil {
				return err
			}
			if n == 0 {
				break
			}
			key = fmt.Sprintf("%s-%d", base, i)
		}
		id := app.NewID()
		if _, err := tx.Exec(ctx, `insert into hub.role_groups (id, org_id, key, name, description) values ($1,$2,$3,$4,$5)`, id, p.Org, key, name, desc); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "role_groups", "id=$1", id); err != nil {
			return err
		}
		if in.Roles != nil {
			if err := s.setGroupRoles(ctx, tx, p.Org, id, in.Roles); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Created the user group "+name, "Roles", p.IP, map[string]any{"group": key})
	})
	if err != nil {
		return "", err
	}
	s.rightsChanged(p.Org)
	return key, nil
}

func (s *Service) UpdateGroup(ctx context.Context, p *tenant.Principal, key string, in GroupIn) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var id uuid.UUID
		var cur string
		if err := tx.QueryRow(ctx, `select id, name from hub.role_groups where key=$1 for update`, key).Scan(&id, &cur); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		if name := strings.TrimSpace(in.Name); name != "" {
			if name, ok := clean(name, 60); !ok {
				return httpx.Bad("A group name is up to 60 characters.")
			} else if !strings.EqualFold(name, cur) {
				var dup bool
				if err := tx.QueryRow(ctx, `select exists (select 1 from hub.role_groups where lower(name)=lower($1) and id<>$2)`, name, id).Scan(&dup); err != nil {
					return err
				}
				if dup {
					return httpx.Err(409, "name_taken", "A group called "+name+" already exists.")
				}
				if _, err := tx.Exec(ctx, `update hub.role_groups set name=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, id, name); err != nil {
					return err
				}
			}
		}
		if in.Desc != nil {
			d := strings.TrimSpace(*in.Desc)
			if len(d) > 300 {
				return httpx.Bad("That description is too long.")
			}
			if _, err := tx.Exec(ctx, `update hub.role_groups set description=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, id, d); err != nil {
				return err
			}
		}
		if err := s.A.Sign.Seal(ctx, tx, "role_groups", "id=$1", id); err != nil {
			return err
		}
		if in.Roles != nil {
			if err := s.setGroupRoles(ctx, tx, p.Org, id, in.Roles); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Changed the user group "+cur, "Roles", p.IP, map[string]any{"group": key})
	})
	if err == nil {
		s.rightsChanged(p.Org)
	}
	return err
}

func (s *Service) DeleteGroup(ctx context.Context, p *tenant.Principal, key string) error {
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var id uuid.UUID
		var name string
		if err := tx.QueryRow(ctx, `select id, name from hub.role_groups where key=$1 for update`, key).Scan(&id, &name); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		for _, q := range []string{`delete from hub.role_group_rights where group_id=$1`, `delete from hub.role_group_roles where group_id=$1`, `delete from hub.role_groups where id=$1`} {
			if _, err := tx.Exec(ctx, q, id); err != nil {
				return err
			}
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Deleted the user group "+name, "Roles", p.IP, map[string]any{"group": key})
	})
	if err == nil {
		s.rightsChanged(p.Org)
	}
	return err
}

// SetGroupRight gives a right to a group or takes it away; it goes through the same approval flow as a role's rights.
func (s *Service) SetGroupRight(ctx context.Context, p *tenant.Principal, group string, c rightChange) (*approvals.Request, bool, error) {
	c.Role, c.Person = group, uuid.Nil
	if err := s.validate(ctx, p.Org, c, false); err != nil {
		return nil, false, err
	}
	var req *approvals.Request
	var applied bool
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var name string
		if err := tx.QueryRow(ctx, `select name from hub.role_groups where key=$1`, group).Scan(&name); errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		verb := "allow"
		if !*c.Allowed {
			verb = "remove"
		}
		var err error
		req, applied, err = s.Ap.Submit(ctx, tx, p, "f8", "rights.group", fmt.Sprintf("Change rights of user group: %s · %s · %s (%s)", name, c.Module, c.Right, verb), "", c)
		return err
	})
	return req, applied, err
}

func (s *Service) applyGroupRight(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var c rightChange
	if err := json.Unmarshal(a.Payload, &c); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `insert into hub.role_group_rights (org_id, group_id, module, right_name, allowed)
		select $5::uuid, g.id, $2::text, $3::text, $4::boolean from hub.role_groups g where g.key=$1
		on conflict (org_id, group_id, module, right_name) do update set allowed=excluded.allowed, `+integrity.BumpOn("role_group_rights"), c.Role, c.Module, c.Right, *c.Allowed, a.Org)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httpx.Err(409, "gone", "That user group no longer exists.")
	}
	if err := s.A.Sign.Seal(ctx, tx, "role_group_rights", "group_id=(select id from hub.role_groups where key=$1) and module=$2 and right_name=$3", c.Role, c.Module, c.Right); err != nil {
		return err
	}
	s.rightsChanged(a.Org)
	return nil
}

func (s *Service) groupRoutes(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	r.Get("/groups", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := s.Groups(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"groups": l})
		return nil
	}))
	r.Post("/groups", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in GroupIn
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		key, err := s.CreateGroup(r.Context(), p, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"id": key})
		return nil
	}))
	r.Get("/groups/{key}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		d, err := s.Group(r.Context(), p, chi.URLParam(r, "key"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, d)
		return nil
	}))
	r.Patch("/groups/{key}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in GroupIn
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		if err := s.UpdateGroup(r.Context(), p, chi.URLParam(r, "key"), in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Delete("/groups/{key}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		if err := s.DeleteGroup(r.Context(), p, chi.URLParam(r, "key")); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Put("/groups/{key}/rights", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in rightChange
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		req, applied, err := s.SetGroupRight(r.Context(), p, chi.URLParam(r, "key"), in)
		if err != nil {
			return err
		}
		reply(w, applied, req)
		return nil
	}))
}
