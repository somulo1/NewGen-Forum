// Package authz answers "may this person do this?" on the server. A personal override beats everything; otherwise the
// person's role decides (hub.role_right: the owner role holds a right until it is switched off for it). With neither, no.
package authz

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
)

// Can must run inside a tenant transaction (db.WithOrg), so row-level security scopes the lookup.
func Can(ctx context.Context, tx pgx.Tx, person uuid.UUID, module, right string) (bool, error) {
	var roleAllowed, personAllowed *bool
	err := tx.QueryRow(ctx, `
		select case when pr.role_id is null then null else hub.role_right(pr.role_id, $2, $3) end,
		       (select x.allowed from hub.person_rights x where x.person_id = $1 and x.module=$2 and x.right_name=$3)
		from hub.people p
		left join hub.person_roles pr on pr.person_id = p.id
		where p.id = $1 and p.status = 'active'`, person, module, right).Scan(&roleAllowed, &personAllowed)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	switch {
	case personAllowed != nil:
		return *personAllowed, nil
	case roleAllowed != nil:
		return *roleAllowed, nil
	}
	return false, nil
}

// Require returns a 403 unless the person holds the right.
func Require(ctx context.Context, tx pgx.Tx, person uuid.UUID, module, right string) error {
	ok, err := Can(ctx, tx, person, module, right)
	if err != nil {
		return err
	}
	if !ok {
		return httpx.ErrForbidden
	}
	return nil
}

// Effective returns every right the person holds, by module: the role's setting with personal overrides applied.
func Effective(ctx context.Context, tx pgx.Tx, person uuid.UUID, all map[string][]string) (map[string][]string, error) {
	var locked bool
	var role *uuid.UUID
	if err := tx.QueryRow(ctx, `select coalesce(r.locked,false), pr.role_id from hub.people p left join hub.person_roles pr on pr.person_id=p.id left join hub.roles r on r.id=pr.role_id where p.id=$1 and p.status='active'`, person).Scan(&locked, &role); err != nil {
		if err == pgx.ErrNoRows {
			return map[string][]string{}, nil
		}
		return nil, err
	}
	// the role's settings, then the person's own overrides on top; the owner role holds what was never switched off for it
	on := map[string]bool{}
	read := func(q string, id uuid.UUID) error {
		rs, err := tx.Query(ctx, q, id)
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
			on[m+"\x00"+r] = a
		}
		return rs.Err()
	}
	if role != nil {
		if err := read(`select module, right_name, allowed from hub.role_rights where role_id=$1`, *role); err != nil {
			return nil, err
		}
	}
	if err := read(`select module, right_name, allowed from hub.person_rights where person_id=$1`, person); err != nil {
		return nil, err
	}
	out := map[string][]string{}
	for m, rights := range all {
		for _, r := range rights {
			if v, set := on[m+"\x00"+r]; v || (!set && locked) {
				out[m] = append(out[m], r)
			}
		}
	}
	return out, nil
}

// PeopleWith lists the active people who hold a right (through their role or a personal grant).
func PeopleWith(ctx context.Context, tx pgx.Tx, module, right string) ([]uuid.UUID, error) {
	rs, err := tx.Query(ctx, `
		select p.id from hub.people p
		left join hub.person_roles pr on pr.person_id = p.id
		where p.status = 'active' and coalesce((select x.allowed from hub.person_rights x where x.person_id = p.id and x.module=$1 and x.right_name=$2),
		              case when pr.role_id is null then null else hub.role_right(pr.role_id, $1, $2) end, false)
		limit 200`, module, right)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	var out []uuid.UUID
	for rs.Next() {
		var id uuid.UUID
		if err := rs.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rs.Err()
}

// CanOwn answers "may this person do this to this item?": with the general right they may do it to anything; with the
// matching "own" right (for example "Edit own posts") only to items they made. Both are rights the organization gives to roles.
func CanOwn(ctx context.Context, tx pgx.Tx, person uuid.UUID, module, anyRight, ownRight string, mine bool) (bool, error) {
	if anyRight != "" {
		if ok, err := Can(ctx, tx, person, module, anyRight); err != nil || ok {
			return ok, err
		}
	}
	if mine && ownRight != "" {
		return Can(ctx, tx, person, module, ownRight)
	}
	return false, nil
}

// RequireOwn is CanOwn that returns a 403 when the answer is no.
func RequireOwn(ctx context.Context, tx pgx.Tx, person uuid.UUID, module, anyRight, ownRight string, mine bool) error {
	ok, err := CanOwn(ctx, tx, person, module, anyRight, ownRight, mine)
	if err != nil {
		return err
	}
	if !ok {
		return httpx.ErrForbidden
	}
	return nil
}
