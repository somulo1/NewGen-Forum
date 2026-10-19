// Package e2ee is the reference implementation of Hub's end-to-end encryption format.
//
// The SERVER never imports the encrypt/decrypt side of this package: it only calls ValidateEnvelope and ValidateWrapped to
// reject malformed blobs. Clients (the web app, mobile apps, the seeder acting as a demo client, tests) use the rest.
//
//	device keys      X25519 (key agreement) + Ed25519 (signatures), generated on the device; private halves never leave it
//	conversation key a random 32-byte key per conversation epoch, chosen by a member's client
//	key wrapping     per recipient: ephemeral X25519 -> ECDH with the recipient's public key -> HKDF-SHA256 -> AES-256-GCM
//	message          AES-256-GCM under the conversation key, bound (AAD) to org, conversation, epoch, sender and message id,
//	                 then signed by the sender's Ed25519 key so members can tell who really wrote it
//
// Membership changes bump the epoch: a removed member never receives the new key, so they cannot read anything new.
package e2ee

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

const (
	Version     = 1
	nonceLen    = 12
	sigLen      = ed25519.SignatureSize
	tagLen      = 16
	wrapMinLen  = 1 + 32 + nonceLen + 32 + tagLen
	MaxEnvelope = 256 << 10
)

var (
	ErrFormat = errors.New("e2ee: malformed envelope")
	ErrOpen   = errors.New("e2ee: cannot open (wrong key, tampered, or not addressed to this context)")
	ErrSig    = errors.New("e2ee: signature does not match the claimed sender")
)

// Device holds one device's private keys.
type Device struct {
	Agree *ecdh.PrivateKey
	Sign  ed25519.PrivateKey
}

// PublicKeys is what a device publishes to the key directory.
type PublicKeys struct {
	Identity []byte // X25519
	Signing  []byte // Ed25519
}

func NewDevice() (*Device, error) {
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	_, s, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &Device{Agree: k, Sign: s}, nil
}

// DeviceFromSeed derives a device deterministically. Used only for demo users in the seeder and for tests.
func DeviceFromSeed(seed []byte) (*Device, error) {
	a, err := hkdf.Key(sha256.New, seed, nil, "hub/e2ee/x25519", 32)
	if err != nil {
		return nil, err
	}
	b, err := hkdf.Key(sha256.New, seed, nil, "hub/e2ee/ed25519", 32)
	if err != nil {
		return nil, err
	}
	k, err := ecdh.X25519().NewPrivateKey(a)
	if err != nil {
		return nil, err
	}
	return &Device{Agree: k, Sign: ed25519.NewKeyFromSeed(b)}, nil
}

func (d *Device) Public() PublicKeys {
	return PublicKeys{Identity: d.Agree.PublicKey().Bytes(), Signing: d.Sign.Public().(ed25519.PublicKey)}
}

// NewConversationKey makes a fresh random conversation key.
func NewConversationKey() ([]byte, error) {
	k := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, k)
	return k, err
}

func wrapKey(shared, eph, rcpt []byte) ([]byte, error) {
	return hkdf.Key(sha256.New, shared, append(append([]byte{}, eph...), rcpt...), "hub/e2ee/wrap/v1", 32)
}

// Wrap encrypts a conversation key for one recipient device. Layout: version | ephemeral public key (32) | nonce | ciphertext+tag.
func Wrap(convKey []byte, recipientIdentity []byte, context []byte) ([]byte, error) {
	pub, err := ecdh.X25519().NewPublicKey(recipientIdentity)
	if err != nil {
		return nil, err
	}
	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	shared, err := eph.ECDH(pub)
	if err != nil {
		return nil, err
	}
	k, err := wrapKey(shared, eph.PublicKey().Bytes(), recipientIdentity)
	if err != nil {
		return nil, err
	}
	out := append([]byte{Version}, eph.PublicKey().Bytes()...)
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	out = append(out, nonce...)
	g, err := gcm(k)
	if err != nil {
		return nil, err
	}
	return g.Seal(out, nonce, convKey, context), nil
}

// Unwrap recovers a conversation key with the device's private key.
func (d *Device) Unwrap(wrapped, context []byte) ([]byte, error) {
	if len(wrapped) < wrapMinLen || wrapped[0] != Version {
		return nil, ErrFormat
	}
	eph := wrapped[1:33]
	pub, err := ecdh.X25519().NewPublicKey(eph)
	if err != nil {
		return nil, ErrFormat
	}
	shared, err := d.Agree.ECDH(pub)
	if err != nil {
		return nil, ErrOpen
	}
	k, err := wrapKey(shared, eph, d.Agree.PublicKey().Bytes())
	if err != nil {
		return nil, err
	}
	g, err := gcm(k)
	if err != nil {
		return nil, err
	}
	pt, err := g.Open(nil, wrapped[33:33+nonceLen], wrapped[33+nonceLen:], context)
	if err != nil {
		return nil, ErrOpen
	}
	return pt, nil
}

func gcm(k []byte) (cipher.AEAD, error) {
	b, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(b)
}

// Context binds an encrypted message to where it belongs.
type Context struct {
	Org, Conversation, Sender, MessageID string
	Epoch                                int
}

func (c Context) aad() []byte {
	return []byte("hub/msg/v1|" + c.Org + "|" + c.Conversation + "|" + c.Sender + "|" + c.MessageID + "|" + itoa(c.Epoch))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// Seal encrypts and signs a message. Layout: version | nonce | ciphertext+tag | signature(64).
func (d *Device) Seal(convKey, plaintext []byte, c Context) ([]byte, error) {
	g, err := gcm(convKey)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	env := append([]byte{Version}, nonce...)
	env = g.Seal(env, nonce, plaintext, c.aad())
	sig := ed25519.Sign(d.Sign, signedBytes(env, c))
	return append(env, sig...), nil
}

func signedBytes(env []byte, c Context) []byte {
	h := sha256.Sum256(env)
	return append(c.aad(), h[:]...)
}

// Open decrypts a message and verifies it was signed by senderSigning.
func Open(convKey, envelope, senderSigning []byte, c Context) ([]byte, error) {
	if err := ValidateEnvelope(envelope); err != nil {
		return nil, err
	}
	body, sig := envelope[:len(envelope)-sigLen], envelope[len(envelope)-sigLen:]
	if !ed25519.Verify(senderSigning, signedBytes(body, c), sig) {
		return nil, ErrSig
	}
	g, err := gcm(convKey)
	if err != nil {
		return nil, err
	}
	pt, err := g.Open(nil, body[1:1+nonceLen], body[1+nonceLen:], c.aad())
	if err != nil {
		return nil, ErrOpen
	}
	return pt, nil
}

// ValidateEnvelope is what the server checks: it can verify the shape and size, nothing more.
func ValidateEnvelope(b []byte) error {
	if len(b) < 1+nonceLen+tagLen+sigLen || len(b) > MaxEnvelope || b[0] != Version {
		return ErrFormat
	}
	return nil
}

// ValidateWrapped checks the shape of a wrapped conversation key.
func ValidateWrapped(b []byte) error {
	if len(b) < wrapMinLen || len(b) > 512 || b[0] != Version {
		return ErrFormat
	}
	return nil
}

// ValidatePublic checks a published key set.
func ValidatePublic(k PublicKeys) error {
	if len(k.Identity) != 32 || len(k.Signing) != ed25519.PublicKeySize {
		return ErrFormat
	}
	if _, err := ecdh.X25519().NewPublicKey(k.Identity); err != nil {
		return ErrFormat
	}
	return nil
}
