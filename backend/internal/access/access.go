// Package access is the starting list of modules (pages) and rights every new organization gets. It is only a starting point:
// the list is copied into each organization's own tables (hub.modules, hub.rights) where it can be changed, and roles are made
// by the organization itself. The server enforces the rights the code checks (they are marked builtin); a right added later is
// just a named permission the pages and approval flows can use.
package access

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

//go:embed catalog.json
var raw []byte

//go:embed demo_roles.json
var demoRaw []byte

type Module struct {
	Key    string   `json:"key"`
	Label  string   `json:"label"`
	Tint   string   `json:"tint"`
	Rights []string `json:"rights"`
}

// Flow is an approval point the code knows about (it submits a request under this key); how many levels it needs and who
// approves each level are the organization's choice.
type Flow struct {
	ID     string `json:"id"`
	Mod    string `json:"mod"`
	Action string `json:"action"`
	Levels int    `json:"levels"`
}

type Catalog struct {
	Modules []Module `json:"modules"`
	Flows   []Flow   `json:"flows"`
}

type Role struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Locked bool   `json:"locked"`
	Needs  string `json:"needs"`
	Scope  string `json:"scope"`
}

// Demo is a sample set of roles with what each may do. It is used only to fill demo data and tests, never for a real
// organization (which starts with one owner role and creates the rest itself).
type Demo struct {
	Roles    []Role                         `json:"roles"`
	Defaults map[string]map[string][]string `json:"defaults"` // role -> module -> allowed rights
	Flows    map[string]DemoFlow            `json:"flows"`    // flow key -> levels and the role that approves each level
}

type DemoFlow struct {
	Levels    int      `json:"levels"`
	Approvers []string `json:"approvers"` // role names; a name with no role is skipped (the owner can always stand in)
}

var (
	once, demoOnce sync.Once
	cat            Catalog
	demo           Demo
)

func Default() *Catalog {
	once.Do(func() {
		if err := json.Unmarshal(raw, &cat); err != nil {
			panic("access: bad catalog: " + err.Error())
		}
	})
	return &cat
}

func DemoRoles() *Demo {
	demoOnce.Do(func() {
		if err := json.Unmarshal(demoRaw, &demo); err != nil {
			panic("access: bad demo roles: " + err.Error())
		}
	})
	return &demo
}

// ApproveRight names the right that lets someone decide level n of a flow in a module.
func ApproveRight(level int) string { return "Approve level " + string(rune('0'+level)) }

// Kinds are the groups the Roles page sorts rights into.
var Kinds = []string{"view", "create", "edit", "delete", "export", "approve", "manage", "other"}

// KindOf guesses the group of a right from its name (only used when a right is first added to the catalog).
func KindOf(name string) string {
	n := strings.ToLower(name)
	has := func(words ...string) bool {
		for _, w := range words {
			if strings.Contains(n, w) {
				return true
			}
		}
		return false
	}
	switch {
	case has("add & remove"):
		return "manage"
	case strings.HasPrefix(n, "approve"):
		return "approve"
	case has("export"):
		return "export"
	case has("delete", "remove"):
		return "delete"
	case strings.HasPrefix(n, "view") || strings.HasPrefix(n, "read") || strings.HasPrefix(n, "see") || strings.HasPrefix(n, "open") || has("who read") || n == "join":
		return "view"
	case has("edit", "update", "move tasks", "resolve", "solve", "close", "assign"):
		return "edit"
	case has("manage", "configure", "set up", "settings", "moderate", "customize", "add & remove"):
		return "manage"
	case has("create", "post", "raise", "upload", "send", "publish", "nominate", "schedule", "invite", "start", "onboard", "add ", "comment", "share", "record", "private"):
		return "create"
	}
	return "other"
}

// Install copies the starting list into an organization's tables. It is safe to run again: rows that exist are left alone, and
// the rights the code relies on are added back if missing. seal signs the rows it wrote (see integrity.Signer.Seal).
func Install(ctx context.Context, tx pgx.Tx, org uuid.UUID, seal func(table, where string, args ...any) error) error {
	c := Default()
	var mods, rights int
	for i, m := range c.Modules {
		tag, err := tx.Exec(ctx, `insert into hub.modules (org_id, key, label, tint, sort, builtin) values ($1,$2,$3,$4,$5,true) on conflict do nothing`, org, m.Key, m.Label, m.Tint, i)
		if err != nil {
			return err
		}
		mods += int(tag.RowsAffected())
		for j, r := range m.Rights {
			tag, err := tx.Exec(ctx, `insert into hub.rights (org_id, module, name, label, kind, sort, builtin) values ($1,$2,$3,$3,$4,$5,true) on conflict do nothing`, org, m.Key, r, KindOf(r), j)
			if err != nil {
				return err
			}
			rights += int(tag.RowsAffected())
		}
	}
	if mods > 0 {
		if err := seal("modules", "org_id=$1", org); err != nil {
			return err
		}
	}
	if rights > 0 {
		if err := seal("rights", "org_id=$1", org); err != nil {
			return err
		}
	}
	return nil
}

// Installed reports whether the organization already has its catalog.
func Installed(ctx context.Context, tx pgx.Tx) (bool, error) {
	var n int
	if err := tx.QueryRow(ctx, `select count(*) from hub.rights where builtin`).Scan(&n); err != nil {
		return false, err
	}
	want := 0
	for _, m := range Default().Modules {
		want += len(m.Rights)
	}
	return n >= want, nil
}

// InstallDemoRoles gives an organization the sample roles (and what each may do) plus a default role. It is for demo data and
// tests only; a real organization makes its own roles. It returns the role ids by key.
func InstallDemoRoles(ctx context.Context, tx pgx.Tx, org uuid.UUID, newID func() uuid.UUID, seal func(table, where string, args ...any) error) (map[string]uuid.UUID, error) {
	d := DemoRoles()
	ids := map[string]uuid.UUID{}
	for _, r := range d.Roles {
		if r.Locked {
			continue // the owner role already exists
		}
		id := newID()
		ids[r.ID] = id
		scope := r.Scope
		if scope == "" {
			scope = "Whole organization"
		}
		if _, err := tx.Exec(ctx, `insert into hub.roles (id, org_id, key, name, description, scope, locked, needs_level) values ($1,$2,$3,$4,$5,$6,false,nullif($7,''))`, id, org, r.ID, r.Name, r.Desc, scope, r.Needs); err != nil {
			return nil, err
		}
	}
	if err := seal("roles", "org_id=$1", org); err != nil {
		return nil, err
	}
	for role, mods := range d.Defaults {
		id, ok := ids[role]
		if !ok {
			continue
		}
		for mod, rights := range mods {
			for _, r := range rights {
				if _, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed) values ($1,$2,$3,$4,true) on conflict do nothing`, org, id, mod, r); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := seal("role_rights", "org_id=$1", org); err != nil {
		return nil, err
	}
	// approval levels, and which sample role may decide each one
	for key, f := range d.Flows {
		var module string
		if err := tx.QueryRow(ctx, `select module from hub.approval_flows where key=$1`, key).Scan(&module); err != nil {
			continue
		}
		if _, err := tx.Exec(ctx, `update hub.approval_flows set levels=$2, updated_at=now(), ver = ver + 1, mac = hub.pending_mac() where key=$1`, key, f.Levels); err != nil {
			return nil, err
		}
		for lvl, who := range f.Approvers {
			if lvl >= f.Levels {
				break
			}
			for _, r := range d.Roles {
				if id, ok := ids[r.ID]; ok && strings.EqualFold(r.Name, who) {
					if _, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed) values ($1,$2,$3,$4,true) on conflict (org_id, role_id, module, right_name) do update set allowed=true`, org, id, module, ApproveRight(lvl+1)); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	if err := seal("approval_flows", "org_id=$1", org); err != nil {
		return nil, err
	}
	if err := seal("role_rights", "org_id=$1", org); err != nil {
		return nil, err
	}
	if id, ok := ids["staff"]; ok {
		if _, err := tx.Exec(ctx, `insert into hub.default_role (org_id, role_id) values ($1,$2) on conflict do nothing`, org, id); err != nil {
			return nil, err
		}
		if err := seal("default_role", "org_id=$1", org); err != nil {
			return nil, err
		}
	}
	return ids, nil
}
