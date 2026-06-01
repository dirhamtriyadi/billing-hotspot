import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { BadgeCheck, CheckCircle2, Copy, Link2, Search } from "lucide-react";
import { api } from "@/lib/api";
import { formatDateTime, formatIDR } from "@/lib/format";
import { errorMessage } from "@/lib/form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { Pagination } from "@/components/Pagination";
import { OrderStatusBadge } from "@/components/StatusBadge";

const STATUSES = ["", "pending", "paid", "failed", "expired", "cancelled"];

export default function OrdersPage() {
  const qc = useQueryClient();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["orders", page, search, status],
    queryFn: () =>
      api.orders.list({
        page,
        per_page: 20,
        search: search || undefined,
        status: status || undefined,
      }),
  });

  const confirmCash = useMutation({
    mutationFn: (id: number) => api.orders.confirmCash(id),
    onSuccess: (order) => {
      toast.success(
        `Pembayaran dikonfirmasi. Voucher: ${order.voucher?.code ?? "-"}`,
      );
      qc.invalidateQueries({ queryKey: ["orders"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  const markPaid = useMutation({
    mutationFn: (id: number) => api.orders.markPaid(id),
    onSuccess: (order) => {
      toast.success(
        `Pesanan dilunaskan. Voucher: ${order.voucher?.code ?? "-"}`,
      );
      qc.invalidateQueries({ queryKey: ["orders"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  // Salin link pembayaran (untuk dikirim ke pelanggan jika halaman tertutup)
  // atau, bila tak ada payment_url, salin link halaman status pesanan.
  const copyPayLink = (o: { payment_url?: string; order_number: string }) => {
    const link =
      o.payment_url && o.payment_url.length > 0
        ? o.payment_url
        : `${window.location.origin}/payment/${o.order_number}`;
    navigator.clipboard?.writeText(link).then(
      () => toast.success("Link pembayaran disalin"),
      () => toast.error("Gagal menyalin"),
    );
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Pesanan</h1>
        <p className="text-muted-foreground">
          Pantau pesanan dan konfirmasi pembayaran tunai.
        </p>
      </div>

      <Card>
        <CardContent className="space-y-4 p-4">
          <div className="flex flex-col gap-3 sm:flex-row">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="pl-9"
                placeholder="Cari no. pesanan / nama / HP…"
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
              />
            </div>
            <Select
              value={status || "all"}
              onValueChange={(v) => {
                setStatus(v === "all" ? "" : v);
                setPage(1);
              }}
            >
              <SelectTrigger className="sm:w-44">
                <SelectValue placeholder="Semua status" />
              </SelectTrigger>
              <SelectContent>
                {STATUSES.map((s) => (
                  <SelectItem key={s || "all"} value={s || "all"}>
                    {s === "" ? "Semua status" : s}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {isLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>No. Pesanan</TableHead>
                    <TableHead>Pelanggan</TableHead>
                    <TableHead>Paket</TableHead>
                    <TableHead>Metode</TableHead>
                    <TableHead>Total</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Voucher</TableHead>
                    <TableHead className="text-right">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.data.map((o) => (
                    <TableRow key={o.id}>
                      <TableCell>
                        <div className="font-mono text-xs">{o.order_number}</div>
                        <div className="text-xs text-muted-foreground">
                          {formatDateTime(o.created_at)}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="font-medium">{o.customer_name}</div>
                        <div className="text-xs text-muted-foreground">
                          {o.customer_phone}
                        </div>
                      </TableCell>
                      <TableCell>{o.package?.name ?? "-"}</TableCell>
                      <TableCell>
                        <Badge variant="outline" className="capitalize">
                          {o.payment_method}
                        </Badge>
                      </TableCell>
                      <TableCell>{formatIDR(o.amount)}</TableCell>
                      <TableCell>
                        <OrderStatusBadge status={o.status} />
                      </TableCell>
                      <TableCell className="font-mono">
                        {o.voucher?.code ?? "-"}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          {/* Salin link bayar/status — untuk order non-cash yg
                              belum lunas, mis. halaman pembayaran tertutup. */}
                          {o.payment_method !== "cash" &&
                            o.status !== "paid" && (
                              <Button
                                size="sm"
                                variant="outline"
                                title="Salin link pembayaran"
                                onClick={() => copyPayLink(o)}
                              >
                                {o.payment_url ? (
                                  <Link2 className="h-4 w-4" />
                                ) : (
                                  <Copy className="h-4 w-4" />
                                )}
                                <span className="hidden sm:inline">Link</span>
                              </Button>
                            )}

                          {/* Konfirmasi tunai (order cash pending). */}
                          {o.payment_method === "cash" &&
                            o.status === "pending" && (
                              <Button
                                size="sm"
                                onClick={() => confirmCash.mutate(o.id)}
                                disabled={confirmCash.isPending}
                              >
                                <BadgeCheck className="h-4 w-4" /> Konfirmasi
                              </Button>
                            )}

                          {/* Lunaskan manual (order gateway yg webhook-nya gagal
                              / PG error): pending, failed, atau expired. */}
                          {o.payment_method !== "cash" &&
                            (o.status === "pending" ||
                              o.status === "failed" ||
                              o.status === "expired") && (
                              <Button
                                size="sm"
                                variant="secondary"
                                title="Tandai lunas manual (sudah verifikasi pembayaran di dashboard PG)"
                                onClick={() => {
                                  if (
                                    confirm(
                                      `Lunaskan manual pesanan ${o.order_number}?\nPastikan pembayaran benar-benar diterima di dashboard payment gateway.`,
                                    )
                                  )
                                    markPaid.mutate(o.id);
                                }}
                                disabled={markPaid.isPending}
                              >
                                <CheckCircle2 className="h-4 w-4" /> Lunaskan
                              </Button>
                            )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                  {data?.data.length === 0 && (
                    <TableRow>
                      <TableCell
                        colSpan={8}
                        className="py-8 text-center text-muted-foreground"
                      >
                        Tidak ada pesanan.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
              <Pagination meta={data?.meta} onPageChange={setPage} />
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
