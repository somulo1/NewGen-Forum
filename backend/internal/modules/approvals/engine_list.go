package approvals

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

func (e *Engine) list(ctx context.Context, tx pgx.Tx, p *tenant.Principal, where string, args []any, limit int) ([]*Request, error) {
	dek, _, err := e.A.OrgDEK(ctx, tx, p.Org)
	if err != nil {
		return nil, err
	}
	rs, err := tx.Query(ctx, `
		select r.id, r.seq, r.flow_key, r.kind, r.title_enc, r.note_enc, r.requester, p.first_name || ' ' || p.last_name, r.level, r.status, r.created_at, r.decided_at,
		       f.levels, f.module, case when r.kind = 'rights.batch' then r.payload_enc end
		from hub.approval_requests r join hub.people p on p.id = r.requester join hub.approval_flows f on f.key = r.flow_key
		where `+where+` order by r.created_at desc, r.id desc limit `+fmt.Sprint(limit), args...)
	if err != nil {
		return nil, err
	}
	out := []*Request{}
	var modules []string
	var seqs []int
	for rs.Next() {
		x := &Request{History: []Step{}}
		var tEnc, nEnc, pEnc []byte
		var mod string
		var sq int
		if err := rs.Scan(&x.ID, &sq, &x.Flow, &x.Kind, &tEnc, &nEnc, &x.Requester, &x.By, &x.Level, &x.Status, &x.CreatedAt, &x.DecidedAt, &x.Levels, &mod, &pEnc); err != nil {
			rs.Close()
			return nil, err
		}
		x.Levels = max(1, x.Levels)
		if len(pEnc) > 0 && itemKinds[x.Kind] {
			if raw, err := crypto.FieldOpen(dek, fieldCtx(p.Org, x.ID, "payload_enc"), pEnc); err == nil {
				x.Items = itemsOf(raw)
			}
		}
		out = append(out, x)
		modules = append(modules, mod)
		seqs = append(seqs, sq)
		if t, err := crypto.FieldOpen(dek, fieldCtx(p.Org, x.ID, "title_enc"), tEnc); err == nil {
			x.Title = string(t)
		}
		if len(nEnc) > 0 {
			if n, err := crypto.FieldOpen(dek, fieldCtx(p.Org, x.ID, "note_enc"), nEnc); err == nil {
				x.Note = string(n)
			}
		}
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return nil, err
	}
	// every request's steps in one query, and each "waiting for" / "may decide" answer worked out once per kind of request
	ids := make([]uuid.UUID, len(out))
	byID := make(map[uuid.UUID]*Request, len(out))
	for i, x := range out {
		ids[i] = x.ID
		byID[x.ID] = x
	}
	if len(ids) > 0 {
		ss, err := tx.Query(ctx, `select s.request_id, s.level, p.first_name || ' ' || p.last_name, s.decision, s.comment_enc, s.created_at from hub.approval_steps s join hub.people p on p.id = s.approver where s.request_id = any($1) order by s.request_id, s.level`, ids)
		if err != nil {
			return nil, err
		}
		for ss.Next() {
			var rid uuid.UUID
			var st Step
			var cEnc []byte
			if err := ss.Scan(&rid, &st.Level, &st.Who, &st.Decision, &cEnc, &st.At); err != nil {
				ss.Close()
				return nil, err
			}
			x := byID[rid]
			if x == nil {
				continue
			}
			if len(cEnc) > 0 {
				if c, err := crypto.FieldOpen(dek, crypto.Context{Org: p.Org.String(), Table: "approval_steps", Column: "comment_enc", RowID: fmt.Sprintf("%s/%d", x.ID, st.Level)}, cEnc); err == nil {
					st.Comment = string(c)
				}
			}
			x.History = append(x.History, st)
		}
		ss.Close()
		if err := ss.Err(); err != nil {
			return nil, err
		}
	}
	type kind struct {
		flow, module string
		level        int
		mine         bool
	}
	waits := map[kind]string{}
	may := map[kind]bool{}
	for i, x := range out {
		x.Ref = ref(seqs[i])
		if x.Status != "Pending" {
			continue
		}
		k := kind{x.Flow, modules[i], x.Level, x.Requester == p.Person}
		w, ok := waits[k]
		if !ok {
			w = e.waiting(ctx, tx, x.Flow, modules[i], x.Level)
			waits[k] = w
		}
		x.Waiting = w
		c, ok := may[k]
		if !ok {
			c, _ = e.mayDecide(ctx, tx, p.Person, x.Requester, &Flow{Module: modules[i]}, x.Level)
			may[k] = c
		}
		x.CanDecide = c
	}
	return out, nil
}

// List returns requests newest first. mine=false lists the ones this person may decide (waiting) or all they may see.
func (e *Engine) List(ctx context.Context, p *tenant.Principal, status string, limit int) ([]*Request, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	out := []*Request{}
	err := e.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		where, args := "true", []any{}
		if status != "" {
			where, args = "r.status = $1", []any{status}
		}
		// People see requests they raised or may decide; people with the approvals right see everything.
		all, err := authz.Can(ctx, tx, p.Person, "roles", "Configure approvals")
		if err != nil {
			return err
		}
		if !all {
		}
		out, err = e.list(ctx, tx, p, where, args, limit)
		if err != nil || all {
			return err
		}
		keep := out[:0]
		for _, x := range out {
			if x.Requester == p.Person || x.CanDecide || len(x.History) > 0 && historyHas(x.History, p.Name) {
				keep = append(keep, x)
			}
		}
		out = keep
		return nil
	})
	return out, err
}

func historyHas(h []Step, name string) bool {
	for _, s := range h {
		if s.Who == name {
			return true
		}
	}
	return false
}

// Cancel lets the requester withdraw a pending request.
func (e *Engine) Cancel(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	return e.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `update hub.approval_requests set status='Cancelled', decided_at=now(), updated_at=now(), `+integrity.Bump+` where id=$1 and requester=$2 and status='Pending'`, id, p.Person)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		if err := e.A.Sign.Seal(ctx, tx, "approval_requests", "id=$1", id); err != nil {
			return err
		}
		return e.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Withdrew an approval request", "Approvals", p.IP, map[string]any{"id": id})
	})
}

// Flows lists the approval flows; SetFlow changes one (it needs the "Configure approvals" right).
func (e *Engine) Flows(ctx context.Context, p *tenant.Principal) ([]Flow, error) {
	out := []Flow{}
	err := e.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select key, module, action, levels from hub.approval_flows order by length(key), key`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var f Flow
			if err := rs.Scan(&f.Key, &f.Module, &f.Action, &f.Levels); err != nil {
				return err
			}
			out = append(out, f)
		}
		if err := rs.Err(); err != nil {
			return err
		}
		rs.Close()
		for i := range out { // names of who decides each level (a second query, so the rows above are closed first)
			out[i].Approvers = []string{}
			for l := 1; l <= max(1, out[i].Levels); l++ {
				out[i].Approvers = append(out[i].Approvers, e.waiting(ctx, tx, out[i].Key, out[i].Module, l))
			}
			out[i].Designees = []Designee{}
		}
		ds, err := tx.Query(ctx, `select d.flow_key, d.level, d.person_id, p.first_name || ' ' || p.last_name, coalesce(r.name, ''), coalesce(r.id::text, '')
			from hub.approval_designees d join hub.people p on p.id = d.person_id left join hub.person_roles pr on pr.person_id = p.id left join hub.roles r on r.id = pr.role_id`)
		if err != nil {
			return err
		}
		defer ds.Close()
		for ds.Next() {
			var k string
			var d Designee
			if err := ds.Scan(&k, &d.Level, &d.Person, &d.Name, &d.Role, &d.RoleID); err != nil {
				return err
			}
			for i := range out {
				if out[i].Key == k {
					out[i].Designees = append(out[i].Designees, d)
				}
			}
		}
		return ds.Err()
	})
	return out, err
}

// SetDesignee chooses who decides one level of a flow: the people given replace whoever was chosen before (none clears it).
// Any one of them may decide the level.
func (e *Engine) SetDesignee(ctx context.Context, p *tenant.Principal, key string, level int, people []uuid.UUID) error {
	if len(people) > 50 {
		return httpx.Bad("Choose up to 50 people for a level.")
	}
	if level < 1 || level > 3 {
		return httpx.Bad("A flow has levels 1 to 3.")
	}
	return e.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Configure approvals"); err != nil {
			return err
		}
		var levels int
		if err := tx.QueryRow(ctx, `select levels from hub.approval_flows where key=$1`, key).Scan(&levels); err != nil {
			return httpx.ErrNotFound
		}
		if level > levels {
			return httpx.Bad("This flow has only " + fmt.Sprint(levels) + " levels.")
		}
		for _, id := range people {
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.people where id=$1 and status='active')`, id).Scan(&ok); err != nil || !ok {
				return httpx.Bad("Choose people who are active in the organization.")
			}
		}
		if _, err := tx.Exec(ctx, `delete from hub.approval_designees where flow_key=$1 and level=$2 and not (person_id = any($3))`, key, level, people); err != nil {
			return err
		}
		for _, id := range people {
			if _, err := tx.Exec(ctx, `insert into hub.approval_designees (org_id, flow_key, level, person_id) values ($1,$2,$3,$4) on conflict do nothing`, p.Org, key, level, id); err != nil {
				return err
			}
		}
		if err := e.A.Sign.Seal(ctx, tx, "approval_designees", "flow_key=$1 and level=$2", key, level); err != nil {
			return err
		}
		return e.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Chose %d approver(s) for level %d of flow %s", len(people), level, key), "Roles", p.IP, map[string]any{"flow": key, "level": level, "people": people})
	})
}

// SetFlow sets how many levels a flow needs. Who approves each level is not stored on the flow: it is whoever holds the right
// "Approve level N" in the flow's module.
func (e *Engine) SetFlow(ctx context.Context, p *tenant.Principal, key string, levels int) error {
	if levels < 0 || levels > 3 {
		return httpx.Bad("A flow has between 0 and 3 levels.")
	}
	return e.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Configure approvals"); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `update hub.approval_flows set levels=$2, updated_at=now(), `+integrity.Bump+` where key=$1`, key, levels)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		if err := e.A.Sign.Seal(ctx, tx, "approval_flows", "key=$1", key); err != nil {
			return err
		}
		return e.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Set approval flow %s to %d levels", key, levels), "Roles", p.IP, map[string]any{"flow": key, "levels": levels})
	})
}

// itemsOf reads the "items" list of a request's payload.
func itemsOf(raw []byte) []Item {
	var pl struct {
		Items []Item `json:"items"`
	}
	if json.Unmarshal(raw, &pl) != nil {
		return nil
	}
	return pl.Items
}
