// Package upkeep brings every organization's data up to date when the API starts: changes that need the system's keys (sealed
// rows, encrypted settings), so a database migration cannot make them. Each step does nothing once it has been done.
package upkeep

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/widgetaccess"
)

const (
	oldOwnerNote = "Full control of Hub. This role cannot be limited."
	ownerNote    = "Full control of Hub. Its rights can be changed; it always keeps managing roles and rights."
)

// Run applies every step to every organization.
func Run(ctx context.Context, a *app.App) {
	var orgs []uuid.UUID
	err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where status <> 'deleted'`)
		if err != nil {
			return err
		}
		orgs, err = pgx.CollectRows(rs, pgx.RowTo[uuid.UUID])
		return err
	})
	if err != nil {
		slog.Error("upkeep", "err", err)
		return
	}
	for _, org := range orgs {
		changed := false
		if err := a.DB.WithOrg(ctx, org, func(tx pgx.Tx) error {
			folded, err := widgetaccess.FoldOrg(ctx, a, tx, org)
			if err != nil {
				return err
			}
			noted, err := ownerNoteOrg(ctx, a, tx, org)
			changed = folded || noted
			return err
		}); err != nil {
			slog.Error("upkeep", "org", org, "err", err)
		} else if changed {
			a.Publish(ctx, realtime.Event{Type: "rights.changed", Channel: realtime.OrgChannel(org), Org: org.String()})
		}
	}
}

// ownerNoteOrg rewrites the owner role's original description, which said its rights could not be changed.
func ownerNoteOrg(ctx context.Context, a *app.App, tx pgx.Tx, org uuid.UUID) (bool, error) {
	tag, err := tx.Exec(ctx, `update hub.roles set description=$1, `+integrity.Bump+` where locked and description=$2`, ownerNote, oldOwnerNote)
	if err != nil || tag.RowsAffected() == 0 {
		return false, err
	}
	return true, a.Sign.Seal(ctx, tx, "roles", "org_id=$1 and locked", org)
}
