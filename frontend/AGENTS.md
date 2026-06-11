# AGENTS.md — Frontend (Storefront + Admin SPA)

> Read this first before touching any code in `frontend/`. It captures the
> stack, structure, and conventions so changes stay consistent.

## What this is

A single React SPA serving TWO audiences:
- **Public storefront** (`/`, `/checkout/:slug`, `/payment/:orderNumber`) — where
  customers pick a hotspot package, pay, and get their voucher code.
- **Admin panel** (`/admin/*`, JWT-protected) — dashboard, reports, packages,
  vouchers, batches, orders, NAS/routers, RADIUS servers, payment gateways,
  settings.

It talks to the billing backend REST API. **All user-facing text is in
Indonesian** — keep it that way.

## Tech stack

- **React 18** + **Vite** + **TypeScript**
- **Tailwind CSS** + **shadcn/ui** (Radix primitives in `components/ui/`) + `tailwindcss-animate`
- **React Router v6** (`react-router-dom`)
- **TanStack Query** (`@tanstack/react-query`) for all server state
- **React Hook Form** + **Zod** (`@hookform/resolvers/zod`) for forms/validation
- **axios** HTTP client
- **sonner** toasts, **lucide-react** icons, **recharts** charts
- **pnpm** package manager (NOT npm/yarn)

## Folder structure (`src/`)

| Path | Responsibility |
|---|---|
| `lib/api.ts` | the axios client + typed `api.*` endpoint groups; `unwrap`/`listRequest`/`request`; `ApiError`; token store |
| `lib/form.ts` | `applyApiErrors` (maps 422 field errors → RHF), `errorMessage` |
| `lib/format.ts` | `formatIDR`, `formatDateTime`, `formatDate`, `formatQuota` |
| `lib/phone.ts` | `normalizeWaNumber`, `isValidWaNumber` (WhatsApp number standard 62…) |
| `lib/hotspot.ts` | captures `?gw=` (hotspot gateway) → `hotspotGateway()` for one-tap login |
| `lib/mikrotik.ts` | generates the Mikrotik `.rsc` setup/teardown scripts + captive `login.html` |
| `lib/utils.ts` | `cn()` classname helper |
| `types/index.ts` | all TS interfaces mirroring the backend envelope + models |
| `schemas/index.ts` | Zod schemas (Indonesian messages) + `z.infer` types |
| `pages/public/` | storefront pages |
| `pages/admin/` | admin pages (one per feature) |
| `components/ui/` | shadcn primitives (button, card, dialog, input, select, table, tabs, …) |
| `components/` | shared app components (StatusBadge, Pagination, ProtectedRoute, …) |
| `context/auth.tsx` | auth provider (token + current user) |
| `layouts/AdminLayout.tsx` | admin shell + sidebar nav |
| `App.tsx` | route table; `main.tsx` wires QueryClient + Router + AuthProvider + Toaster |

## API client conventions (`lib/api.ts`)

- Base URL = `VITE_API_BASE_URL` (**HOST ONLY**, e.g. `http://host:8080`, no path)
  + `API_VERSION` constant (`/api/v1`) added in code. To bump API version, edit
  `API_VERSION` — do NOT put the version in the env var.
- Three request helpers: `request<T>` (returns `{data, meta}`), `unwrap<T>`
  (returns `data`), `listRequest<T>` (returns `{data: T[], meta}`).
- All endpoints grouped under `api`: `api.auth`, `api.public`, `api.dashboard`,
  `api.packages`, `api.vouchers`, `api.batches`, `api.orders`, `api.settings`,
  `api.paymentGateways`, `api.nas`, `api.reports`, … Add new calls to the right group.
- Errors throw `ApiError { message, code, status, details }`. JWT is auto-attached
  from `tokenStore`; a 401 on an `/admin` route clears the token and redirects to login.

## Form / error conventions

- Forms: `useForm` + `zodResolver(schema)`. Schemas live in `schemas/index.ts`
  with Indonesian messages.
- On submit error: `if (!applyApiErrors(e, form.setError)) toast.error(errorMessage(e))`
  — maps backend 422 field errors onto the form; otherwise shows a toast.
- Server state: TanStack Query (`useQuery`/`useMutation`); invalidate the relevant
  `queryKey` on success. Public settings key = `["public-settings"]`.

## UI conventions

- Use shadcn components from `components/ui/` — don't hand-roll buttons/inputs.
- **All copy in Indonesian.** Currency via `formatIDR`, dates via `formatDateTime`.
- Status badges via `components/StatusBadge`. Lists paginate via `components/Pagination`.

## Env vars (baked at BUILD time by Vite)

| Var | Meaning |
|---|---|
| `VITE_API_BASE_URL` | backend host only, e.g. `http://192.168.x.x:8080` (no `/api/v1`) |
| `VITE_HOTSPOT_GATEWAY` | fallback Mikrotik gateway for the one-tap WiFi login button (default `10.5.50.1`) |
| `FRONTEND_PORT` | nginx host port (docker compose only) |

Because these are baked at build, **changing them requires a frontend rebuild**.

## Run / build

```bash
pnpm install
pnpm dev                          # dev server (Vite)
node_modules/.bin/tsc --noEmit    # typecheck (do before commit)
node_modules/.bin/vite build      # production build
# Docker:
docker compose up -d --build
```

## Gotchas (learned the hard way)

1. **Stale Docker bundle:** a plain `docker compose up -d --build frontend` can
   reuse a CACHED `pnpm build` layer and serve OLD code. If a change doesn't
   appear, rebuild clean: `docker compose build --no-cache frontend && docker compose up -d frontend`.
2. **pnpm build-script approval:** esbuild needs its postinstall allowed —
   handled via `pnpm-workspace.yaml` (`onlyBuiltDependencies`/`allowBuilds`). If
   `pnpm install` aborts with `ERR_PNPM_IGNORED_BUILDS`, that file is why.
3. **Mikrotik `.rsc` generator** (`lib/mikrotik.ts`): every statement must be on
   its own line (RouterOS terminal breaks multi-line `:foreach`/array blocks when
   pasted). Never whitelist broad Google ranges in the generated walled garden —
   it breaks the captive-portal detection.
4. **WhatsApp number** is normalized to `62…` at input (Settings) AND when building
   `wa.me` links — both via `lib/phone.ts`. Keep them consistent.

## Hard rules

1. Keep all UI text **Indonesian**.
2. New API calls go in the right `api.*` group in `lib/api.ts` — never call axios ad-hoc in a page.
3. Forms use Zod + `applyApiErrors`/`errorMessage` for backend error mapping.
4. `VITE_API_BASE_URL` is host-only; version lives in `API_VERSION`.
5. Run `tsc --noEmit` + `vite build` before committing; rebuild Docker `--no-cache` if a change won't show.
