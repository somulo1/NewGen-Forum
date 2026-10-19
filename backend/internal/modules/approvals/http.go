package approvals

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

func (e *Engine) Public(chi.Router) {}

func (e *Engine) Private(r chi.Router) {
	r.Get("/approvals", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		list, err := e.List(r.Context(), p, r.URL.Query().Get("status"), n)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"requests": list})
		return nil
	}))
	decide := func(approve bool) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			id, err := uuid.Parse(chi.URLParam(r, "id"))
			if err != nil {
				return httpx.ErrNotFound
			}
			var in struct{ Comment string }
			if err := httpx.Decode(r, &in, 8<<10); err != nil {
				return err
			}
			x, err := e.Decide(r.Context(), p, id, approve, in.Comment)
			if err != nil {
				return err
			}
			httpx.JSON(w, 200, x)
			return nil
		})
	}
	r.Post("/approvals/{id}/approve", decide(true))
	r.Post("/approvals/{id}/reject", decide(false))
	r.Post("/approvals/{id}/cancel", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		if err := e.Cancel(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/approvals/flows", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		f, err := e.Flows(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"flows": f})
		return nil
	}))
	r.Put("/approvals/flows/{key}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in struct {
			Levels int `json:"levels"`
		}
		if err := httpx.Decode(r, &in, 8<<10); err != nil {
			return err
		}
		if err := e.SetFlow(r.Context(), p, chi.URLParam(r, "key"), in.Levels); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}
