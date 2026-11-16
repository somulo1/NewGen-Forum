package files

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

func (s *Service) Shares(ctx context.Context, p *tenant.Principal) ([]*Share, error) {
	out := []*Share{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "secureshare", "View link history"); err != nil {
			return err
		}
		rs, err := tx.Query(ctx, `select id from hub.shares order by created_at desc limit 200`)
		if err != nil {
			return err
		}
		var ids []uuid.UUID
		for rs.Next() {
			var id uuid.UUID
			if err := rs.Scan(&id); err != nil {
				rs.Close()
				return err
			}
			ids = append(ids, id)
		}
		rs.Close()
		// the whole page in one query (not one per link), in the same order
		xs, err := s.sharesByID(ctx, tx, ids)
		out = xs
		return err
	})
	return out, err
}

func (s *Service) Revoke(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	return s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var by uuid.UUID
		if err := tx.QueryRow(ctx, `select created_by from hub.shares where id=$1 and status <> 'Revoked'`, id).Scan(&by); err != nil {
			return httpx.ErrNotFound
		}
		if by != p.Person {
			if err := authz.Require(ctx, tx, p.Person, "secureshare", "View link history"); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `update hub.shares set status='Revoked', `+integrity.Bump+` where id=$1`, id); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "shares", "id=$1", id); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Revoked a share link", "Secure share", p.IP, map[string]any{"share": id})
	})
}

// ---- the public side of a link ----

type linkInfo struct {
	Org, Share, File uuid.UUID
	KeyEnvelope      []byte
	Size             int64
	Protected        bool
	BlobKey          string
	Pass             string
}

var errGone = httpx.Err(404, "link_gone", "This link has expired, was used up, or was removed.")

func (s *Service) link(ctx context.Context, token string) (*linkInfo, error) {
	h := crypto.Hash([]byte(token))
	l := &linkInfo{}
	var pass *string
	// The token hash is globally unique, so the link can be found without knowing the organization. The lookup reads
	// across organizations, so it runs as a narrow, read-only query through a function owned by the database owner.
	err := s.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select org_id, share_id, file_id, key_envelope, size, pass_hash, blob_key from hub.share_lookup($1)`, h).Scan(&l.Org, &l.Share, &l.File, &l.KeyEnvelope, &l.Size, &pass, &l.BlobKey)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errGone
	}
	if err != nil {
		return nil, err
	}
	l.Protected = pass != nil
	if pass != nil {
		l.Pass = *pass
	}
	return l, nil
}

// Open returns what the page needs before asking for a password.
func (s *Service) Open(ctx context.Context, token string) (map[string]any, error) {
	l, err := s.link(ctx, token)
	if err != nil {
		return nil, err
	}
	out := map[string]any{"size": l.Size, "password_protected": l.Protected}
	if !l.Protected {
		out["key_envelope"] = base64.StdEncoding.EncodeToString(l.KeyEnvelope)
	}
	return out, nil
}

// Unlock checks the link password (rate limited per link and per address) and hands back a short download pass and the key package.
func (s *Service) Unlock(ctx context.Context, token, password, ip string) (map[string]any, error) {
	l, err := s.link(ctx, token)
	if err != nil {
		return nil, err
	}
	if ok, retry := s.A.Limit.Allow(ctx, "share-pw", l.Share.String()+"|"+ip, 5, 15*time.Minute, true); !ok {
		return nil, httpx.Err(429, "rate_limited", "Too many wrong passwords. Try again in "+retry.Round(time.Minute).String()+".")
	}
	if l.Protected && !crypto.VerifyPassword(password, l.Pass) {
		return nil, httpx.Err(401, "wrong_password", "That password is not right.")
	}
	pass := crypto.RandomToken(24)
	if err := s.A.Rdb.Set(ctx, "dlpass:"+pass, l.Share.String(), 2*time.Minute).Err(); err != nil {
		return nil, err
	}
	return map[string]any{"download_pass": pass, "key_envelope": base64.StdEncoding.EncodeToString(l.KeyEnvelope), "size": l.Size}, nil
}

// Content streams the ciphertext once per download allowed. The count is taken before any bytes are sent.
func (s *Service) Content(ctx context.Context, token, pass, ip string, w http.ResponseWriter, r *http.Request) error {
	l, err := s.link(ctx, token)
	if err != nil {
		return err
	}
	if l.Protected || pass != "" {
		got, err := s.A.Rdb.GetDel(ctx, "dlpass:"+pass).Result()
		if err != nil || got != l.Share.String() {
			return httpx.ErrUnauthorized
		}
	}
	var used bool
	err = s.A.DB.WithOrg(ctx, l.Org, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `update hub.shares set downloads = downloads + 1, `+integrity.Bump+` where id=$1 and status='Active' and expires_at > now() and downloads < max_downloads`, l.Share)
		if err != nil {
			return err
		}
		used = tag.RowsAffected() == 1
		if !used {
			return nil
		}
		if err := s.A.Sign.Seal(ctx, tx, "shares", "id=$1", l.Share); err != nil {
			return err
		}
		return s.A.AuditEntry(ctx, tx, &app.Entry{Org: l.Org, Name: "External viewer", Action: "Opened a secure link", Module: "Secure share", IP: ip, Result: "Success"})
	})
	if err != nil {
		return err
	}
	if !used {
		return errGone
	}
	return s.serve(ctx, l.BlobKey, w, r)
}

func (s *Service) Public(r chi.Router) {
	r.Get("/share/{token}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		if ok, retry := s.A.Limit.Allow(r.Context(), "share-open", httpx.ClientIP(r), 60, time.Minute, false); !ok {
			return httpx.Err(429, "rate_limited", "Slow down. Try again in "+retry.String()+".")
		}
		out, err := s.Open(r.Context(), chi.URLParam(r, "token"))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	r.Post("/share/{token}/unlock", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		var in struct{ Password string }
		if err := httpx.Decode(r, &in, 1<<10); err != nil {
			return err
		}
		out, err := s.Unlock(r.Context(), chi.URLParam(r, "token"), in.Password, httpx.ClientIP(r))
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, out)
		return nil
	}))
	r.Get("/share/{token}/content", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		return s.Content(r.Context(), chi.URLParam(r, "token"), r.URL.Query().Get("pass"), httpx.ClientIP(r), w, r)
	}))
}

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
	s.quotaRoutes(r, org)
	uid := func(r *http.Request, k string) (uuid.UUID, error) {
		v, err := uuid.Parse(chi.URLParam(r, k))
		if err != nil {
			return v, httpx.ErrNotFound
		}
		return v, nil
	}
	r.Post("/folders/{id}/files", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		folder, err := uid(r, "id")
		if err != nil {
			return err
		}
		var in CreateIn
		if err := httpx.Decode(r, &in, 64<<10); err != nil {
			return err
		}
		f, err := s.Begin(r.Context(), p, folder, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, f)
		return nil
	}))
	r.Get("/folders/{id}/files", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		folder, err := uid(r, "id")
		if err != nil {
			return err
		}
		l, err := s.List(r.Context(), p, folder)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"files": l})
		return nil
	}))
	r.Put("/files/{id}/content", func(w http.ResponseWriter, r *http.Request) {
		httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
			p, err := tenant.MustOrg(r.Context())
			if err != nil {
				return err
			}
			id, err := uid(r, "id")
			if err != nil {
				return err
			}
			if r.ContentLength <= 0 {
				return httpx.Err(411, "length_required", "Send the file size with the upload.")
			}
			// the global 2 MB body cap does not apply to file uploads
			f, err := s.Upload(r.Context(), p, id, http.MaxBytesReader(w, r.Body, s.A.Cfg.MaxUpload+1), r.ContentLength)
			if err != nil {
				return err
			}
			httpx.JSON(w, 200, f)
			return nil
		})(w, r)
	})
	r.Get("/files/{id}/content", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uid(r, "id")
		if err != nil {
			return err
		}
		return s.Download(r.Context(), p, id, w, r)
	}))
	r.Delete("/files/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uid(r, "id")
		if err != nil {
			return err
		}
		if err := s.Delete(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
	r.Post("/files/{id}/shares", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uid(r, "id")
		if err != nil {
			return err
		}
		var in ShareIn
		if err := httpx.Decode(r, &in, 16<<10); err != nil {
			return err
		}
		sh, req, err := s.CreateShare(r.Context(), p, id, in)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, map[string]any{"share": sh, "approval": req})
		return nil
	}))
	r.Get("/shares", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		l, err := s.Shares(r.Context(), p)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"shares": l})
		return nil
	}))
	r.Delete("/shares/{id}", org(func(w http.ResponseWriter, r *http.Request, p *tenant.Principal) error {
		id, err := uid(r, "id")
		if err != nil {
			return err
		}
		if err := s.Revoke(r.Context(), p, id); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	}))
}

var _ = strings.TrimSpace

// filesByID loads many files at once, in the order of ids.
func (s *Service) filesByID(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) ([]*File, error) {
	out := []*File{}
	if len(ids) == 0 {
		return out, nil
	}
	rs, err := tx.Query(ctx, `select f.id, f.folder_id, f.uploader, p.first_name || ' ' || p.last_name, f.author_device, f.epoch, f.envelope, f.size, f.status, f.created_at
		from hub.files f join hub.people p on p.id = f.uploader where f.id = any($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	by := map[uuid.UUID]*File{}
	for rs.Next() {
		f := &File{}
		var env []byte
		if err := rs.Scan(&f.ID, &f.Folder, &f.Uploader, &f.Name, &f.Device, &f.Epoch, &env, &f.Size, &f.Status, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.Envelope = base64.StdEncoding.EncodeToString(env)
		by[f.ID] = f
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if f := by[id]; f != nil {
			out = append(out, f)
		}
	}
	return out, nil
}

// sharesByID loads many secure links at once, in the order of ids.
func (s *Service) sharesByID(ctx context.Context, tx pgx.Tx, ids []uuid.UUID) ([]*Share, error) {
	out := []*Share{}
	if len(ids) == 0 {
		return out, nil
	}
	rs, err := tx.Query(ctx, `select s.id, s.file_id, s.status, s.expires_at, s.max_downloads, s.downloads, s.pass_hash, p.first_name || ' ' || p.last_name, s.created_at
		from hub.shares s join hub.people p on p.id = s.created_by where s.id = any($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	by := map[uuid.UUID]*Share{}
	for rs.Next() {
		x := &Share{}
		var pass *string
		if err := rs.Scan(&x.ID, &x.File, &x.Status, &x.ExpiresAt, &x.MaxDownloads, &x.Downloads, &pass, &x.CreatedBy, &x.CreatedAt); err != nil {
			return nil, err
		}
		x.Protected = pass != nil
		if x.Status == "Active" && time.Now().After(x.ExpiresAt) {
			x.Status = "Expired"
		}
		by[x.ID] = x
	}
	if err := rs.Err(); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if x := by[id]; x != nil {
			out = append(out, x)
		}
	}
	return out, nil
}
