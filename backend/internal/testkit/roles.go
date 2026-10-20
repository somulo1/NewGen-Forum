package testkit

import (
	"context"
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
)

// A sample set of roles with what each may do, for tests only. A real organization starts with its owner role and either builds
// its own roles or applies the starter pack (access.StarterPack, `bootstrap org -roles starter`).
//
//go:embed test_roles.json
var testRolesRaw []byte

var (
	testRolesOnce sync.Once
	testRoles     access.Pack
)

func TestRoles() *access.Pack {
	testRolesOnce.Do(func() {
		if err := json.Unmarshal(testRolesRaw, &testRoles); err != nil {
			panic("testkit: bad test roles: " + err.Error())
		}
	})
	return &testRoles
}

func installTestRoles(ctx context.Context, tx pgx.Tx, org uuid.UUID, newID func() uuid.UUID, seal func(table, where string, args ...any) error) error {
	_, err := access.InstallPack(ctx, tx, org, TestRoles(), newID, seal)
	return err
}
