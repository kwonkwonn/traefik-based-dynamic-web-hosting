package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/redis/go-redis/v9"
	"traefik-based-dynamic-web-hosting/internal/handler"
	rdb "traefik-based-dynamic-web-hosting/internal/redis"
)

func NewMux(redisClient *redis.Client, repo *rdb.Repository) *http.ServeMux {
	mux := http.NewServeMux()

	routes := handler.NewRouteHandler(repo)
	health := handler.NewHealthHandler(redisClient)

	mux.HandleFunc("POST /routes", routes.Create)
	mux.HandleFunc("DELETE /routes/{name}", routes.Delete)
	mux.HandleFunc("GET /routes", routes.List)
	mux.HandleFunc("GET /routes/{name}", routes.Get)
	mux.HandleFunc("PUT /routes/{name}", routes.Update)
	mux.HandleFunc("GET /health", health.Check)

	return mux
}

func StartServer(ctx context.Context, port string, redisClient *redis.Client, repo *rdb.Repository) error {
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: NewMux(redisClient, repo),
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	slog.Info("server listening", "port", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
