// Package keys is the public key directory for end-to-end encryption. It stores PUBLIC keys and one opaque, client-encrypted
// backup per account. It never holds anything that could decrypt a message.
package keys

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/e2ee"
	"hub/internal/httpx"
	"hub/internal/integrity"
	"hub/internal/tenant"
)

type Module struct{ A *app.App }

func New(a *app.App) *Module { return &Module{A: a} }

func (m *Module) Public(chi.Router) {}

type Device struct {
	ID          uuid.UUID `json:"id"`
	Account     uuid.UUID `json:"account,omitempty"`
	Label       string    `json:"label"`
	Identity    string    `json:"identity_key"`
	Signing     string    `json:"signing_key"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
	Revoked     bool      `json:"revoked,omitempty"`
}

// Fingerprint is what two people compare (in person, or over a call) to be sure the directory did not swap a key.
func Fingerprint(identity, signing []byte) string {
	h := sha256.Sum256(append(append([]byte("hub/fp/v1"), identity...), signing...))
	s := hex.EncodeToString(h[:15])
	out := ""
	for i := 0; i < len(s); i += 5 {
		if i > 0 {
			out += " "
		}
		out += s[i : i+5]
	}
	return out
}

// BindingMessage is what a device signs to prove its two keys belong together and to this account.
func BindingMessage(identity []byte, account uuid.UUID) []byte {
	return append(append([]byte("hub/device/v1|"), identity...), account[:]...)
}

func b64(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) }

func (m *Module) Register(ctx context.Context, p *tenant.Principal, label, identityB64, signingB64, bindingB64 string) (*Device, error) {
	id, e1 := b64(identityB64)
	sg, e2 := b64(signingB64)
	bd, e3 := b64(bindingB64)
	if e1 != nil || e2 != nil || e3 != nil {
		return nil, httpx.Bad("Keys must be base64.")
	}
	if e2ee.ValidatePublic(e2ee.PublicKeys{Identity: id, Signing: sg}) != nil {
		return nil, httpx.Bad("Those are not valid device keys.")
	}
	if !ed25519.Verify(sg, BindingMessage(id, p.Account), bd) {
		return nil, httpx.Bad("The device could not prove it owns these keys.")
	}
	if len(label) > 60 {
		label = label[:60]
	}
	d := &Device{ID: app.NewID(), Label: label, Identity: identityB64, Signing: signingB64, Fingerprint: Fingerprint(id, sg), CreatedAt: app.Now()}
	err := m.A.DB.Tx(ctx, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from hub.devices where account_id=$1 and revoked_at is null`, p.Account).Scan(&n); err != nil {
			return err
		}
		if n >= 10 {
			return httpx.Err(409, "too_many_devices", "Remove an old device before adding another.")
		}
		if _, err := tx.Exec(ctx, `insert into hub.devices (id, account_id, label, identity_key, signing_key, binding_sig) values ($1,$2,$3,$4,$5,$6)`, d.ID, p.Account, label, id, sg, bd); err != nil {
			return err
		}
		return m.A.Sign.Seal(ctx, tx, "devices", "id=$1", d.ID)
	})
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (m *Module) Devices(r *http.Request, account uuid.UUID) ([]Device, error) {
	out := []Device{}
	err := m.A.DB.Tx(r.Context(), func(tx pgx.Tx) error {
		rs, err := tx.Query(r.Context(), `select id, label, identity_key, signing_key, created_at, revoked_at is not null from hub.devices where account_id=$1 order by created_at`, account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var d Device
			var id, sg []byte
			if err := rs.Scan(&d.ID, &d.Label, &id, &sg, &d.CreatedAt, &d.Revoked); err != nil {
				return err
			}
			d.Identity, d.Signing, d.Fingerprint = base64.StdEncoding.EncodeToString(id), base64.StdEncoding.EncodeToString(sg), Fingerprint(id, sg)
			out = append(out, d)
		}
		return rs.Err()
	})
	return out, err
}

func (m *Module) Private(r chi.Router) {
	r.Post("/e2ee/devices", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		var in struct {
			Label    string `json:"label"`
			Identity string `json:"identity_key"`
			Signing  string `json:"signing_key"`
			Binding  string `json:"binding_sig"`
		}
		if err := httpx.Decode(r, &in, 4<<10); err != nil {
			return err
		}
		d, err := m.Register(r.Context(), p, in.Label, in.Identity, in.Signing, in.Binding)
		if err != nil {
			return err
		}
		httpx.JSON(w, 201, d)
		return nil
	}))
	r.Get("/e2ee/devices", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		l, err := m.Devices(r, p.Account)
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"devices": l})
		return nil
	}))
	r.Delete("/e2ee/devices/{id}", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		return m.A.DB.Tx(r.Context(), func(tx pgx.Tx) error {
			tag, err := tx.Exec(r.Context(), `update hub.devices set revoked_at=now(), `+integrity.Bump+` where id=$1 and account_id=$2 and revoked_at is null`, id, p.Account)
			if err != nil {
				return err
			}
			if tag.RowsAffected() > 0 {
				if err := m.A.Sign.Seal(r.Context(), tx, "devices", "id=$1", id); err != nil {
					return err
				}
			}
			if tag.RowsAffected() == 0 {
				return httpx.ErrNotFound
			}
			w.WriteHeader(204)
			return nil
		})
	}))
	// Public keys of a colleague's devices (same organization only), so a client can wrap a conversation key for them.
	r.Get("/e2ee/people/{id}/devices", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		pid, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			return httpx.ErrNotFound
		}
		var acct *uuid.UUID
		if err := m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			return tx.QueryRow(r.Context(), `select account_id from hub.people where id=$1 and status='active'`, pid).Scan(&acct)
		}); errors.Is(err, pgx.ErrNoRows) || err == nil && acct == nil {
			return httpx.ErrNotFound
		} else if err != nil {
			return err
		}
		l, err := m.Devices(r, *acct)
		if err != nil {
			return err
		}
		active := l[:0]
		for _, d := range l {
			if !d.Revoked {
				active = append(active, d)
			}
		}
		httpx.JSON(w, 200, map[string]any{"person": pid, "account": acct, "devices": active})
		return nil
	}))
	// Public keys of several colleagues' devices in one call (to wrap a conversation key for a whole group).
	r.Post("/e2ee/people/devices", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.MustOrg(r.Context())
		if err != nil {
			return err
		}
		var in struct {
			People []uuid.UUID `json:"people"`
		}
		if err := httpx.Decode(r, &in, 64<<10); err != nil {
			return err
		}
		if len(in.People) == 0 || len(in.People) > 5000 {
			return httpx.Bad("Give between 1 and 5000 people.")
		}
		type row struct {
			Person uuid.UUID `json:"person"`
			Device
		}
		out := []row{}
		err = m.A.DB.WithOrg(r.Context(), p.Org, func(tx pgx.Tx) error {
			rs, err := tx.Query(r.Context(), `select pe.id, d.id, d.account_id, d.label, d.identity_key, d.signing_key, d.created_at
				from hub.people pe join hub.devices d on d.account_id = pe.account_id and d.revoked_at is null
				where pe.id = any($1) and pe.status='active' order by pe.id, d.created_at`, in.People)
			if err != nil {
				return err
			}
			defer rs.Close()
			for rs.Next() {
				var x row
				var id, sg []byte
				if err := rs.Scan(&x.Person, &x.ID, &x.Account, &x.Label, &id, &sg, &x.CreatedAt); err != nil {
					return err
				}
				x.Identity, x.Signing, x.Fingerprint = base64.StdEncoding.EncodeToString(id), base64.StdEncoding.EncodeToString(sg), Fingerprint(id, sg)
				out = append(out, x)
			}
			return rs.Err()
		})
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"devices": out})
		return nil
	}))
	r.Put("/e2ee/backup", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		var in struct {
			Blob string         `json:"blob"`
			KDF  map[string]any `json:"kdf"`
		}
		if err := httpx.Decode(r, &in, 256<<10); err != nil {
			return err
		}
		blob, err := b64(in.Blob)
		if err != nil || len(blob) < 48 || len(blob) > 128<<10 || in.KDF == nil {
			return httpx.Bad("That backup is not valid.")
		}
		return m.A.DB.Tx(r.Context(), func(tx pgx.Tx) error {
			if _, err := tx.Exec(r.Context(), `insert into hub.key_backups (account_id, blob, kdf) values ($1,$2,$3)
				on conflict (account_id) do update set blob=excluded.blob, kdf=excluded.kdf, updated_at=now(), `+integrity.BumpOn("key_backups"), p.Account, blob, in.KDF); err != nil {
				return err
			}
			if err := m.A.Sign.Seal(r.Context(), tx, "key_backups", "account_id=$1", p.Account); err != nil {
				return err
			}
			w.WriteHeader(204)
			return nil
		})
	}))
	r.Get("/e2ee/backup", httpx.Wrap(func(w http.ResponseWriter, r *http.Request) error {
		p, err := tenant.Must(r.Context())
		if err != nil {
			return err
		}
		var blob []byte
		var kdf map[string]any
		err = m.A.DB.Tx(r.Context(), func(tx pgx.Tx) error {
			return tx.QueryRow(r.Context(), `select blob, kdf from hub.key_backups where account_id=$1`, p.Account).Scan(&blob, &kdf)
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.ErrNotFound
		}
		if err != nil {
			return err
		}
		httpx.JSON(w, 200, map[string]any{"blob": base64.StdEncoding.EncodeToString(blob), "kdf": kdf})
		return nil
	}))
}
