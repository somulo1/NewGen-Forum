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

func (s *Service) Private(r chi.Router) {
	s.conversationRoutes(r)
	s.keyRoutes(r)
	s.wallAndRecordRoutes(r)
}

// conversationRoutes: conversations: create, list, open, join and their people.
func (s *Service) conversationRoutes(r chi.Router) {
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
}

// keyRoutes: keys and messages: end-to-end keys, reading, sending, clearing.
func (s *Service) keyRoutes(r chi.Router) {
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
	r.Post("/e2ee/keys/request", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		n, err := s.RequestKeys(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"waiting": n})
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
	r.Get("/conversations/cleared", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		ids, err := s.Cleared(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"ids": ids})
		return nil
	}))
	// Delete a private chat or forum for me only (see Clear).
	r.Post("/conversations/{id}/clear", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		id, err := cid(r)
		if err != nil {
			return err
		}
		return s.clearRoutes(w, r, p, id)
	}))
	// the community wall (server-side): like / dislike / love ("" takes it back), a final poll vote, and plain attachments
}
