package projects

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

type orgHandler func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc

// githubPublic: GitHub's deliveries (no Hub sign-in; the signature with the project's secret is the proof).
func (s *Service) githubPublic(r chi.Router) {
	r.Post("/hooks/github/{hook}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		hook, err := uuid.Parse(chi.URLParam(r, "hook"))
		if err != nil {
			return httpx.ErrNotFound
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 5<<20))
		if err != nil {
			return httpx.Err(413, "too_large", "That delivery is too large.")
		}
		if err := s.HandleHook(r.Context(), hook, r.Header.Get("X-GitHub-Event"), r.Header.Get("X-GitHub-Delivery"), r.Header.Get("X-Hub-Signature-256"), r.Header.Get("Content-Type"), body); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}

func (s *Service) githubRoutes(r chi.Router, org orgHandler) {
	pid := func(r *http.Request) (uuid.UUID, error) {
		v, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return v, httpx.ErrNotFound
		}
		return v, nil
	}
	r.Get("/projects/{id}/github", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := pid(r)
		if err != nil {
			return err
		}
		var only []uuid.UUID
		if v, err := uuid.Parse(r.URL.Query().Get("task")); err == nil {
			only = append(only, v)
		}
		g, err := s.GitHub(r.Context(), p, id, only...)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, g)
		return nil
	}))
	r.Put("/projects/{id}/github", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := pid(r)
		if err != nil {
			return err
		}
		var in LinkIn
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		g, err := s.Link(r.Context(), p, id, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, g)
		return nil
	}))
	r.Delete("/projects/{id}/github", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := pid(r)
		if err != nil {
			return err
		}
		if err := s.Unlink(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/projects/{id}/github/branches", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := pid(r)
		if err != nil {
			return err
		}
		l, err := s.Branches(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"branches": l})
		return nil
	}))
	r.Post("/projects/{id}/github/events/{event}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := pid(r)
		if err != nil {
			return err
		}
		ev, err := uuid.Parse(chi.URLParam(r, "event"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in struct {
			Decision string `json:"decision"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.Decide(r.Context(), p, id, ev, in.Decision); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}
