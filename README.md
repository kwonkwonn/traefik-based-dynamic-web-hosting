# 🐾 CUTE WEB Proxy

> Traefik + Redis 기반 동적 라우팅 컨트롤 플레인.
> API 한 번 호출로 서브도메인이 생기고, 80번 포트로 즉시 트래픽이 흐릅니다.

```
POST /routes  →  Redis  →  Traefik  →  subdomain.yourdomain.com:80
                                              ↓
                                        VM IP:Port
```

---

## 빠른 시작

```bash
# 1. Traefik + Redis 시작
BASE_DOMAIN=yourdomain.com docker compose up -d --build

# 2. 라우트 등록
curl -X POST http://localhost:8090/routes \
  -H "Content-Type: application/json" \
  -d '{"name":"myapp","subdomain":"myapp","target_ip":"1.2.3.4","target_port":3000}'

# 3. 접근
curl http://myapp.yourdomain.com
```

---

## 로컬 테스트

```bash
# Python 서버 띄우기
python3 -m http.server 9999

# 라우트 등록
curl -X POST http://localhost:8090/routes \
  -H "Content-Type: application/json" \
  -d '{"name":"test","subdomain":"test","target_ip":"host.docker.internal","target_port":9999}'

# 브라우저에서 접근 (Chrome은 바로 됨)
open http://test.localhost
```

---

## API

| Method | Path | Body | 설명 |
|---|---|---|---|
| `POST` | `/routes` | `{name, subdomain, target_ip, target_port}` | 라우트 등록 |
| `GET` | `/routes` | — | 전체 목록 |
| `GET` | `/routes/{name}` | — | 단일 조회 |
| `PUT` | `/routes/{name}` | `{subdomain, target_ip, target_port}` | 수정 |
| `DELETE` | `/routes/{name}` | — | 삭제 |
| `GET` | `/health` | — | Redis 연결 상태 |

### 요청 예시

```bash
# 등록
curl -X POST http://localhost:8090/routes \
  -H "Content-Type: application/json" \
  -d '{"name":"blog","subdomain":"blog","target_ip":"192.168.1.10","target_port":8080}'

# 조회
curl http://localhost:8090/routes/blog

# 수정 (IP 변경)
curl -X PUT http://localhost:8090/routes/blog \
  -H "Content-Type: application/json" \
  -d '{"subdomain":"blog","target_ip":"192.168.1.20","target_port":8080}'

# 삭제
curl -X DELETE http://localhost:8090/routes/blog
```

---

## 환경변수

| 변수 | 기본값 | 설명 |
|---|---|---|
| `SERVER_PORT` | `8090` | API 서버 포트 |
| `REDIS_ADDR` | `localhost:6379` | Redis 주소 |
| `BASE_DOMAIN` | `localhost` | 서브도메인 베이스 |

---

## 구조

```
internal/
  config/      환경변수 로드
  handler/     HTTP 핸들러
  redis/       Redis 레포지토리
pkg/
  traefik/     Traefik KV 키 빌더
server/        HTTP 서버 + 라우팅
```

---

## 포트

| 포트 | 용도 |
|---|---|
| `80` | Traefik — 실제 트래픽 |
| `8080` | Traefik 대시보드 |
| `8090` | CUTE WEB Proxy API |
