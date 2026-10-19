package chat

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
	"hub/internal/tenant"
)

func (s *Service) Public(chi.Router) {}

func cid(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, httpx.ErrNotFound
	}
	return id, nil
}

type keysIn struct {
	Device uuid.UUID `json:"device"`
	Keys   []KeyWrap `json:"keys"`
}

func (s *Service) Private(r chi.Router) {
	r.Post("/conversations", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in CreateIn
		if err := httpx.Decode(r, &in, 8<<20); err != nil {
			return err
		}
		c, existed, err := s.Create(r.Context(), p, in)
		if err != nil {
			return err
		}
		code := 201
		if existed {
			code = 200
		}
		httpx.JSON(w, code, c)
		return nil
	}))
	r.Get("/conversations", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		q := r.URL.Query()
		n, _ := strconv.Atoi(q.Get("limit"))
		var before *time.Time
		if v := q.Get("before"); v != "" {
			t, err := time.Parse(time.RFC3339Nano, v)
			if err != nil {
				return httpx.Bad("Bad page marker.")
			}
			before = &t
		}
		l, err := s.List(r.Context(), p, q.Get("kind"), q.Get("open") == "1", n, before)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"conversations": l})
		return nil
	}))
	// Who a wall audience would cover (so a client can wrap the key for them before creating the wall).
	r.Get("/conversations/audience-members", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var ref *uuid.UUID
		if v := r.URL.Query().Get("ref"); v != "" {
			u, err := uuid.Parse(v)
			if err != nil {
				return httpx.Bad("Bad reference.")
			}
			ref = &u
		}
		var ids []uuid.UUID
		err = s.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			var e error
			ids, e = s.audienceMembers(r.Context(), tx, r.URL.Query().Get("audience"), ref)
			return e
		})
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"people": ids})
		return nil
	}))
	r.Get("/conversations/{id}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		c, err := s.Get(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, c)
		return nil
	}))
	r.Post("/conversations/{id}/join", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		if err := s.Join(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/conversations/{id}/members", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var in struct {
			keysIn
			People []uuid.UUID `json:"people"`
		}
		if err := httpx.Decode(r, &in, 8<<20); err != nil {
			return err
		}
		m, err := s.AddMembers(r.Context(), p, id, in.People, in.Keys, in.Device)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"added": m})
		return nil
	}))
	r.Delete("/conversations/{id}/members/{person}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		person, err := uuid.Parse(chi.URLParam(r, "person"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in keysIn
		if err := httpx.Decode(r, &in, 8<<20); err != nil {
			return err
		}
		if err := s.RemoveMember(r.Context(), p, id, person, in.Keys, in.Device); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Patch("/conversations/{id}/members/{person}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		person, err := uuid.Parse(chi.URLParam(r, "person"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var in struct {
			Role string `json:"role"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.SetMemberRole(r.Context(), p, id, person, in.Role); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Get("/conversations/{id}/keys", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		k, err := s.Keys(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"keys": k})
		return nil
	}))
	r.Post("/conversations/{id}/keys", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var in keysIn
		if err := httpx.Decode(r, &in, 8<<20); err != nil {
			return err
		}
		n, err := s.AddKeys(r.Context(), p, id, in.Keys, in.Device)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"stored": n})
		return nil
	}))
	r.Get("/conversations/{id}/missing-keys", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		m, err := s.MissingKeys(r.Context(), p, id)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"missing": m})
		return nil
	}))
	r.Get("/e2ee/todo", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		l, err := s.Todo(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"todo": l})
		return nil
	}))
	r.Post("/conversations/{id}/rotate", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var in keysIn
		if err := httpx.Decode(r, &in, 8<<20); err != nil {
			return err
		}
		if err := s.Rotate(r.Context(), p, id, in.Keys, in.Device); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/conversations/{id}/read", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		var in struct {
			Message uuid.UUID `json:"message"`
		}
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		if err := s.MarkRead(r.Context(), p, id, in.Message); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/conversations/{id}/messages", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		if ok, retry := s.A.Limit.Allow(r.Context(), "msg", p.Account.String(), 120, time.Minute, false); !ok {
			w.Header().Set("Retry-After", retry.String())
			return httpx.Err(429, "rate_limited", "You are sending messages too fast.")
		}
		var in SendIn
		if err := httpx.Decode(r, &in, 512<<10); err != nil {
			return err
		}
		m, err := s.Send(r.Context(), p, id, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, m)
		return nil
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
