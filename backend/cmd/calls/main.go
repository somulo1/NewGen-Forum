// Command calls is the calls microservice (signaling and room control for voice and video). It scales separately from the API.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"hub/internal/callsvc"
	"hub/internal/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(cfg.RedisClientOptions())
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("redis", "err", err)
		os.Exit(1)
	}
	addr := os.Getenv("HUB_CALLS_ADDR")
	if addr == "" {
		addr = ":8090"
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	srv := &http.Server{Addr: addr, Handler: callsvc.New(cfg.CallsSecret, rdb, cfg.AllowedOrigins).Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 2 * time.Minute}
	go func() {
		<-ctx.Done()
		sc, c := context.WithTimeout(context.Background(), 10*time.Second)
		defer c()
		_ = srv.Shutdown(sc)
	}()
	slog.Info("hub calls listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "err", err)
		os.Exit(1)
	}
}
