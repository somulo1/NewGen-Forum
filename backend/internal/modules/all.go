// Package modules lists every widget module in the order the API mounts them.
package modules

import (
	"context"

	"hub/internal/app"
	"hub/internal/modules/approvals"
	"hub/internal/modules/auth"
	"hub/internal/modules/calls"
	"hub/internal/modules/chat"
	"hub/internal/modules/comms"
	"hub/internal/modules/files"
	"hub/internal/modules/keys"
	"hub/internal/modules/live"
	"hub/internal/modules/org"
	"hub/internal/modules/people"
	"hub/internal/modules/projects"
	"hub/internal/modules/roles"
	"hub/internal/modules/security"
	"hub/internal/modules/settings"
	"hub/internal/modules/tickets"
	"hub/internal/modules/tones"
	"hub/internal/notify"
	"hub/internal/server"
)

func All(a *app.App, au *auth.Service) []server.Module {
	ap := approvals.New(a)
	fs := files.New(a, ap)
	ch := chat.New(a)
	ch.Purge = func(ctx context.Context) { _, _, _ = fs.PurgeDue(ctx, 200) } // a deleted post's files leave storage at once
	return []server.Module{
		org.New(a, au),
		ap,
		people.New(a, au, ap),
		roles.New(a, ap),
		live.New(a),
		keys.New(a),
		ch,
		calls.New(a),
		tickets.New(a),
		projects.New(a, ap),
		comms.New(a, ap),
		fs,
		security.New(a),
		settings.New(a),
		notify.New(a),
		tones.New(a),
	}
}
