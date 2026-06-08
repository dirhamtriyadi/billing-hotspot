import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { api } from "@/lib/api";
import { applyApiErrors, errorMessage } from "@/lib/form";
import {
  radiusServerSchema,
  type RadiusServerValues,
} from "@/schemas";
import type { RadiusServer } from "@/types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Pencil, Plus, Server, Trash2 } from "lucide-react";
import type { ReactNode } from "react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

const defaults: RadiusServerValues = {
  name: "",
  api_url: "",
  api_key: "",
  radius_ip: "",
  coa_port: "3799",
  description: "",
  is_default: false,
};

export default function RadiusServersPage() {
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<RadiusServer | null>(null);

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["radius-servers"],
    queryFn: api.radiusServers.list,
  });

  const form = useForm<RadiusServerValues>({
    resolver: zodResolver(radiusServerSchema),
    defaultValues: defaults,
  });

  const openCreate = () => {
    setEditing(null);
    form.reset(defaults);
    setOpen(true);
  };

  const openEdit = (row: RadiusServer) => {
    setEditing(row);
    form.reset({
      name: row.name,
      api_url: row.api_url,
      api_key: row.api_key,
      radius_ip: row.radius_ip,
      coa_port: row.coa_port || "3799",
      description: row.description,
      is_default: row.is_default,
    });
    setOpen(true);
  };

  const save = useMutation({
    mutationFn: (values: RadiusServerValues) =>
      editing
        ? api.radiusServers.update(editing.id, values)
        : api.radiusServers.create(values),
    onSuccess: () => {
      toast.success(editing ? "Radius server diperbarui" : "Radius server ditambahkan");
      qc.invalidateQueries({ queryKey: ["radius-servers"] });
      setOpen(false);
    },
    onError: (e) => {
      if (!applyApiErrors(e, form.setError)) toast.error(errorMessage(e));
    },
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.radiusServers.remove(id),
    onSuccess: () => {
      toast.success("Radius server dihapus");
      qc.invalidateQueries({ queryKey: ["radius-servers"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Radius Server</h1>
          <p className="text-muted-foreground">
            Kelola endpoint radius-api per cabang untuk provisioning paket,
            voucher, NAS, dan script Mikrotik.
          </p>
        </div>
        <Button onClick={openCreate}>
          <Plus className="h-4 w-4" /> Tambah
        </Button>
      </div>

      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="space-y-2 p-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : isError ? (
            <div className="p-8 text-center text-muted-foreground">
              <Server className="mx-auto mb-2 h-8 w-8 opacity-50" />
              <p>Gagal memuat radius server.</p>
              <p className="text-sm">{errorMessage(error)}</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nama</TableHead>
                  <TableHead>Radius API URL</TableHead>
                  <TableHead>IP RADIUS</TableHead>
                  <TableHead>CoA</TableHead>
                  <TableHead>Default</TableHead>
                  <TableHead>Deskripsi</TableHead>
                  <TableHead className="text-right">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.map((row) => (
                  <TableRow key={row.id}>
                    <TableCell className="font-medium">{row.name}</TableCell>
                    <TableCell className="font-mono text-sm">
                      {row.api_url}
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {row.radius_ip || "-"}
                    </TableCell>
                    <TableCell>{row.coa_port || "3799"}</TableCell>
                    <TableCell>
                      {row.is_default ? (
                        <Badge>Default</Badge>
                      ) : (
                        <span className="text-muted-foreground">-</span>
                      )}
                    </TableCell>
                    <TableCell className="max-w-[220px] truncate text-muted-foreground">
                      {row.description || "-"}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        title="Edit"
                        onClick={() => openEdit(row)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        title="Hapus"
                        onClick={() => {
                          if (confirm(`Hapus radius server "${row.name}"?`))
                            remove.mutate(row.id);
                        }}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                {data?.length === 0 && (
                  <TableRow>
                    <TableCell
                      colSpan={7}
                      className="py-8 text-center text-muted-foreground"
                    >
                      Belum ada radius server. Tambahkan endpoint radius-api
                      cabang terlebih dahulu.
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
            <DialogTitle>
              {editing ? "Edit Radius Server" : "Tambah Radius Server"}
            </DialogTitle>
          </DialogHeader>
          <form
            onSubmit={form.handleSubmit((v) => save.mutate(v))}
            className="space-y-4"
            noValidate
          >
            <Field label="Nama" error={form.formState.errors.name?.message}>
              <Input placeholder="radius-bandung" {...form.register("name")} />
            </Field>
            <Field
              label="Radius API URL"
              error={form.formState.errors.api_url?.message}
            >
              <Input
                placeholder="https://radius-bandung.example.com"
                {...form.register("api_url")}
              />
            </Field>
            <Field
              label="Radius API Key"
              error={form.formState.errors.api_key?.message}
            >
              <Input {...form.register("api_key")} />
            </Field>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field
                label="IP Server RADIUS"
                error={form.formState.errors.radius_ip?.message}
              >
                <Input
                  placeholder="10.10.1.2"
                  {...form.register("radius_ip")}
                />
              </Field>
              <Field
                label="CoA Port"
                error={form.formState.errors.coa_port?.message}
              >
                <Input placeholder="3799" {...form.register("coa_port")} />
              </Field>
            </div>
            <Field
              label="Deskripsi"
              error={form.formState.errors.description?.message}
            >
              <Textarea rows={2} {...form.register("description")} />
            </Field>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                className="h-4 w-4"
                {...form.register("is_default")}
              />
              Jadikan default
            </label>

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
  children: ReactNode;
}) {
  return (
    <div className="space-y-1">
      <Label>{label}</Label>
      {children}
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}
