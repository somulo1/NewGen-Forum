package chat

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

// wallAndRecordRoutes: the community wall, history and catch-up, records, the calendar and conversation rules.
func (s *Service) wallAndRecordRoutes(r chi.Router) {
	r.Put("/wall/{id}/reaction", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var in struct {
			Kind string `json:"kind"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.React(r.Context(), p, id, in.Kind); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/wall/counts", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var ids []uuid.UUID
		for _, v := range strings.Split(r.URL.Query().Get("ids"), ",") {
			if id, err := uuid.Parse(strings.TrimSpace(v)); err == nil {
				ids = append(ids, id)
			}
		}
		l, err := s.WallCounts(r.Context(), p, ids)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"items": l})
		return nil
	}))
	r.Put("/wall/{id}/vote", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var in struct {
			Option string `json:"option"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.Vote(r.Context(), p, id, in.Option); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Put("/conversations/{id}/plain-files", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		r.Body = http.MaxBytesReader(w, r.Body, s.A.Cfg.MaxUpload+1)
		f, err := s.UploadPlain(r.Context(), p, id, r.URL.Query().Get("name"), r.URL.Query().Get("type"), r.Body, r.ContentLength)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, f)
		return nil
	}))
	r.Get("/plain-files/{id}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		return s.ServePlain(r.Context(), p, id, w, r)
	}))
	r.Get("/conversations/{id}/messages", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var before *uuid.UUID
		if v := r.URL.Query().Get("before"); v != "" {
			b, err := uuid.Parse(v)
			if err != nil {
				return httpx.Bad("Bad page marker.")
			}
			before = &b
		}
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		m, err := s.History(r.Context(), p, id, before, n)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"messages": m})
		return nil
	}))
	// What changed since a point (new and removed messages), for keeping an open conversation current cheaply.
	r.Get("/conversations/{id}/changes", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		q := r.URL.Query()
		since, err := time.Parse(time.RFC3339Nano, q.Get("since"))
		if err != nil {
			return httpx.Bad("Bad sync marker.")
		}
		var after *uuid.UUID
		if v := q.Get("after"); v != "" {
			a, err := uuid.Parse(v)
			if err != nil {
				return httpx.Bad("Bad sync marker.")
			}
			after = &a
		}
		n, _ := strconv.Atoi(q.Get("limit"))
		page, err := s.Changes(r.Context(), p, id, since, after, n)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, page)
		return nil
	}))
	r.Delete("/conversations/{id}/messages/{mid}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		mid, err := uuid.Parse(chi.URLParam(r, "mid"))
		if err != nil {
			return httpx.ErrNotFound
		}
		if err := s.Delete(r.Context(), p, id, mid); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))

	rec := func(update bool) http.HandlerFunc {
		return httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			id, err := cid(r)
			if err != nil {
				return err
			}
			var in RecordIn
			if err := httpx.Decode(r, &in, 1<<20); err != nil {
				return err
			}
			if update {
				rid, err := uuid.Parse(chi.URLParam(r, "rid"))
				if err != nil {
					return httpx.ErrNotFound
				}
				in.ID = rid
			}
			x, err := s.SaveRecord(r.Context(), p, id, in, update)
			if err != nil {
				return err
			}
			code := 201
			if update {
				code = 200
			}
			httpx.JSON(w, code, x)
			return nil
		})
	}
	r.Post("/containers/{id}/records", rec(false))
	r.Put("/containers/{id}/records/{rid}", rec(true))
	r.Delete("/containers/{id}/records/{rid}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		rid, err := uuid.Parse(chi.URLParam(r, "rid"))
		if err != nil {
			return httpx.ErrNotFound
		}
		if err := s.DeleteRecord(r.Context(), p, id, rid); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/containers/{id}/records", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var since *time.Time
		if v := r.URL.Query().Get("since"); v != "" {
			t, err := time.Parse(time.RFC3339Nano, v)
			if err != nil {
				return httpx.Bad("Bad time.")
			}
			since = &t
		}
		n, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		l, err := s.Records(r.Context(), p, id, r.URL.Query().Get("collection"), since, n)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"records": l})
		return nil
	}))
	r.Get("/calendar", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		l, err := s.Calendar(r.Context(), p, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"entries": l})
		return nil
	}))
	r.Get("/conversations/{id}/rules", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		ru, err := s.Rules(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, ru)
		return nil
	}))
	r.Put("/conversations/{id}/rules", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var in ForumRules
		if err := httpx.Decode(r, &in, 16<<10); err != nil {
			return err
		}
		if err := s.SetRules(r.Context(), p, id, in); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}
