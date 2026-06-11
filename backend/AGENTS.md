# AGENTS.md — Billing Backend

> Read this first before touching any code in `backend/`. It captures the stack,
> structure, and the conventions you MUST follow so changes stay consistent.

## What this service is

The **billing API** of a hotspot voucher system. It owns the business domain:
packages, vouchers, voucher batches, self-service orders, payment gateways,
revenue reports, admin auth, and storefront settings. It persists to
**PostgreSQL**. It never touches the FreeRADIUS database directly — it provisions
hotspot credentials by calling the **radius-api** microservice over HTTP.

```
frontend (SPA) ──REST /api/v1──► backend (this) ──HTTP X-API-Key──► radius-api ──SQL──► MariaDB/FreeRADIUS
                                      │
                                      └──► PostgreSQL (billing data)
```

## Tech stack

- **Go** + **Gin** (`github.com/gin-gonic/gin`) HTTP framework
- **GORM** (`gorm.io/gorm`, `gorm.io/driver/postgres`) ORM → PostgreSQL
- **goose** (`github.com/pressly/goose/v3`) SQL migrations, embedded + auto-run on startup
- **go-playground/validator/v10** request validation (via Gin binding tags)
- **golang-jwt/jwt/v5** admin auth tokens
- **swaggo/swag** + **gin-swagger** API docs at `/swagger/index.html`
- **godotenv** loads `.env`
- Payment gateways (Xendit, Midtrans, Tripay) via plain `net/http` (no SDKs)

## Folder structure (`internal/`)

| Package | Responsibility |
|---|---|
| `config/` | `config.Load()` reads env into typed `Config` (App, DB, JWT, Radius, Payment, CORS) |
| `database/` | GORM connect + goose `Migrate` (runs `migrations/*.sql` embedded) |
| `models/` | GORM entities mirroring the schema (User, Package, Voucher, Batch, Order, Setting, NAS configs, RadiusServer) |
| `dto/` | request/response payloads with `binding` tags; never expose models with secrets raw |
| `response/` | the single JSON envelope + helpers (see below) |
| `apperror/` | typed `AppError{Status, Code, Message, Err}` + constructors |
| `handlers/` | Gin handlers: translate HTTP ↔ service calls, render via `response` |
| `services/` | business logic; depend on `*gorm.DB` + clients; return `apperror.AppError` |
| `payment/` | `Gateway` interface + `Registry` (Xendit/Midtrans/Tripay) + DB-creds overlay |
| `radius/` | HTTP client to radius-api (provision users, NAS, CoA) |
| `middleware/` | RequestID, Logger, Recovery, CORS, **Auth** (JWT) |
| `token/` | JWT issue/verify |
| `util/` | code generation (`codegen.go`), string helpers |
| `validatorx/` | translate validator errors → structured field errors |
| `server/` | `server.New(cfg, db)` wires everything + registers ALL routes |

`cmd/api/main.go` = entrypoint: `config.Load()` → `database.Connect` → `database.Migrate` → `validatorx.Setup()` → `server.New` → listen.

## Response envelope (MANDATORY — every endpoint)

`response.Envelope`:
```json
{ "success": true, "message": "OK", "data": {...}, "meta": {...},
  "error": { "code": "VALIDATION_ERROR", "details": [{"field":"x","message":"...","tag":"required"}] } }
```
Helpers (use these, never write `c.JSON` directly):
- `response.OK(c, msg, data)` — 200
- `response.Created(c, msg, data)` — 201
- `response.Paginated(c, msg, data, response.NewMeta(page, perPage, total))` — 200 + list meta
- `response.NoContent(c, msg)` — 200, no payload
- `response.Error(c, status, code, msg)` — failure
- `response.ValidationError(c, msg, details)` — 422 with field details

## Error handling (MANDATORY)

Services return `apperror.AppError`. Constructors: `BadRequest`, `Unauthorized`,
`Forbidden`, `NotFound`, `Conflict`, `Unprocessable`, `Internal`,
`ServiceUnavailable`. Handlers call `fail(c, err)` which maps an AppError to its
status/code (logs cause), else 500. DB errors go through `mapDBError(err)` in
`services/service.go` (translates GORM not-found/duplicate/FK).

## Handler conventions

Helpers in `handlers/handler.go` + `params.go`:
- `bindJSON(c, &dto) bool` — bind+validate body; writes 422/400 and returns false on failure
- `bindQuery(c, &dto) bool` — same for query params
- `fail(c, err)` — render a service error
- `idParam(c) (uint, bool)` — parse `:id`

Every handler has swaggo annotations (`@Summary`, `@Router`, `@Security BearerAuth`).

## Service conventions

Each service is a struct with `New*Service(deps...)` (no shared deps struct).
Take `context.Context` as first arg. Return `apperror.AppError`. Example shape:
```go
type FooService struct { db *gorm.DB }
func NewFooService(db *gorm.DB) *FooService { return &FooService{db: db} }
func (s *FooService) Get(ctx context.Context, id uint) (*models.Foo, error) { ... mapDBError(err) }
```

## Routing (in `server/server.go`)

- Health: `GET /health`; Swagger: `GET /swagger/*any`
- `api := r.Group("/api/v1")`
- `pub := api.Group("/public")` — unauthenticated storefront (packages, checkout, order status, settings, `hotspot/login.html`)
- `api.POST("/webhooks/:provider", ...)` — gateway callbacks (signature-verified in service)
- `authed := api.Group(""); authed.Use(middleware.Auth(cfg.JWT.Secret))` — all admin endpoints

When adding an endpoint: add service method → handler (with swaggo) → register in the right group → **regenerate swagger** (see below).

## Settings & payment gateways

`Setting` is a key/value store (`models.Setting{Key, Value}`). Payment gateway
credentials are editable from the admin UI: stored in `settings`, overlaid on the
env defaults via `payment.OverlayConfig`, and the live `payment.Registry` is
hot-reloaded (`Reload`) at startup and after every gateway-settings update.
Secrets are masked on read.

## NAS / multi-RADIUS

A NAS (Mikrotik) has a FreeRADIUS client record (managed via radius-api) PLUS a
local `NASHotspotConfig` (billing DB) holding deployment settings used for
Mikrotik `.rsc` script generation. Multiple radius-api endpoints are supported:
`RadiusServer` rows + `radius_directory.go` resolve which radius-api a NAS talks
to. The Go `radius.Client` is created per-endpoint (`radius.NewClientWith`).

## Migrations

goose format (`-- +goose Up` / `-- +goose Down`), in `migrations/`, embedded via
go:embed, **run automatically on startup**. To add one: create the next numbered
file `0000N_name.sql`. Current highest: `00008`. Seed admin = `admin`/`admin123`
(bcrypt in `00002_seed.sql`) — must be changed in production.

## Run / build / test

```bash
go build ./...           # compile
go vet ./...             # static check
gofmt -w internal/ cmd/  # format (always before commit)
go run ./cmd/api         # run locally (needs .env + Postgres)
# Docker: from repo root or backend/: docker compose up -d --build
```

## Swagger (regenerate after ANY handler annotation change)

```bash
~/go/bin/swag init -g cmd/api/main.go -o docs   # (swag CLI not on PATH; lives in ~/go/bin)
# or: make backend-swagger (if swag is on PATH)
```
The `docs/` package is committed; stale docs = wrong Swagger UI. `docs.go` is
imported for side effect in `server.go`.

## Hard rules

1. NEVER return bare `c.JSON` — always go through `response` helpers.
2. NEVER leak secrets/internal errors to clients — `fail()` logs cause, returns safe message.
3. Business logic lives in `services/`, not handlers. Handlers only bind + call + render.
4. Webhook URL is `{APP_BASE_URL}/api/v1/webhooks/{provider}` — gateways must reach it.
5. Keep `RADIUS_API_KEY` (here) == radius-api `API_KEY`.
6. Run `gofmt` + `go vet` + regenerate swagger before committing handler/route changes.
