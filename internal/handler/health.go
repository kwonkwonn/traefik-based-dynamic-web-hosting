package handler

import (
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	redis *redis.Client
}

func NewHealthHandler(redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{redis: redisClient}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	redisOK := h.redis.Ping(r.Context()).Err() == nil

	status := "ok"
	code := http.StatusOK
	if !redisOK {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"redis":  redisOK,
	})
}
