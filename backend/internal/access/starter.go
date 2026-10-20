package access

import (
	"context"
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// The starter pack is what a real organization can start from: ready-made roles (Admin, Head of department, IT support,
// Finance approver, Staff, Guest, and the sector, branch and unit managers) with sensible rights, and Staff as the default role.
// It is applied on request with `bootstrap org` and only adds what is missing; the organization changes it freely afterwards on
// the Roles page. Approval levels are left at none: each organization decides who must approve what.
// It is kept apart from the demo pack (demo.go) so changing demo data can never change what a real organization receives.
//
//go:embed starter_roles.json
var starterRaw []byte

var (
	starterOnce sync.Once
	starter     Pack
)

func StarterPack() *Pack {
	starterOnce.Do(func() {
		if err := json.Unmarshal(starterRaw, &starter); err != nil {
			panic("access: bad starter roles: " + err.Error())
		}
	})
	return &starter
}

// InstallStarterRoles adds the starter roles an organization does not have yet. Safe to run again.
func InstallStarterRoles(ctx context.Context, tx pgx.Tx, org uuid.UUID, newID func() uuid.UUID, seal func(table, where string, args ...any) error) (map[string]uuid.UUID, error) {
	return InstallPack(ctx, tx, org, StarterPack(), newID, seal)
}
