// Command worker runs the background jobs on their own (the API already runs them unless HUB_WITH_WORKER=off): ticket timing
// sweeps, call history from the calls service, erasing deleted files, and the integrity scan. Run one or several; every job is
// safe to run on more than one node at once.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"hub/internal/app"
	"hub/internal/config"
	"hub/internal/mail"
	"hub/internal/runapi"
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
	slog.Info("hub worker running")
	runapi.RunWorker(ctx, a)
}
