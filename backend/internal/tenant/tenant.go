// Package tenant carries "who is calling, inside which organization" through a request.
package tenant

import (
	"context"

	"github.com/google/uuid"

	"hub/internal/httpx"
)

type Principal struct {
	Account uuid.UUID
	Org     uuid.UUID // uuid.Nil until the caller belongs to an organization
	Person  uuid.UUID
	Session uuid.UUID
	Name    string
	IP      string
}

type key struct{}

func With(ctx context.Context, p *Principal) context.Context { return context.WithValue(ctx, key{}, p) }

func From(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(key{}).(*Principal)
	return p, ok
}

// Must returns the caller or the 401 error to return.
func Must(ctx context.Context) (*Principal, error) {
	p, ok := From(ctx)
	if !ok {
		return nil, httpx.ErrUnauthorized
	}
	return p, nil
}

// MustOrg additionally requires the caller to be inside an organization.
func MustOrg(ctx context.Context) (*Principal, error) {
	p, err := Must(ctx)
	if err != nil {
		return nil, err
	}
	if p.Org == uuid.Nil {
		return nil, httpx.Err(403, "no_organization", "Join or create an organization first.")
	}
	return p, nil
}
