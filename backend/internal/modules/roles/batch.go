package roles

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/modules/approvals"
	"hub/internal/tenant"
)

// Several rights marked together go to approval as ONE request. Each right in it has its own id, so an approver can approve some,
// reject some and leave the rest waiting (the engine splits the request: see approvals.Decide with items). The request is called
// "rights.batch"; once approved, each right is applied exactly as a single "rights.role", "rights.group" or "rights.person" change.

type batchItem struct {
	ID     string      `json:"id"`
	Label  string      `json:"label"`
	Verb   string      `json:"verb"` // allow | remove | reset
	Change rightChange `json:"change"`
}

type batchPayload struct {
	Scope  string      `json:"scope"`  // role | group | person
	Target string      `json:"target"` // role key, group key or person id
	Name   string      `json:"name"`
	Items  []batchItem `json:"items"`
}

type batchIn struct {
	Changes []struct {
		Module  string `json:"module"`
		Right   string `json:"right"`
		Allowed *bool  `json:"allowed"` // nil on a person removes the personal override
	} `json:"changes"`
}

const maxBatch = 300

// SetRightsBatch marks many rights of one role, user group or person at once and sends them as one approval request (or applies them
// at once when the rights flow needs no approval).
func (s *Service) SetRightsBatch(ctx context.Context, p *tenant.Principal, scope, target string, in batchIn) (*approvals.Request, bool, error) {
	if len(in.Changes) == 0 || len(in.Changes) > maxBatch {
		return nil, false, httpx.Bad(fmt.Sprintf("Mark between 1 and %d rights.", maxBatch))
	}
	seen := map[string]bool{}
	items := make([]batchItem, 0, len(in.Changes))
	for _, c := range in.Changes {
		rc := rightChange{Module: c.Module, Right: c.Right, Allowed: c.Allowed}
		switch scope {
		case "role", "group":
			rc.Role = target
		case "person":
			id, err := uuid.Parse(target)
			if err != nil {
				return nil, false, httpx.ErrNotFound
			}
			rc.Person = id
		default:
			return nil, false, httpx.Bad("Unknown kind of change.")
		}
		if err := s.validate(ctx, p.Org, rc, scope == "person"); err != nil {
			return nil, false, err
		}
		k := c.Module + "\x00" + c.Right
		if seen[k] {
			return nil, false, httpx.Bad("A right is marked twice.")
		}
		seen[k] = true
		verb := "reset"
		if rc.Allowed != nil {
			verb = map[bool]string{true: "allow", false: "remove"}[*rc.Allowed]
		}
		items = append(items, batchItem{ID: uuid.NewString(), Label: c.Module + " · " + c.Right, Verb: verb, Change: rc})
	}
	var req *approvals.Request
	var applied bool
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "roles", "Manage roles & rights"); err != nil {
			return err
		}
		var name string
		switch scope {
		case "role":
			var locked bool
			if err := tx.QueryRow(ctx, `select locked, name from hub.roles where key=$1`, target).Scan(&locked, &name); errors.Is(err, pgx.ErrNoRows) {
				return httpx.ErrNotFound
			} else if err != nil {
				return err
			}
			if locked {
				return httpx.Err(409, "locked_role", name+" cannot be limited.")
			}
		case "group":
			if err := tx.QueryRow(ctx, `select name from hub.role_groups where key=$1`, target).Scan(&name); errors.Is(err, pgx.ErrNoRows) {
				return httpx.ErrNotFound
			} else if err != nil {
				return err
			}
		case "person":
			var locked bool
			if err := tx.QueryRow(ctx, `select p.first_name || ' ' || p.last_name, coalesce(r.locked,false) from hub.people p left join hub.person_roles pr on pr.person_id=p.id left join hub.roles r on r.id=pr.role_id where p.id=$1 and p.status <> 'deleted'`, items[0].Change.Person).Scan(&name, &locked); errors.Is(err, pgx.ErrNoRows) {
				return httpx.ErrNotFound
			} else if err != nil {
				return err
			}
			if locked {
				return httpx.Err(409, "locked_role", "A super admin cannot be limited.")
			}
		}
		label := map[string]string{"role": "Role", "group": "User group", "person": "Person"}[scope]
		payload := batchPayload{Scope: scope, Target: target, Name: name, Items: items}
		var err error
		req, applied, err = s.Ap.Submit(ctx, tx, p, "f8", "rights.batch", fmt.Sprintf("Change rights · %s: %s", label, name), "", payload)
		return err
	})
	return req, applied, err
}

// applyBatch applies each right of an approved batch the way its single request would be applied.
func (s *Service) applyBatch(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var b batchPayload
	if err := json.Unmarshal(a.Payload, &b); err != nil {
		return err
	}
	for _, it := range b.Items {
		raw, _ := json.Marshal(it.Change)
		one := a
		one.Payload = raw
		var err error
		switch b.Scope {
		case "role":
			err = s.applyRoleRight(ctx, tx, one)
		case "group":
			err = s.applyGroupRight(ctx, tx, one)
		case "person":
			err = s.applyPersonRight(ctx, tx, one)
		default:
			err = fmt.Errorf("roles: unknown batch scope %q", b.Scope)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) batchRoutes(r chi.Router) {
	h := func(scope, param string) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			var in batchIn
			if err := httpx.Decode(r, &in, 64<<10); err != nil {
				return err
			}
			req, applied, err := s.SetRightsBatch(r.Context(), p, scope, chi.URLParam(r, param), in)
			if err != nil {
				return err
			}
			reply(w, applied, req)
			return nil
		})
	}
	r.Put("/roles/{key}/rights/batch", h("role", "key"))
	r.Put("/groups/{key}/rights/batch", h("group", "key"))
	r.Put("/people/{id}/rights/batch", h("person", "id"))
}
