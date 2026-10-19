package org

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/chat"
	"hub/internal/tenant"
)

// Moving and merging parts of the structure. People carry their whole placement (sector, branch, unit, department), so
// after either change everyone underneath is re-placed from their most specific level upwards, and joins or leaves the
// walls that follow from it, in the same transaction.

// allowedParent reports whether a unit of `kind` may sit under a unit of `parentKind`.
func allowedParent(kind, parentKind string) bool {
	for _, k := range parentKinds[kind] {
		if k == parentKind {
			return true
		}
	}
	return false
}

// MoveUnit puts a sector, branch, unit or department under another parent (or at the top level when parent is nil).
func (s *Service) MoveUnit(ctx context.Context, p *tenant.Principal, id uuid.UUID, parent *uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		var kind, name string
		var cur *uuid.UUID
		if err := tx.QueryRow(ctx, `select kind, name, parent_id from hub.org_units where id=$1 for update`, id).Scan(&kind, &name, &cur); err != nil {
			return httpx.ErrNotFound
		}
		dest := "the top level"
		if parent != nil {
			if *parent == id {
				return httpx.Bad("A " + kind + " cannot sit under itself.")
			}
			var pk, pn string
			if err := tx.QueryRow(ctx, `select kind, name from hub.org_units where id=$1`, *parent).Scan(&pk, &pn); err != nil {
				return httpx.Bad("That parent does not exist.")
			}
			if !allowedParent(kind, pk) {
				return httpx.Bad("A " + kind + " cannot sit under a " + pk + ".")
			}
			// kinds only nest downwards (sector > branch > unit > department), but check anyway: no loops
			var loop bool
			if err := tx.QueryRow(ctx, `with recursive up(id, parent_id) as (select id, parent_id from hub.org_units where id=$1
				union all select u.id, u.parent_id from hub.org_units u join up on u.id = up.parent_id) select exists (select 1 from up where id=$2)`, *parent, id).Scan(&loop); err != nil {
				return err
			}
			if loop {
				return httpx.Bad("A " + kind + " cannot sit under one of its own parts.")
			}
			dest = pk + " " + pn
		}
		if (cur == nil && parent == nil) || (cur != nil && parent != nil && *cur == *parent) {
			return nil
		}
		if _, err := tx.Exec(ctx, `update hub.org_units set parent_id=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, id, parent); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "org_units", "id=$1", id); err != nil {
			return err
		}
		if err := s.replace(ctx, tx, p.Org, []uuid.UUID{id}); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Moved "+kind+" "+name+" to "+dest, "Settings", p.IP, map[string]any{"unit": id, "parent": parent})
	})
}

// MergeUnits folds `from` (all of the same kind as `into`) into `into`: their parts, people, tickets, ticket categories
// and walls move over, then they are removed.
func (s *Service) MergeUnits(ctx context.Context, p *tenant.Principal, into uuid.UUID, from []uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "settings", "Edit organization profile"); err != nil {
			return err
		}
		var kind, name string
		if err := tx.QueryRow(ctx, `select kind, name from hub.org_units where id=$1 for update`, into).Scan(&kind, &name); err != nil {
			return httpx.Bad("Choose what to merge into.")
		}
		seen := map[uuid.UUID]bool{into: true}
		src := []uuid.UUID{}
		names := []string{}
		for _, f := range from {
			if seen[f] {
				continue
			}
			seen[f] = true
			var k, n string
			if err := tx.QueryRow(ctx, `select kind, name from hub.org_units where id=$1 for update`, f).Scan(&k, &n); err != nil {
				return httpx.Bad("One of the places to merge does not exist any more.")
			}
			if k != kind {
				return httpx.Bad("Only the same kind can be merged: " + n + " is a " + k + ", " + name + " is a " + kind + ".")
			}
			src = append(src, f)
			names = append(names, n)
		}
		if len(src) == 0 {
			return httpx.Bad("Choose at least one other " + kind + " to merge.")
		}
		col := kind + "_id"
		steps := []struct{ q, table, where string }{
			{`update hub.org_units set parent_id=$1, updated_at=now(), ` + integrity.Bump + ` where parent_id = any($2)`, "org_units", "parent_id=$1"},
			{`update hub.people set ` + col + `=$1, ` + integrity.Bump + ` where ` + col + ` = any($2)`, "people", col + "=$1"},
			{`update hub.tickets set to_ref=$1, ` + integrity.Bump + ` where to_ref = any($2)`, "tickets", "to_ref=$1"},
			{`update hub.conversations set audience_ref=$1, ` + integrity.Bump + ` where audience_ref = any($2)`, "conversations", "audience_ref=$1"},
		}
		for _, st := range steps {
			if _, err := tx.Exec(ctx, st.q, into, src); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, st.table, st.where, into); err != nil {
				return err
			}
		}
		// ticket categories: the merged-in ones move over, except where the target already has one by that name
		if _, err := tx.Exec(ctx, `delete from hub.ticket_categories c where c.unit_ref = any($2)
			and exists (select 1 from hub.ticket_categories t where t.unit_ref=$1 and lower(t.name)=lower(c.name))`, into, src); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.ticket_categories set unit_ref=$1, `+integrity.Bump+` where unit_ref = any($2)`, into, src); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "ticket_categories", "unit_ref=$1", into); err != nil {
			return err
		}
		for _, q := range []string{`delete from hub.unit_targets where unit_id = any($1)`, `delete from hub.org_units where id = any($1)`} {
			if _, err := tx.Exec(ctx, q, src); err != nil {
				return err
			}
		}
		if err := s.replace(ctx, tx, p.Org, []uuid.UUID{into}); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Merged %s %s into %s", kind, strings.Join(names, ", "), name), "Settings", p.IP, map[string]any{"into": into, "from": src})
	})
}

// replace re-places everyone at or under the given units from their most specific level upwards, then brings their wall
// memberships in line with where they now are.
func (s *Service) replace(ctx context.Context, tx pgx.Tx, org uuid.UUID, roots []uuid.UUID) error {
	rs, err := tx.Query(ctx, `
		with recursive sub(id) as (select unnest($1::uuid[]) union select u.id from hub.org_units u join sub on u.parent_id = sub.id),
		who as (select p.id, coalesce(p.department_id, p.unit_id, p.branch_id, p.sector_id) as leaf from hub.people p
		        where p.department_id in (select id from sub) or p.unit_id in (select id from sub) or p.branch_id in (select id from sub) or p.sector_id in (select id from sub)),
		up(person, id, kind, parent_id) as (
			select w.id, u.id, u.kind, u.parent_id from who w join hub.org_units u on u.id = w.leaf
			union all select up.person, u.id, u.kind, u.parent_id from up join hub.org_units u on u.id = up.parent_id),
		placed as (select person,
			max(id::text) filter (where kind='sector')::uuid s, max(id::text) filter (where kind='branch')::uuid b,
			max(id::text) filter (where kind='unit')::uuid u, max(id::text) filter (where kind='department')::uuid d
			from up group by person)
		update hub.people p set sector_id=pl.s, branch_id=pl.b, unit_id=pl.u, department_id=pl.d, `+integrity.Bump+`
		  from placed pl where p.id = pl.person
		   and (p.sector_id, p.branch_id, p.unit_id, p.department_id) is distinct from (pl.s, pl.b, pl.u, pl.d)
		returning p.id, p.account_id`, roots)
	if err != nil {
		return err
	}
	var ids []uuid.UUID
	var accounts []uuid.UUID
	for rs.Next() {
		var id uuid.UUID
		var acct *uuid.UUID
		if err := rs.Scan(&id, &acct); err != nil {
			rs.Close()
			return err
		}
		ids = append(ids, id)
		if acct != nil {
			accounts = append(accounts, *acct)
		}
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	if len(ids) > 0 {
		if err := s.A.Sign.Seal(ctx, tx, "people", "id = any($1)", ids); err != nil {
			return err
		}
	}
	// people whose placement column was merged also need their walls checked, even if re-placing changed nothing more
	extra, err := tx.Query(ctx, `with recursive sub(id) as (select unnest($1::uuid[]) union select u.id from hub.org_units u join sub on u.parent_id = sub.id)
		select distinct account_id from hub.people where account_id is not null and status='active'
		  and (department_id in (select id from sub) or unit_id in (select id from sub) or branch_id in (select id from sub) or sector_id in (select id from sub))`, roots)
	if err != nil {
		return err
	}
	have := map[uuid.UUID]bool{}
	for _, a := range accounts {
		have[a] = true
	}
	for extra.Next() {
		var a uuid.UUID
		if err := extra.Scan(&a); err != nil {
			extra.Close()
			return err
		}
		if !have[a] {
			have[a] = true
			accounts = append(accounts, a)
		}
	}
	extra.Close()
	for _, a := range accounts {
		if err := chat.SyncAudience(ctx, tx, s.A.Sign, org, a); err != nil {
			return err
		}
	}
	return nil
}
