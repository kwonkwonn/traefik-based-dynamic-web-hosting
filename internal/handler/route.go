package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/redis/go-redis/v9"
	rdb "traefik-based-dynamic-web-hosting/internal/redis"
)

type RouteHandler struct {
	repo *rdb.Repository
}

func NewRouteHandler(repo *rdb.Repository) *RouteHandler {
	return &RouteHandler{repo: repo}
}

type routeRequest struct {
	Name       string `json:"name"`
	Subdomain  string `json:"subdomain"`
	TargetIP   string `json:"target_ip"`
	TargetPort int    `json:"target_port"`
}

func (h *RouteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req routeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Subdomain == "" || req.TargetIP == "" || req.TargetPort == 0 {
		http.Error(w, "name, subdomain, target_ip, target_port required", http.StatusBadRequest)
		return
	}
	route := rdb.Route{Name: req.Name, Subdomain: req.Subdomain, TargetIP: req.TargetIP, TargetPort: req.TargetPort}
	if err := h.repo.Create(r.Context(), route); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, route)
}

func (h *RouteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := h.repo.Delete(r.Context(), name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RouteHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	route, err := h.repo.Get(r.Context(), name)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, route)
}

func (h *RouteHandler) List(w http.ResponseWriter, r *http.Request) {
	routes, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, routes)
}

func (h *RouteHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req routeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	route := rdb.Route{Name: name, Subdomain: req.Subdomain, TargetIP: req.TargetIP, TargetPort: req.TargetPort}
	if err := h.repo.Update(r.Context(), route); err != nil {
		if errors.Is(err, redis.Nil) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, route)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
