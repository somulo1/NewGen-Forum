package roles

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

func (s *Service) Public(chi.Router) {}

func reply(w http.ResponseWriter, applied bool, req any) {
	code := 202
	if applied {
		code = 200
	}
	httpx.JSON(w, code, map[string]any{"applied": applied, "request": req})
}

func (s *Service) Private(r chi.Router) {
	s.groupRoutes(r)
	r.Get("/me", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		m, err := s.Me(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, m)
		return nil
	}))
	r.Patch("/me", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in ProfileEdit
		if err := httpx.Decode(r, &in, 1<<20); err != nil {
			return err
		}
		if err := s.EditMe(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/me/prefs", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		raw, err := s.Prefs(r.Context(), p)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write(raw)
		return nil
	}))
	r.Put("/me/prefs", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var raw json.RawMessage
		if err := httpx.Decode(r, &raw, 16<<10); err != nil {
			return err
		}
		if err := s.SetPrefs(r.Context(), p, raw); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/access/catalog", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		c, err := s.Catalog(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, c)
		return nil
	}))
	r.Post("/access/modules", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in RightIn
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		if err := s.AddModule(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/access/rights", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in RightIn
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		if err := s.AddRight(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Patch("/access/rights", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in RightIn
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		if err := s.UpdateRight(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Delete("/access/rights", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		if err := s.RemoveRight(r.Context(), p, r.URL.Query().Get("module"), r.URL.Query().Get("name")); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/roles", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in RoleIn
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		key, err := s.CreateRole(r.Context(), p, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"id": key})
		return nil
	}))
	r.Patch("/roles/{key}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in RoleIn
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		if err := s.UpdateRole(r.Context(), p, chi.URLParam(r, "key"), in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Delete("/roles/{key}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		if err := s.DeleteRole(r.Context(), p, chi.URLParam(r, "key")); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/roles", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		l, err := s.Roles(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"roles": l})
		return nil
	}))
	r.Get("/roles/{key}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		d, err := s.Role(r.Context(), p, chi.URLParam(r, "key"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, d)
		return nil
	}))
	r.Put("/roles/{key}/rights", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in rightChange
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		req, applied, err := s.SetRoleRight(r.Context(), p, chi.URLParam(r, "key"), in)
		if err != nil {
			return err
		}
		reply(w, applied, req)
		return nil
	}))
	r.Put("/people/{id}/rights", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in rightChange
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		req, applied, err := s.SetPersonRight(r.Context(), p, id, in)
		if err != nil {
			return err
		}
		reply(w, applied, req)
		return nil
	}))
	r.Put("/roles/reporting", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in Reporting
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		if err := s.SetRoleReport(r.Context(), p, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}
