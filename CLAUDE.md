# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

A Go REST API server that dynamically manages Docker containers, designed to work alongside Traefik as a reverse proxy. Containers are declared via YAML specs; the API handles pulling images and launching containers with Traefik-compatible labels.

## Commands

```bash
# Run the API server (requires Docker daemon accessible)
go run main.go

# Build binary
go build -o webhosting .

# Tidy dependencies
go mod tidy

# Start Traefik alongside the API
docker compose up -d

# Pull an image via the API
curl -X POST http://localhost:8090/images/pull -d '{"name":"nginx:latest"}'
```

## Architecture

**Entry point**: `main.go` — initializes the Docker client, starts the HTTP server goroutine on port 8090, then blocks.

**`Handler/` package** — wraps the Docker API (`github.com/moby/moby/client`):
- `client.go`: `DockClient` struct; `InitDockClient()` builds the client from environment
- `container.go`: `ContainerSpec` (YAML schema) → `RunFromSpec()` / `RunFromFile()` pipeline. Port bindings are parsed from `"hostPort:containerPort[/proto]"` strings; restart policies mapped from string names.
- `image.go`: `ImageHandler` — HTTP handler for `POST /images/pull`

**`server/` package** — `server.go` sets up the HTTP mux (port 8090) and registers handlers.

**`docker-compose.yaml`** — runs Traefik v3.6 with the Docker provider. The Go server creates containers with Traefik labels so they get picked up automatically for routing.

## Container Spec YAML

Containers are defined as YAML files loaded by `RunFromFile()`:

```yaml
name: my-service
image: nginx:latest
ports:
  - "8080:80"
labels:
  traefik.enable: "true"
  traefik.http.routers.my-service.rule: "Host(`my-service.localhost`)"
env:
  - FOO=bar
restart: always
```

`ContainerSpec` fields: `name`, `image`, `command`, `ports`, `volumes`, `env`, `labels`, `restart`, `network_mode`. All optional except `image`.
