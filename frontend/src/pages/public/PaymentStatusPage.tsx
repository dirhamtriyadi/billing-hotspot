import { Link, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { useState } from "react";
import {
  CheckCircle2,
  Clock,
  Copy,
  Loader2,
  LogIn,
  Wifi,
  XCircle,
} from "lucide-react";
import { api } from "@/lib/api";
import { formatIDR } from "@/lib/format";
import { hotspotGateway } from "@/lib/hotspot";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";

export default function PaymentStatusPage() {
  const { orderNumber } = useParams<{ orderNumber: string }>();

  const { data: order, isLoading } = useQuery({
    queryKey: ["order", orderNumber],
    queryFn: () => api.public.orderStatus(orderNumber!),
    enabled: !!orderNumber,
    refetchInterval: (query) =>
      query.state.data?.status === "pending" ? 5000 : false,
  });

  const copy = (text: string) => {
    navigator.clipboard?.writeText(text).then(
      () => toast.success("Kode disalin"),
      () => toast.error("Gagal menyalin"),
    );
  };

  if (isLoading || !order) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-md">
        <CardContent className="p-8 text-center">
          {order.status === "paid" && order.voucher ? (
            <>
              <CheckCircle2 className="mx-auto mb-4 h-16 w-16 text-emerald-500" />
              <h1 className="text-2xl font-bold">Pembayaran Berhasil!</h1>
              <p className="mt-1 text-muted-foreground">
                Voucher untuk{" "}
                <strong>{order.package?.name ?? "paket"}</strong> sudah aktif.
              </p>

              <div className="my-6 rounded-2xl border-2 border-dashed border-primary/40 bg-primary/5 p-6">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">
                  Kode Voucher
                </p>
                <div className="mt-1 flex items-center justify-center gap-3">
                  <span className="font-mono text-3xl font-extrabold tracking-widest">
                    {order.voucher.code}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => copy(order.voucher!.code)}
                  >
                    <Copy className="h-5 w-5" />
                  </Button>
                </div>
              </div>

              <ConnectNow code={order.voucher.code} />

              <div className="mt-4 rounded-lg bg-muted p-4 text-left text-sm">
                <p className="mb-2 flex items-center gap-2 font-semibold">
                  <Wifi className="h-4 w-4" /> Kalau tombol di atas tidak jalan
                </p>
                <ol className="list-decimal space-y-1 pl-5 text-muted-foreground">
                  <li>Pastikan perangkat terhubung ke WiFi hotspot.</li>
                  <li>Buka halaman login (otomatis muncul / buka browser).</li>
                  <li>
                    Masukkan kode <strong>{order.voucher.code}</strong> lalu
                    klik Hubungkan.
                  </li>
                </ol>
              </div>
            </>
          ) : order.status === "pending" ? (
            <>
              <Clock className="mx-auto mb-4 h-16 w-16 text-amber-500" />
              <h1 className="text-2xl font-bold">
                {order.payment_method === "cash"
                  ? "Menunggu Konfirmasi"
                  : "Menunggu Pembayaran"}
              </h1>
              {order.payment_method === "cash" ? (
                <p className="mt-2 text-muted-foreground">
                  Tunjukkan nomor pesanan berikut ke operator untuk membayar
                  tunai. Voucher akan aktif setelah dikonfirmasi.
                </p>
              ) : (
                <p className="mt-2 text-muted-foreground">
                  Selesaikan pembayaran kamu. Halaman ini akan diperbarui
                  otomatis.
                </p>
              )}

              <div className="my-5 rounded-lg bg-muted p-4">
                <p className="text-xs uppercase text-muted-foreground">
                  Nomor Pesanan
                </p>
                <p className="font-mono text-lg font-bold">
                  {order.order_number}
                </p>
              </div>

              {order.payment_url && (
                <Button className="w-full" size="lg" asChild>
                  <a href={order.payment_url}>Lanjutkan Pembayaran</a>
                </Button>
              )}
              <div className="mt-3 flex items-center justify-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" /> Memeriksa status…
              </div>
            </>
          ) : (
            <>
              <XCircle className="mx-auto mb-4 h-16 w-16 text-destructive" />
              <h1 className="text-2xl font-bold">Pembayaran Belum Selesai</h1>
              <p className="mt-2 text-muted-foreground">
                Status pesanan: <strong>{order.status}</strong>. Silakan coba
                pesan kembali.
              </p>
              <Button className="mt-6 w-full" asChild>
                <Link to="/">Pesan Paket Lagi</Link>
              </Button>
            </>
          )}

          <div className="mt-6 border-t pt-4 text-sm text-muted-foreground">
            Total dibayar: <strong>{formatIDR(order.amount)}</strong>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * One-tap connect for customers still on the hotspot WiFi. Mikrotik's hotspot
 * accepts a login at http://<gateway>/login with username=password=voucher code
 * (single-code login, login-by=http-pap). The gateway comes from the router
 * itself via hotspotGateway() (captured from the captive portal's ?gw=), so it
 * is correct per-NAS, not hardcoded.
 *
 * If the user did not enter through the captive portal, the gateway is unknown
 * and the manual steps below remain the fallback.
 */
function ConnectNow({ code }: { code: string }) {
  const [submitted, setSubmitted] = useState(false);
  const gw = hotspotGateway();

  const connect = () => {
    if (!gw) return;
    setSubmitted(true);
    // Mikrotik accepts GET to /login?username=&password= when login-by=http-pap.
    // dst sends the user to the storefront root after a successful login.
    const url =
      `http://${gw}/login` +
      `?username=${encodeURIComponent(code)}` +
      `&password=${encodeURIComponent(code)}` +
      `&dst=${encodeURIComponent(window.location.origin)}`;
    window.location.href = url;
  };

  if (!gw) {
    return (
      <p className="my-4 rounded-md border border-dashed p-3 text-sm text-muted-foreground">
        Hubungkan otomatis tersedia saat halaman dibuka dari portal hotspot.
        Gunakan kode voucher di halaman login WiFi.
      </p>
    );
  }

  return (
    <div className="my-4">
      <Button size="lg" className="w-full gap-2" onClick={connect}>
        <LogIn className="h-5 w-5" />
        Hubungkan Sekarang
      </Button>
      {submitted && (
        <p className="mt-2 text-xs text-muted-foreground">
          Jika tidak otomatis online, pakai cara manual di bawah.
        </p>
      )}
    </div>
  );
}
