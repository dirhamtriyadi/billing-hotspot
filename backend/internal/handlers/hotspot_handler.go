package handlers

import (
	"net/http"
	"strings"

	"github.com/dirhamt/billing-hotspot/backend/internal/config"
	"github.com/gin-gonic/gin"
)

// HotspotHandler serves the Mikrotik hotspot captive-portal login page so a
// router can fetch it directly (via /tool fetch) into its flash/hotspot folder.
// The page guides a newly-connected customer: a "Beli Paket" call-to-action to
// the storefront plus a voucher-code login form.
type HotspotHandler struct {
	app config.AppConfig
}

// NewHotspotHandler builds a HotspotHandler.
func NewHotspotHandler(app config.AppConfig) *HotspotHandler {
	return &HotspotHandler{app: app}
}

// LoginPage godoc
// @Summary  Mikrotik hotspot login.html (captive portal)
// @Tags     Public
// @Produce  html
// @Param    store query string false "Storefront URL the 'Beli Paket' button opens"
// @Success  200 {string} string "HTML login page"
// @Router   /public/hotspot/login.html [get]
func (h *HotspotHandler) LoginPage(c *gin.Context) {
	store := strings.TrimSpace(c.Query("store"))
	if store == "" {
		store = h.app.FrontendURL
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, loginHTML(store))
}

// loginHTML renders the captive-portal page. Mikrotik template variables
// (e.g. $(link-login-only), $(if chap-id)) are emitted verbatim for the router
// to process; only storeURL is substituted server-side.
func loginHTML(storeURL string) string {
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

    <a class="cta" id="beli" href="` + storeURL + `" target="_blank" rel="noopener">
      🛒 Beli Paket Internet
      <small>Belum punya kode? Pilih &amp; bayar paket di sini</small>
    </a>
    <script type="text/javascript">
      // Sisipkan gateway hotspot ASLI router ini ke link storefront, supaya
      // tombol "Hubungkan Sekarang" di halaman voucher pakai gateway yang benar
      // (otomatis per-NAS, tidak di-hardcode). $(link-login-only) = URL login
      // lengkap dgn IP gateway router; kita ambil host-nya saja.
      (function () {
        var loginUrl = "$(link-login-only)";
        var m = loginUrl.match(/^https?:\/\/([^\/]+)/);
        var a = document.getElementById("beli");
        if (m && a) {
          a.href += (a.href.indexOf("?") < 0 ? "?" : "&") + "gw=" + encodeURIComponent(m[1]);
        }
      })();
    </script>

    <div class="divider">SUDAH PUNYA KODE?</div>

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
`
}
