// Generates ready-to-paste Mikrotik artifacts from a NAS record, so operators
// never hand-edit files in a separate folder:
//   • a RouterOS setup .rsc  (RADIUS + CoA + pool/DHCP + hotspot + walled garden)
//   • a teardown .rsc        (removes every billing-* object)
//   • a captive-portal login.html (guides the customer: buy a package, then
//     enter the voucher code)

import type { NAS } from "@/types";

// Range IP gateway pembayaran yang diizinkan SEBELUM login, supaya pelanggan
// bisa membuka & membayar di halaman Midtrans Snap langsung dari hotspot tanpa
// "blank putih"/ERR_CONNECTION_CLOSED. Pakai IP (bukan hostname) karena rule
// dst-host gagal di TLS/SNI modern. KOMBINASI INI TERVERIFIKASI di lapangan
// 2026-05-31: Snap tampil penuh DAN portal login tetap muncul.
//
// PENTING: JANGAN tambahkan range Google (74.125.0.0/16, 142.250.0.0/15) atau
// 8.8.8.8 di sini — itu menelan host deteksi captive-portal Android
// (connectivitycheck.gstatic.com) sehingga HP mengira sudah online dan portal
// login TIDAK MUNCUL. Snap tetap tampil penuh tanpa Google (font default saja).
const PAYMENT_IP_RANGES = [
  "8.215.0.0/16", // Alibaba — Midtrans app/api/snap (incl. sandbox)
  "147.139.0.0/16", // Alibaba — Midtrans snap-assets CDN
  "43.109.0.0/16", // Alibaba — Midtrans payment channels
  "34.101.0.0/16", // GCP Jakarta — Veritrans/Midtrans API
  "23.0.0.0/8", // Akamai CDN — Snap static assets (lebar; pre-login only)
];

export interface MikrotikParams {
  /** Public IP of the FreeRADIUS server as the Mikrotik can reach it. */
  radiusIP: string;
  /** Shared secret — must equal the NAS `secret` registered here. */
  radiusSecret: string;
  /** CoA / Disconnect port (radius-api COA_PORT, default 3799). */
  coaPort: string;
  /** WAN interface = the SOURCE of internet (uplink to the modem/ISP), e.g.
   *  ether1. Used to masquerade (NAT) hotspot clients out to the internet. */
  wanInterface: string;
  /** Interface the hotspot runs on. In bridge mode this is the bridge NAME the
   *  script creates (e.g. bridge-hotspot); otherwise an EXISTING interface —
   *  a single one (wlan1) or the router's built-in bridge (e.g. `bridge`) so
   *  whatever is already in that bridge (the stock WiFi) becomes the hotspot. */
  hsInterface: string;
  /** When non-empty, the script first creates `hsInterface` as a bridge and
   *  adds these member ports to it (comma-separated, e.g. "wlan1,wlan2"). Use
   *  this to serve several WiFi bands (2.4G+5G) under one hotspot. Empty = run
   *  the hotspot directly on the existing `hsInterface` (single port, or the
   *  router's built-in bridge / stock WiFi). */
  bridgePorts: string;
  /** Hotspot subnet in CIDR (e.g. 10.5.50.0/24). */
  hsNetwork: string;
  /** Hotspot gateway = router IP on that interface (e.g. 10.5.50.1). */
  hsGateway: string;
  /** DHCP pool range (e.g. 10.5.50.10-10.5.50.254). */
  hsPoolRange: string;
  /** Upstream DNS servers (comma-separated). */
  hsDNS: string;
  /** Host serving the frontend + backend API (IP or domain, no scheme). */
  feHost: string;
}

const defaultMikrotikParams: Omit<
  MikrotikParams,
  "radiusIP" | "radiusSecret" | "feHost"
> = {
  coaPort: "3799",
  // WAN uplink to the internet (ISP/modem side). ether1 is the hAP ac2 default.
  wanInterface: "ether1",
  // Default to a WLAN bridge — the common hAP ac2 case (2.4G + 5G under one
  // hotspot). The script creates `bridge-hotspot` from wlan1+wlan2.
  hsInterface: "bridge-hotspot",
  bridgePorts: "wlan1,wlan2",
  hsNetwork: "10.5.50.0/24",
  hsGateway: "10.5.50.1",
  hsPoolRange: "10.5.50.10-10.5.50.254",
  hsDNS: "8.8.8.8,1.1.1.1",
};

function configuredBackendURL(): string {
  const fromWindow =
    typeof window !== "undefined" && window.location.hostname
      ? `http://${window.location.hostname}:8080`
      : "";

  const raw = import.meta.env.VITE_API_BASE_URL || "";
  try {
    if (raw) {
      const url = new URL(raw);
      if (!isLocalHost(url.hostname)) return trimURLPath(url);
    }
  } catch {
    const endpoint = normalizeHTTPBase(raw);
    if (!isLocalEndpoint(endpoint)) return endpoint;
  }

  if (
    typeof window !== "undefined" &&
    window.location.hostname &&
    !isLocalHost(window.location.hostname)
  ) {
    return fromWindow;
  }

  return fromWindow || "http://localhost:8080";
}

function configuredBackendHost(): string {
  return new URL(configuredBackendURL()).hostname;
}

function isLocalHost(host: string): boolean {
  const normalized = host.trim().toLowerCase();
  return (
    normalized === "localhost" ||
    normalized === "::1" ||
    normalized.startsWith("127.")
  );
}

function isLocalEndpoint(endpoint: string): boolean {
  try {
    return isLocalHost(new URL(endpoint).hostname);
  } catch {
    return false;
  }
}

function trimURLPath(url: URL): string {
  return `${url.protocol}//${url.host}`;
}

function normalizeHTTPBase(value: string): string {
  const raw = value.trim().replace(/\/+$/, "");
  if (!raw) return "";
  if (raw.startsWith("http://") || raw.startsWith("https://")) {
    const url = new URL(raw);
    return trimURLPath(url);
  }
  return `http://${raw.replace(/\/.*$/, "")}`;
}

function routerReachableHost(host?: string): string {
  const normalized = (host || "").trim();
  if (!normalized) return configuredBackendHost();
  try {
    if (isLocalHost(new URL(normalizeHTTPBase(normalized)).hostname)) {
      return configuredBackendHost();
    }
  } catch {
    if (isLocalHost(normalized.replace(/:\d+$/, ""))) {
      return configuredBackendHost();
    }
  }
  return normalized;
}

/** Derive sensible network defaults + the secret from a NAS record. */
export function paramsFromNas(
  nas: NAS,
  overrides: Partial<MikrotikParams> = {},
): MikrotikParams {
  const cfg = nas.hotspot_config;
  const host = configuredBackendHost();
  return {
    radiusIP: routerReachableHost(cfg?.radius_ip) || host,
    radiusSecret: nas.secret,
    feHost: routerReachableHost(cfg?.frontend_host) || host,
    ...defaultMikrotikParams,
    coaPort: cfg?.coa_port || defaultMikrotikParams.coaPort,
    wanInterface: cfg?.wan_interface || defaultMikrotikParams.wanInterface,
    hsInterface:
      cfg?.hotspot_interface || defaultMikrotikParams.hsInterface,
    bridgePorts: cfg?.bridge_ports ?? defaultMikrotikParams.bridgePorts,
    hsNetwork: cfg?.hotspot_network || defaultMikrotikParams.hsNetwork,
    hsGateway: cfg?.hotspot_gateway || defaultMikrotikParams.hsGateway,
    hsPoolRange:
      cfg?.hotspot_pool_range || defaultMikrotikParams.hsPoolRange,
    hsDNS: cfg?.hotspot_dns || defaultMikrotikParams.hsDNS,
    ...overrides,
  };
}

/** Best-effort storefront URL from the walled-garden host (nginx serves :8088). */
export function storeUrlFromParams(p: { feHost: string }): string {
  const endpoint = normalizeHTTPBase(p.feHost);
  if (!endpoint) return `http://${configuredBackendHost()}:8088`;
  const url = new URL(endpoint);
  return `${url.protocol}//${url.hostname}:8088`;
}

/** Backend API base the router fetches login.html from. */
function backendUrlFromParams(p: { feHost: string }): string {
  const endpoint = normalizeHTTPBase(p.feHost);
  return endpoint || configuredBackendURL();
}

/** Build the full setup .rsc script text. */
export function generateMikrotikScript(p: MikrotikParams): string {
  // Plain URL with NO query string: RouterOS terminal treats "?" as its help
  // key, which mangles a pasted URL. The backend already defaults the "Beli
  // Paket" target to its FRONTEND_URL, so no ?store= is needed.
  const loginFetchUrl = `${backendUrlFromParams(p)}/api/v1/public/hotspot/login.html`;
  const ports = (p.bridgePorts || "")
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
  const bridgeMode = ports.length > 0;

  // Prefix length for the gateway address, taken from the hotspot network CIDR
  // (e.g. "10.5.50.0/24" → "24") so a non-/24 subnet still gets a matching mask
  // instead of a hardcoded /24.
  const prefix = (p.hsNetwork.split("/")[1] || "24").trim();

  // The walled garden must allow the storefront/backend host BEFORE login. When
  // feHost is an IP, RouterOS dst-host matching does not catch it, so we add it
  // by dst-address (IP rule). When it's a hostname we add it as dst-host.
  const fe = (p.feHost || "")
    .trim()
    .replace(/^https?:\/\//, "")
    .replace(/[:/].*$/, "");
  const feIsIP = /^\d{1,3}(\.\d{1,3}){3}$/.test(fe);
  // Step labels are 1-based; the total adjusts to whether the optional bridge
  // step is present (8 with bridge, 7 without) so "[n/total]" reads correctly.
  // s(n) takes the bridge-mode step number (2..8) and shifts it down by 1 when
  // there is no bridge step.
  const total = bridgeMode ? 8 : 7;
  const s = (n: number) => `${bridgeMode ? n : n - 1}/${total}`;

  // Step 1 (bridge): only emitted in bridge mode. Creates the bridge and moves
  // each WLAN/ether port into it (removing it from any existing bridge first so
  // the move is clean). Wrapped in :do/on-error so a Winbox-over-WiFi session
  // dropping mid-move doesn't abort the whole script.
  const bridgeBlock = bridgeMode
    ? `# --- 1. Bridge WLAN (gabung beberapa interface jadi satu hotspot) ------------
:put "[1/${total}] Membuat bridge $hsInterface untuk: ${ports.join(", ")}..."
:if ([:len [/interface bridge find where name=$hsInterface]] = 0) do={
  /interface bridge add name=$hsInterface
}
${ports
  .map(
    (port) => `:do {
  /interface bridge port remove [find where interface=${port}]
} on-error={}
:if ([:len [/interface bridge port find where bridge=$hsInterface and interface=${port}]] = 0) do={
  /interface bridge port add bridge=$hsInterface interface=${port}
}`,
  )
  .join("\n")}

`
    : "";

  return `# =============================================================================
#  Billing Hotspot — Mikrotik setup (RouterOS 6.4x / 7.x)
#  Dihasilkan otomatis dari panel admin. Salin-tempel SELURUH isi ini ke
#  terminal Winbox/SSH, atau simpan sebagai file lalu:
#     /import file-name=hotspot-billing.rsc
#
#  PENTING: radiusSecret di bawah HARUS sama dengan secret NAS di panel admin
#  dan dengan yang dikirim FreeRADIUS. Sudah terisi otomatis.
# =============================================================================

:global wanInterface  "${p.wanInterface}"
:global hsInterface   "${p.hsInterface}"
:global hsNetwork     "${p.hsNetwork}"
:global hsGateway     "${p.hsGateway}"
:global hsPoolRange   "${p.hsPoolRange}"
:global hsDNS         "${p.hsDNS}"
:global radiusIP      "${p.radiusIP}"
:global radiusSecret  "${p.radiusSecret}"
:global radiusCoAPort "${p.coaPort}"

:put "[*] Setup Billing Hotspot..."

${bridgeBlock}# --- RADIUS ------------------------------------------------------------------
:put "[${s(2)}] Daftar server RADIUS..."
/radius
:if ([:len [find where address=$radiusIP service~"hotspot"]] = 0) do={
  add service=hotspot address=$radiusIP secret=$radiusSecret \\
      authentication-port=1812 accounting-port=1813 timeout=3000ms comment="billing-hotspot"
} else={
  set [find where address=$radiusIP] secret=$radiusSecret service=hotspot \\
      authentication-port=1812 accounting-port=1813 timeout=3000ms
}
/radius incoming set accept=yes port=$radiusCoAPort

# --- IP Pool & DHCP ----------------------------------------------------------
:put "[${s(3)}] IP pool & DHCP..."
/ip pool
:if ([:len [find where name="hs-pool"]] = 0) do={
  add name=hs-pool ranges=$hsPoolRange
} else={ set [find where name="hs-pool"] ranges=$hsPoolRange }

/ip address
:if ([:len [find where interface=$hsInterface and address~$hsGateway]] = 0) do={
  add address=($hsGateway . "/${prefix}") interface=$hsInterface comment="hotspot-gw"
}

/ip dhcp-server
:if ([:len [find where name="hs-dhcp"]] = 0) do={
  add name=hs-dhcp interface=$hsInterface address-pool=hs-pool lease-time=1h disabled=no
}
/ip dhcp-server network
:if ([:len [find where address=$hsNetwork]] = 0) do={
  add address=$hsNetwork gateway=$hsGateway dns-server=$hsDNS comment="billing-hotspot"
}

# --- NAT / Masquerade (sumber internet) --------------------------------------
# Agar klien hotspot bisa keluar ke internet, trafik dari subnet hotspot di-NAT
# (masquerade) keluar lewat WAN ($wanInterface = sumber internet). Tanpa ini,
# klien dapat IP & bisa login tapi tidak ada koneksi internet.
:put "[${s(4)}] NAT masquerade lewat WAN $wanInterface..."
/ip firewall nat
:if ([:len [find where comment="billing-nat"]] = 0) do={
  add chain=srcnat src-address=$hsNetwork out-interface=$wanInterface \\
      action=masquerade comment="billing-nat"
} else={
  set [find where comment="billing-nat"] src-address=$hsNetwork out-interface=$wanInterface
}

# --- Hotspot profile (login via RADIUS) --------------------------------------
:put "[${s(5)}] Hotspot profile (RADIUS)..."
/ip hotspot profile
:if ([:len [find where name="hsprof-billing"]] = 0) do={
  add name=hsprof-billing hotspot-address=$hsGateway \\
      login-by=http-chap,http-pap use-radius=yes radius-accounting=yes \\
      radius-interim-update=5m
} else={
  set [find where name="hsprof-billing"] hotspot-address=$hsGateway \\
      login-by=http-chap,http-pap use-radius=yes radius-accounting=yes radius-interim-update=5m
}

# --- Hotspot server ----------------------------------------------------------
:put "[${s(6)}] Aktifkan hotspot di $hsInterface..."
/ip hotspot
:if ([:len [find where name="hs-billing"]] = 0) do={
  add name=hs-billing interface=$hsInterface address-pool=hs-pool profile=hsprof-billing \\
      addresses-per-mac=2 idle-timeout=5m keepalive-timeout=2m disabled=no
} else={
  set [find where name="hs-billing"] interface=$hsInterface address-pool=hs-pool \\
      profile=hsprof-billing disabled=no
}

# --- Walled Garden (akses sebelum login) -------------------------------------
:put "[${s(7)}] Walled garden (storefront + Midtrans)..."
# Bersihkan rule lama (kalau script dijalankan ulang) supaya tidak menumpuk.
# Termasuk sisa rule payment-gateway lama (billing-wg-auto / range CDN) bila ada.
/ip hotspot walled-garden remove [find where comment="billing-wg"]
/ip hotspot walled-garden ip remove [find where comment="billing-wg"]
/ip hotspot walled-garden ip remove [find where comment="billing-wg-auto"]
/ip hotspot walled-garden ip remove [find where comment="billing-wg-dns"]
/system scheduler remove [find where name="billing-resolve"]
/system script remove [find where name="billing-resolve"]

# (1) Server storefront — supaya pelanggan bisa lihat paket & terima voucher.
${
  feIsIP
    ? `/ip hotspot walled-garden ip add dst-address=${fe} action=accept comment="billing-wg"`
    : `/ip hotspot walled-garden add dst-host="${fe}" action=allow comment="billing-wg"`
}

# (2) Range IP gateway pembayaran (Midtrans) — agar halaman Snap bisa dibuka &
# dibayar langsung dari hotspot tanpa blank/connection-closed. Pakai IP karena
# rule hostname gagal di TLS modern. Terverifikasi: Snap full + portal tetap
# muncul. JANGAN tambah range Google di sini (memecah deteksi captive portal).
# Catatan: sebagian channel e-wallet/bank lanjutan bisa redirect ke domain lain
# di luar range ini — pelanggan tetap bisa pakai kuota HP untuk channel itu.
${PAYMENT_IP_RANGES.map(
  (r) =>
    `/ip hotspot walled-garden ip add dst-address=${r} action=accept comment="billing-wg"`,
).join("\n")}

# --- Halaman login custom (ambil otomatis dari server) -----------------------
# Hotspot setup di atas membuat folder flash/hotspot berisi file default
# (termasuk md5.js). Kita unduh halaman custom ke file SEMENTARA dulu, dan baru
# menimpa login.html bila unduhan BERHASIL — supaya bila server tak terjangkau,
# login.html bawaan tetap utuh (tidak jadi 404).
:put "[${s(8)}] Memasang halaman login custom..."
:do {
  /tool fetch url="${loginFetchUrl}" mode=http dst-path=flash/hotspot/login.custom
  :delay 1s
  :if ([:len [/file find where name="flash/hotspot/login.custom"]] > 0) do={
    /file remove [find where name="flash/hotspot/login.html"]
    /file set [find where name="flash/hotspot/login.custom"] name="flash/hotspot/login.html"
    :put "  login.html custom terpasang."
  }
} on-error={
  :put "  GAGAL mengunduh login.html dari ${backendUrlFromParams(p)}."
  :put "  login.html bawaan dibiarkan utuh. Setelah server siap, ulangi:"
  :put "  /tool fetch url=\\"${loginFetchUrl}\\" mode=http dst-path=flash/hotspot/login.html"
}

:put "Selesai! Hotspot Billing siap."
:put ("  WAN    : " . $wanInterface . "  (sumber internet)")
:put ("  Hotspot: " . $hsInterface . "  Net : " . $hsNetwork)
:put ("  RADIUS : " . $radiusIP . "  CoA : " . $radiusCoAPort)
:put "Login hotspot pakai KODE VOUCHER (username = password)."
`;
}

/** Build the teardown .rsc script text (removes all billing-* objects). */
export function generateMikrotikTeardown(): string {
  return `# =============================================================================
#  Billing Hotspot — Mikrotik teardown (hapus semua objek billing-*)
# =============================================================================
:put "[*] Menghapus konfigurasi Billing Hotspot..."
/ip hotspot remove [find where name="hs-billing"]
/ip hotspot profile remove [find where name="hsprof-billing"]
/ip hotspot walled-garden remove [find where comment="billing-wg"]
/ip hotspot walled-garden ip remove [find where comment="billing-wg"]
/ip hotspot walled-garden ip remove [find where comment="billing-wg-auto"]
/ip hotspot walled-garden ip remove [find where comment="billing-wg-dns"]
/system scheduler remove [find where name="billing-resolve"]
/system script remove [find where name="billing-resolve"]
/ip dhcp-server remove [find where name="hs-dhcp"]
/ip dhcp-server network remove [find where comment="billing-hotspot"]
/ip address remove [find where comment="hotspot-gw"]
/ip pool remove [find where name="hs-pool"]
/ip firewall nat remove [find where comment="billing-nat"]
/radius remove [find where comment="billing-hotspot"]
/radius incoming set accept=no
# Lepas port dari bridge-hotspot lalu hapus bridge-nya (abaikan bila tak ada).
:do {
  /interface bridge port remove [find where bridge=bridge-hotspot]
  /interface bridge remove [find where name=bridge-hotspot]
} on-error={}
:put "[*] Selesai."
`;
}

/**
 * Build the Mikrotik hotspot `login.html` captive portal page — shown the
 * moment a customer connects to the WiFi. It guides them: a clear "Beli Paket"
 * call-to-action that opens the storefront (already in the walled garden), plus
 * a voucher-code form for customers who already bought one. The form posts to
 * the Mikrotik `$(link-login-only)` endpoint with the voucher code as both
 * username and password (the project's single-code login style).
 *
 * Upload the file as `login.html` into the router's hotspot directory (Files →
 * `hotspot/`). `storeUrl` is the storefront URL reachable from the walled
 * garden, e.g. http://billing.local:8088
 */
export function generateLoginHtml(storeUrl: string): string {
  return `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>WiFi Hotspot — Login</title>
  $(if chap-id)
  <script type="text/javascript" src="/md5.js"></script>
  $(endif)
  <style>
    * { box-sizing: border-box; }
    body { margin:0; min-height:100vh; font-family: system-ui,-apple-system,Segoe UI,Roboto,sans-serif;
      background: linear-gradient(160deg,#1e3a8a 0%,#2563eb 55%,#0ea5e9 100%);
      display:flex; align-items:center; justify-content:center; padding:20px; color:#0f172a; }
    .card { background:#fff; width:100%; max-width:380px; border-radius:18px; padding:28px 24px;
      box-shadow:0 20px 50px rgba(0,0,0,.25); }
    .logo { width:56px; height:56px; border-radius:14px; background:#2563eb; color:#fff;
      display:flex; align-items:center; justify-content:center; margin:0 auto 14px; font-size:26px; }
    h1 { font-size:20px; text-align:center; margin:0 0 4px; }
    p.sub { text-align:center; color:#64748b; font-size:14px; margin:0 0 22px; }
    .cta { display:block; text-align:center; text-decoration:none; font-weight:700;
      background:#f59e0b; color:#fff; padding:14px; border-radius:12px; font-size:16px; }
    .cta small { display:block; font-weight:400; opacity:.9; font-size:12px; margin-top:2px; }
    .divider { display:flex; align-items:center; gap:10px; color:#94a3b8; font-size:12px; margin:20px 0; }
    .divider::before,.divider::after { content:""; flex:1; height:1px; background:#e2e8f0; }
    label { font-size:13px; font-weight:600; color:#334155; }
    input { width:100%; padding:13px 14px; margin:6px 0 0; border:1px solid #cbd5e1; border-radius:12px;
      font-size:16px; text-transform:uppercase; letter-spacing:1px; text-align:center; }
    button { width:100%; margin-top:14px; padding:14px; border:0; border-radius:12px; cursor:pointer;
      background:#2563eb; color:#fff; font-size:16px; font-weight:700; }
    .err { background:#fee2e2; color:#b91c1c; border-radius:10px; padding:10px 12px; font-size:13px;
      margin-bottom:14px; text-align:center; }
    .foot { text-align:center; color:#94a3b8; font-size:12px; margin-top:18px; }
  </style>
</head>
<body>
  <div class="card">
    <div class="logo">📶</div>
    <h1>Selamat Datang</h1>
    <p class="sub">Pilih paket internet, lalu masukkan kode voucher untuk online.</p>

    $(if error)<div class="err">$(error)</div>$(endif)

    <!-- Langkah 1: beli paket (storefront ada di walled garden) -->
    <a class="cta" href="${storeUrl}" target="_blank" rel="noopener">
      🛒 Beli Paket Internet
      <small>Belum punya kode? Pilih &amp; bayar paket di sini</small>
    </a>

    <div class="divider">SUDAH PUNYA KODE?</div>

    <!-- Langkah 2: login pakai kode voucher (username = password = kode) -->
    <form name="login" action="$(link-login-only)" method="post"
          $(if chap-id)onsubmit="return doLogin()"$(endif)>
      <input type="hidden" name="dst" value="$(link-orig)" />
      <input type="hidden" name="popup" value="true" />
      <label for="code">Kode Voucher</label>
      <input id="code" name="username" type="text" autocomplete="off"
             autocapitalize="characters" placeholder="MASUKKAN KODE" required />
      <input name="password" type="hidden" />
      <button type="submit">Hubungkan</button>
    </form>

    <p class="foot">Kode voucher otomatis menjadi username &amp; password.</p>
  </div>

  $(if chap-id)
  <script type="text/javascript">
    function doLogin() {
      var f = document.login;
      f.password.value = hexMD5('$(chap-id)' + f.username.value + '$(chap-challenge)');
      return true;
    }
  </script>
  $(else)
  <script type="text/javascript">
    document.login.onsubmit = function () {
      document.login.password.value = document.login.username.value;
      return true;
    };
  </script>
  $(endif)
</body>
</html>
`;
}
