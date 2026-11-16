package projects

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

// SweepAll runs for every active organization (the worker calls this on a timer).
func (s *Service) SweepAll(ctx context.Context) {
	var orgs []uuid.UUID
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select id from hub.orgs where status='active'`)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				return err
			}
			orgs = append(orgs, id)
		}
		return rs.Err()
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("project sweep", "err", err)
	}
	for _, o := range orgs {
		if err := s.Sweep(ctx, o, time.Now()); err != nil {
			slog.Error("project sweep", "org", o, "err", err)
		}
	}
}

func (s *Service) automationRoutes(r chi.Router, org func(func(http.ResponseWriter, *http.Request, *tenant.Principal) error) http.HandlerFunc, ids func(*http.Request, string) (uuid.UUID, error)) {
	r.Post("/projects/{id}/sprints/{item}/close", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		if err := s.CloseSprint(r.Context(), p, id, item); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/projects/{id}/gates/{item}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := ids(r, "id")
		if err != nil {
			return err
		}
		item, err := ids(r, "item")
		if err != nil {
			return err
		}
		var in struct {
			Comment string `json:"comment"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		req, err := s.PassGate(r.Context(), p, id, item, in.Comment)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"approval": req})
		return nil
	}))
}
