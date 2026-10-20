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

var (
	once sync.Once
	cat  Catalog
)

func Default() *Catalog {
	once.Do(func() {
		if err := json.Unmarshal(raw, &cat); err != nil {
			panic("access: bad catalog: " + err.Error())
		}
	})
	return &cat
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
	return LinkLockedRoles(ctx, tx, org, seal)
}

// LinkLockedRoles ticks every right of the catalog for each locked role (the owner), as stored rows. A locked role holds every
// right anyway; the stored rows make that visible and queryable (the Roles page, reports, the database itself), and a right
// added to the catalog later is linked to it the next time this runs. Safe to run again.
func LinkLockedRoles(ctx context.Context, tx pgx.Tx, org uuid.UUID, seal func(table, where string, args ...any) error) error {
	tag, err := tx.Exec(ctx, `insert into hub.role_rights (org_id, role_id, module, right_name, allowed)
		select $1, r.id, g.module, g.name, true from hub.roles r cross join hub.rights g where r.locked
		on conflict (org_id, role_id, module, right_name) do update set allowed = true, ver = hub.role_rights.ver + 1, mac = hub.pending_mac()
		where not hub.role_rights.allowed`, org)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return seal("role_rights", "org_id=$1", org)
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
