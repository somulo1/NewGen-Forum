package runapi

import (
	"context"
	"log/slog"
	"os"
	"time"

	"hub/internal/app"
	"hub/internal/modules/approvals"
	"hub/internal/modules/calls"
	"hub/internal/modules/chat"
	"hub/internal/modules/files"
	"hub/internal/modules/security"
	"hub/internal/modules/projects"
	"hub/internal/modules/settings"
	"hub/internal/modules/tickets"
)

// RunWorker runs the background jobs until ctx ends: ticket timing sweeps, call history from the calls service, erasing deleted
// files and deleted organizations, and the integrity scan that alerts on records changed outside the system. Every job is safe
// to run on more than one node at once.
func RunWorker(ctx context.Context, a *app.App) {
	go calls.New(a).Consume(ctx)
	ts := tickets.New(a)
	ps := projects.New(a, approvals.New(a))
	fs := files.New(a, approvals.New(a))
	cs := chat.New(a)
	cs.Purge = func(ctx context.Context) { _, _, _ = fs.PurgeDue(ctx, 500) }
	// Moments last 24 hours; at 00:02 (Nairobi time) every day those past it are erased for good, with their pictures and videos.
	// A node that was down at that minute runs it as soon as it is back (the lists already leave expired moments out meanwhile).
	nairobi, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		nairobi = time.FixedZone("EAT", 3*3600)
	}
	lastSweep := ""
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
			if now := time.Now().In(nairobi); lastSweep != now.Format("2006-01-02") && (now.Hour() > 0 || now.Minute() >= 2) {
				lastSweep = now.Format("2006-01-02")
				if n, err := cs.ExpireMoments(ctx); err != nil {
					slog.Error("expired moments", "err", err)
					lastSweep = ""
				} else if n > 0 {
					slog.Info("expired moments erased", "count", n)
				}
			}
			// uploads that were started and never finished, once an hour
			if ticks%60 == 1 {
				if n, err := fs.SweepUploads(ctx); err != nil {
					slog.Error("unfinished uploads", "err", err)
				} else if n > 0 {
					slog.Info("unfinished uploads closed", "count", n)
				}
			}
			ps.SweepAll(ctx)
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
