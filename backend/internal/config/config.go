// Package config loads service settings from the environment. Secrets are never defaulted in production.
package config

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"hub/internal/blob"
)

type Config struct {
	Env         string // dev | staging | prod
	HTTPAddr    string
	DatabaseURL string // runtime role: DML only, no DDL, subject to row level security
	MigrateURL  string // owner role, used only by cmd/migrate
	RedisAddr   string
	RedisPass   string

	// Key material. In production these come from a KMS/HSM; here they are 32-byte keys supplied as base64.
	KEK       []byte // key-encryption key: wraps per-organization data keys
	MACKey    []byte // row integrity MAC key (never stored in the database)
	BlindKey  []byte // blind-index key for searchable encrypted fields (email, phone)
	TokenKey  []byte // signs access tokens
	LedgerKey []byte // signs audit ledger checkpoints

	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	AllowedOrigins []string

	CallsURLs   []string // base URLs of the calls microservice instances (a room is pinned to one by hashing its id)
	CallsSecret []byte   // shared secret used to mint call-room tokens
	// Media servers (LiveKit SFU). Empty: calls fall back to the small peer-to-peer rooms on the calls service.
	SFUNodes   []SFUNode
	CallLimits CallLimits
	RateRPS    int
	DBGuard    string // proves to the database that a write comes from the Hub system

	BlobDir string // this server's own disk for encrypted file contents (used when no file server is configured or answers)
	// EmailDomains limits who can sign up and sign in to these email domains (HUB_ALLOWED_EMAIL_DOMAINS, comma separated);
	// empty allows any domain.
	EmailDomains []string
	FileServers  []string // S3-compatible file servers, tried in order: s3://KEY:SECRET@host/bucket?region=..
	MaxUpload    int64    // bytes of ciphertext per file
}

// SFUNode is one media server (or a load-balanced region of them): its HTTP base URL and the API key pair shared with it.
type SFUNode struct{ URL, Key, Secret string }

// CallLimits are the admission caps for calls. Zero never means "unlimited": a missing value takes the default.
type CallLimits struct {
	Voice      int // people in one voice room
	Video      int // people in one video room
	OpenMic    int // up to this many people everyone may speak; beyond it newcomers join as listeners until the host brings them on stage
	Stage      int // people who may speak at the same time in a big room
	RoomsOrg   int // rooms running at once in one organization
	RoomsTotal int // rooms running at once on this deployment
}

func Load() (*Config, error) {
	c := &Config{
		Env:         get("HUB_ENV", "dev"),
		HTTPAddr:    get("HUB_HTTP_ADDR", ":8080"),
		DatabaseURL: get("HUB_DATABASE_URL", "postgres://hub_app:hub_app_dev_pw@127.0.0.1:5433/hub?sslmode=disable"),
		MigrateURL:  get("HUB_MIGRATE_URL", "postgres://hub_owner:hub_owner_dev_pw@127.0.0.1:5433/hub?sslmode=disable"),
		RedisAddr:   get("HUB_REDIS_ADDR", "127.0.0.1:6390"),
		RedisPass:   os.Getenv("HUB_REDIS_PASS"),
		AccessTTL:   dur("HUB_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:  dur("HUB_REFRESH_TTL", 30*24*time.Hour),
		RateRPS:     int(num("HUB_RATE_RPS", 50)),
		DBGuard:     get("HUB_DB_GUARD", "dev-db-guard-token-change-me"),
		BlobDir:     get("HUB_BLOB_DIR", "./data/blobs"),
		MaxUpload:   num("HUB_MAX_UPLOAD", 100<<20),
	}
	if c.Env == "prod" && os.Getenv("HUB_DB_GUARD") == "" {
		return nil, fmt.Errorf("HUB_DB_GUARD is required in production")
	}
	c.FileServers = strings.FieldsFunc(os.Getenv("HUB_FILE_SERVERS"), func(r rune) bool { return r == ' ' || r == '\n' })
	for _, fs := range c.FileServers {
		if _, err := blob.ParseS3(fs); err != nil {
			return nil, fmt.Errorf("HUB_FILE_SERVERS: %w", err)
		}
	}
	for _, d := range strings.Split(os.Getenv("HUB_ALLOWED_EMAIL_DOMAINS"), ",") {
		if d = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(d), "@"))); d != "" {
			c.EmailDomains = append(c.EmailDomains, d)
		}
	}
	c.CallsURLs = strings.Split(get("HUB_CALLS_URL", "http://127.0.0.1:8090"), ",")
	// HUB_SFU_NODES="https://media-eu.example.com|APIKEY|SECRET,https://media-af.example.com|APIKEY|SECRET"
	for _, n := range strings.Split(os.Getenv("HUB_SFU_NODES"), ",") {
		if f := strings.Split(strings.TrimSpace(n), "|"); len(f) == 3 {
			c.SFUNodes = append(c.SFUNodes, SFUNode{URL: strings.TrimRight(f[0], "/"), Key: f[1], Secret: f[2]})
		} else if strings.TrimSpace(n) != "" {
			return nil, fmt.Errorf("HUB_SFU_NODES: each node is url|key|secret")
		}
	}
	if len(c.SFUNodes) == 0 && c.Env != "prod" {
		// development: use the local media server from scripts/dev-media.sh when it is running
		if conn, err := net.DialTimeout("tcp", "127.0.0.1:7880", 300*time.Millisecond); err == nil {
			conn.Close()
			c.SFUNodes = []SFUNode{{URL: "http://127.0.0.1:7880", Key: "hubkey", Secret: "hubsecret-hubsecret-hubsecret-hubsecret"}}
		}
	}
	c.CallLimits = CallLimits{
		Voice: int(num("HUB_CALL_MAX_VOICE", 800)), Video: int(num("HUB_CALL_MAX_VIDEO", 800)),
		OpenMic: int(num("HUB_CALL_OPEN_MIC", 25)), Stage: int(num("HUB_CALL_STAGE", 25)),
		RoomsOrg: int(num("HUB_CALL_ROOMS_ORG", 100)), RoomsTotal: int(num("HUB_CALL_ROOMS_TOTAL", 100000)),
	}
	c.AllowedOrigins = strings.Split(get("HUB_ALLOWED_ORIGINS", "http://localhost:3000"), ",")
	var err error
	for name, dst := range map[string]*[]byte{"HUB_KEK": &c.KEK, "HUB_MAC_KEY": &c.MACKey, "HUB_BLIND_KEY": &c.BlindKey, "HUB_TOKEN_KEY": &c.TokenKey, "HUB_LEDGER_KEY": &c.LedgerKey, "HUB_CALLS_SECRET": &c.CallsSecret} {
		if *dst, err = key(name, c.Env); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// key reads a base64 32-byte key. Dev derives a stable throwaway key so a fresh checkout runs; prod refuses to start without one.
func key(name, env string) ([]byte, error) {
	v := os.Getenv(name)
	if v == "" {
		if env == "prod" {
			return nil, fmt.Errorf("%s is required in production", name)
		}
		b := make([]byte, 32)
		copy(b, []byte("dev-only-"+strings.ToLower(name)+"-0123456789abcdef"))
		return b, nil
	}
	b, err := base64.StdEncoding.DecodeString(v)
	if err != nil || len(b) != 32 {
		return nil, fmt.Errorf("%s must be 32 bytes, base64 encoded", name)
	}
	return b, nil
}

func get(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func num(k string, d int64) int64 {
	if v, err := strconv.ParseInt(os.Getenv(k), 10, 64); err == nil {
		return v
	}
	return d
}
func dur(k string, d time.Duration) time.Duration {
	if v, err := time.ParseDuration(os.Getenv(k)); err == nil {
		return v
	}
	return d
}
