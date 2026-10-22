// Package approvals is the approvals engine: a request waits at each level of its flow until the right person decides.
// Other modules register what to do once the last level approves; nothing is applied before that.
package approvals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/access"
	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// Applied is what a handler receives when a request has been fully approved (or needed no approval).
type Applied struct {
	Org       uuid.UUID
	Request   uuid.UUID
	Ref       string
	Requester uuid.UUID
	Payload   json.RawMessage
	DEK       []byte
}

type Handler func(ctx context.Context, tx pgx.Tx, a Applied) error

type Engine struct {
	A        *app.App
	mu       sync.RWMutex
	handlers map[string]Handler
}

func New(a *app.App) *Engine { return &Engine{A: a, handlers: map[string]Handler{}} }

// Register says what to do when a request of this kind is approved. Call it once per kind at startup.
func (e *Engine) Register(kind string, h Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[kind] = h
}

type Step struct {
	Level    int       `json:"level"`
	Who      string    `json:"who"`
	Decision string    `json:"decision"`
	Comment  string    `json:"comment,omitempty"`
	At       time.Time `json:"at"`
}

type Request struct {
	ID        uuid.UUID  `json:"id"`
	Ref       string     `json:"ref"`
	Flow      string     `json:"flow"`
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Note      string     `json:"note,omitempty"`
	By        string     `json:"by"`
	Requester uuid.UUID  `json:"requester"`
	Level     int        `json:"level"`
	Levels    int        `json:"levels"`
	Waiting   string     `json:"waiting_for,omitempty"`
	Status    string     `json:"status"`
	History   []Step     `json:"history"`
	CreatedAt time.Time  `json:"created_at"`
	DecidedAt *time.Time `json:"decided_at,omitempty"`
	CanDecide bool       `json:"can_decide"`
}

type Flow struct {
	Key       string     `json:"id"`
	Module    string     `json:"mod"`
	Action    string     `json:"action"`
	Levels    int        `json:"levels"`
	Approvers []string   `json:"approvers"`
	Designees []Designee `json:"designees"` // the person chosen for a level, where one was chosen
}

type Designee struct {
	Level  int       `json:"level"`
	Person uuid.UUID `json:"person"`
	Name   string    `json:"name"`
	Role   string    `json:"role"`
	RoleID string    `json:"role_id"`
}

func fieldCtx(org uuid.UUID, id uuid.UUID, col string) crypto.Context {
	return crypto.Context{Org: org.String(), Table: "approval_requests", Column: col, RowID: id.String()}
}

func (e *Engine) flow(ctx context.Context, tx pgx.Tx, key string) (*Flow, error) {
	f := &Flow{Key: key}
	err := tx.QueryRow(ctx, `select module, action, levels from hub.approval_flows where key=$1`, key).Scan(&f.Module, &f.Action, &f.Levels)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.Bad("That action has no approval flow.")
	}
	return f, err
}

// Submit starts approval for a change. When the flow needs no approval (0 levels) the change is applied right away.
// It must run inside the caller's tenant transaction, so the request and the caller's own writes commit together.
func (e *Engine) Submit(ctx context.Context, tx pgx.Tx, p *tenant.Principal, flowKey, kind, title, note string, payload any) (*Request, bool, error) {
	f, err := e.flow(ctx, tx, flowKey)
	if err != nil {
		return nil, false, err
	}
	dek, _, err := e.A.OrgDEK(ctx, tx, p.Org)
	if err != nil {
		return nil, false, err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, false, err
	}
	id := app.NewID()
	var seq int
	if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'approvals',1) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org).Scan(&seq); err != nil {
		return nil, false, err
	}
	tEnc, err := crypto.Field(dek, fieldCtx(p.Org, id, "title_enc"), []byte(title))
	if err != nil {
		return nil, false, err
	}
	var nEnc []byte
	if note != "" {
		if nEnc, err = crypto.Field(dek, fieldCtx(p.Org, id, "note_enc"), []byte(note)); err != nil {
			return nil, false, err
		}
	}
	pEnc, err := crypto.Field(dek, fieldCtx(p.Org, id, "payload_enc"), raw)
	if err != nil {
		return nil, false, err
	}
	status, applied := "Pending", false
	if f.Levels == 0 {
		status, applied = "Approved", true
	}
	if _, err := tx.Exec(ctx, `insert into hub.approval_requests (id, org_id, seq, flow_key, kind, title_enc, note_enc, payload_enc, requester, status, decided_at) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10, case when $10='Approved' then now() end)`,
		id, p.Org, seq, flowKey, kind, tEnc, nEnc, pEnc, p.Person, status); err != nil {
		return nil, false, err
	}
	if err := e.A.Sign.Seal(ctx, tx, "approval_requests", "id=$1", id); err != nil {
		return nil, false, err
	}
	if applied {
		if err := e.apply(ctx, tx, Applied{Org: p.Org, Request: id, Ref: ref(seq), Requester: p.Person, Payload: raw, DEK: dek}, kind); err != nil {
			return nil, false, err
		}
	}
	if err := e.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Asked for approval: "+title, f.Module, p.IP, map[string]any{"ref": ref(seq), "flow": flowKey, "direct": applied}); err != nil {
		return nil, false, err
	}
	e.notifyLater(p.Org, id, ref(seq), status)
	made := &Request{ID: id, Ref: ref(seq), Flow: flowKey, Kind: kind, Title: title, Note: note, Requester: p.Person, By: p.Name, Level: 1, Levels: max(1, f.Levels), Status: status, History: []Step{}, CreatedAt: app.Now()}
	if status == "Pending" {
		e.mailWaiting(ctx, tx, p.Org, f, made)
	}
	return made, applied, nil
}

func ref(seq int) string { return fmt.Sprintf("AP-%d", seq) }

func (e *Engine) apply(ctx context.Context, tx pgx.Tx, a Applied, kind string) error {
	e.mu.RLock()
	h := e.handlers[kind]
	e.mu.RUnlock()
	if h == nil {
		return fmt.Errorf("approvals: no handler for %q", kind)
	}
	return h(ctx, tx, a)
}

// pending notifications are sent after commit by the caller's HTTP layer; here we only queue them on the bus right away
// because a late event for a rolled-back request is harmless (clients re-read the list).
func (e *Engine) notifyLater(org, id uuid.UUID, r, status string) {
	data, _ := json.Marshal(map[string]any{"id": id, "ref": r, "status": status})
	go e.A.Publish(context.Background(), realtime.Event{Type: "approval.changed", Channel: realtime.OrgChannel(org), Org: org.String(), Data: data})
}

// Decide records one person's decision at the request's current level.
func (e *Engine) Decide(ctx context.Context, p *tenant.Principal, id uuid.UUID, approve bool, comment string) (*Request, error) {
	var out *Request
	err := e.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var level, seq int
		var status, flowKey, kind string
		var requester uuid.UUID
		var pEnc []byte
		err := tx.QueryRow(ctx, `select level, seq, status, flow_key, kind, requester, payload_enc from hub.approval_requests where id=$1 for update`, id).Scan(&level, &seq, &status, &flowKey, &kind, &requester, &pEnc)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		if status != "Pending" {
			return httpx.Err(409, "already_decided", "This request has already been decided.")
		}
		f, err := e.flow(ctx, tx, flowKey)
		if err != nil {
			return err
		}
		if ok, err := e.mayDecide(ctx, tx, p.Person, requester, f, level); err != nil {
			return err
		} else if !ok {
			return httpx.Err(403, "not_your_turn", "This request is waiting for someone else.")
		}
		dek, _, err := e.A.OrgDEK(ctx, tx, p.Org)
		if err != nil {
			return err
		}
		decision := "Rejected"
		if approve {
			decision = "Approved"
		}
		var cEnc []byte
		if comment = strings.TrimSpace(comment); comment != "" {
			if cEnc, err = crypto.Field(dek, crypto.Context{Org: p.Org.String(), Table: "approval_steps", Column: "comment_enc", RowID: fmt.Sprintf("%s/%d", id, level)}, []byte(comment)); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `insert into hub.approval_steps (org_id, request_id, level, approver, decision, comment_enc) values ($1,$2,$3,$4,$5,$6)`, p.Org, id, level, p.Person, decision, cEnc); err != nil {
			return err
		}
		if err := e.A.Sign.Seal(ctx, tx, "approval_steps", "request_id=$1 and level=$2", id, level); err != nil {
			return err
		}
		newStatus, newLevel := "Pending", level
		switch {
		case !approve:
			newStatus = "Rejected"
		case level >= max(1, f.Levels):
			newStatus = "Approved"
		default:
			newLevel = level + 1
		}
		if _, err := tx.Exec(ctx, `update hub.approval_requests set level=$2, status=$3, updated_at=now(), decided_at = case when $3<>'Pending' then now() end, `+integrity.Bump+` where id=$1`, id, newLevel, newStatus); err != nil {
			return err
		}
		if err := e.A.Sign.Seal(ctx, tx, "approval_requests", "id=$1", id); err != nil {
			return err
		}
		if newStatus == "Approved" {
			raw, err := crypto.FieldOpen(dek, fieldCtx(p.Org, id, "payload_enc"), pEnc)
			if err != nil {
				return err
			}
			if err := e.apply(ctx, tx, Applied{Org: p.Org, Request: id, Ref: ref(seq), Requester: requester, Payload: raw, DEK: dek}, kind); err != nil {
				return err
			}
		}
		if err := e.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("%s %s at level %d", decision, ref(seq), level), f.Module, p.IP, map[string]any{"ref": ref(seq), "decision": decision}); err != nil {
			return err
		}
		out, err = e.get(ctx, tx, p, id)
		if err == nil && out != nil {
			if newStatus == "Pending" {
				e.mailWaiting(ctx, tx, p.Org, f, out)
			} else {
				e.mailDecided(ctx, tx, p.Org, out, p.Name, comment)
			}
		}
		e.notifyLater(p.Org, id, ref(seq), newStatus)
		return err
	})
	return out, err
}

// mayDecide: the requester never decides their own request. A level is decided by anyone holding the right "Approve level N" in
// the flow's module (given to roles on the Roles page). Locked (owner) roles can stand in at any level, so a vacant position
// never blocks the organization.
func (e *Engine) mayDecide(ctx context.Context, tx pgx.Tx, who, requester uuid.UUID, f *Flow, level int) (bool, error) {
	if who == requester {
		return false, nil
	}
	var chosen uuid.UUID
	err := tx.QueryRow(ctx, `select person_id from hub.approval_designees where flow_key=$1 and level=$2`, f.Key, level).Scan(&chosen)
	if err == nil {
		if who == chosen {
			return true, nil
		}
		var locked bool
		err = tx.QueryRow(ctx, `select coalesce(bool_or(r.locked), false) from hub.person_roles pr join hub.roles r on r.id = pr.role_id where pr.person_id=$1`, who).Scan(&locked)
		return locked, err
	}
	if err != pgx.ErrNoRows {
		return false, err
	}
	return authz.Can(ctx, tx, who, f.Module, access.ApproveRight(level))
}

// designee names the person chosen for a level, or "" when none was.
func (e *Engine) designee(ctx context.Context, tx pgx.Tx, key string, level int) string {
	var n string
	if tx.QueryRow(ctx, `select p.first_name || ' ' || p.last_name from hub.approval_designees d join hub.people p on p.id = d.person_id where d.flow_key=$1 and d.level=$2`, key, level).Scan(&n) == nil {
		return n
	}
	return ""
}

// waiting names the roles that can decide a level (for "waiting for ..." labels).
func (e *Engine) waiting(ctx context.Context, tx pgx.Tx, key, module string, level int) string {
	if n := e.designee(ctx, tx, key, level); n != "" {
		return n
	}
	rs, err := tx.Query(ctx, `select r.name from hub.roles r
		where r.locked or coalesce((select rr.allowed from hub.role_rights rr where rr.role_id = r.id and rr.module=$1 and rr.right_name=$2), false)
		order by r.locked, r.name`, module, access.ApproveRight(level))
	if err != nil {
		return ""
	}
	defer rs.Close()
	var names []string
	for rs.Next() {
		var n string
		if rs.Scan(&n) == nil {
			names = append(names, n)
		}
	}
	if len(names) > 3 {
		names = append(names[:3], "…")
	}
	return strings.Join(names, ", ")
}

func (e *Engine) get(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID) (*Request, error) {
	list, err := e.list(ctx, tx, p, "r.id = $1", []any{id}, 1)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, httpx.ErrNotFound
	}
	return list[0], nil
}

func (e *Engine) list(ctx context.Context, tx pgx.Tx, p *tenant.Principal, where string, args []any, limit int) ([]*Request, error) {
	dek, _, err := e.A.OrgDEK(ctx, tx, p.Org)
	if err != nil {
		return nil, err
	}
	rs, err := tx.Query(ctx, `
		select r.id, r.seq, r.flow_key, r.kind, r.title_enc, r.note_enc, r.requester, p.first_name || ' ' || p.last_name, r.level, r.status, r.created_at, r.decided_at,
		       f.levels, f.module
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
		var tEnc, nEnc []byte
		var mod string
		var sq int
		if err := rs.Scan(&x.ID, &sq, &x.Flow, &x.Kind, &tEnc, &nEnc, &x.Requester, &x.By, &x.Level, &x.Status, &x.CreatedAt, &x.DecidedAt, &x.Levels, &mod); err != nil {
			rs.Close()
			return nil, err
		}
		x.Levels = max(1, x.Levels)
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

// SetDesignee chooses (or, with a nil person, clears) who decides one level of a flow.
func (e *Engine) SetDesignee(ctx context.Context, p *tenant.Principal, key string, level int, person *uuid.UUID) error {
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
		if person == nil {
			if _, err := tx.Exec(ctx, `delete from hub.approval_designees where flow_key=$1 and level=$2`, key, level); err != nil {
				return err
			}
		} else {
			var ok bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from hub.people where id=$1 and status='active')`, *person).Scan(&ok); err != nil || !ok {
				return httpx.Bad("Choose someone who is active in the organization.")
			}
			if _, err := tx.Exec(ctx, `insert into hub.approval_designees (org_id, flow_key, level, person_id) values ($1,$2,$3,$4)
				on conflict (org_id, flow_key, level) do update set person_id=excluded.person_id, updated_at=now(), `+integrity.Bump, p.Org, key, level, *person); err != nil {
				return err
			}
			if err := e.A.Sign.Seal(ctx, tx, "approval_designees", "flow_key=$1 and level=$2", key, level); err != nil {
				return err
			}
		}
		return e.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, fmt.Sprintf("Chose the approver of level %d of flow %s", level, key), "Roles", p.IP, map[string]any{"flow": key, "level": level, "person": person})
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
