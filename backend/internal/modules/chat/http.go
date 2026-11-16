package chat

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"hub/internal/httpx"
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
