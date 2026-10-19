package tickets

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

func (s *Service) Public(chi.Router) {}

func (s *Service) Private(r chi.Router) {
	org := func(h func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			return h(w, r, p)
		})
	}
	id := func(r *http.Request) (uuid.UUID, error) {
		v, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return v, httpx.ErrNotFound
		}
		return v, nil
	}
	r.Post("/tickets", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var in CreateIn
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		t, err := s.Create(r.Context(), p, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, t)
		return nil
	}))
	r.Get("/tickets", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		q := r.URL.Query()
		n, _ := strconv.Atoi(q.Get("limit"))
		lq := ListQuery{Mine: q.Get("mine") == "1", Open: q.Get("open") == "1", Late: q.Get("late") == "1", Limit: n}
		if h, err := uuid.Parse(q.Get("handler")); err == nil {
			lq.Handler = &h
		}
		l, err := s.List(r.Context(), p, lq)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"tickets": l})
		return nil
	}))
	r.Get("/tickets/settings", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		c, err := s.Settings(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, c)
		return nil
	}))
	r.Put("/tickets/settings", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var c Config
		if err := httpx.Decode(r, &c, 16<<10); err != nil {
			return err
		}
		if err := s.SetSettings(r.Context(), p, c); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/tickets/categories", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var ref *uuid.UUID
		if v, err := uuid.Parse(r.URL.Query().Get("to_ref")); err == nil {
			ref = &v
		}
		t := r.URL.Query().Get("to_type")
		if t == "" {
			t = "desk"
		}
		l, err := s.Categories(r.Context(), p, t, ref)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"categories": l})
		return nil
	}))
	r.Post("/tickets/categories", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		var c Category
		if err := httpx.Decode(r, &c, 8<<10); err != nil {
			return err
		}
		out, err := s.SaveCategory(r.Context(), p, nil, c)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, out)
		return nil
	}))
	r.Put("/tickets/categories/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		cid, err := id(r)
		if err != nil {
			return err
		}
		var c Category
		if err := httpx.Decode(r, &c, 8<<10); err != nil {
			return err
		}
		out, err := s.SaveCategory(r.Context(), p, &cid, c)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	r.Delete("/tickets/categories/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		cid, err := id(r)
		if err != nil {
			return err
		}
		if err := s.DeleteCategory(r.Context(), p, cid); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/tickets/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		tid, err := id(r)
		if err != nil {
			return err
		}
		t, err := s.Get(r.Context(), p, tid)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, t)
		return nil
	}))
	r.Post("/tickets/{id}/act", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		tid, err := id(r)
		if err != nil {
			return err
		}
		var a Action
		if err := httpx.Decode(r, &a, 8<<10); err != nil {
			return err
		}
		t, err := s.Act(r.Context(), p, tid, a)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, t)
		return nil
	}))
}
