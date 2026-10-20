// Command worker runs the background jobs: ticket timing sweeps, call history from the calls service, and erasing the
// contents of deleted files from storage, and the integrity scan that alerts on records changed outside the system. Run one or several;
// every job is safe to run on more than one node at once.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/mail"
	"hub/internal/modules/approvals"
	"hub/internal/modules/calls"
	"hub/internal/modules/files"
	"hub/internal/modules/security"
	"hub/internal/modules/settings"
	"hub/internal/modules/tickets"
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
	a, err := app.New(ctx, cfg, &mail.Dev{})
	if err != nil {
		slog.Error("startup", "err", err)
		os.Exit(1)
	}
	defer a.Close()
	go calls.New(a).Consume(ctx)
	ts := tickets.New(a)
	fs := files.New(a, approvals.New(a))
	security.New(a) // audit entries written here raise their security alerts
	auditURL := os.Getenv("HUB_AUDIT_URL")
	if auditURL == "" {
		auditURL = "postgres://hub_audit:hub_audit_dev_pw@127.0.0.1:5433/hub?sslmode=disable"
	}
	every, err := time.ParseDuration(os.Getenv("HUB_INTEGRITY_EVERY"))
	if err != nil || every <= 0 {
		every = time.Hour
	}
	scan := func() {
		n, err := security.ScanIntegrity(ctx, a, auditURL)
		if err != nil {
			slog.Error("integrity scan", "err", err)
		} else if n > 0 {
			slog.Warn("integrity scan found records changed outside the system", "problems", n)
		}
	}
	go func() {
		scan()
		it := time.NewTicker(every)
		defer it.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-it.C:
				scan()
			}
		}
	}()
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	slog.Info("hub worker running")
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := ts.SweepAll(ctx); err != nil {
				slog.Error("ticket sweep", "err", err)
			}
			// organizations whose 30-day deletion wait is over: their files first, then every record and key
			if due, err := settings.New(a).DeletionDue(ctx); err == nil {
				for _, org := range due {
					if err := fs.EraseOrg(ctx, org, settings.New(a)); err != nil {
						slog.Error("organization deletion", "org", org, "err", err)
					}
				}
			}
			// deleted files whose contents could not be erased straight away
			for {
				done, failed, err := fs.PurgeDue(ctx, 500)
				if err != nil {
					slog.Error("file purge", "err", err)
				}
				if done > 0 || failed > 0 {
					slog.Info("file purge", "erased", done, "waiting", failed)
				}
				if err != nil || done+failed < 500 {
					break
				}
			}
		}
	}
}
