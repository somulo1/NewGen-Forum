// Package modules lists every widget module in the order the API mounts them.
package modules

import (
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
	return []server.Module{
		org.New(a, au),
		ap,
		people.New(a, au, ap),
		roles.New(a, ap),
		live.New(a),
		keys.New(a),
		chat.New(a),
		calls.New(a),
		tickets.New(a),
		projects.New(a, ap),
		comms.New(a, ap),
		files.New(a, ap),
		security.New(a),
		settings.New(a),
		notify.New(a),
		tones.New(a),
	}
}
