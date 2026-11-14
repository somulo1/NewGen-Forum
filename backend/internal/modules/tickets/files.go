package tickets

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/blob"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

// A ticket's attachments (evidence, screenshots) are plain files the server keeps for the people on the ticket. They are not
// end-to-end encrypted: nothing has to be sealed, unsealed or re-keyed to raise, open or hand on a ticket.
type File struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	By        string    `json:"by"`
	CreatedAt time.Time `json:"created_at"`
}

func ticketBlobKey(org, id uuid.UUID) string { return "t/" + org.String() + "/" + id.String() }

// only types that cannot run script in the browser keep their type; everything else is served as a download
var safeType = regexp.MustCompile(`^(image/(png|jpeg|gif|webp|avif|bmp)|video/(mp4|webm|ogg|quicktime)|audio/(mpeg|ogg|wav|webm|mp4)|application/pdf|text/(plain|csv))$`)

func cleanType(t string) string {
	t = strings.ToLower(strings.TrimSpace(strings.SplitN(t, ";", 2)[0]))
	if safeType.MatchString(t) {
		return t
	}
	return "application/octet-stream"
}

// mayRead: the caller can view tickets and is on this one.
func (s *Service) mayRead(ctx context.Context, tx pgx.Tx, p *tenant.Principal, ticket uuid.UUID) error {
	if err := authz.Require(ctx, tx, p.Person, "tickets", "View"); err != nil {
		return err
	}
	var exists bool
	if err := tx.QueryRow(ctx, `select exists (select 1 from hub.tickets where conversation_id=$1)`, ticket).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return httpx.ErrNotFound
	}
	ok, err := s.participant(ctx, tx, p, ticket)
	if err != nil {
		return err
	}
	if !ok {
		return httpx.ErrNotFound
	}
	return nil
}

// AddFile stores an attachment on a ticket.
func (s *Service) AddFile(ctx context.Context, p *tenant.Principal, ticket uuid.UUID, name, typ string, body io.Reader, declared int64) (*File, error) {
	if s.Store == nil {
		return nil, httpx.Err(503, "no_storage", "Files cannot be stored right now.")
	}
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > 255 || strings.ContainsAny(name, "/\\\x00") {
		return nil, httpx.Bad("Give the file a name.")
	}
	if declared <= 0 || declared > s.A.Cfg.MaxUpload {
		return nil, httpx.Err(413, "too_large", fmt.Sprintf("Files can be up to %d MB.", s.A.Cfg.MaxUpload>>20))
	}
	if err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := s.mayRead(ctx, tx, p, ticket); err != nil {
			return err
		}
		var step int
		if err := tx.QueryRow(ctx, `select step from hub.tickets where conversation_id=$1`, ticket).Scan(&step); err != nil {
			return err
		}
		if step >= 4 {
			return httpx.Err(409, "ticket_closed", "This ticket is closed.")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	id := app.NewID()
	key, n, _, err := s.Store.Put(ctx, ticketBlobKey(p.Org, id), body, declared)
	if errors.Is(err, blob.ErrTooLarge) {
		return nil, httpx.Err(413, "too_large", "That file is larger than you said it was.")
	}
	if err != nil {
		return nil, err
	}
	if n != declared {
		_ = s.Store.Delete(ctx, key)
		return nil, httpx.Bad("The upload was cut short. Try again.")
	}
	out := &File{ID: id, Name: name, Type: cleanType(typ), Size: n, By: p.Name}
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `insert into hub.ticket_files (org_id, id, ticket_id, uploader, name, type, size, blob_key) values ($1,$2,$3,$4,$5,$6,$7,$8) returning created_at`,
			p.Org, id, ticket, p.Person, name, out.Type, n, key).Scan(&out.CreatedAt); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select first_name || ' ' || last_name from hub.people where id=$1`, p.Person).Scan(&out.By)
	})
	if err != nil {
		_ = s.Store.Delete(ctx, key)
		return nil, err
	}
	return out, nil
}

// files lists the attachments of a ticket, oldest first.
func (s *Service) files(ctx context.Context, tx pgx.Tx, ticket uuid.UUID) ([]File, error) {
	rs, err := tx.Query(ctx, `select f.id, f.name, f.type, f.size, coalesce(p.first_name || ' ' || p.last_name, ''), f.created_at
		from hub.ticket_files f left join hub.people p on p.id = f.uploader where f.ticket_id=$1 order by f.created_at, f.id`, ticket)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	out := []File{}
	for rs.Next() {
		var f File
		if err := rs.Scan(&f.ID, &f.Name, &f.Type, &f.Size, &f.By, &f.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rs.Err()
}

// ServeFile sends an attachment to someone on the ticket.
func (s *Service) ServeFile(ctx context.Context, p *tenant.Principal, ticket, id uuid.UUID, w http.ResponseWriter, r *http.Request) error {
	if s.Store == nil {
		return httpx.ErrNotFound
	}
	var key, typ string
	if err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := s.mayRead(ctx, tx, p, ticket); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `select blob_key, type from hub.ticket_files where id=$1 and ticket_id=$2`, id, ticket).Scan(&key, &typ)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		return err
	}); err != nil {
		return err
	}
	f, _, err := s.Store.Get(ctx, key)
	if errors.Is(err, blob.ErrNotFound) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	defer f.Close()
	w.Header().Set("Content-Type", cleanType(typ))
	w.Header().Set("Cache-Control", "private, max-age=86400, immutable")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, "", time.Time{}, f)
	return nil
}
