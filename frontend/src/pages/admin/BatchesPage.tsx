import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Eye, Loader2, Plus, Printer, Trash2 } from "lucide-react";
import { api } from "@/lib/api";
import { formatDateTime } from "@/lib/format";
import { applyApiErrors, errorMessage } from "@/lib/form";
import { batchSchema, type BatchValues } from "@/schemas";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import type { VoucherBatch } from "@/types";

export default function BatchesPage() {
  const qc = useQueryClient();
  const [page, setPage] = useState(1);
  const [open, setOpen] = useState(false);
  const [viewing, setViewing] = useState<VoucherBatch | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["batches", page],
    queryFn: () => api.batches.list({ page, per_page: 20 }),
  });

  const { data: packages } = useQuery({
    queryKey: ["packages-min"],
    queryFn: () => api.packages.list({ per_page: 100 }),
  });

  const form = useForm<BatchValues>({
    resolver: zodResolver(batchSchema),
    defaultValues: { name: "", quantity: 10, prefix: "", code_length: 8 },
  });

  const create = useMutation({
    mutationFn: (values: BatchValues) => api.batches.create(values),
    onSuccess: () => {
      toast.success("Batch voucher dibuat");
      qc.invalidateQueries({ queryKey: ["batches"] });
      qc.invalidateQueries({ queryKey: ["vouchers"] });
      setOpen(false);
    },
    onError: (e) => {
      if (!applyApiErrors(e, form.setError)) toast.error(errorMessage(e));
    },
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.batches.remove(id),
    onSuccess: () => {
      toast.success("Batch dihapus");
      qc.invalidateQueries({ queryKey: ["batches"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  const view = async (id: number) => {
    try {
      setViewing(await api.batches.get(id));
    } catch (e) {
      toast.error(errorMessage(e));
    }
  };

  const printBatch = (batch: VoucherBatch) => {
    const win = window.open("", "_blank");
    if (!win) return;
    const cards = (batch.vouchers ?? [])
      .map(
        (v) => `
        <div class="card">
          <div class="title">${batch.package?.name ?? "Voucher"}</div>
          <div class="code">${v.code}</div>
          <div class="muted">Username & Password sama dengan kode</div>
        </div>`,
      )
      .join("");
    win.document.write(`
      <html><head><title>Cetak Voucher — ${batch.name}</title>
      <style>
        body{font-family:Inter,system-ui,sans-serif;padding:16px}
        .grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px}
        .card{border:1px dashed #94a3b8;border-radius:12px;padding:14px;text-align:center}
        .title{font-size:12px;color:#475569}
        .code{font-size:22px;font-weight:800;letter-spacing:3px;margin:6px 0;font-family:monospace}
        .muted{font-size:10px;color:#94a3b8}
        @media print{.no-print{display:none}}
      </style></head><body>
      <button class="no-print" onclick="window.print()">Cetak</button>
      <h3>${batch.name}</h3>
      <div class="grid">${cards}</div>
      </body></html>`);
    win.document.close();
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Batch Voucher</h1>
          <p className="text-muted-foreground">
            Generate banyak voucher sekaligus untuk dijual tunai.
          </p>
        </div>
        <Button
          onClick={() => {
            form.reset({
              name: "",
              quantity: 10,
              prefix: "",
              code_length: 8,
              package_id: packages?.data[0]?.id,
            });
            setOpen(true);
          }}
        >
          <Plus className="h-4 w-4" /> Generate Voucher
        </Button>
      </div>

      <Card>
        <CardContent className="space-y-4 p-4">
          {isLoading ? (
            <div className="space-y-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Nama Batch</TableHead>
                    <TableHead>Paket</TableHead>
                    <TableHead>Jumlah</TableHead>
                    <TableHead>Dibuat</TableHead>
                    <TableHead className="text-right">Aksi</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {data?.data.map((b) => (
                    <TableRow key={b.id}>
                      <TableCell className="font-medium">{b.name}</TableCell>
                      <TableCell>{b.package?.name ?? "-"}</TableCell>
                      <TableCell>{b.quantity}</TableCell>
                      <TableCell className="text-muted-foreground">
                        {formatDateTime(b.created_at)}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          title="Lihat & cetak"
                          onClick={() => view(b.id)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          title="Hapus"
                          onClick={() => {
                            if (confirm(`Hapus batch "${b.name}" beserta vouchernya?`))
                              remove.mutate(b.id);
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
                        colSpan={5}
                        className="py-8 text-center text-muted-foreground"
                      >
                        Belum ada batch.
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

      {/* Create dialog */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Generate Voucher</DialogTitle>
          </DialogHeader>
          <form
            onSubmit={form.handleSubmit((v) => create.mutate(v))}
            className="space-y-4"
            noValidate
          >
            <div className="space-y-2">
              <Label>Paket</Label>
              <Select
                value={form.watch("package_id")?.toString()}
                onValueChange={(v) =>
                  form.setValue("package_id", Number(v), {
                    shouldValidate: true,
                  })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Pilih paket" />
                </SelectTrigger>
                <SelectContent>
                  {packages?.data.map((p) => (
                    <SelectItem key={p.id} value={p.id.toString()}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {form.formState.errors.package_id && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.package_id.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label>Nama Batch (opsional)</Label>
              <Input
                placeholder="mis. Voucher Warung Bu Sri"
                {...form.register("name")}
              />
            </div>

            <div className="grid grid-cols-3 gap-4">
              <div className="space-y-2">
                <Label>Jumlah</Label>
                <Input type="number" {...form.register("quantity")} />
                {form.formState.errors.quantity && (
                  <p className="text-sm text-destructive">
                    {form.formState.errors.quantity.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label>Prefix</Label>
                <Input placeholder="WIFI" {...form.register("prefix")} />
              </div>
              <div className="space-y-2">
                <Label>Panjang Kode</Label>
                <Input type="number" {...form.register("code_length")} />
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={create.isPending}>
                {create.isPending && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                Generate
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* View dialog */}
      <Dialog open={!!viewing} onOpenChange={(o) => !o && setViewing(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center justify-between gap-4">
              <span>{viewing?.name}</span>
              {viewing && (
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => printBatch(viewing)}
                >
                  <Printer className="h-4 w-4" /> Cetak
                </Button>
              )}
            </DialogTitle>
          </DialogHeader>
          <div className="grid max-h-[60vh] grid-cols-2 gap-2 overflow-y-auto sm:grid-cols-3">
            {viewing?.vouchers?.map((v) => (
              <div
                key={v.id}
                className="rounded-lg border border-dashed p-3 text-center"
              >
                <div className="font-mono text-lg font-bold tracking-wider">
                  {v.code}
                </div>
              </div>
            ))}
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
