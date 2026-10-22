package runapi

import (
	"context"
	"log/slog"
	"os"
	"time"

	"hub/internal/app"
	"hub/internal/modules/approvals"
	"hub/internal/modules/calls"
	"hub/internal/modules/files"
	"hub/internal/modules/security"
	"hub/internal/modules/settings"
	"hub/internal/modules/tickets"
)

// RunWorker runs the background jobs until ctx ends: ticket timing sweeps, call history from the calls service, erasing deleted
// files and deleted organizations, and the integrity scan that alerts on records changed outside the system. Every job is safe
// to run on more than one node at once.
func RunWorker(ctx context.Context, a *app.App) {
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
	ticks := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			ticks++
			// uploads that were started and never finished, once an hour
			if ticks%60 == 1 {
				if n, err := fs.SweepUploads(ctx); err != nil {
					slog.Error("unfinished uploads", "err", err)
				} else if n > 0 {
					slog.Info("unfinished uploads closed", "count", n)
				}
			}
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
