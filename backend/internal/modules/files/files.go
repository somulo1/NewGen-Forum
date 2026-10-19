// Package files is encrypted file storage and secure sharing. A file's bytes are encrypted on the sender's device with a
// random key; the server stores ciphertext and never sees that key, the file name or its type.
package files

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/authz"
	"hub/internal/blob"
	"hub/internal/crypto"
	"hub/internal/e2ee"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/modules/approvals"
	"hub/internal/orgsettings"
	"hub/internal/realtime"
	"hub/internal/tenant"
)

type Service struct {
	A     *app.App
	Ap    *approvals.Engine
	Store blob.Store
}

func New(a *app.App, ap *approvals.Engine) *Service {
	s := &Service{A: a, Ap: ap, Store: blob.Dir{Root: a.Cfg.BlobDir}}
	ap.Register("share.activate", s.applyShare)
	return s
}

type File struct {
	ID        uuid.UUID `json:"id"`
	Folder    uuid.UUID `json:"folder"`
	Uploader  uuid.UUID `json:"uploader"`
	Name      string    `json:"uploader_name"`
	Device    uuid.UUID `json:"author_device"`
	Epoch     int       `json:"epoch"`
	Envelope  string    `json:"envelope"` // name, type and file key, sealed for the folder's members
	Size      *int64    `json:"size,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateIn struct {
	ID       uuid.UUID `json:"id"`
	Epoch    int       `json:"epoch"`
	Device   uuid.UUID `json:"device"`
	Envelope string    `json:"envelope"`
}

// Begin registers an upload. The content follows with PUT /files/{id}/content.
func (s *Service) Begin(ctx context.Context, p *tenant.Principal, folder uuid.UUID, in CreateIn) (*File, error) {
	env, err := base64.StdEncoding.DecodeString(in.Envelope)
	if err != nil || e2ee.ValidateEnvelope(env) != nil {
		return nil, httpx.Bad("That is not a valid encrypted file description.")
	}
	if in.ID.Version() != 7 {
		return nil, httpx.Bad("Files need a time-ordered id (UUID v7) made by the sender.")
	}
	var out *File
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var kind string
		var epoch int
		if err := tx.QueryRow(ctx, `select c.kind, c.epoch from hub.conversations c join hub.conversation_members m on m.conversation_id=c.id and m.account_id=$2 and m.left_at is null where c.id=$1`, folder, p.Account).Scan(&kind, &epoch); err != nil {
			return httpx.ErrNotFound
		}
		// folders hold the Files library; the other containers hold attachments (wall photos, chat files, ticket evidence)
		// that only their members can open
		switch kind {
		case "folder":
			if err := authz.Require(ctx, tx, p.Person, "files", "Upload"); err != nil {
				return err
			}
		case "wall", "forum", "dm", "ticket", "project", "event":
		default:
			return httpx.Bad("Files cannot be attached here.")
		}
		if in.Epoch != epoch {
			return httpx.Err(409, "epoch_changed", "The folder's key changed. Fetch it and try again.")
		}
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.devices where id=$1 and account_id=$2 and revoked_at is null)`, in.Device, p.Account).Scan(&ok); err != nil || !ok {
			return httpx.Bad("That device is not yours or was removed.")
		}
		if _, err := tx.Exec(ctx, `insert into hub.files (org_id, id, folder_id, uploader, author_account, author_device, epoch, envelope) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
			p.Org, in.ID, folder, p.Person, p.Account, in.Device, epoch, env); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "files", "id=$1 and org_id=$2", in.ID, p.Org); err != nil {
			return err
		}
		out = &File{ID: in.ID, Folder: folder, Uploader: p.Person, Name: p.Name, Device: in.Device, Epoch: epoch, Envelope: in.Envelope, Status: "uploading", CreatedAt: app.Now()}
		return nil
	})
	return out, err
}

func blobKey(org, id uuid.UUID) string { return "o/" + org.String() + "/" + id.String() }

// Upload stores the ciphertext. Only the uploader can add content, once.
func (s *Service) Upload(ctx context.Context, p *tenant.Principal, id uuid.UUID, body io.Reader, declared int64) (*File, error) {
	if declared <= 0 || declared > s.A.Cfg.MaxUpload {
		return nil, httpx.Err(413, "too_large", fmt.Sprintf("Files can be up to %d MB.", s.A.Cfg.MaxUpload>>20))
	}
	var folder uuid.UUID
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var status string
		var up uuid.UUID
		if err := tx.QueryRow(ctx, `select folder_id, uploader, status from hub.files where id=$1`, id).Scan(&folder, &up, &status); err != nil {
			return httpx.ErrNotFound
		}
		if up != p.Person || status != "uploading" {
			return httpx.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	key := blobKey(p.Org, id)
	n, sum, err := s.Store.Put(ctx, key, body, declared)
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
	var out *File
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `update hub.files set size=$2, sha256=$3, blob_key=$4, status='ready', updated_at=now(), `+integrity.Bump+` where id=$1 and uploader=$5 and status='uploading'`, id, n, sum, key, p.Person)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		if err := s.A.Sign.Seal(ctx, tx, "files", "id=$1 and org_id=$2", id, p.Org); err != nil {
			return err
		}
		out, err = s.file(ctx, tx, id)
		return err
	})
	if err != nil {
		_ = s.Store.Delete(ctx, key)
		return nil, err
	}
	d, _ := json.Marshal(out)
	s.A.Publish(ctx, realtime.Event{Type: "file.added", Channel: realtime.ConvChannel(folder), Org: p.Org.String(), Actor: p.Person.String(), Data: d})
	return out, nil
}

func (s *Service) file(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*File, error) {
	f := &File{}
	var env []byte
	err := tx.QueryRow(ctx, `select f.id, f.folder_id, f.uploader, p.first_name || ' ' || p.last_name, f.author_device, f.epoch, f.envelope, f.size, f.status, f.created_at
		from hub.files f join hub.people p on p.id = f.uploader where f.id=$1`, id).Scan(&f.ID, &f.Folder, &f.Uploader, &f.Name, &f.Device, &f.Epoch, &env, &f.Size, &f.Status, &f.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpx.ErrNotFound
	}
	f.Envelope = base64.StdEncoding.EncodeToString(env)
	return f, err
}

// List returns a folder's ready files. Only members of the folder can see anything about them.
func (s *Service) List(ctx context.Context, p *tenant.Principal, folder uuid.UUID) ([]*File, error) {
	out := []*File{}
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "files", "View"); err != nil {
			return err
		}
		var first int
		if err := tx.QueryRow(ctx, `select first_epoch from hub.conversation_members where conversation_id=$1 and account_id=$2 and left_at is null`, folder, p.Account).Scan(&first); err != nil {
			return httpx.ErrNotFound
		}
		rs, err := tx.Query(ctx, `select id from hub.files where folder_id=$1 and status='ready' and epoch >= $2 order by created_at desc limit 500`, folder, first)
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
		for _, id := range ids {
			f, err := s.file(ctx, tx, id)
			if err != nil {
				return err
			}
			out = append(out, f)
		}
		return nil
	})
	return out, err
}

// Download streams the ciphertext to a member of the file's folder.
func (s *Service) Download(ctx context.Context, p *tenant.Principal, id uuid.UUID, w http.ResponseWriter, r *http.Request) error {
	var key string
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		if err := authz.Require(ctx, tx, p.Person, "files", "View"); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `select f.blob_key from hub.files f join hub.conversation_members m on m.conversation_id = f.folder_id and m.account_id=$2 and m.left_at is null and m.first_epoch <= f.epoch
			where f.id=$1 and f.status='ready'`, id, p.Account).Scan(&key)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	s.countDownload(ctx, p)
	return s.serve(ctx, key, w, r)
}

// countDownload watches for someone pulling a lot of files in a short time (security rule r3).
func (s *Service) countDownload(ctx context.Context, p *tenant.Principal) {
	k := fmt.Sprintf("dl:%s:%d", p.Person, time.Now().Unix()/900)
	n, err := s.A.Rdb.Incr(ctx, k).Result()
	if err != nil {
		return
	}
	if n == 1 {
		s.A.Rdb.Expire(ctx, k, 20*time.Minute)
	}
	if n == 51 {
		_ = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
			return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Downloaded 51 files in 15 minutes", "Files", p.IP, nil)
		})
	}
}

func (s *Service) serve(ctx context.Context, key string, w http.ResponseWriter, r *http.Request) error {
	f, _, err := s.Store.Get(ctx, key)
	if errors.Is(err, blob.ErrNotFound) {
		return httpx.ErrNotFound
	}
	if err != nil {
		return err
	}
	defer f.Close()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Disposition", "attachment")
	http.ServeContent(w, r, "", time.Time{}, f)
	return nil
}

func (s *Service) Delete(ctx context.Context, p *tenant.Principal, id uuid.UUID) error {
	var key *string
	err := s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		var up uuid.UUID
		if err := tx.QueryRow(ctx, `select uploader, blob_key from hub.files where id=$1 and status <> 'deleted' for update`, id).Scan(&up, &key); err != nil {
			return httpx.ErrNotFound
		}
		if up != p.Person {
			if err := authz.Require(ctx, tx, p.Person, "files", "Delete"); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `update hub.files set status='deleted', envelope='\x00'::bytea, blob_key=null, updated_at=now(), `+integrity.Bump+` where id=$1`, id); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "files", "id=$1 and org_id=$2", id, p.Org); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update hub.shares set status='Revoked', `+integrity.Bump+` where file_id=$1 and status <> 'Revoked'`, id); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "shares", "file_id=$1", id); err != nil {
			return err
		}
		return s.A.Audit(ctx, tx, p.Org, &p.Person, p.Name, "Deleted a file", "Files", p.IP, map[string]any{"file": id})
	})
	if err == nil && key != nil {
		_ = s.Store.Delete(ctx, *key)
	}
	return err
}

// ---- secure share ----

type Share struct {
	ID           uuid.UUID `json:"id"`
	File         uuid.UUID `json:"file"`
	Status       string    `json:"status"`
	ExpiresAt    time.Time `json:"expires_at"`
	MaxDownloads int       `json:"max_downloads"`
	Downloads    int       `json:"downloads"`
	Protected    bool      `json:"password_protected"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	Token        string    `json:"token,omitempty"` // shown once, at creation
}

type ShareIn struct {
	ExpiresInHours int    `json:"expires_in_hours"`
	MaxDownloads   int    `json:"max_downloads"`
	Password       string `json:"password"`
	KeyEnvelope    string `json:"key_envelope"` // made on the device with the link key; the link key itself never reaches us
}

// CreateShare makes a link to a file for someone outside Hub. Flow f3 (outside sharing) and f6 (links lasting a week or
// more) may need approval before the link works.
func (s *Service) CreateShare(ctx context.Context, p *tenant.Principal, file uuid.UUID, in ShareIn) (*Share, *approvals.Request, error) {
	if in.ExpiresInHours < 1 || in.ExpiresInHours > 24*30 {
		return nil, nil, httpx.Bad("A link can last from 1 hour to 30 days.")
	}
	if in.MaxDownloads < 1 || in.MaxDownloads > 1000 {
		return nil, nil, httpx.Bad("Choose between 1 and 1,000 downloads.")
	}
	if in.Password != "" && (len(in.Password) < 6 || len(in.Password) > 128) {
		return nil, nil, httpx.Bad("A link password needs 6 to 128 characters.")
	}
	ke, err := base64.StdEncoding.DecodeString(in.KeyEnvelope)
	if err != nil || e2ee.ValidateEnvelope(ke) != nil && len(ke) < 48 {
		return nil, nil, httpx.Bad("The link's key package is not valid.")
	}
	token := crypto.RandomToken(24)
	var passHash *string
	if in.Password != "" {
		h := crypto.HashPassword(in.Password, s.A.PW)
		passHash = &h
	}
	var out *Share
	var req *approvals.Request
	err = s.A.DB.WithOrg(ctx, p.Org, func(tx pgx.Tx) error {
		// either the Secure share widget's right or the Files page's "share outside Hub" right is enough
		a, err := authz.Can(ctx, tx, p.Person, "secureshare", "Create link")
		if err != nil {
			return err
		}
		b, err := authz.Can(ctx, tx, p.Person, "files", "Share outside Hub")
		if err != nil {
			return err
		}
		if !a && !b {
			return httpx.ErrForbidden
		}
		sec := orgsettings.DefaultSecurity()
		if err := orgsettings.Get(ctx, s.A, tx, p.Org, "security", &sec); err != nil {
			return err
		}
		if !sec.ExternalSharing {
			return httpx.Err(403, "sharing_off", "Your organization has turned off sharing outside Hub.")
		}
		var ok bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from hub.files f join hub.conversation_members m on m.conversation_id=f.folder_id and m.account_id=$2 and m.left_at is null where f.id=$1 and f.status='ready')`, file, p.Account).Scan(&ok); err != nil || !ok {
			return httpx.ErrNotFound
		}
		id := app.NewID()
		if _, err := tx.Exec(ctx, `insert into hub.shares (org_id, id, file_id, token_hash, created_by, status, expires_at, max_downloads, pass_hash, key_envelope) values ($1,$2,$3,$4,$5,'Pending approval',$6,$7,$8,$9)`,
			p.Org, id, file, crypto.Hash([]byte(token)), p.Person, app.Now().Add(time.Duration(in.ExpiresInHours)*time.Hour), in.MaxDownloads, passHash, ke); err != nil {
			return err
		}
		if err := s.A.Sign.Seal(ctx, tx, "shares", "id=$1", id); err != nil {
			return err
		}
		flow := "f3"
		if in.ExpiresInHours >= 24*7 {
			flow = "f6"
		}
		r, applied, err := s.Ap.Submit(ctx, tx, p, flow, "share.activate", "Share a file outside Hub", "", map[string]any{"share": id})
		if err != nil {
			return err
		}
		if !applied {
			req = r
		}
		out, err = s.share(ctx, tx, id)
		if err == nil {
			out.Token = token
		}
		return err
	})
	return out, req, err
}

func (s *Service) applyShare(ctx context.Context, tx pgx.Tx, a approvals.Applied) error {
	var pl struct {
		Share uuid.UUID `json:"share"`
	}
	if err := json.Unmarshal(a.Payload, &pl); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `update hub.shares set status='Active', `+integrity.Bump+` where id=$1 and status='Pending approval'`, pl.Share); err != nil {
		return err
	}
	return s.A.Sign.Seal(ctx, tx, "shares", "id=$1", pl.Share)
}

func (s *Service) share(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Share, error) {
	x := &Share{}
	var pass *string
	err := tx.QueryRow(ctx, `select s.id, s.file_id, s.status, s.expires_at, s.max_downloads, s.downloads, s.pass_hash, p.first_name || ' ' || p.last_name, s.created_at
		from hub.shares s join hub.people p on p.id = s.created_by where s.id=$1`, id).Scan(&x.ID, &x.File, &x.Status, &x.ExpiresAt, &x.MaxDownloads, &x.Downloads, &pass, &x.CreatedBy, &x.CreatedAt)
	x.Protected = pass != nil
	if x.Status == "Active" && time.Now().After(x.ExpiresAt) {
		x.Status = "Expired"
	}
	return x, err
}

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
		for _, id := range ids {
			x, err := s.share(ctx, tx, id)
			if err != nil {
				return err
			}
			out = append(out, x)
		}
		return nil
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
