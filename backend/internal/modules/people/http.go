package people

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

func (s *Service) Public(chi.Router) {}

func pid(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, httpx.ErrNotFound
	}
	return id, nil
}

func (s *Service) Private(r chi.Router) {
	r.Get("/people", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		q := r.URL.Query()
		n, _ := strconv.Atoi(q.Get("limit"))
		page, err := s.List(r.Context(), p, ListQuery{Q: q.Get("q"), Dept: q.Get("department"), Branch: q.Get("branch"), Status: q.Get("status"), AvailableOnly: q.Get("available") == "1", Cursor: q.Get("cursor"), Limit: n})
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, page)
		return nil
	}))
	r.Get("/people/{id}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := pid(r)
		if err != nil {
			return err
		}
		x, err := s.Get(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, x)
		return nil
	}))
	r.Post("/people", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in struct {
			People []NewPerson `json:"people"`
		}
		if err := httpx.Decode(r, &in, 1<<20); err != nil {
			return err
		}
		res, err := s.Onboard(r.Context(), p, in.People)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"results": res})
		return nil
	}))
	r.Patch("/people/{id}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := pid(r)
		if err != nil {
			return err
		}
		var c Changes
		if err := httpx.Decode(r, &c, 16<<10); err != nil {
			return err
		}
		req, applied, err := s.Update(r.Context(), p, id, c)
		if err != nil {
			return err
		}
		code := 202
		if applied {
			code = 200
		}
		httpx.JSON(w, code, map[string]any{"applied": applied, "request": req})
		return nil
	}))
	r.Post("/people/{id}/status", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := pid(r)
		if err != nil {
			return err
		}
		var in struct{ Action string }
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.SetStatus(r.Context(), p, id, in.Action); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/people/{id}/reset-password", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := pid(r)
		if err != nil {
			return err
		}
		temp, err := s.ResetPassword(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"temporary_password": temp})
		return nil
	}))
	r.Post("/invites/accept", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		var in struct{ Token string }
		if err := httpx.Decode(r, &in, 2<<10); err != nil {
			return err
		}
		org, err := s.Accept(r.Context(), p.Account, in.Token)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"org": org})
		return nil
	}))
}
