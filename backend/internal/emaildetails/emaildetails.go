// Package emaildetails decides whether emails may show the text of a ticket or message. The text is end-to-end encrypted, so
// it comes from the sender's app inside the request that causes the email, is written into the email and is not stored.
package emaildetails

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/orgsettings"
)

// Allowed reports whether the organization lets emails carry details.
func Allowed(ctx context.Context, a *app.App, tx pgx.Tx, org uuid.UUID) bool {
	sec := orgsettings.DefaultSecurity()
	if err := orgsettings.Get(ctx, a, tx, org, "security", &sec); err != nil {
		return false
	}
	return sec.EmailDetails
}

// Clip cuts text to n characters (not bytes).
func Clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
