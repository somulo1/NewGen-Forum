package push

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/app"
	"hub/internal/crypto"
	"hub/internal/httpx"
	"hub/internal/tenant"
)

const maxDevicesPerAccount = 10

type device struct {
	hash     []byte
	platform string
	token    string
}

func hashOf(token string) []byte { h := sha256.Sum256([]byte(token)); return h[:] }

func sealCtx(hash []byte) crypto.Context {
	return crypto.Context{Org: "platform", Table: "push_devices", Column: "token_enc", RowID: hex.EncodeToString(hash)}
}

// Module serves the endpoints the phone app uses to turn its pushes on and off.
type Module struct{ A *app.App }

func NewModule(a *app.App) *Module { return &Module{A: a} }

func (m *Module) Public(chi.Router) {}

func (m *Module) Private(r chi.Router) {
	r.Post("/push/devices", httpx.Wrap(m.register))
	r.Delete("/push/devices", httpx.Wrap(m.unregister))
}

type tokenIn struct {
	Token    string `json:"token"`
	Platform string `json:"platform"`
}

func readToken(r *http.Request) (tokenIn, error) {
	var in tokenIn
	if err := httpx.Decode(r, &in, 8<<10); err != nil {
		return in, err
	}
	in.Token, in.Platform = strings.TrimSpace(in.Token), strings.ToLower(strings.TrimSpace(in.Platform))
	if len(in.Token) < 16 || len(in.Token) > 4096 || strings.ContainsAny(in.Token, " \n\r\t") {
		return in, httpx.Bad("That is not a push address.")
	}
	if in.Platform != "android" && in.Platform != "ios" {
		return in, httpx.Bad("Push notifications are for Android and iOS.")
	}
	return in, nil
}

// register binds the phone's push address to this sign-in. Only sign-ins made from the phone app (bound to a key in the phone)
// may receive pushes.
func (m *Module) register(w http.ResponseWriter, r *http.Request) error {
	p, err := tenant.MustOrg(r.Context())
	if err != nil {
		return err
	}
	in, err := readToken(r)
	if err != nil {
		return err
	}
	hash := hashOf(in.Token)
	enc, err := crypto.Field(m.A.PlatformKey(), sealCtx(hash), []byte(in.Token))
	if err != nil {
		return err
	}
	return m.A.DB.Tx(r.Context(), func(tx pgx.Tx) error {
		var family uuid.UUID
		var bound bool
		if err := tx.QueryRow(r.Context(), `select s.family_id, exists (select 1 from hub.session_devices d where d.family_id = s.family_id)
			from hub.sessions s where s.id=$1 and s.account_id=$2 and s.revoked_at is null`, p.Session, p.Account).Scan(&family, &bound); err != nil {
			return httpx.Err(401, "signed_out", "Sign in to continue.")
		}
		if !bound {
			return httpx.Err(403, "not_app", "Push notifications are for the Hub phone app.")
		}
		if _, err := tx.Exec(r.Context(), `insert into hub.push_devices (token_hash, account_id, family_id, platform, token_enc) values ($1,$2,$3,$4,$5)
			on conflict (token_hash) do update set account_id=excluded.account_id, family_id=excluded.family_id, platform=excluded.platform,
			token_enc=excluded.token_enc, seen_at=now()`, hash, p.Account, family, in.Platform, enc); err != nil {
			return err
		}
		// addresses of sign-ins that have ended, and any beyond the newest few, are dropped
		if _, err := tx.Exec(r.Context(), `delete from hub.push_devices pd where pd.account_id=$1 and (
			not exists (select 1 from hub.sessions s where s.family_id = pd.family_id and s.revoked_at is null and s.expires_at > now())
			or pd.token_hash not in (select token_hash from hub.push_devices where account_id=$1 order by seen_at desc limit $2))`, p.Account, maxDevicesPerAccount); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	})
}

func (m *Module) unregister(w http.ResponseWriter, r *http.Request) error {
	p, err := tenant.Must(r.Context())
	if err != nil {
		return err
	}
	in, err := readToken(r)
	if err != nil {
		return err
	}
	return m.A.DB.Tx(r.Context(), func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(), `delete from hub.push_devices where token_hash=$1 and account_id=$2`, hashOf(in.Token), p.Account); err != nil {
			return err
		}
		w.WriteHeader(204)
		return nil
	})
}

// devicesOf lists the live push addresses of these accounts, for sign-ins currently in this organization.
func devicesOf(ctx context.Context, a *app.App, org uuid.UUID, accounts []uuid.UUID) ([]device, error) {
	var out []device
	err := a.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select pd.token_hash, pd.platform, pd.token_enc from hub.push_devices pd
			where pd.account_id = any($1) and exists (select 1 from hub.sessions s where s.family_id = pd.family_id
			and s.revoked_at is null and s.expires_at > now() and s.org_id = $2)`, accounts, org)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var d device
			var enc []byte
			if err := rs.Scan(&d.hash, &d.platform, &enc); err != nil {
				return err
			}
			t, err := crypto.FieldOpen(a.PlatformKey(), sealCtx(d.hash), enc)
			if err != nil {
				continue
			}
			d.token = string(t)
			out = append(out, d)
		}
		return rs.Err()
	})
	return out, err
}

// forget drops an address the platform says no longer exists (the app was removed, or the phone reset).
func forget(ctx context.Context, a *app.App, hash []byte) {
	_ = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `delete from hub.push_devices where token_hash=$1`, hash)
		return err
	})
}
