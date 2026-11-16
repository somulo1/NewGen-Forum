package approvals

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

// splitItems keeps the chosen items in this request and moves the others into a new request that waits exactly where this one is:
// same flow, same level, same requester, with the decisions already taken at lower levels copied. It returns the chosen payload and
// the new request's reference. The decision then goes on to the chosen ones; the rest are left for later.
func (e *Engine) splitItems(ctx context.Context, tx pgx.Tx, p *tenant.Principal, dek []byte, id uuid.UUID, level int, raw []byte, chosen map[string]bool) ([]byte, string, error) {
	var pl map[string]json.RawMessage
	if err := json.Unmarshal(raw, &pl); err != nil {
		return nil, "", err
	}
	var all []json.RawMessage
	if err := json.Unmarshal(pl["items"], &all); err != nil {
		return nil, "", err
	}
	var keep, rest []json.RawMessage
	for _, it := range all {
		var head struct{ ID string }
		_ = json.Unmarshal(it, &head)
		if chosen[head.ID] {
			keep = append(keep, it)
		} else {
			rest = append(rest, it)
		}
	}
	if len(keep) != len(chosen) {
		return nil, "", httpx.Bad("One of the chosen items is not part of this request.")
	}
	if len(rest) == 0 {
		return raw, "", nil
	}
	mk := func(items []json.RawMessage) ([]byte, error) {
		cp := map[string]json.RawMessage{}
		for k, v := range pl {
			cp[k] = v
		}
		b, err := json.Marshal(items)
		if err != nil {
			return nil, err
		}
		cp["items"] = b
		return json.Marshal(cp)
	}
	keepRaw, err := mk(keep)
	if err != nil {
		return nil, "", err
	}
	restRaw, err := mk(rest)
	if err != nil {
		return nil, "", err
	}
	newID := app.NewID()
	var seq int
	if err := tx.QueryRow(ctx, `insert into hub.counters (org_id, name, value) values ($1,'approvals',1) on conflict (org_id, name) do update set value = hub.counters.value + 1 returning value`, p.Org).Scan(&seq); err != nil {
		return nil, "", err
	}
	var tEnc, nEnc []byte
	var flowKey, kind string
	var requester uuid.UUID
	var title, note string
	var tOld, nOld []byte
	if err := tx.QueryRow(ctx, `select flow_key, kind, requester, title_enc, note_enc from hub.approval_requests where id=$1`, id).Scan(&flowKey, &kind, &requester, &tOld, &nOld); err != nil {
		return nil, "", err
	}
	if b, err := crypto.FieldOpen(dek, fieldCtx(p.Org, id, "title_enc"), tOld); err == nil {
		title = string(b)
	}
	if len(nOld) > 0 {
		if b, err := crypto.FieldOpen(dek, fieldCtx(p.Org, id, "note_enc"), nOld); err == nil {
			note = string(b)
		}
	}
	if tEnc, err = crypto.Field(dek, fieldCtx(p.Org, newID, "title_enc"), []byte(title)); err != nil {
		return nil, "", err
	}
	if note != "" {
		if nEnc, err = crypto.Field(dek, fieldCtx(p.Org, newID, "note_enc"), []byte(note)); err != nil {
			return nil, "", err
		}
	}
	pEnc, err := crypto.Field(dek, fieldCtx(p.Org, newID, "payload_enc"), restRaw)
	if err != nil {
		return nil, "", err
	}
	if _, err := tx.Exec(ctx, `insert into hub.approval_requests (id, org_id, seq, flow_key, kind, title_enc, note_enc, payload_enc, requester, status, level) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,'Pending',$10)`,
		newID, p.Org, seq, flowKey, kind, tEnc, nEnc, pEnc, requester, level); err != nil {
		return nil, "", err
	}
	if err := e.A.Sign.Seal(ctx, tx, "approval_requests", "id=$1", newID); err != nil {
		return nil, "", err
	}
	// what lower levels already decided about these items stands for the part that is left
	hs, err := tx.Query(ctx, `select level, approver, decision, comment_enc, created_at from hub.approval_steps where request_id=$1 and level < $2 order by level`, id, level)
	if err != nil {
		return nil, "", err
	}
	type prior struct {
		level    int
		approver uuid.UUID
		decision string
		comment  string
		at       time.Time
	}
	var priors []prior
	for hs.Next() {
		var x prior
		var cEnc []byte
		if err := hs.Scan(&x.level, &x.approver, &x.decision, &cEnc, &x.at); err != nil {
			hs.Close()
			return nil, "", err
		}
		if len(cEnc) > 0 {
			if b, err := crypto.FieldOpen(dek, crypto.Context{Org: p.Org.String(), Table: "approval_steps", Column: "comment_enc", RowID: fmt.Sprintf("%s/%d", id, x.level)}, cEnc); err == nil {
				x.comment = string(b)
			}
		}
		priors = append(priors, x)
	}
	hs.Close()
	for _, x := range priors {
		var cEnc []byte
		if x.comment != "" {
			if cEnc, err = crypto.Field(dek, crypto.Context{Org: p.Org.String(), Table: "approval_steps", Column: "comment_enc", RowID: fmt.Sprintf("%s/%d", newID, x.level)}, []byte(x.comment)); err != nil {
				return nil, "", err
			}
		}
		if _, err := tx.Exec(ctx, `insert into hub.approval_steps (org_id, request_id, level, approver, decision, comment_enc, created_at) values ($1,$2,$3,$4,$5,$6,$7)`, p.Org, newID, x.level, x.approver, x.decision, cEnc, x.at); err != nil {
			return nil, "", err
		}
	}
	if len(priors) > 0 {
		if err := e.A.Sign.Seal(ctx, tx, "approval_steps", "request_id=$1", newID); err != nil {
			return nil, "", err
		}
	}
	// this request now holds only the chosen items
	pEnc2, err := crypto.Field(dek, fieldCtx(p.Org, id, "payload_enc"), keepRaw)
	if err != nil {
		return nil, "", err
	}
	if _, err := tx.Exec(ctx, `update hub.approval_requests set payload_enc=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, id, pEnc2); err != nil {
		return nil, "", err
	}
	if err := e.A.Sign.Seal(ctx, tx, "approval_requests", "id=$1", id); err != nil {
		return nil, "", err
	}
	e.notifyLater(p.Org, newID, ref(seq), "Pending")
	return keepRaw, ref(seq), nil
}
