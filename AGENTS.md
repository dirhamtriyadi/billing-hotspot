# AGENTS.md — Billing Hotspot (monorepo)

> **Read this first.** This is a monorepo for a hotspot voucher billing system.
> Each service has its OWN detailed `AGENTS.md` — read the one for the area you
> are changing before writing code.

| Service | Path | Stack | Detailed guide |
|---|---|---|---|
| Billing backend | `backend/` | Go + Gin + GORM → **PostgreSQL** | [`backend/AGENTS.md`](backend/AGENTS.md) |
| RADIUS API | `radius-api/` | Go + Gin + GORM → **MariaDB/FreeRADIUS** (+ `freeradius/` config) | [`radius-api/AGENTS.md`](radius-api/AGENTS.md) |
| Frontend SPA | `frontend/` | React + Vite + TS + Tailwind + shadcn | [`frontend/AGENTS.md`](frontend/AGENTS.md) |

## How the pieces fit

```
customer browser
   │  REST /api/v1 (HTTPS/HTTP)
   ▼
frontend (SPA, nginx)            backend (Go)  ──HTTP X-API-Key──►  radius-api (Go)  ──SQL──►  MariaDB
                                    │                                                            ▲
                                    └──► PostgreSQL (billing)                FreeRADIUS daemon ──┘ reads same tables
                                                                                   ▲
                                                          Mikrotik NAS ──RADIUS 1812/1813──┘  ◄── CoA 3799 (from radius-api)
```

- **backend** owns business logic (packages, vouchers, orders, payments, reports, auth, settings) on PostgreSQL. It provisions hotspot credentials only via radius-api HTTP calls.
- **radius-api** is the only writer of the FreeRADIUS MariaDB schema; it also sends CoA disconnects and can reload the FreeRADIUS container.
- **freeradius** (under `radius-api/freeradius/`) is the official FreeRADIUS daemon, configured to read the same MariaDB tables; it authenticates Mikrotik hotspot logins.
- **frontend** is the storefront + admin UI. All UI text is **Indonesian**.

## Deployment model (per-service Docker Compose)

There is **no root docker-compose**. Each service has its own `docker-compose.yml`
so they start/stop independently, joined by an external network `hotspot-net`:

```bash
docker network create hotspot-net            # once
cd radius-api && docker compose up -d --build   # mariadb + radius-api + freeradius
cd ../backend  && docker compose up -d --build  # postgres + backend
cd ../frontend && docker compose up -d --build  # nginx SPA
# or, in order, from repo root: make up   (make down / make radius-up / backend-up / frontend-up)
```

Ports: frontend `:8088`, backend `:8080`, radius-api `:8081`, Postgres `:5432`,
MariaDB `:3306`, FreeRADIUS UDP `1812/1813`, CoA `3799`.

## Cross-cutting conventions

- **Consistent JSON envelope** on both Go services: `{success, message, data, meta?, error?{code, details?}}`. Use the `response` package helpers, never bare `c.JSON`.
- **Shared secrets must match:** backend `RADIUS_API_KEY` == radius-api `API_KEY`; Mikrotik `/radius secret` == `NAS_SHARED_SECRET`.
- **`.env` is gitignored** (holds secrets). Each service ships a `.env.example` —
  copy it and fill values per environment. Never commit a real `.env`.
- **Migrations** (goose) run automatically on each Go service's startup.
- **Swagger** docs are committed and must be regenerated after handler changes
  (`~/go/bin/swag init -g cmd/api/main.go -o docs` in each Go service).
- **UI is Indonesian**; the Mikrotik `.rsc` script + captive `login.html` are
  generated from the frontend (admin → Router/NAS), not hand-edited files.

## Before you commit (any service)

- Go: `gofmt -w ...` + `go build ./...` + `go vet ./...` + regenerate swagger if handlers changed.
- Frontend: `tsc --noEmit` + `vite build`; if a deployed change won't show, rebuild Docker with `--no-cache`.
- Never let `.env`, `node_modules/`, or `dist/` into a commit (already gitignored).

## Key operational notes (hard-won)

- The Mikrotik **walled garden** must NOT whitelist broad Google ranges
  (`74.125/16`, `142.250/15`, `8.8.8.8`) — that defeats Android captive-portal
  detection and the login page stops appearing. Allowed payment ranges are
  IP-based (see `frontend/src/lib/mikrotik.ts`).
- Payment-gateway **webhooks** need to reach the backend from the internet
  (`{APP_BASE_URL}/api/v1/webhooks/{provider}`). If unreachable, an admin can
  settle a stuck gateway order manually in Admin → Orders ("Lunaskan").
- `VITE_API_BASE_URL` is **host only** (no `/api/v1`) and is baked at build time —
  changing it requires a frontend rebuild.
