# AGENTS.md — radius-api

> Read this first before touching any code in `radius-api/`. It captures the
> stack, structure, and conventions so changes stay consistent.

## What this service is

A thin **management API over the FreeRADIUS SQL schema**. It is the ONLY thing
that writes/reads the FreeRADIUS database (**MariaDB/MySQL**). The billing
backend calls it over HTTP (authenticated with a shared `X-API-Key`) to:
provision/remove voucher credentials, define bandwidth/quota profiles, list
active sessions, manage NAS (RADIUS client) records, and disconnect users via
**CoA / Packet-of-Disconnect**.

It does NOT hold business logic (no packages/orders/payments). That is the
billing backend's job. This service only manipulates RADIUS rows + talks CoA.

```
billing backend ──HTTP X-API-Key──► radius-api (this) ──SQL──► MariaDB ◄──reads── FreeRADIUS daemon
                                          │                                            ▲
                                          └── CoA/Disconnect (UDP 3799) ───────────────┘ (to the Mikrotik NAS)
```

The **FreeRADIUS daemon itself** is NOT this Go program — it is the official
`freeradius-server` image, configured by the files in `radius-api/freeradius/`
(`Dockerfile`, `docker-entrypoint.sh`, `raddb/clients.conf`,
`raddb/mods-available/sql`). FreeRADIUS reads the same MariaDB tables this API
writes. The two meet only at the database.

## Tech stack

- **Go** + **Gin** HTTP framework
- **GORM** (`gorm.io/gorm`, `gorm.io/driver/mysql`) → MariaDB/MySQL
- **goose** embedded SQL migrations (auto-run on startup) — the FreeRADIUS schema
- **layeh.com/radius** — sends CoA/Disconnect-Request packets to NAS devices
- **swaggo/swag** + **gin-swagger** — docs at `/swagger/index.html`
- **godotenv** loads `.env`

## Folder structure (`internal/`)

| Package | Responsibility |
|---|---|
| `config/` | `config.Load()` → typed `Config` (App, DB, Auth, CoA, Reload, CORS) |
| `database/` | GORM connect + goose `Migrate` (FreeRADIUS schema) |
| `models/` | GORM structs mapped to FreeRADIUS tables via `TableName()` — `radcheck`, `radreply`, `radgroupcheck`, `radgroupreply`, `radusergroup`, `radacct`, `nas` |
| `dto/` | request/response payloads (`ProfileRequest`, `UserRequest`, `BulkUsersRequest`, `NASRequest`, `UserDetail`) |
| `radiussql/` | `Service` — all the SQL operations (upsert users/profiles, list sessions, NAS CRUD, disconnect) |
| `coa/` | `Disconnector` — builds & sends CoA/PoD packets (layeh.com/radius) |
| `radiusreload/` | restarts the FreeRADIUS container via the Docker socket after a NAS change |
| `handlers/` | Gin handlers (one file `handlers.go`) — bind + call `radiussql` + render |
| `response/` | JSON envelope + helpers |
| `middleware/` | RequestID, Logger, Recovery, CORS, **APIKey** (X-API-Key) |
| `server/` | `server.New(cfg, db)` wires service + handlers + routes |

`cmd/api/main.go` = entrypoint (config → connect → migrate → server → listen).
`migrations/` has a single file: `00001_radius_schema.sql` (standard FreeRADIUS schema).

## Response envelope

Matches the backend contract (`response.Envelope`):
```json
{ "success": true, "message": "OK", "data": {...}, "error": { "code": "..." } }
```
Helpers: `response.OK(c,msg,data)` (200), `Created` (201), `NoContent` (200),
`Error(c, status, code, msg)`. NOTE: the error body here is **simpler** than the
backend — just `{code}`, no per-field details array.

## Handler conventions

In `handlers/handlers.go`:
- `bind(c, &dto) bool` — `ShouldBindJSON`; on failure writes `422 VALIDATION_ERROR` (err text) and returns false
- `fail(c, err)` — `radiussql.ErrNotFound` → `404 NOT_FOUND`; anything else → logged `500 INTERNAL_ERROR`
- Every handler has swaggo annotations (`@Security ApiKeyAuth`)

## Auth

All `/api/v1/*` routes require header `X-API-Key` matching `API_KEY` env
(`middleware.APIKey`). This value MUST equal the billing backend's
`RADIUS_API_KEY`.

## Routes (`server/server.go`)

`api := r.Group("/api/v1")` then `api.Use(middleware.APIKey(...))`:
- `POST /profiles` — upsert a FreeRADIUS group (bandwidth/quota attrs)
- `POST /users`, `POST /users/bulk`, `GET/DELETE /users/{username}`, `POST /users/{username}/disconnect`
- `GET /sessions` — active sessions (`radacct`)
- `GET /nas`, `POST /nas`, `DELETE /nas/{id}` — RADIUS clients

## FreeRADIUS attribute model

A package "profile" → a FreeRADIUS group (`radgroupcheck`/`radgroupreply`):
`Mikrotik-Rate-Limit`, `Session-Timeout`, `Mikrotik-Total-Limit`,
`Simultaneous-Use`. A voucher → a user (`radcheck` Cleartext-Password = code +
optional `Expiration`) + `radusergroup` link. Single-code login: username ==
password == voucher code.

## Run / build / test

```bash
go build ./... && go vet ./...
gofmt -w internal/ cmd/
go run ./cmd/api          # needs .env + MariaDB
# Docker (this folder = the whole RADIUS stack: mariadb + radius-api + freeradius):
docker compose up -d --build
```

## Swagger (regenerate after handler changes)

```bash
~/go/bin/swag init -g cmd/api/main.go -o docs
```

## FreeRADIUS config gotchas (in `freeradius/`)

- `docker-entrypoint.sh` renders bind-mounted templates (`/etc/raddb/templates/`)
  into real raddb paths. `envsubst` MUST be scoped to OUR vars only — a plain
  `envsubst` would blank FreeRADIUS's own `${...}` config and break the parser.
- Must symlink `mods-enabled/sql` or FreeRADIUS logs `Ignoring "sql"` and no
  DB-backed auth works.
- `read_clients = yes` + `client_table = nas` → FreeRADIUS loads NAS clients from
  SQL at boot. Adding a NAS needs a reload (`radiusreload`) to take effect.

## Hard rules

1. Keep `API_KEY` == backend `RADIUS_API_KEY`.
2. This service is stateless business-wise — do NOT add billing logic here.
3. Models map to fixed FreeRADIUS table names via `TableName()` — don't rename.
4. UDP 1812/1813 (auth/acct) belong to the FreeRADIUS daemon, not this API (8081).
5. `gofmt` + `go vet` + regenerate swagger before committing.
