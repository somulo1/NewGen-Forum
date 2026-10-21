// Package runapi is the one-process Hub server. Starting it is the whole deployment: it applies any database migrations that are
// missing, installs the reference data (sounds), creates the first owner and organization when the settings for them are given and
// the database is empty, runs the background jobs, and serves the API. Nothing else has to be run first.
//
// Switches (all optional):
//
//	HUB_AUTO_MIGRATE=off   do not touch the schema at start-up (run cmd/migrate yourself)
//	HUB_WITH_WORKER=off    do not run the background jobs here (run cmd/worker elsewhere)
//	HUB_WITH_CALLS=on      also serve the calls signalling service on this port under /calls-signal (set HUB_CALLS_URL to
//	                       wss://<this address>/calls-signal), so small calls need no second service
//	HUB_OWNER_NAME, HUB_OWNER_EMAIL, HUB_OWNER_PASSWORD, HUB_ORG_NAME, HUB_ORG_SLUG   first owner and organization, created once
//	                       when no account exists yet. The password must be changed at first sign-in unless HUB_OWNER_KEEP_PASSWORD=1.
//	PORT                   the port to listen on when HUB_HTTP_ADDR is not set (platforms such as Render set it)
package runapi

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"

	"hub/internal/app"
	"hub/internal/bootstrap"
	"hub/internal/callsvc"
	"hub/internal/config"
	"hub/internal/db"
	"hub/internal/mail"
	"hub/internal/modules"
	"hub/internal/modules/auth"
	"hub/internal/modules/tones"
	"hub/internal/server"
	"hub/migrations"
)

func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	os.Exit(1)
}

func Run() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg, err := config.Load()
	if err != nil {
		fatal("config", err)
	}
	if os.Getenv("HUB_HTTP_ADDR") == "" && os.Getenv("PORT") != "" {
		cfg.HTTPAddr = ":" + os.Getenv("PORT")
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	slog.Info("hub starting", "env", cfg.Env, "addr", cfg.HTTPAddr)
	if os.Getenv("RENDER") != "" && cfg.Env != "prod" {
		slog.Warn("running on Render without HUB_ENV=prod: none of the production settings reached this service. Add them under Environment (and save), or give them as a Secret File named .env")
	}

	// 1. the schema
	if os.Getenv("HUB_AUTO_MIGRATE") != "off" {
		applied, err := db.Migrate(ctx, cfg.MigrateURL, cfg.DBGuard, migrations.FS)
		if err != nil {
			fatal("migrate", err)
		}
		if len(applied) > 0 {
			slog.Info("database migrated", "applied", len(applied), "last", applied[len(applied)-1])
		}
	}

	dev := &mail.Dev{}
	var mailer mail.Mailer = dev // real email when HUB_SMTP_HOST is set; otherwise messages stay in memory (development)
	if sm := mail.FromEnv(); sm != nil {
		mailer = sm
		slog.Info("email through SMTP", "host", sm.Host, "port", sm.Port, "security", sm.Security)
	} else if cfg.Env == "prod" {
		slog.Warn("no HUB_SMTP_HOST set: password reset codes and notifications will not be delivered")
	}
	a, err := app.New(ctx, cfg, mailer)
	if err != nil {
		fatal("startup", err)
	}
	defer a.Close()
	go a.Bus.Run(ctx, a.Hub.Deliver)

	// 2. reference data every deployment needs
	if err := tones.New(a).Install(ctx); err != nil {
		fatal("install tones", err)
	}

	// 3. the first owner and organization, once, when asked for and the database has no accounts
	seedOwner(ctx, a, cfg)

	// 4. background jobs
	if os.Getenv("HUB_WITH_WORKER") != "off" {
		go RunWorker(ctx, a)
	}

	au := auth.New(a)
	var h http.Handler = server.Router(server.Deps{App: a, Auth: au, Dev: dev, Modules: modules.All(a, au)})
	if os.Getenv("HUB_WITH_CALLS") == "on" {
		rdb := redis.NewClient(cfg.RedisClientOptions())
		calls := http.StripPrefix("/calls-signal", callsvc.New(cfg.CallsSecret, rdb, cfg.AllowedOrigins).Handler())
		api := h
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/calls-signal/") {
				calls.ServeHTTP(w, r)
				return
			}
			api.ServeHTTP(w, r)
		})
	}
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 120 * time.Second, MaxHeaderBytes: 32 << 10}
	go func() {
		<-ctx.Done()
		sc, c := context.WithTimeout(context.Background(), 20*time.Second)
		defer c()
		_ = srv.Shutdown(sc)
	}()
	slog.Info("hub api listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fatal("server", err)
	}
}

func seedOwner(ctx context.Context, a *app.App, cfg *config.Config) {
	name, email, pw := os.Getenv("HUB_OWNER_NAME"), os.Getenv("HUB_OWNER_EMAIL"), os.Getenv("HUB_OWNER_PASSWORD")
	org, slug := os.Getenv("HUB_ORG_NAME"), os.Getenv("HUB_ORG_SLUG")
	if name == "" || email == "" || pw == "" || org == "" || slug == "" {
		return
	}
	var accounts int
	if err := a.DB.Tx(ctx, func(tx pgx.Tx) error { return tx.QueryRow(ctx, `select count(*) from hub.accounts`).Scan(&accounts) }); err != nil {
		slog.Error("first owner: could not check for accounts", "err", err)
		return
	}
	if accounts > 0 {
		return
	}
	if err := bootstrap.Owner(ctx, a, cfg, name, email, pw, org, slug, false, os.Getenv("HUB_OWNER_KEEP_PASSWORD") == "1"); err != nil {
		slog.Error("first owner", "err", err)
		return
	}
	slog.Info("first owner and organization created", "organization", slug)
}
