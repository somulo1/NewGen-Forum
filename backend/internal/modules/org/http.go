package org

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/app"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

func uuidNew() uuid.UUID { return app.NewID() }

func (s *Service) Public(chi.Router) {}

func (s *Service) Private(r chi.Router) {
	r.Post("/orgs", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		var in struct{ Name, Slug string }
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if ok, retry := s.A.Limit.Allow(r.Context(), "org-create", p.Account.String(), 5, 24*60*60*1e9, true); !ok {
			return httpx.Err(429, "rate_limited", "Too many organizations created. Try again in "+retry.String()+".")
		}
		c, err := s.Create(r.Context(), p.Account, p.Name, in.Name, in.Slug, p.IP)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, c)
		return nil
	}))
	r.Get("/org/structure", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		u, err := s.Structure(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"units": u})
		return nil
	}))
	r.Post("/org/units", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in Unit
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		u, err := s.SaveUnit(r.Context(), p, nil, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, u)
		return nil
	}))
	r.Patch("/org/units/{id}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in Unit
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		u, err := s.SaveUnit(r.Context(), p, &id, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, u)
		return nil
	}))
	r.Get("/org/units/{id}/dashboard", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		days, _ := strconv.Atoi(r.URL.Query().Get("days"))
		d, err := s.Dashboard(r.Context(), p, id, days)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, d)
		return nil
	}))
	r.Put("/org/units/{id}/target", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in struct {
			Target int `json:"target"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.SetTarget(r.Context(), p, id, in.Target); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/org/performance", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		l, err := s.Performance(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"units": l})
		return nil
	}))
}
