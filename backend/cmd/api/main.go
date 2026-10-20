// Command api is the Hub HTTP + realtime API node. Run as many as you need behind a load balancer.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/mail"
	"hub/internal/modules"
	"hub/internal/modules/auth"
	"hub/internal/modules/tones"
	"hub/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	dev := &mail.Dev{}
	// real email when HUB_SMTP_HOST is set; otherwise messages are kept in memory (development)
	var mailer mail.Mailer = dev
	if sm := mail.FromEnv(); sm != nil {
		mailer = sm
		slog.Info("email through SMTP", "host", sm.Host, "port", sm.Port, "security", sm.Security)
	}
	a, err := app.New(ctx, cfg, mailer)
	if err != nil {
		slog.Error("startup", "err", err)
		os.Exit(1)
	}
	defer a.Close()
	go a.Bus.Run(ctx, a.Hub.Deliver)
	// the app's sounds are reference data every deployment needs, so they are installed here, not by the demo seeder
	if err := tones.New(a).Install(ctx); err != nil {
		slog.Error("install tones", "err", err)
		os.Exit(1)
	}

	au := auth.New(a)
	h := server.Router(server.Deps{App: a, Auth: au, Dev: dev, Modules: modules.All(a, au)})
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: h, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 60 * time.Second, IdleTimeout: 120 * time.Second, MaxHeaderBytes: 32 << 10}
	go func() {
		<-ctx.Done()
		sc, c := context.WithTimeout(context.Background(), 20*time.Second)
		defer c()
		_ = srv.Shutdown(sc)
	}()
	slog.Info("hub api listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
