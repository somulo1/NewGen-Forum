// Command worker runs the background jobs: ticket timing sweeps, call history from the calls service, and erasing the
// contents of deleted files from storage. Run one or several;
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
