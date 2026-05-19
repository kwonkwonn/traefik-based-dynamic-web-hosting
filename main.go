package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"traefik-based-dynamic-web-hosting/internal/config"
	rdb "traefik-based-dynamic-web-hosting/internal/redis"
	"traefik-based-dynamic-web-hosting/server"
)

func main() {
	cfg := config.Load()

	redisClient := rdb.NewClient(cfg.RedisAddr)
	defer redisClient.Close()

	if err := rdb.Ping(context.Background(), redisClient); err != nil {
		slog.Error("redis connection failed", "err", err)
		os.Exit(1)
	}
	slog.Info("redis connected", "addr", cfg.RedisAddr)

	repo := rdb.NewRepository(redisClient, cfg.BaseDomain)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := server.StartServer(ctx, cfg.ServerPort, redisClient, repo); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
	slog.Info("shutdown complete")
}
