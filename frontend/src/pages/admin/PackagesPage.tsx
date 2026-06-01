import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Loader2, Pencil, Plus, Trash2 } from "lucide-react";
import { api } from "@/lib/api";
import { formatIDR } from "@/lib/format";
import { applyApiErrors, errorMessage } from "@/lib/form";
import { packageSchema, type PackageValues } from "@/schemas";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
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
import type { Package } from "@/types";

const defaults: PackageValues = {
  name: "",
  description: "",
  price: 5000,
  rate_down_kbps: 10000,
  rate_up_kbps: 3000,
  burst_enabled: true,
  validity_value: 1,
  validity_unit: "day",
  session_timeout_secs: 0,
  data_quota_mb: 0,
  simultaneous_use: 1,
  highlight: false,
  badge_text: "",
  color: "#2563eb",
  icon: "wifi",
  sort_order: 0,
  is_active: true,
};

export default function PackagesPage() {
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<Package | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["packages"],
    queryFn: () => api.packages.list({ per_page: 100 }),
  });

  const form = useForm<PackageValues>({
    resolver: zodResolver(packageSchema),
    defaultValues: defaults,
  });

  const openCreate = () => {
    setEditing(null);
    form.reset(defaults);
    setOpen(true);
  };

  const openEdit = (p: Package) => {
    setEditing(p);
    form.reset({
      name: p.name,
      description: p.description,
      price: p.price,
      rate_down_kbps: p.rate_down_kbps,
      rate_up_kbps: p.rate_up_kbps,
      burst_enabled: p.burst_enabled,
      validity_value: p.validity_value,
      validity_unit: p.validity_unit,
      session_timeout_secs: p.session_timeout_secs,
      data_quota_mb: p.data_quota_mb,
      simultaneous_use: p.simultaneous_use,
      highlight: p.highlight,
      badge_text: p.badge_text,
      color: p.color,
      icon: p.icon,
      sort_order: p.sort_order,
      is_active: p.is_active,
    });
    setOpen(true);
  };

  const save = useMutation({
    mutationFn: (values: PackageValues) =>
      editing
        ? api.packages.update(editing.id, values)
        : api.packages.create(values),
    onSuccess: () => {
      toast.success(editing ? "Paket diperbarui" : "Paket dibuat");
      qc.invalidateQueries({ queryKey: ["packages"] });
      setOpen(false);
    },
    onError: (e) => {
      if (!applyApiErrors(e, form.setError)) toast.error(errorMessage(e));
    },
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.packages.remove(id),
    onSuccess: () => {
      toast.success("Paket dihapus");
      qc.invalidateQueries({ queryKey: ["packages"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Paket</h1>
          <p className="text-muted-foreground">
            Kelola paket internet yang dijual.
          </p>
        </div>
        <Button onClick={openCreate}>
          <Plus className="h-4 w-4" /> Tambah Paket
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nama</TableHead>
                  <TableHead>Harga</TableHead>
                  <TableHead>Kecepatan</TableHead>
                  <TableHead>Masa Aktif</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.data.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell>
                      <div className="font-medium">{p.name}</div>
                      <div className="text-xs text-muted-foreground">
                        {p.slug}
                      </div>
                    </TableCell>
                    <TableCell>{formatIDR(p.price)}</TableCell>
                    <TableCell>
                      {(p.rate_down_kbps / 1000).toFixed(0)}/
                      {(p.rate_up_kbps / 1000).toFixed(0)} Mbps
                    </TableCell>
                    <TableCell>
                      {p.validity_value} {p.validity_unit}
                    </TableCell>
                    <TableCell>
                      {p.is_active ? (
                        <Badge variant="success">Aktif</Badge>
                      ) : (
                        <Badge variant="secondary">Nonaktif</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => openEdit(p)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => {
                          if (confirm(`Hapus paket "${p.name}"?`))
                            remove.mutate(p.id);
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
                      Belum ada paket.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>{editing ? "Edit Paket" : "Tambah Paket"}</DialogTitle>
          </DialogHeader>
          <form
            onSubmit={form.handleSubmit((v) => save.mutate(v))}
            className="space-y-4"
            noValidate
          >
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="Nama Paket" error={form.formState.errors.name?.message}>
                <Input {...form.register("name")} />
              </Field>
              <Field label="Harga (IDR)" error={form.formState.errors.price?.message}>
                <Input type="number" {...form.register("price")} />
              </Field>
            </div>

            <Field
              label="Deskripsi"
              error={form.formState.errors.description?.message}
            >
              <Textarea rows={2} {...form.register("description")} />
            </Field>

            <div className="grid gap-4 sm:grid-cols-2">
              <Field
                label="Download (kbps)"
                error={form.formState.errors.rate_down_kbps?.message}
              >
                <Input type="number" {...form.register("rate_down_kbps")} />
              </Field>
              <Field
                label="Upload (kbps)"
                error={form.formState.errors.rate_up_kbps?.message}
              >
                <Input type="number" {...form.register("rate_up_kbps")} />
              </Field>
            </div>

            <div className="grid gap-4 sm:grid-cols-3">
              <Field
                label="Masa Aktif"
                error={form.formState.errors.validity_value?.message}
              >
                <Input type="number" {...form.register("validity_value")} />
              </Field>
              <Field label="Satuan">
                <Select
                  value={form.watch("validity_unit")}
                  onValueChange={(v) =>
                    form.setValue("validity_unit", v as PackageValues["validity_unit"])
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="minute">Menit</SelectItem>
                    <SelectItem value="hour">Jam</SelectItem>
                    <SelectItem value="day">Hari</SelectItem>
                    <SelectItem value="month">Bulan</SelectItem>
                  </SelectContent>
                </Select>
              </Field>
              <Field
                label="Login Bersamaan"
                error={form.formState.errors.simultaneous_use?.message}
              >
                <Input type="number" {...form.register("simultaneous_use")} />
              </Field>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
              <Field
                label="Kuota Data (MB, 0 = unlimited)"
                error={form.formState.errors.data_quota_mb?.message}
              >
                <Input type="number" {...form.register("data_quota_mb")} />
              </Field>
              <Field
                label="Session Timeout (detik, 0 = unlimited)"
                error={form.formState.errors.session_timeout_secs?.message}
              >
                <Input type="number" {...form.register("session_timeout_secs")} />
              </Field>
            </div>

            <div className="grid gap-4 sm:grid-cols-3">
              <Field label="Ikon">
                <Select
                  value={form.watch("icon")}
                  onValueChange={(v) => form.setValue("icon", v)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {["wifi", "zap", "sun", "rocket", "calendar", "crown", "gauge", "star"].map(
                      (ic) => (
                        <SelectItem key={ic} value={ic}>
                          {ic}
                        </SelectItem>
                      ),
                    )}
                  </SelectContent>
                </Select>
              </Field>
              <Field label="Warna">
                <Input type="color" className="h-10 p-1" {...form.register("color")} />
              </Field>
              <Field label="Urutan" error={form.formState.errors.sort_order?.message}>
                <Input type="number" {...form.register("sort_order")} />
              </Field>
            </div>

            <Field label="Teks Badge (mis. TERLARIS)">
              <Input {...form.register("badge_text")} />
            </Field>

            <div className="flex items-center gap-8">
              <label className="flex items-center gap-2 text-sm">
                <Switch
                  checked={form.watch("burst_enabled")}
                  onCheckedChange={(c) => form.setValue("burst_enabled", c)}
                />
                Burst
              </label>
              <label className="flex items-center gap-2 text-sm">
                <Switch
                  checked={form.watch("highlight")}
                  onCheckedChange={(c) => form.setValue("highlight", c)}
                />
                Unggulan
              </label>
              <label className="flex items-center gap-2 text-sm">
                <Switch
                  checked={form.watch("is_active")}
                  onCheckedChange={(c) => form.setValue("is_active", c)}
                />
                Aktif
              </label>
            </div>

            <div className="flex justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => setOpen(false)}
              >
                Batal
              </Button>
              <Button type="submit" disabled={save.isPending}>
                {save.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Simpan
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function Field({
  label,
  error,
  children,
}: {
  label: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      {children}
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  );
}
