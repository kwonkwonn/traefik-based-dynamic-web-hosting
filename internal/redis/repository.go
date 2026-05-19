package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	traefikkv "traefik-based-dynamic-web-hosting/pkg/traefik"
)

const (
	routeIndexKey = "routes:index"
	routeMetaFmt  = "routes:%s"
)

type Route struct {
	Name       string    `json:"name"`
	Subdomain  string    `json:"subdomain"`
	TargetIP   string    `json:"target_ip"`
	TargetPort int       `json:"target_port"`
	CreatedAt  time.Time `json:"created_at"`
}

type Repository struct {
	client     *redis.Client
	baseDomain string
}

func NewRepository(client *redis.Client, baseDomain string) *Repository {
	return &Repository{client: client, baseDomain: baseDomain}
}

func (r *Repository) set(ctx context.Context, route Route) error {
	meta, err := json.Marshal(route)
	if err != nil {
		return err
	}

	host := fmt.Sprintf("Host(`%s.%s`)", route.Subdomain, r.baseDomain)
	targetURL := fmt.Sprintf("http://%s:%d", route.TargetIP, route.TargetPort)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, traefikkv.RouterRule(route.Name), host, 0)
	pipe.Set(ctx, traefikkv.RouterEntrypoint(route.Name), "web", 0)
	pipe.Set(ctx, traefikkv.RouterService(route.Name), route.Name, 0)
	pipe.Set(ctx, traefikkv.ServiceURL(route.Name), targetURL, 0)
	pipe.Set(ctx, fmt.Sprintf(routeMetaFmt, route.Name), meta, 0)
	pipe.SAdd(ctx, routeIndexKey, route.Name)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *Repository) Create(ctx context.Context, route Route) error {
	route.CreatedAt = time.Now()
	return r.set(ctx, route)
}

func (r *Repository) Update(ctx context.Context, route Route) error {
	existing, err := r.Get(ctx, route.Name)
	if err != nil {
		return err
	}
	route.CreatedAt = existing.CreatedAt
	return r.set(ctx, route)
}

func (r *Repository) Delete(ctx context.Context, name string) error {
	keys := traefikkv.AllRouterKeys(name)
	keys = append(keys, fmt.Sprintf(routeMetaFmt, name))

	pipe := r.client.Pipeline()
	for _, k := range keys {
		pipe.Del(ctx, k)
	}
	pipe.SRem(ctx, routeIndexKey, name)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Repository) Get(ctx context.Context, name string) (*Route, error) {
	data, err := r.client.Get(ctx, fmt.Sprintf(routeMetaFmt, name)).Bytes()
	if err != nil {
		return nil, err
	}
	var route Route
	if err := json.Unmarshal(data, &route); err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *Repository) List(ctx context.Context) ([]Route, error) {
	names, err := r.client.SMembers(ctx, routeIndexKey).Result()
	if err != nil {
		return nil, err
	}
	routes := make([]Route, 0, len(names))
	for _, name := range names {
		route, err := r.Get(ctx, name)
		if err != nil {
			continue
		}
		routes = append(routes, *route)
	}
	return routes, nil
}
