// Package app wires the shared services every module uses.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"hub/internal/config"
	"hub/internal/crypto"
	"hub/internal/db"
	"hub/internal/integrity"
	"hub/internal/mail"
	"hub/internal/realtime"
)

type App struct {
	Cfg    *config.Config
	DB     *db.DB
	Keys   *crypto.OrgKeys
	Sign   *integrity.Signer
	Ledger *integrity.Ledger
	Rdb    *redis.Client
	Bus    *realtime.Bus
	Hub    *realtime.Hub
	Limit  *realtime.Limiter
	Mail   mail.Mailer
	Hooks  []Hook
	// Geo maps an IP address to "City, CC". Nil when no geo-IP database is configured (rules that need it then stay quiet).
	Geo func(ip string) string
	// Fast, cheap password hashing in tests; production keeps crypto.DefaultPassword.
	PW crypto.PasswordParams

	mem *miniredis.Miniredis // the in-memory stand-in for Redis when HUB_REDIS=off
}

func New(ctx context.Context, cfg *config.Config, m mail.Mailer) (*App, error) {
	d, err := db.Open(ctx, cfg.DatabaseURL, cfg.DBGuard, 40)
	if err != nil {
		return nil, err
	}
	rdb, mem, err := openRedis(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &App{Cfg: cfg, DB: d, Keys: crypto.NewOrgKeys(cfg.KEK), Sign: integrity.NewSigner(cfg.MACKey), Ledger: integrity.NewLedger(cfg.LedgerKey),
		Rdb: rdb, Bus: realtime.NewBus(rdb), Hub: realtime.NewHub(), Limit: realtime.NewLimiter(rdb), Mail: m, PW: crypto.DefaultPassword, mem: mem}, nil
}

func (a *App) Close() {
	a.DB.Close()
	_ = a.Rdb.Close()
	if a.mem != nil {
		a.mem.Close()
	}
}

// Now is the clock used for stored times: UTC, microsecond precision, so MACs computed before and after a round trip agree.
func Now() time.Time { return time.Now().UTC().Truncate(time.Microsecond) }

// NewID returns a time-ordered UUID (v7): good for index locality at very large row counts.
func NewID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

// PlatformKey encrypts account-level fields (they belong to no single organization).
func (a *App) PlatformKey() []byte { return crypto.MAC(a.Cfg.KEK, []byte("hub/platform-key/v1")) }

// OrgDEK returns the newest data key for an organization, creating the first one when the organization has none.
func (a *App) OrgDEK(ctx context.Context, tx pgx.Tx, org uuid.UUID) ([]byte, int, error) {
	var ver int
	var wrapped []byte
	err := tx.QueryRow(ctx, "select version, wrapped_dek from hub.org_keys where org_id=$1 order by version desc limit 1", org).Scan(&ver, &wrapped)
	if errors.Is(err, pgx.ErrNoRows) {
		dek, w, err := a.Keys.NewDEK(org.String(), 1)
		if err != nil {
			return nil, 0, err
		}
		if _, err = tx.Exec(ctx, "insert into hub.org_keys (org_id, version, wrapped_dek) values ($1,1,$2)", org, w); err != nil {
			return nil, 0, err
		}
		return dek, 1, nil
	}
	if err != nil {
		return nil, 0, err
	}
	dek, err := a.Keys.DEK(org.String(), ver, wrapped)
	return dek, ver, err
}

// DEKVersion loads a specific historical data key version.
func (a *App) DEKVersion(ctx context.Context, tx pgx.Tx, org uuid.UUID, ver int) ([]byte, error) {
	var wrapped []byte
	if err := tx.QueryRow(ctx, "select wrapped_dek from hub.org_keys where org_id=$1 and version=$2", org, ver).Scan(&wrapped); err != nil {
		return nil, err
	}
	return a.Keys.DEK(org.String(), ver, wrapped)
}

// Entry is one line of the audit log.
type Entry struct {
	Org      uuid.UUID
	Actor    *uuid.UUID
	Name     string
	Action   string
	Module   string
	IP       string
	Location string
	Result   string // Success | Failed | Blocked
	Payload  any
	Rule     string // set by a security rule hook when this entry trips one
	// At backdates the entry. Only the seeder sets it, to load history; live code leaves it nil (the database clock is used).
	At  *time.Time
	Seq int64
}

// Hook sees every audit entry in the transaction that writes it (security rules use this to raise alerts).
type Hook func(ctx context.Context, a *App, tx pgx.Tx, e *Entry) error

// Audit records an action twice: in the organization's tamper-evident ledger and in the audit log shown on the Security page.
// It runs inside the caller's transaction, so the change and its record commit together or not at all.
func (a *App) Audit(ctx context.Context, tx pgx.Tx, org uuid.UUID, actor *uuid.UUID, actorName, action, module, ip string, payload any) error {
	return a.AuditEntry(ctx, tx, &Entry{Org: org, Actor: actor, Name: actorName, Action: action, Module: module, IP: ip, Result: "Success", Payload: payload})
}

func (a *App) AuditEntry(ctx context.Context, tx pgx.Tx, e *Entry) error {
	raw, _ := json.Marshal(e.Payload)
	seq, err := a.Ledger.Append(ctx, tx, e.Org, e.Actor, e.Action, e.Module, raw)
	if err != nil {
		return err
	}
	e.Seq = seq
	for _, h := range a.Hooks {
		if err := h(ctx, a, tx, e); err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `insert into hub.audit_logs (org_id, ts, actor_name, action, module, ip, location, result, rule, ledger_seq) values ($1,coalesce($2, now()),$3,$4,$5,$6,$7,$8,$9,$10)`,
		e.Org, e.At, e.Name, e.Action, e.Module, e.IP, e.Location, e.Result, e.Rule, seq)
	return err
}

// Publish sends a realtime event. A failed publish is logged by callers; the database already holds the truth.
func (a *App) Publish(ctx context.Context, e realtime.Event) { _ = a.Bus.Publish(ctx, e) }

// AuditAccount writes a sign-in style event (failed password, lockout, successful sign-in) into the audit log of every
// organization the account belongs to, so each organization's security team sees what happens to its people's accounts.
// It is best effort: a logging problem must never block a sign-in.
func (a *App) AuditAccount(ctx context.Context, account uuid.UUID, name, action, result, ip string) {
	var orgs []uuid.UUID
	_ = a.DB.Tx(ctx, func(tx pgx.Tx) error {
		rs, err := tx.Query(ctx, `select org_id from hub.account_orgs where account_id=$1`, account)
		if err != nil {
			return err
		}
		defer rs.Close()
		for rs.Next() {
			var o uuid.UUID
			if err := rs.Scan(&o); err != nil {
				return err
			}
			orgs = append(orgs, o)
		}
		return rs.Err()
	})
	for _, o := range orgs {
		_ = a.DB.WithOrg(ctx, o, func(tx pgx.Tx) error {
			var pid *uuid.UUID
			_ = tx.QueryRow(ctx, `select id from hub.people where account_id=$1`, account).Scan(&pid)
			loc := ""
			if a.Geo != nil {
				loc = a.Geo(ip)
			}
			return a.AuditEntry(ctx, tx, &Entry{Org: o, Actor: pid, Name: name, Action: action, Module: "Auth", IP: ip, Location: loc, Result: result})
		})
	}
}

// openRedis connects to Redis. With HUB_REDIS=off no Redis server is needed: a small one runs inside this process (rate limits, session
// markers, realtime notices and call passes all work, but only for this one instance, and they start empty after a restart). Turn
// it back on by removing HUB_REDIS=off and setting HUB_REDIS_ADDR.
func openRedis(ctx context.Context, cfg *config.Config) (*redis.Client, *miniredis.Miniredis, error) {
	if cfg.RedisOff {
		mr, err := miniredis.Run()
		if err != nil {
			return nil, nil, fmt.Errorf("in-memory redis: %w", err)
		}
		go func() { // the in-memory server keeps its own clock: move it forward so expiries happen
			t := time.NewTicker(time.Second)
			defer t.Stop()
			for range t.C {
				mr.FastForward(time.Second)
			}
		}()
		slog.Warn("HUB_REDIS=off: using an in-memory Redis inside this process; run one instance only, and state is lost on restart")
		return redis.NewClient(&redis.Options{Addr: mr.Addr()}), mr, nil
	}
	// A hosted Redis allows few connections at once (Redis Cloud's free plan: 30, shared by the old and new version while a deploy
	// runs), so the pool is kept small and idle ones are let go.
	pool := 8
	if n, err := strconv.Atoi(os.Getenv("HUB_REDIS_POOL")); err == nil && n > 0 {
		pool = n
	}
	ro := cfg.RedisClientOptions(pool)
	ro.ConnMaxIdleTime = 2 * time.Minute
	ro.PoolTimeout = 10 * time.Second
	rdb := redis.NewClient(ro)
	// While a new version starts the old one still holds its connections, and a full Redis closes new ones ("EOF"): wait and try
	// again instead of giving up on the first refusal.
	var perr error
	for i := 0; i < 12; i++ {
		if perr = rdb.Ping(ctx).Err(); perr == nil {
			break
		}
		slog.Warn("redis not ready; trying again", "attempt", i+1, "err", perr)
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
	if perr != nil {
		return nil, nil, fmt.Errorf("redis: %w", perr)
	}
	return rdb, nil, nil
}
