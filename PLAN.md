# Traefik-based Dynamic Web Hosting — Implementation Plan

## Final Goal
Traefik + Redis 기반 라우팅 컨트롤 플레인.
외부 플랫폼이 `POST /routes`를 호출하면 Redis에 Traefik KV 설정이 실시간 반영되어
`{subdomain}.domain.com → 기존 VM의 IP:Port` 로 트래픽이 라우팅된다.

```
외부 플랫폼 ──POST /routes──▶ Go API ──▶ Redis ──▶ Traefik ──▶ {subdomain}.domain.com
                              GET/PUT/DELETE                          │
                                                                      ▼
                                                              기존 VM의 IP:Port
```

---

## Phase 1 — 기반 정리 [✅ 완료]
- [x] Docker 관련 코드 전부 제거 (`Handler/` 디렉토리 삭제)
- [x] Go 코드 layered architecture 재구조화
  ```
  internal/config/    # 환경변수 기반 설정
  internal/handler/   # HTTP 핸들러 (route.go, health.go)
  internal/redis/     # Redis 레포지토리 (client.go, repository.go)
  pkg/traefik/        # Traefik KV 키 빌더 (keys.go)
  ```
- [x] Redis 클라이언트 추가 (`github.com/redis/go-redis/v9`)
- [x] 환경변수 config 패키지 구현 (`SERVER_PORT`, `REDIS_ADDR`, `BASE_DOMAIN`)

## Phase 2 — Traefik Redis 연동 [✅ 완료]
- [x] `docker-compose.yaml`에 Redis 서비스 추가 (redis:7-alpine, 영구 볼륨)
- [x] Traefik에서 Docker provider 제거 → Redis provider 설정 (`--providers.redis.endpoints=redis:6379`)
- [x] Docker 소켓 마운트 제거
- [x] `pkg/traefik/keys.go`: 라우트 → Redis KV 키셋 변환 빌더 구현
  ```
  traefik/http/routers/{name}/rule
  traefik/http/routers/{name}/entrypoints/0
  traefik/http/routers/{name}/service
  traefik/http/services/{name}/loadbalancer/servers/0/url
  ```

## Phase 3 — 라우팅 CRUD API [✅ 완료]
- [x] `POST /routes` — Traefik KV + 메타데이터 원자적 등록 (Redis Pipeline)
- [x] `DELETE /routes/{name}` — 관련 키 전부 삭제
- [x] `GET /routes` — 전체 라우트 목록
- [x] `GET /routes/{name}` — 단일 라우트 조회
- [x] `PUT /routes/{name}` — 라우트 수정 (IP/Port 변경, created_at 보존)

## Phase 4 — 운영 품질 [✅ 완료]
- [x] 구조화 로깅 (`log/slog`)
- [x] Graceful shutdown (`signal.NotifyContext` + `http.Server.Shutdown`)
- [x] `GET /health` — Redis 연결 상태 체크

---

## Redis 데이터 모델

### Traefik KV (Traefik이 읽음)
```
traefik/http/routers/{name}/rule                         = "Host(`{subdomain}.yourdomain.com`)"
traefik/http/routers/{name}/entrypoints/0                = "web"
traefik/http/routers/{name}/service                      = "{name}"
traefik/http/services/{name}/loadbalancer/servers/0/url  = "http://{ip}:{port}"
```

### 라우트 메타데이터 (API가 관리)
```
routes:{name}   = JSON { name, subdomain, ip, port, created_at }
routes:index    = Redis Set (name 목록)
```

## API 명세

| Method | Path | 설명 |
|---|---|---|
| POST | /routes | 라우트 등록 |
| DELETE | /routes/{name} | 라우트 제거 |
| GET | /routes | 전체 목록 |
| GET | /routes/{name} | 단일 조회 |
| PUT | /routes/{name} | 수정 |
| GET | /health | Redis 연결 상태 |

## 환경변수

| 변수 | 기본값 | 설명 |
|---|---|---|
| `SERVER_PORT` | `8090` | API 서버 포트 |
| `REDIS_ADDR` | `localhost:6379` | Redis 주소 |
| `BASE_DOMAIN` | `localhost` | 서브도메인 베이스 도메인 |

## 미결 사항
- TLS 전략 (권장: 와일드카드 인증서 `*.domain.com`)
- 실제 도메인명 확정
