import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
  generateLoginHtml,
  generateMikrotikScript,
  generateMikrotikTeardown,
  paramsFromNas,
  storeUrlFromParams,
  type MikrotikParams,
} from "@/lib/mikrotik";
import { nasSchema, type NasValues } from "@/schemas";
import type { NAS } from "@/types";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Download,
  FileCode,
  Loader2,
  Pencil,
  Plus,
  Router as RouterIcon,
  Trash2,
} from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

const defaults: NasValues = {
  nasname: "",
  shortname: "",
  secret: "",
  type: "mikrotik",
  description: "",
  ports: "",
  hotspot_config: {
    radius_api_url: "",
    radius_api_key: "",
    radius_ip: "",
    frontend_host: "",
    coa_port: "3799",
    wan_interface: "ether1",
    hotspot_interface: "bridge-hotspot",
    bridge_ports: "wlan1,wlan2",
    hotspot_network: "10.5.50.0/24",
    hotspot_gateway: "10.5.50.1",
    hotspot_pool_range: "10.5.50.10-10.5.50.254",
    hotspot_dns: "8.8.8.8,1.1.1.1",
  },
};

export default function NasPage() {
  const qc = useQueryClient();
  const [open, setOpen] = useState(false);
  const [editing, setEditing] = useState<NAS | null>(null);
  const [scriptFor, setScriptFor] = useState<NAS | null>(null);

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["nas"],
    queryFn: api.nas.list,
  });

  const form = useForm<NasValues>({
    resolver: zodResolver(nasSchema),
    defaultValues: defaults,
  });

  const openCreate = () => {
    setEditing(null);
    form.reset(defaults);
    setOpen(true);
  };

  const openEdit = (n: NAS) => {
    setEditing(n);
    form.reset({
      nasname: n.nasname,
      shortname: n.shortname,
      secret: n.secret,
      type: n.type || "mikrotik",
      description: n.description,
      ports: n.ports ?? "",
      hotspot_config: {
        ...defaults.hotspot_config,
        ...n.hotspot_config,
      },
    });
    setOpen(true);
  };

  const save = useMutation({
    mutationFn: (values: NasValues) =>
      api.nas.upsert({
        ...values,
        ports:
          values.ports === "" || values.ports == null
            ? undefined
            : Number(values.ports),
      }),
    onSuccess: () => {
      toast.success(editing ? "Router diperbarui" : "Router ditambahkan");
      qc.invalidateQueries({ queryKey: ["nas"] });
      setOpen(false);
    },
    onError: (e) => {
      if (!applyApiErrors(e, form.setError)) toast.error(errorMessage(e));
    },
  });

  const remove = useMutation({
    mutationFn: (id: number) => api.nas.remove(id),
    onSuccess: () => {
      toast.success("Router dihapus");
      qc.invalidateQueries({ queryKey: ["nas"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Router (NAS)</h1>
          <p className="text-muted-foreground">
            Daftarkan router/Mikrotik sebagai klien RADIUS, lalu generate script
            &amp; halaman login siap-pakai.
          </p>
        </div>
        <Button onClick={openCreate}>
          <Plus className="h-4 w-4" /> Tambah Router
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
              <RouterIcon className="mx-auto mb-2 h-8 w-8 opacity-50" />
              <p>Gagal memuat daftar router.</p>
              <p className="text-sm">{errorMessage(error)}</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>NAS Name (IP/Identitas)</TableHead>
                  <TableHead>Nama Pendek</TableHead>
                  <TableHead>Tipe</TableHead>
                  <TableHead>Gateway Hotspot</TableHead>
                  <TableHead>Secret</TableHead>
                  <TableHead>Deskripsi</TableHead>
                  <TableHead className="text-right">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.map((n) => (
                  <TableRow key={n.id}>
                    <TableCell className="font-mono text-sm">
                      {n.nasname}
                    </TableCell>
                    <TableCell>{n.shortname || "-"}</TableCell>
                    <TableCell>
                      <Badge variant="secondary">{n.type || "other"}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {n.hotspot_config?.hotspot_gateway || "-"}
                    </TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground">
                      ••••{n.secret.slice(-4)}
                    </TableCell>
                    <TableCell className="max-w-[200px] truncate text-muted-foreground">
                      {n.description || "-"}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        variant="ghost"
                        size="icon"
                        title="Generate script & login.html"
                        onClick={() => setScriptFor(n)}
                      >
                        <FileCode className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        title="Edit"
                        onClick={() => openEdit(n)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        title="Hapus"
                        onClick={() => {
                          if (confirm(`Hapus router "${n.nasname}"?`))
                            remove.mutate(n.id);
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
                      Belum ada router terdaftar.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Tentang NAS</CardTitle>
        </CardHeader>
        <CardContent className="space-y-1 text-sm text-muted-foreground">
          <p>
            <b>NAS Name</b> = alamat IP router sebagaimana dilihat server RADIUS
            (mis. <code>192.168.88.253</code>).
          </p>
          <p>
            <b>Secret</b> harus sama dengan yang dikonfigurasi di router. Pakai
            tombol <b>Generate</b> agar secret-nya otomatis cocok.
          </p>
          <p className="text-emerald-600">
            ✓ Menambah/mengubah/menghapus router otomatis tersimpan ke RADIUS
            dan langsung aktif — FreeRADIUS dimuat ulang otomatis, tanpa restart
            manual.
          </p>
        </CardContent>
      </Card>

      {/* Add / edit dialog */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editing ? "Edit Router" : "Tambah Router"}
            </DialogTitle>
          </DialogHeader>
          <form
            onSubmit={form.handleSubmit((v) => save.mutate(v))}
            className="space-y-4"
            noValidate
          >
            <Field
              label="NAS Name (IP / identitas)"
              error={form.formState.errors.nasname?.message}
            >
              <Input
                placeholder="192.168.88.253"
                {...form.register("nasname")}
                disabled={!!editing}
              />
            </Field>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field
                label="Nama Pendek"
                error={form.formState.errors.shortname?.message}
              >
                <Input
                  placeholder="mikrotik-utama"
                  {...form.register("shortname")}
                />
              </Field>
              <Field label="Tipe" error={form.formState.errors.type?.message}>
                <Input placeholder="mikrotik" {...form.register("type")} />
              </Field>
            </div>
            <Field
              label="RADIUS Secret"
              error={form.formState.errors.secret?.message}
            >
              <Input
                placeholder="secret-yang-sama-dengan-router"
                {...form.register("secret")}
              />
            </Field>
            <div className="grid gap-4 sm:grid-cols-2">
              <Field
                label="Ports (opsional)"
                error={form.formState.errors.ports?.message}
              >
                <Input
                  type="number"
                  placeholder="1812"
                  {...form.register("ports")}
                />
              </Field>
            </div>
            <Field
              label="Deskripsi"
              error={form.formState.errors.description?.message}
            >
              <Textarea rows={2} {...form.register("description")} />
            </Field>

            <div className="space-y-3 rounded-md border p-3">
              <div>
                <p className="text-sm font-medium">Konfigurasi Hotspot</p>
                <p className="text-xs text-muted-foreground">
                  Disimpan per router dan dipakai sebagai default saat generate
                  script Mikrotik.
                </p>
              </div>
              <div className="grid gap-3 sm:grid-cols-2">
                <Field
                  label="Radius API URL Cabang"
                  error={
                    form.formState.errors.hotspot_config?.radius_api_url
                      ?.message
                  }
                >
                  <Input
                    placeholder="https://radius-bandung.example.com / kosong = env default"
                    {...form.register("hotspot_config.radius_api_url")}
                  />
                </Field>
                <Field
                  label="Radius API Key Cabang"
                  error={
                    form.formState.errors.hotspot_config?.radius_api_key
                      ?.message
                  }
                >
                  <Input
                    placeholder="kosong = env RADIUS_API_KEY"
                    {...form.register("hotspot_config.radius_api_key")}
                  />
                </Field>
                <Field
                  label="IP Server RADIUS"
                  error={
                    form.formState.errors.hotspot_config?.radius_ip?.message
                  }
                >
                  <Input
                    placeholder="kosong = host backend saat ini"
                    {...form.register("hotspot_config.radius_ip")}
                  />
                </Field>
                <Field
                  label="Host Frontend/Backend"
                  error={
                    form.formState.errors.hotspot_config?.frontend_host?.message
                  }
                >
                  <Input
                    placeholder="kosong = host backend saat ini"
                    {...form.register("hotspot_config.frontend_host")}
                  />
                </Field>
                <Field
                  label="Gateway Hotspot"
                  error={
                    form.formState.errors.hotspot_config?.hotspot_gateway
                      ?.message
                  }
                >
                  <Input
                    placeholder="10.5.50.1"
                    {...form.register("hotspot_config.hotspot_gateway")}
                  />
                </Field>
                <Field
                  label="Subnet Hotspot"
                  error={
                    form.formState.errors.hotspot_config?.hotspot_network
                      ?.message
                  }
                >
                  <Input
                    placeholder="10.5.50.0/24"
                    {...form.register("hotspot_config.hotspot_network")}
                  />
                </Field>
                <Field
                  label="Range DHCP"
                  error={
                    form.formState.errors.hotspot_config?.hotspot_pool_range
                      ?.message
                  }
                >
                  <Input
                    placeholder="10.5.50.10-10.5.50.254"
                    {...form.register("hotspot_config.hotspot_pool_range")}
                  />
                </Field>
                <Field
                  label="DNS"
                  error={
                    form.formState.errors.hotspot_config?.hotspot_dns?.message
                  }
                >
                  <Input
                    placeholder="8.8.8.8,1.1.1.1"
                    {...form.register("hotspot_config.hotspot_dns")}
                  />
                </Field>
                <Field
                  label="WAN / Sumber Internet"
                  error={
                    form.formState.errors.hotspot_config?.wan_interface?.message
                  }
                >
                  <Input
                    placeholder="ether1"
                    {...form.register("hotspot_config.wan_interface")}
                  />
                </Field>
                <Field
                  label="CoA Port"
                  error={
                    form.formState.errors.hotspot_config?.coa_port?.message
                  }
                >
                  <Input
                    placeholder="3799"
                    {...form.register("hotspot_config.coa_port")}
                  />
                </Field>
                <Field
                  label="Interface Hotspot"
                  error={
                    form.formState.errors.hotspot_config?.hotspot_interface
                      ?.message
                  }
                >
                  <Input
                    placeholder="bridge-hotspot / bridge / wlan1"
                    {...form.register("hotspot_config.hotspot_interface")}
                  />
                </Field>
                <Field
                  label="Bridge Ports"
                  error={
                    form.formState.errors.hotspot_config?.bridge_ports?.message
                  }
                >
                  <Input
                    placeholder="wlan1,wlan2"
                    {...form.register("hotspot_config.bridge_ports")}
                  />
                </Field>
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
              <Button type="submit" disabled={save.isPending}>
                {save.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
                Simpan
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Script + login.html generator */}
      {scriptFor && (
        <ScriptDialog nas={scriptFor} onClose={() => setScriptFor(null)} />
      )}
    </div>
  );
}

function ScriptDialog({ nas, onClose }: { nas: NAS; onClose: () => void }) {
  const [p, setP] = useState<MikrotikParams>(() => paramsFromNas(nas));
  const [activeScript, setActiveScript] = useState<
    "setup" | "teardown" | "login"
  >("setup");
  const script = generateMikrotikScript(p);
  const teardownScript = generateMikrotikTeardown();
  const storeUrl = storeUrlFromParams(p);
  const loginHtml = generateLoginHtml(storeUrl);
  const activeArtifact = {
    setup: {
      label: "Script setup",
      fileName: "hotspot-billing.rsc",
      content: script,
    },
    teardown: {
      label: "Script teardown",
      fileName: "hotspot-billing-teardown.rsc",
      content: teardownScript,
    },
    login: {
      label: "login.html",
      fileName: "login.html",
      content: loginHtml,
    },
  }[activeScript];

  // Derive the WiFi-source mode from params: bridge ports set → new bridge;
  // hsInterface "bridge" with no ports → built-in bridge; otherwise single.
  const hsMode = p.bridgePorts
    ? "bridge"
    : p.hsInterface === "bridge"
      ? "builtin"
      : "single";
  const feHostIsLocal =
    p.feHost.trim().toLowerCase() === "localhost" ||
    p.feHost.trim().startsWith("127.");

  const set = (k: keyof MikrotikParams) => (v: string) =>
    setP((prev) => ({ ...prev, [k]: v }));

  const download = (text: string, name: string) => {
    const blob = new Blob([text], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = name;
    document.body.appendChild(a);
    a.click();
    a.remove();
    URL.revokeObjectURL(url);
  };

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-h-[90vh] max-w-3xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Konfigurasi Mikrotik — {nas.nasname}</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Sesuaikan parameter jaringan, lalu salin/unduh artefak di bawah.
            Secret otomatis sama dengan NAS ini.
          </p>

          <div className="grid gap-3 sm:grid-cols-2">
            <SmallField label="IP Server RADIUS">
              <Input
                value={p.radiusIP}
                onChange={(e) => set("radiusIP")(e.target.value)}
                placeholder="IP server billing/RADIUS"
              />
            </SmallField>
            <SmallField label="RADIUS Secret (= secret NAS)">
              <Input
                value={p.radiusSecret}
                onChange={(e) => set("radiusSecret")(e.target.value)}
              />
            </SmallField>
            <SmallField label="Host/URL Backend (walled garden)">
              <Input
                value={p.feHost}
                onChange={(e) => set("feHost")(e.target.value)}
                placeholder="192.168.1.3:8082 / http://192.168.1.3:8082"
              />
              {feHostIsLocal && (
                <p className="text-xs text-destructive">
                  Jangan pakai localhost untuk script router. Isi IP mini PC
                  yang bisa dijangkau Mikrotik.
                </p>
              )}
            </SmallField>
            <SmallField label="CoA Port">
              <Input
                value={p.coaPort}
                onChange={(e) => set("coaPort")(e.target.value)}
              />
            </SmallField>
            <SmallField label="WAN / Sumber Internet">
              <Input
                value={p.wanInterface}
                onChange={(e) => set("wanInterface")(e.target.value)}
                placeholder="ether1"
              />
            </SmallField>
            <SmallField label="Sumber Hotspot (WiFi)">
              <select
                className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
                value={hsMode}
                onChange={(e) => {
                  const mode = e.target.value;
                  if (mode === "bridge") {
                    setP((prev) => ({
                      ...prev,
                      hsInterface: "bridge-hotspot",
                      bridgePorts: "wlan1,wlan2",
                    }));
                  } else if (mode === "builtin") {
                    // Pakai bridge bawaan Mikrotik (stok WiFi sudah di dalamnya).
                    setP((prev) => ({
                      ...prev,
                      hsInterface: "bridge",
                      bridgePorts: "",
                    }));
                  } else {
                    setP((prev) => ({
                      ...prev,
                      hsInterface: "wlan1",
                      bridgePorts: "",
                    }));
                  }
                }}
              >
                <option value="bridge">
                  Gabung beberapa interface ke bridge baru
                </option>
                <option value="builtin">
                  Pakai WiFi/bridge bawaan Mikrotik
                </option>
                <option value="single">Satu interface saja</option>
              </select>
            </SmallField>
            {hsMode === "bridge" ? (
              <>
                <SmallField label="Nama Bridge">
                  <Input
                    value={p.hsInterface}
                    onChange={(e) => set("hsInterface")(e.target.value)}
                    placeholder="bridge-hotspot"
                  />
                </SmallField>
                <SmallField label="Interface anggota (ether & wlan, pisah koma)">
                  <Input
                    value={p.bridgePorts}
                    onChange={(e) => set("bridgePorts")(e.target.value)}
                    placeholder="wlan1,wlan2,ether3"
                  />
                </SmallField>
              </>
            ) : (
              <SmallField
                label={
                  hsMode === "builtin"
                    ? "Nama Bridge Bawaan"
                    : "Interface Hotspot (ether / wlan)"
                }
              >
                <Input
                  value={p.hsInterface}
                  onChange={(e) => set("hsInterface")(e.target.value)}
                  placeholder={
                    hsMode === "builtin" ? "bridge" : "wlan1 / ether3"
                  }
                />
              </SmallField>
            )}
            <div className="sm:col-span-2 -mt-1 text-xs text-muted-foreground">
              {hsMode === "bridge"
                ? "Isi interface yang dijadikan hotspot, mis. wlan1,wlan2 (WiFi 2.4G+5G) atau campur dengan kabel: wlan1,ether3. Semua digabung ke bridge di atas."
                : hsMode === "builtin"
                  ? "Pakai bridge bawaan Mikrotik (cek di menu Bridge). Interface yang sudah ada di dalamnya — termasuk WiFi & port ether — otomatis jadi hotspot."
                  : "Pilih satu interface saja sebagai hotspot, mis. wlan1 (WiFi) atau ether3 (kabel)."}
            </div>
            <SmallField label="Subnet Hotspot (CIDR)">
              <Input
                value={p.hsNetwork}
                onChange={(e) => set("hsNetwork")(e.target.value)}
                placeholder="10.5.50.0/24"
              />
            </SmallField>
            <SmallField label="Gateway Hotspot">
              <Input
                value={p.hsGateway}
                onChange={(e) => set("hsGateway")(e.target.value)}
                placeholder="10.5.50.1"
              />
            </SmallField>
            <SmallField label="Range DHCP">
              <Input
                value={p.hsPoolRange}
                onChange={(e) => set("hsPoolRange")(e.target.value)}
                placeholder="10.5.50.10-10.5.50.254"
              />
            </SmallField>
            <SmallField label="DNS">
              <Input
                value={p.hsDNS}
                onChange={(e) => set("hsDNS")(e.target.value)}
                placeholder="8.8.8.8,1.1.1.1"
              />
            </SmallField>
          </div>

          <div className="space-y-3 rounded-lg border border-dashed p-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Tabs
                value={activeScript}
                onValueChange={(v) =>
                  setActiveScript(v as "setup" | "teardown" | "login")
                }
              >
                <TabsList>
                  <TabsTrigger value="setup">Setup .rsc</TabsTrigger>
                  <TabsTrigger value="teardown">Teardown .rsc</TabsTrigger>
                  <TabsTrigger value="login">login.html</TabsTrigger>
                </TabsList>
              </Tabs>
              <Button
                variant="outline"
                onClick={() =>
                  download(activeArtifact.content, activeArtifact.fileName)
                }
              >
                <Download className="h-4 w-4" /> Unduh {activeArtifact.label}
              </Button>
            </div>

            <p className="text-xs text-muted-foreground">
              Pilih tab untuk melihat artefak. Script setup mengunduh{" "}
              <code>login.html</code> otomatis dari <code>{storeUrl}</code>.
            </p>
          </div>

          <Textarea
            readOnly
            value={activeArtifact.content}
            className="h-72 resize-y whitespace-pre-wrap font-mono text-xs leading-relaxed"
          />
        </div>
      </DialogContent>
    </Dialog>
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

function SmallField({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-1">
      <Label className="text-xs">{label}</Label>
      {children}
    </div>
  );
}
