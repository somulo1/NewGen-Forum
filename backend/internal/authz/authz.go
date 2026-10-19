// Package authz answers "may this person do this?" on the server. A personal override beats the role's setting;
// with neither, the answer is no. Locked roles (super admin) are always yes.
package authz

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
)

// Can must run inside a tenant transaction (db.WithOrg), so row-level security scopes the lookup.
func Can(ctx context.Context, tx pgx.Tx, person uuid.UUID, module, right string) (bool, error) {
	var locked bool
	var roleAllowed, personAllowed *bool
	err := tx.QueryRow(ctx, `
		select coalesce(r.locked, false),
		       (select rr.allowed from hub.role_rights rr where rr.role_id = pr.role_id and rr.module=$2 and rr.right_name=$3),
		       (select x.allowed from hub.person_rights x where x.person_id = $1 and x.module=$2 and x.right_name=$3)
		from hub.people p
		left join hub.person_roles pr on pr.person_id = p.id
		left join hub.roles r on r.id = pr.role_id
		where p.id = $1 and p.status = 'active'`, person, module, right).Scan(&locked, &roleAllowed, &personAllowed)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	switch {
	case locked:
		return true, nil
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
	if locked {
		return all, nil
	}
	on := map[string]bool{}
	if role != nil {
		rs, err := tx.Query(ctx, `select module, right_name from hub.role_rights where role_id=$1 and allowed`, *role)
		if err != nil {
			return nil, err
		}
		for rs.Next() {
			var m, r string
			if err := rs.Scan(&m, &r); err != nil {
				rs.Close()
				return nil, err
			}
			on[m+"\x00"+r] = true
		}
		rs.Close()
	}
	rs, err := tx.Query(ctx, `select module, right_name, allowed from hub.person_rights where person_id=$1`, person)
	if err != nil {
		return nil, err
	}
	for rs.Next() {
		var m, r string
		var a bool
		if err := rs.Scan(&m, &r, &a); err != nil {
			rs.Close()
			return nil, err
		}
		on[m+"\x00"+r] = a
	}
	rs.Close()
	out := map[string][]string{}
	for m, rights := range all {
		for _, r := range rights {
			if on[m+"\x00"+r] {
				out[m] = append(out[m], r)
			}
		}
	}
	return out, nil
}

// PeopleWith lists the active people who hold a right (through their role, a personal grant, or a protected role).
func PeopleWith(ctx context.Context, tx pgx.Tx, module, right string) ([]uuid.UUID, error) {
	rs, err := tx.Query(ctx, `
		select p.id from hub.people p
		left join hub.person_roles pr on pr.person_id = p.id
		left join hub.roles r on r.id = pr.role_id
		where p.status = 'active' and (
		  coalesce(r.locked,false)
		  or coalesce((select x.allowed from hub.person_rights x where x.person_id = p.id and x.module=$1 and x.right_name=$2),
		              (select rr.allowed from hub.role_rights rr where rr.role_id = pr.role_id and rr.module=$1 and rr.right_name=$2), false))
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
