package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/authz"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

// patchPlan changes a project's type and workflow. A status still holding tasks cannot be removed.
func (s *Service) patchPlan(ctx context.Context, tx pgx.Tx, p *tenant.Principal, id uuid.UUID, in Patch) error {
	if in.Method != nil {
		if err := authz.Require(ctx, tx, p.Person, "projects", "Change project type"); err != nil {
			return err
		}
		if !methods[*in.Method] {
			return httpx.Bad("Choose agile, waterfall or hybrid.")
		}
		if _, err := tx.Exec(ctx, `update hub.projects set method=$2, `+integrity.Bump+` where conversation_id=$1`, id, *in.Method); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "projects", "conversation_id=$1", id); err != nil {
			return err
		}
	}
	if in.Workflow == nil {
		return nil
	}
	if err := authz.Require(ctx, tx, p.Person, "projects", "Manage workflow"); err != nil {
		return err
	}
	w := in.Workflow
	if len(w) < 2 || len(w) > 12 {
		return httpx.Bad("A workflow has 2 to 12 statuses.")
	}
	if w[0].Cat == 3 || w[len(w)-1].Cat != 3 {
		return httpx.Bad("A workflow starts with an unfinished status and ends with a finished one.")
	}
	keys := map[string]int{}
	var current []Status
	var raw []byte
	if err := tx.QueryRow(ctx, `select workflow from hub.projects where conversation_id=$1`, id).Scan(&raw); err != nil {
		return err
	}
	_ = json.Unmarshal(raw, &current)
	for _, st := range current {
		kept := false
		for _, x := range w {
			kept = kept || x.Key == st.Key
		}
		if !kept {
			if err := authz.Require(ctx, tx, p.Person, "projects", "Remove workflow statuses"); err != nil {
				return err
			}
			break
		}
	}
	for i := range w {
		w[i].Name = strings.TrimSpace(w[i].Name)
		if w[i].Name == "" || len(w[i].Name) > 40 || w[i].Key == "" || w[i].Cat < 0 || w[i].Cat > 3 || w[i].WIP < 0 {
			return httpx.Bad("Every status needs a name of up to 40 characters.")
		}
		if _, dup := keys[w[i].Key]; dup {
			return httpx.Bad("Two statuses are the same.")
		}
		keys[w[i].Key] = w[i].Cat
	}
	rs, err := tx.Query(ctx, `select id, data from hub.project_items where project_id=$1 and collection='task'`, id)
	if err != nil {
		return err
	}
	type row struct {
		id   uuid.UUID
		data map[string]any
	}
	var tasks []row
	for rs.Next() {
		var r row
		var b []byte
		if err := rs.Scan(&r.id, &b); err != nil {
			rs.Close()
			return err
		}
		_ = json.Unmarshal(b, &r.data)
		tasks = append(tasks, r)
	}
	rs.Close()
	if err := rs.Err(); err != nil {
		return err
	}
	for _, t := range tasks {
		cat, ok := keys[text(t.data, "status")]
		if !ok {
			return httpx.Err(409, "status_in_use", fmt.Sprintf("Move %q out of its status before removing that status.", text(t.data, "title")))
		}
		// a status's meaning changed: its tasks follow
		if c, _ := t.data["col"].(float64); int(c) != cat {
			t.data["col"] = cat
			b, _ := json.Marshal(t.data)
			if _, err := tx.Exec(ctx, `update hub.project_items set data=$2, updated_at=now(), `+integrity.Bump+` where id=$1`, t.id, b); err != nil {
				return err
			}
			if err := s.A.Sign.Seal(ctx, tx, "project_items", "id=$1", t.id); err != nil {
				return err
			}
		}
	}
	b, _ := json.Marshal(w)
	_, err = tx.Exec(ctx, `update hub.projects set workflow=$2, `+integrity.Bump+` where conversation_id=$1`, id, b)
	return err
}

func (s *Service) planChanged(ctx context.Context, p *tenant.Principal, project uuid.UUID) {
	d, _ := json.Marshal(map[string]any{"project": project, "by": p.Person})
	s.A.Publish(ctx, realtime.Event{Type: "project.plan", Channel: realtime.ConvChannel(project), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
}

func (s *Service) planRoutes(r chi.Router, org func(func(http.ResponseWriter, *http.Request, *tenant.Principal) error) http.HandlerFunc) {
	ids := func(r *http.Request, name string) (uuid.UUID, error) {
		id, err := uuid.Parse(chi.URLParam(r, name))
		if err != nil {
			return id, httpx.ErrNotFound
		}
		return id, nil
	}
	s.automationRoutes(r, org, ids)
	r.Post("/projects/{id}/pages/{item}/live", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		var in LiveIn
		if err := httpx.Decode(r, &in, 6<<20); err != nil {
			return err
		}
		if err := s.Live(r.Context(), p, id, item, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/projects/eligible", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := s.Eligible(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"people": l})
		return nil
	}))
	r.Delete("/projects/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		if err := s.DeleteProject(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/projects/{id}/plan", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		items, changes, err := s.Plan(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"items": items, "changes": changes})
		return nil
	}))
	r.Put("/projects/{id}/items/{item}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		var in struct {
			Collection string         `json:"collection"`
			Data       map[string]any `json:"data"`
		}
		if err := httpx.Decode(r, &in, 8<<20); err != nil {
			return err
		}
		out, err := s.Save(r.Context(), p, id, item, in.Collection, in.Data)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	r.Delete("/projects/{id}/items/{item}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		if err := s.Delete(r.Context(), p, id, item); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/projects/{id}/import", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		var in struct {
			Items []ImportItem `json:"items"`
		}
		if err := httpx.Decode(r, &in, 32<<20); err != nil {
			return err
		}
		if err := s.Import(r.Context(), p, id, in.Items); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}
