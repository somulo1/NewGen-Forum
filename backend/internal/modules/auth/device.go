package auth

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"hub/internal/httpx"
)

// Phone binding. The Hub phone app makes an Ed25519 key pair the first time it runs and keeps the private half in the phone's
// own secure storage (Keychain on iOS, the Android Keystore), marked "this device only": it is never sent anywhere, never backed
// up and never reaches the web pages the app shows. Signing in from the app sends the public half with a signature, and the
// whole sign-in (every refresh token that follows from it) is tied to that key. A refresh token from a bound sign-in is only
// accepted with a fresh signature from the same key, so a copied token is useless on any other phone or computer: presenting
// one without the signature ends the sign-in at once, as a reused refresh token does.

// DeviceProof is what the app sends: its public key, the time it signed at (Unix seconds) and the signature.
type DeviceProof struct {
	Key      string `json:"key"`
	TS       int64  `json:"ts"`
	Sig      string `json:"sig"`
	Platform string `json:"platform"`
}

var (
	errDeviceProof = httpx.Err(400, "device_proof", "This phone could not prove its key. Close the app, open it again and sign in.")
	errDeviceRetry = httpx.Err(400, "device_retry", "This phone could not prove its key just now. Hub tries again by itself.")
)

// a signature older (or further in the future) than this is refused; a day, so a phone whose clock is wrong still works
const proofWindow = 24 * time.Hour

func loginProofMessage(email string) string    { return "hub-device-login|" + email }
func refreshProofMessage(sid uuid.UUID) string { return "hub-device-refresh|" + sid.String() }

// verifyProof checks a signature over msg|ts made by pub within the allowed window.
func verifyProof(pub []byte, msg string, ts int64, sig64 string) bool {
	if len(pub) != ed25519.PublicKeySize || ts == 0 {
		return false
	}
	if d := time.Since(time.Unix(ts, 0)); d > proofWindow || d < -proofWindow {
		return false
	}
	sig, err := base64.StdEncoding.DecodeString(sig64)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	return ed25519.Verify(ed25519.PublicKey(pub), []byte(msg+"|"+strconv.FormatInt(ts, 10)), sig)
}

// loginKey checks the proof sent with a sign-in and returns the public key to bind (nil: a browser, nothing to bind).
func loginKey(p *DeviceProof, email string) ([]byte, error) {
	if p == nil || (p.Key == "" && p.Sig == "") {
		return nil, nil
	}
	pub, err := base64.StdEncoding.DecodeString(p.Key)
	if err != nil || !verifyProof(pub, loginProofMessage(email), p.TS, p.Sig) {
		return nil, errDeviceProof
	}
	return pub, nil
}

// bindFamily ties a new sign-in (its refresh-token family) to the phone's key.
func bindFamily(ctx context.Context, tx pgx.Tx, family, account uuid.UUID, pub []byte, platform string) error {
	if len(platform) > 40 {
		platform = platform[:40]
	}
	_, err := tx.Exec(ctx, `insert into hub.session_devices (family_id, account_id, public_key, platform) values ($1,$2,$3,$4)`, family, account, pub, strings.TrimSpace(platform))
	return err
}

// refreshProof is the signature that came with a refresh (headers X-Hub-Device-TS and X-Hub-Device-Sig).
type refreshProof struct {
	TS  int64
	Sig string
}

type proofKey struct{}

// WithRefreshProof reads the device signature headers of a refresh request into the context.
func WithRefreshProof(ctx context.Context, ts, sig string) context.Context {
	n, _ := strconv.ParseInt(strings.TrimSpace(ts), 10, 64)
	return context.WithValue(ctx, proofKey{}, refreshProof{TS: n, Sig: strings.TrimSpace(sig)})
}

// deviceOK reports whether a refresh of session sid (in family) may go ahead, and whether the sign-in is bound to a phone: true
// for a browser sign-in, or when the signature that came with the request was made by the bound key.
func deviceOK(ctx context.Context, tx pgx.Tx, family, sid uuid.UUID) (ok, bound bool, err error) {
	var pub []byte
	err = tx.QueryRow(ctx, `select public_key from hub.session_devices where family_id=$1`, family).Scan(&pub)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, false, nil
	}
	if err != nil {
		return false, false, err
	}
	p, _ := ctx.Value(proofKey{}).(refreshProof)
	return verifyProof(pub, refreshProofMessage(sid), p.TS, p.Sig), true, nil
}
