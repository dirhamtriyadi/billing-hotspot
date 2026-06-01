import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Ban, CheckCircle, Copy, Search, Trash2 } from "lucide-react";
import { api } from "@/lib/api";
import { formatDate, formatIDR } from "@/lib/format";
import { errorMessage } from "@/lib/form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { VoucherStatusBadge } from "@/components/StatusBadge";

const STATUSES = ["", "unused", "active", "used", "expired", "disabled"];

export default function VouchersPage() {
  const qc = useQueryClient();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["vouchers", page, search, status],
    queryFn: () =>
      api.vouchers.list({
        page,
        per_page: 20,
        search: search || undefined,
        status: status || undefined,
      }),
  });

  const setStatusMut = useMutation({
    mutationFn: ({ id, s }: { id: number; s: "active" | "disabled" }) =>
      api.vouchers.setStatus(id, s),
    onSuccess: () => {
      toast.success("Status voucher diperbarui");
      qc.invalidateQueries({ queryKey: ["vouchers"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.vouchers.remove(id),
    onSuccess: () => {
      toast.success("Voucher dihapus");
      qc.invalidateQueries({ queryKey: ["vouchers"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  const copy = (code: string) => {
    navigator.clipboard?.writeText(code);
    toast.success(`Kode ${code} disalin`);
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Voucher</h1>
        <p className="text-muted-foreground">
          Daftar semua kode voucher hotspot.
        </p>
      </div>

      <Card>
        <CardContent className="space-y-4 p-4">
          <div className="flex flex-col gap-3 sm:flex-row">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                className="pl-9"
                placeholder="Cari kode voucher…"
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
              <SelectTrigger className="sm:w-48">
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
                    <TableHead>Kode</TableHead>
                    <TableHead>Paket</TableHead>
                    <TableHead>Harga</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Kadaluarsa</TableHead>
                    <TableHead className="text-right">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.data.map((v) => (
                    <TableRow key={v.id}>
                      <TableCell>
                        <button
                          className="inline-flex items-center gap-2 font-mono font-semibold hover:text-primary"
                          onClick={() => copy(v.code)}
                        >
                          {v.code}
                          <Copy className="h-3.5 w-3.5 opacity-50" />
                        </button>
                      </TableCell>
                      <TableCell>{v.package?.name ?? "-"}</TableCell>
                      <TableCell>{formatIDR(v.price)}</TableCell>
                      <TableCell>
                        <VoucherStatusBadge status={v.status} />
                      </TableCell>
                      <TableCell className="text-muted-foreground">
                        {formatDate(v.expires_at)}
                      </TableCell>
                      <TableCell className="text-right">
                        {v.status === "disabled" ? (
                          <Button
                            variant="ghost"
                            size="icon"
                            title="Aktifkan"
                            onClick={() =>
                              setStatusMut.mutate({ id: v.id, s: "active" })
                            }
                          >
                            <CheckCircle className="h-4 w-4 text-emerald-500" />
                          </Button>
                        ) : (
                          <Button
                            variant="ghost"
                            size="icon"
                            title="Nonaktifkan"
                            onClick={() =>
                              setStatusMut.mutate({ id: v.id, s: "disabled" })
                            }
                          >
                            <Ban className="h-4 w-4 text-amber-500" />
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="icon"
                          title="Hapus"
                          onClick={() => {
                            if (confirm(`Hapus voucher ${v.code}?`))
                              remove.mutate(v.id);
                          }}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                  {data?.data.length === 0 && (
                    <TableRow>
                      <TableCell
                        colSpan={6}
                        className="py-8 text-center text-muted-foreground"
                      >
                        Tidak ada voucher.
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
