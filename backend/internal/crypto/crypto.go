// Package crypto holds the server-side cryptography.
//
// Two kinds of data exist in Hub and they are protected differently:
//
//  1. End-to-end encrypted content (chats, forums, wall, notes, file contents, ticket threads, call media).
//     The server never holds a key for it. Clients encrypt on their device, and the server stores and forwards
//     opaque envelopes plus the per-recipient wrapped keys. Nothing in this package can open them.
//
//  2. Server-readable business fields (people's phone/date of birth/email, billing, organization profile).
//     The server must read these to do its job, so they are encrypted at rest with per-organization data keys
//     (envelope encryption: a KMS/HSM key wraps each organization's DEK). Every ciphertext is bound to the exact
//     organization, table, column and row it belongs to, so it cannot be copied to another row and still decrypt.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	versionByte = 1
	nonceSize   = 12
)

var ErrDecrypt = errors.New("crypto: cannot decrypt (wrong key, wrong context or tampered data)")

// Seal encrypts with AES-256-GCM. Output: version(1) | nonce(12) | ciphertext+tag. aad binds the context.
func Seal(key, plaintext, aad []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("crypto: key must be 32 bytes")
	}
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	g, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 1+nonceSize, 1+nonceSize+len(plaintext)+g.Overhead())
	out[0] = versionByte
	if _, err := rand.Read(out[1 : 1+nonceSize]); err != nil {
		return nil, err
	}
	return g.Seal(out, out[1:1+nonceSize], plaintext, aad), nil
}

// Open reverses Seal.
func Open(key, sealed, aad []byte) ([]byte, error) {
	if len(key) != 32 || len(sealed) < 1+nonceSize+16 || sealed[0] != versionByte {
		return nil, ErrDecrypt
	}
	blk, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	g, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}
	pt, err := g.Open(nil, sealed[1:1+nonceSize], sealed[1+nonceSize:], aad)
	if err != nil {
		return nil, ErrDecrypt
	}
	return pt, nil
}

// Context describes exactly where a ciphertext lives. It is authenticated, not secret.
type Context struct {
	Org    string
	Table  string
	Column string
	RowID  string
}

func (c Context) aad() []byte {
	return []byte(fmt.Sprintf("hub/v1|%s|%s|%s|%s", c.Org, c.Table, c.Column, c.RowID))
}

// OrgKeys wraps and unwraps per-organization data keys with the KEK, and caches unwrapped keys briefly.
type OrgKeys struct {
	kek   []byte
	mu    sync.Mutex
	cache map[string]cached
	ttl   time.Duration
}

type cached struct {
	key []byte
	at  time.Time
}

func NewOrgKeys(kek []byte) *OrgKeys {
	return &OrgKeys{kek: kek, cache: map[string]cached{}, ttl: 5 * time.Minute}
}

// NewDEK makes a fresh data key and returns it with its KEK-wrapped form (store the wrapped form only).
func (o *OrgKeys) NewDEK(org string, version int) (dek, wrapped []byte, err error) {
	dek = make([]byte, 32)
	if _, err = rand.Read(dek); err != nil {
		return nil, nil, err
	}
	wrapped, err = Seal(o.kek, dek, dekAAD(org, version))
	return dek, wrapped, err
}

func dekAAD(org string, version int) []byte {
	return []byte(fmt.Sprintf("hub/dek|%s|%d", org, version))
}

// DEK returns the unwrapped data key for an organization version.
func (o *OrgKeys) DEK(org string, version int, wrapped []byte) ([]byte, error) {
	k := fmt.Sprintf("%s/%d", org, version)
	o.mu.Lock()
	defer o.mu.Unlock()
	if c, ok := o.cache[k]; ok && time.Since(c.at) < o.ttl {
		return c.key, nil
	}
	dek, err := Open(o.kek, wrapped, dekAAD(org, version))
	if err != nil {
		return nil, err
	}
	o.cache[k] = cached{dek, time.Now()}
	return dek, nil
}

// Forget drops cached keys (on rotation or logout of a service).
func (o *OrgKeys) Forget() { o.mu.Lock(); o.cache = map[string]cached{}; o.mu.Unlock() }

// Field encrypts one value for one exact place.
func Field(dek []byte, c Context, plaintext []byte) ([]byte, error) {
	return Seal(dek, plaintext, c.aad())
}

// FieldOpen decrypts it; it fails if the ciphertext was moved to another row, column, table or organization.
func FieldOpen(dek []byte, c Context, sealed []byte) ([]byte, error) {
	return Open(dek, sealed, c.aad())
}

// Blind returns a keyed hash so encrypted values (like email) can still be looked up by equality without storing them in clear.
func Blind(key []byte, scope, value string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(scope))
	m.Write([]byte{0})
	m.Write([]byte(value))
	return m.Sum(nil)[:20]
}

// MAC is HMAC-SHA256 used for row integrity and token signing.
func MAC(key []byte, parts ...[]byte) []byte {
	m := hmac.New(sha256.New, key)
	for _, p := range parts {
		var l [8]byte
		binary.BigEndian.PutUint64(l[:], uint64(len(p)))
		m.Write(l[:]) // length-prefix: ("ab","c") can never equal ("a","bc")
		m.Write(p)
	}
	return m.Sum(nil)
}

func Equal(a, b []byte) bool { return subtle.ConstantTimeCompare(a, b) == 1 }

// RandomBytes / RandomToken give unguessable values for sessions, links and codes.
func RandomBytes(n int) []byte { b := make([]byte, n); _, _ = rand.Read(b); return b }
func RandomToken(n int) string { return base64.RawURLEncoding.EncodeToString(RandomBytes(n)) }

// Digits returns an n-digit numeric one-time code, uniformly distributed.
func Digits(n int) string {
	out := make([]byte, n)
	for i := range out {
		var b [1]byte
		for {
			_, _ = rand.Read(b[:])
			if b[0] < 250 { // reject to avoid modulo bias
				out[i] = '0' + b[0]%10
				break
			}
		}
	}
	return string(out)
}

// Hash is SHA-256.
func Hash(parts ...[]byte) []byte {
	h := sha256.New()
	for _, p := range parts {
		h.Write(p)
	}
	return h.Sum(nil)
}

// --- passwords: Argon2id with a per-password salt, in the standard PHC string format ---

type PasswordParams struct {
	Time, MemoryKiB uint32
	Threads         uint8
}

var DefaultPassword = PasswordParams{Time: 3, MemoryKiB: 64 * 1024, Threads: 2}

func HashPassword(pw string, p PasswordParams) string {
	salt := RandomBytes(16)
	sum := argon2.IDKey([]byte(pw), salt, p.Time, p.MemoryKiB, p.Threads, 32)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", p.MemoryKiB, p.Time, p.Threads, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(sum))
}

func VerifyPassword(pw, encoded string) bool {
	var m, t uint32
	var p uint8
	var saltB, sumB string
	var v int
	if _, err := fmt.Sscanf(encoded, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s", &v, &m, &t, &p, &saltB); err != nil {
		return false
	}
	// Sscanf %s consumed "salt$sum"; split it.
	for i := 0; i < len(saltB); i++ {
		if saltB[i] == '$' {
			sumB = saltB[i+1:]
			saltB = saltB[:i]
			break
		}
	}
	salt, e1 := base64.RawStdEncoding.DecodeString(saltB)
	want, e2 := base64.RawStdEncoding.DecodeString(sumB)
	if e1 != nil || e2 != nil || len(want) == 0 {
		return false
	}
	got := argon2.IDKey([]byte(pw), salt, t, m, p, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}
