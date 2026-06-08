# 📶 Billing Hotspot

Sistem billing hotspot lengkap: pelanggan membeli paket internet lewat halaman
yang menarik (bayar **cash** atau **payment gateway**), sistem otomatis membuat
voucher dan menyuntikkannya ke **FreeRADIUS**, lalu **Mikrotik** mengautentikasi
login hotspot terhadap RADIUS tersebut.

Mendukung **3 payment gateway**: **Xendit**, **Midtrans**, dan **Tripay**.

---

## ✨ Fitur

- **Storefront publik** — landing page pemilihan paket yang menarik, checkout,
  dan halaman status pembayaran (voucher tampil otomatis setelah lunas).
- **Hybrid billing** — self-service (bayar online) **dan** batch voucher untuk
  dijual tunai. Login hotspot pakai **satu kode voucher** (username = password).
- **3 payment gateway** dengan verifikasi webhook/signature: Xendit, Midtrans, Tripay.
- **Pembayaran cash** dengan konfirmasi operator.
- **Panel admin** — dashboard, **laporan pendapatan** (grafik tren, rincian per
  metode & per paket, filter tanggal, ekspor CSV), CRUD paket, voucher, batch
  (cetak voucher), pesanan, **manajemen router/NAS**, **pengaturan payment
  gateway** (kredensial dikelola dari UI, tersamar saat ditampilkan), dan
  pengaturan storefront.
- **Integrasi FreeRADIUS** lewat microservice `radius-api` (provisioning user,
  profil bandwidth/kuota, sesi aktif, **disconnect via CoA/PoD**).
- **Response API konsisten** di kedua service, termasuk **error validasi**
  terstruktur per-field.
- **Swagger** (swaggo) untuk backend & radius-api.
- **Script Mikrotik hAP ac2** siap salin-tempel (hotspot + RADIUS + walled garden).

---

## 🏗️ Arsitektur

```
                ┌───────────────────────────────────────────────────┐
                │                   Pelanggan                        │
                │   (browser di balik hotspot Mikrotik hAP ac2)      │
                └───────────────┬───────────────────────────────────┘
                                │ pilih paket, bayar
                                ▼
        ┌──────────────┐   REST    ┌──────────────────┐   SQL   ┌────────────┐
        │  Frontend     │◄────────►│   Backend (Go)    │◄───────►│ PostgreSQL │
        │ React+Vite+TS │  /api/v1 │ Gin · GORM · JWT  │         │  (billing) │
        │ shadcn·RHF·Zod│          │ Xendit/Midtrans/  │         └────────────┘
        └──────────────┘          │ Tripay · validator │
                                   └────────┬──────────┘
                                            │ HTTP (X-API-Key)
                                            ▼
                                   ┌──────────────────┐   SQL   ┌────────────┐
                                   │   radius-api (Go) │◄───────►│  MariaDB   │
                                   │ Gin · goose · CoA │         │ (FreeRADIUS)│
                                   └────────┬─────────┘         └─────▲──────┘
                                            │ Disconnect (3799)        │ SQL
                                            ▼                          │
                                   ┌──────────────────┐                │
                                   │   Mikrotik hAP    │   RADIUS       │
                                   │   ac2 (NAS)       │◄──1812/1813───►│ FreeRADIUS
                                   └──────────────────┘                └──────────┘
```

**Dua database (sesuai permintaan):**
- **PostgreSQL** → data billing (paket, voucher, pesanan, user admin).
- **MariaDB/MySQL** → skema standar FreeRADIUS (default).

---

## 🧰 Tech stack

| Bagian | Teknologi |
|--------|-----------|
| Backend billing | Go, Gin, GORM, **goose** (migrasi), **go-playground/validator**, JWT, **swaggo** |
| radius-api | Go, Gin, GORM, **goose**, `layeh.com/radius` (CoA), **swaggo** |
| Frontend | React, **Vite**, TypeScript, Tailwind, **shadcn/ui**, **React Hook Form**, **Zod**, TanStack Query |
| Payment | Xendit (Invoice), Midtrans (Snap), Tripay (closed transaction) |
| Infra | Docker Compose, PostgreSQL, MariaDB, FreeRADIUS 3.2 |

---

## 🚀 Menjalankan dengan Docker (tiap service terpisah)

Setiap service punya **`docker-compose.yml` sendiri** supaya bisa di-start/stop &
di-maintenance independen. Database menempel ke service pemiliknya:

| Stack | Folder | Isi |
|-------|--------|-----|
| **RADIUS** | `radius-api/` | MariaDB + radius-api + FreeRADIUS |
| **Backend** | `backend/` | PostgreSQL + backend billing |
| **Frontend** | `frontend/` | SPA (nginx) |

Mereka berkomunikasi lewat satu **network bersama** `hotspot-net`. Jalankan
sesuai urutan dependensi (RADIUS → backend → frontend):

```bash
# 1) Buat network bersama — cukup SEKALI
docker network create hotspot-net

# 2) RADIUS stack (mariadb + radius-api + freeradius)
cd radius-api && cp .env.example .env && docker compose up -d --build

# 3) Backend stack (postgres + backend)
cd ../backend && cp .env.example .env && docker compose up -d --build

# 4) Frontend
cd ../frontend && cp .env.example .env && docker compose up -d --build
```

Atau lewat `make` (otomatis membuat network + urutan benar):

```bash
make radius-up      # atau: make backend-up / make frontend-up
make up             # semua stack sekaligus, tetap sebagai compose terpisah
make down           # hentikan semua
make radius-logs    # tail log per-stack
```

> **Penting:** `API_KEY` di `radius-api/.env` didaftarkan di menu
> **Admin → Radius Server → Radius API Key**. Backend tidak lagi menyimpan
> endpoint/key radius-api di `.env`.

> Untuk start/stop satu service saja, cukup `cd <folder> && docker compose
> up -d` / `docker compose down`. Database service ikut di stack-nya.

Service yang aktif:

| Service | URL |
|---------|-----|
| Frontend (storefront + admin) | http://localhost:8088 |
| Backend API | http://localhost:8080 |
| Backend Swagger | http://localhost:8080/swagger/index.html |
| radius-api | http://localhost:8081 |
| radius-api Swagger | http://localhost:8081/swagger/index.html |
| PostgreSQL | localhost:5432 |
| MariaDB | localhost:3306 |
| FreeRADIUS | UDP 1812/1813 |

> Migrasi database (goose) berjalan **otomatis** saat backend & radius-api start.
> Data awal (admin + paket contoh) ikut ter-seed.

### 🔐 Kredensial default admin

```
username: admin
password: admin123
```

Login di **http://localhost:8088/admin/login** → **ganti password** lewat menu
Pengaturan.

---

## 💻 Menjalankan secara lokal (tanpa Docker untuk app)

Tetap butuh PostgreSQL & MariaDB. Cara cepat: jalankan **hanya database** dari
stack masing-masing —
`cd radius-api && docker compose up -d mariadb` dan
`cd backend && docker compose up -d postgres` — atau pakai instalasi native.

```bash
# 1) radius-api  (port 8081)
cd radius-api && cp .env.example .env && go run ./cmd/api

# 2) backend     (port 8080)
cd backend && cp .env.example .env && go run ./cmd/api

# 3) frontend    (port 5173)
cd frontend && cp .env.example .env && pnpm install && pnpm dev
```

Atau pakai `make`: `make radius-run`, `make backend-run`, `make frontend-dev`.

---

## 💳 Konfigurasi payment gateway

Kredensial bisa diisi lewat **`.env`** *atau* langsung dari panel admin di menu
**Payment Gateway** (`/admin/payment-gateways`). Nilai dari UI disimpan di
database dan **menimpa** nilai `.env` (env tetap jadi fallback); perubahan
langsung berlaku tanpa restart. Setiap provider yang **kosong** otomatis
dinonaktifkan (tidak muncul di checkout). Daftarkan **callback/webhook URL**
berikut di
dashboard masing-masing provider (ganti `APP_BASE_URL` dengan domain publik —
gunakan tunnel seperti ngrok saat development karena gateway harus bisa
menjangkau backend kamu):

| Provider | Env yang diisi | Webhook / Callback URL |
|----------|----------------|------------------------|
| **Midtrans** | `MIDTRANS_SERVER_KEY`, `MIDTRANS_CLIENT_KEY`, `MIDTRANS_IS_PRODUCTION` | `{APP_BASE_URL}/api/v1/webhooks/midtrans` |
| **Xendit** | `XENDIT_SECRET_KEY`, `XENDIT_CALLBACK_TOKEN` | `{APP_BASE_URL}/api/v1/webhooks/xendit` |
| **Tripay** | `TRIPAY_API_KEY`, `TRIPAY_PRIVATE_KEY`, `TRIPAY_MERCHANT_CODE`, `TRIPAY_IS_PRODUCTION` | `{APP_BASE_URL}/api/v1/webhooks/tripay` |

Verifikasi keamanan webhook yang diterapkan:
- **Midtrans** — SHA-512 `order_id + status_code + gross_amount + ServerKey`.
- **Xendit** — header `x-callback-token` dicocokkan dengan `XENDIT_CALLBACK_TOKEN`.
- **Tripay** — HMAC-SHA256 body callback dengan `TRIPAY_PRIVATE_KEY`.

Saat status **PAID** diterima (dan signature valid), sistem otomatis membuat
voucher dan menyuntikkannya ke FreeRADIUS. Webhook bersifat **idempoten**.

---

## 📡 Setup Mikrotik hAP ac2

Script Mikrotik **dibuat otomatis dari panel admin** — tidak ada lagi file
`.rsc` statis yang perlu diedit manual.

1. **Admin → Router (NAS)** → tambah router (NAS Name = IP router seperti
   dilihat RADIUS, mis. `192.168.88.253`; isi RADIUS Secret unik).
2. Klik ikon **Generate Script**, sesuaikan parameter jaringan (IP server
   RADIUS, interface, subnet, dll) di dialog.
3. **Salin seluruh script** → tempel di terminal Winbox/SSH, atau **Unduh .rsc**
   lalu `/import file-name=hotspot-billing.rsc`.

Script otomatis: daftar RADIUS, aktifkan CoA (3799), IP pool + DHCP, hotspot
`use-radius=yes`, dan **walled garden** (frontend + payment gateway agar bisa
bayar sebelum login). Secret di script otomatis sama dengan NAS yang dipilih.
Tersedia juga tombol **Salin Teardown** untuk menghapus konfigurasi `billing-*`.

Jika frontend dan backend dipublikasikan di domain/subdomain berbeda, isi field
NAS secara eksplisit:

- **URL Frontend / Storefront**: URL yang dibuka pelanggan hotspot untuk beli
  paket, mis. `https://wifi.example.com`.
- **URL Backend API**: URL backend yang bisa dijangkau Mikrotik untuk mengambil
  `login.html`, mis. `https://api.example.com`.

Jika dikosongkan saat menyimpan NAS, backend mengisinya dari env:
`FRONTEND_URL` untuk storefront dan `APP_BASE_URL` untuk API. Generator script
akan memasukkan keduanya ke walled garden sebelum login. Untuk redirect setelah
pembayaran dan webhook gateway, env yang sama tetap dipakai:
`FRONTEND_URL=https://wifi.example.com` dan
`APP_BASE_URL=https://api.example.com`.

> Menambahkan router lewat panel **otomatis tersimpan ke tabel `nas`** di
> database RADIUS (dibaca FreeRADIUS via `read_clients`). Lihat catatan di bawah
> soal kapan FreeRADIUS perlu di-reload.

---

## 🔄 Alur provisioning voucher

```
Pesanan (paid) ─► backend mints Voucher ─► POST radius-api /users
                                               │
                                               ▼
                          radcheck (Cleartext-Password = kode)
                          radusergroup (group = profil paket)
                          radgroupreply (Mikrotik-Rate-Limit, Session-Timeout, …)
                                               │
                          user login di hotspot ─► FreeRADIUS Access-Accept
```

- **Profil paket** → FreeRADIUS *group* (`radgroupcheck`/`radgroupreply`):
  bandwidth (`Mikrotik-Rate-Limit`), timeout (`Session-Timeout`), kuota
  (`Mikrotik-Total-Limit`), login bersamaan (`Simultaneous-Use`).
- **Voucher** → user (`radcheck` + `radusergroup`), dengan `Expiration` bila paket
  punya masa aktif.
- **Disable / hapus voucher** → user dihapus dari RADIUS + **Disconnect-Request**
  (CoA) dikirim ke Mikrotik untuk memutus sesi yang sedang aktif.

### Multi-cabang dengan RADIUS lokal

Untuk cabang yang masing-masing punya FreeRADIUS lokal, daftarkan dulu endpoint
di **Admin → Radius Server**:

- **Radius API URL**: URL `radius-api` cabang, mis.
  `https://radius-bandung.example.com`.
- **Radius API Key**: API key `radius-api` cabang.
- **IP Server RADIUS**: IP/host FreeRADIUS yang bisa dijangkau Mikrotik cabang.

Lalu di **Admin → Router (NAS)** pilih **Radius Server** yang sesuai. Field
Radius API URL/key/IP akan terisi dari master data itu dan masih bisa dioverride
per router bila perlu.

Backend akan menyebarkan profil paket dan voucher ke semua endpoint
`radius-api` unik yang terdaftar di **Radius Server** dan NAS, sehingga voucher
yang sama bisa dipakai di semua cabang.

`radius-api` cabang adalah HTTP API, jadi boleh dipublikasikan lewat HTTPS /
Cloudflare Tunnel. Trafik RADIUS operasional tetap lokal di cabang:
Mikrotik cabang → FreeRADIUS cabang UDP `1812/1813`, dan CoA/Disconnect UDP
`3799` dari `radius-api` cabang ke Mikrotik cabang.

---

## 📁 Struktur proyek

```
billing-hotspot/
├── backend/            # Billing API (Go, Gin, GORM, goose, validator, swaggo) → PostgreSQL
│   ├── cmd/api/            # main.go
│   ├── internal/
│   │   ├── config/ database/ server/ middleware/
│   │   ├── models/ dto/ response/ apperror/ validatorx/
│   │   ├── handlers/ services/          # HTTP + business logic
│   │   ├── payment/                      # Xendit, Midtrans, Tripay
│   │   └── radius/                       # HTTP client ke radius-api
│   └── migrations/                       # goose SQL (schema + seed)
├── radius-api/         # FreeRADIUS management API (Go, Gin, goose, CoA, swaggo) → MariaDB
│   ├── cmd/api/
│   ├── internal/ (radiussql/, coa/, handlers/, …)
│   ├── migrations/                       # skema FreeRADIUS
│   ├── freeradius/                       # image + config FreeRADIUS (sql module, clients.conf)
│   └── docker-compose.yml                # stack radius: mariadb + radius-api + freeradius
├── frontend/           # React + Vite + TS + Tailwind + shadcn + RHF + Zod
│   ├── src/ (pages/public, pages/admin, components/ui, lib, schemas, …)
│   └── docker-compose.yml                # SPA (nginx)
└── Makefile                              # (script Mikrotik .rsc di-generate dari panel admin)
```

---

## 🧾 Response API yang konsisten

Semua endpoint mengembalikan amplop yang sama:

```jsonc
// Sukses
{ "success": true, "message": "OK", "data": { /* ... */ }, "meta": { /* paginasi */ } }

// Error
{ "success": false, "message": "Validation failed",
  "error": { "code": "VALIDATION_ERROR",
             "details": [ { "field": "customer_phone", "message": "...", "tag": "required" } ] } }
```

Kode error stabil: `VALIDATION_ERROR`, `BAD_REQUEST`, `UNAUTHORIZED`,
`FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `UNPROCESSABLE`, `SERVICE_UNAVAILABLE`,
`INTERNAL_ERROR`. Frontend memetakan `details` langsung ke field React Hook Form.

---

## 📚 Regenerasi Swagger (opsional)

Spesifikasi placeholder sudah disertakan agar langsung jalan. Untuk membangun
spec lengkap dari anotasi:

```bash
make install-swag      # go install swaggo/swag CLI
make backend-swagger
make radius-swagger
```

---

## 🔒 Catatan produksi

- Ganti **semua** secret di `.env` (`JWT_SECRET`, `API_KEY` radius-api,
  `NAS_SHARED_SECRET`, password DB).
- Ganti password admin default.
- Set `APP_ENV=production`, `*_IS_PRODUCTION=true`, dan `APP_BASE_URL`/
  `FRONTEND_URL` ke domain publik (HTTPS).
- Persempit `NAS_CLIENT_SUBNET` & `CORS_ALLOWED_ORIGINS`.
- Taruh backend/frontend di belakang reverse proxy TLS.
