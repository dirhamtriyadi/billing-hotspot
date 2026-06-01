// Captures the hotspot gateway passed by the Mikrotik captive portal.
//
// The portal's "Beli Paket" link appends ?gw=<router-gateway-host> (derived from
// Mikrotik's $(link-login-only)). We persist it on first load — for ANY entry
// route — so the "Hubungkan Sekarang" one-tap login on the voucher page uses the
// gateway of the actual NAS the customer is on (multi-NAS safe), instead of a
// build-time hardcoded IP. Import for side effect once at app start.
try {
  const gw = new URLSearchParams(window.location.search).get("gw");
  if (gw) sessionStorage.setItem("hotspot_gw", gw);
} catch {
  /* sessionStorage/URL unavailable — fall back to the build default downstream */
}

/** The hotspot gateway to use for one-tap login (stored gw → build default). */
export function hotspotGateway(): string {
  const fallback = import.meta.env.VITE_HOTSPOT_GATEWAY || "10.5.50.1";
  try {
    return sessionStorage.getItem("hotspot_gw") || fallback;
  } catch {
    return fallback;
  }
}
