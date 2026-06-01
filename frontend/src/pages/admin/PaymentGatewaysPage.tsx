import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { api } from "@/lib/api";
import { errorMessage } from "@/lib/form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { GatewaySettings } from "@/types";

interface FormState {
  default_provider: string;
  enable_cash: boolean;
  midtrans: {
    enabled: boolean;
    production: boolean;
    server_key: string;
    client_key: string;
  };
  xendit: {
    enabled: boolean;
    secret_key: string;
    callback_token: string;
  };
  tripay: {
    enabled: boolean;
    production: boolean;
    api_key: string;
    private_key: string;
    merchant_code: string;
  };
}

function toForm(d: GatewaySettings): FormState {
  // Secret fields start blank — the masked current value is shown as a
  // placeholder; leaving it blank keeps the stored secret. Public fields
  // (client key, merchant code) are prefilled with their real values.
  return {
    default_provider: d.default_provider || "midtrans",
    enable_cash: d.enable_cash,
    midtrans: {
      enabled: d.midtrans.enabled,
      production: d.midtrans.production,
      server_key: "",
      client_key: d.midtrans.client_key,
    },
    xendit: {
      enabled: d.xendit.enabled,
      secret_key: "",
      callback_token: "",
    },
    tripay: {
      enabled: d.tripay.enabled,
      production: d.tripay.production,
      api_key: "",
      private_key: "",
      merchant_code: d.tripay.merchant_code,
    },
  };
}

export default function PaymentGatewaysPage() {
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ["payment-gateways"],
    queryFn: api.paymentGateways.get,
  });
  const [form, setForm] = useState<FormState | null>(null);

  useEffect(() => {
    if (data) setForm(toForm(data));
  }, [data]);

  const save = useMutation({
    mutationFn: (body: FormState) => api.paymentGateways.update(body),
    onSuccess: () => {
      toast.success("Pengaturan gateway disimpan");
      qc.invalidateQueries({ queryKey: ["payment-gateways"] });
      qc.invalidateQueries({ queryKey: ["public-settings"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  if (isLoading || !form || !data) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold">Payment Gateway</h1>
          <p className="text-muted-foreground">Kelola kredensial pembayaran.</p>
        </div>
        <Skeleton className="h-96 rounded-xl" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Payment Gateway</h1>
          <p className="text-muted-foreground">
            Kelola kredensial dan aktivasi metode pembayaran.
          </p>
        </div>
        <Button onClick={() => save.mutate(form)} disabled={save.isPending}>
          {save.isPending && <Loader2 className="h-4 w-4 animate-spin" />}
          Simpan
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Umum</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-2">
            <Label>Provider Default</Label>
            <Select
              value={form.default_provider}
              onValueChange={(v) =>
                setForm({ ...form, default_provider: v })
              }
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="midtrans">Midtrans</SelectItem>
                <SelectItem value="xendit">Xendit</SelectItem>
                <SelectItem value="tripay">Tripay</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <label className="flex items-end gap-3 pb-2 text-sm">
            <Switch
              checked={form.enable_cash}
              onCheckedChange={(c) => setForm({ ...form, enable_cash: c })}
            />
            Aktifkan pembayaran Tunai (Cash)
          </label>
        </CardContent>
      </Card>

      <Tabs defaultValue="midtrans">
        <TabsList>
          <TabsTrigger value="midtrans">Midtrans</TabsTrigger>
          <TabsTrigger value="xendit">Xendit</TabsTrigger>
          <TabsTrigger value="tripay">Tripay</TabsTrigger>
        </TabsList>

        {/* ── Midtrans ── */}
        <TabsContent value="midtrans">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle className="text-base">Midtrans (Snap)</CardTitle>
              <ConfiguredBadge ok={data.midtrans.configured} />
            </CardHeader>
            <CardContent className="space-y-4">
              <ToggleRow
                label="Tampilkan ke pelanggan"
                checked={form.midtrans.enabled}
                onChange={(c) =>
                  setForm({
                    ...form,
                    midtrans: { ...form.midtrans, enabled: c },
                  })
                }
              />
              <ToggleRow
                label="Mode Production (live)"
                hint="Nonaktif = Sandbox"
                checked={form.midtrans.production}
                onChange={(c) =>
                  setForm({
                    ...form,
                    midtrans: { ...form.midtrans, production: c },
                  })
                }
              />
              <SecretField
                label="Server Key"
                placeholder={data.midtrans.server_key || "Belum diatur"}
                value={form.midtrans.server_key}
                onChange={(v) =>
                  setForm({
                    ...form,
                    midtrans: { ...form.midtrans, server_key: v },
                  })
                }
              />
              <div className="space-y-2">
                <Label>Client Key</Label>
                <Input
                  value={form.midtrans.client_key}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      midtrans: {
                        ...form.midtrans,
                        client_key: e.target.value,
                      },
                    })
                  }
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ── Xendit ── */}
        <TabsContent value="xendit">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle className="text-base">Xendit</CardTitle>
              <ConfiguredBadge ok={data.xendit.configured} />
            </CardHeader>
            <CardContent className="space-y-4">
              <ToggleRow
                label="Tampilkan ke pelanggan"
                checked={form.xendit.enabled}
                onChange={(c) =>
                  setForm({ ...form, xendit: { ...form.xendit, enabled: c } })
                }
              />
              <SecretField
                label="Secret Key"
                placeholder={data.xendit.secret_key || "Belum diatur"}
                value={form.xendit.secret_key}
                onChange={(v) =>
                  setForm({
                    ...form,
                    xendit: { ...form.xendit, secret_key: v },
                  })
                }
              />
              <SecretField
                label="Callback Verification Token"
                placeholder={data.xendit.callback_token || "Belum diatur"}
                value={form.xendit.callback_token}
                onChange={(v) =>
                  setForm({
                    ...form,
                    xendit: { ...form.xendit, callback_token: v },
                  })
                }
              />
            </CardContent>
          </Card>
        </TabsContent>

        {/* ── Tripay ── */}
        <TabsContent value="tripay">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle className="text-base">Tripay</CardTitle>
              <ConfiguredBadge ok={data.tripay.configured} />
            </CardHeader>
            <CardContent className="space-y-4">
              <ToggleRow
                label="Tampilkan ke pelanggan"
                checked={form.tripay.enabled}
                onChange={(c) =>
                  setForm({ ...form, tripay: { ...form.tripay, enabled: c } })
                }
              />
              <ToggleRow
                label="Mode Production (live)"
                hint="Nonaktif = Sandbox"
                checked={form.tripay.production}
                onChange={(c) =>
                  setForm({
                    ...form,
                    tripay: { ...form.tripay, production: c },
                  })
                }
              />
              <SecretField
                label="API Key"
                placeholder={data.tripay.api_key || "Belum diatur"}
                value={form.tripay.api_key}
                onChange={(v) =>
                  setForm({ ...form, tripay: { ...form.tripay, api_key: v } })
                }
              />
              <SecretField
                label="Private Key"
                placeholder={data.tripay.private_key || "Belum diatur"}
                value={form.tripay.private_key}
                onChange={(v) =>
                  setForm({
                    ...form,
                    tripay: { ...form.tripay, private_key: v },
                  })
                }
              />
              <div className="space-y-2">
                <Label>Merchant Code</Label>
                <Input
                  value={form.tripay.merchant_code}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      tripay: {
                        ...form.tripay,
                        merchant_code: e.target.value,
                      },
                    })
                  }
                />
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <p className="text-sm text-muted-foreground">
        Kolom rahasia ditampilkan tersamar. Biarkan kosong untuk
        mempertahankan nilai yang tersimpan; isi hanya jika ingin menggantinya.
      </p>
    </div>
  );
}

function ConfiguredBadge({ ok }: { ok: boolean }) {
  return ok ? (
    <Badge variant="success" className="gap-1">
      <CheckCircle2 className="h-3 w-3" /> Terkonfigurasi
    </Badge>
  ) : (
    <Badge variant="secondary" className="gap-1">
      <XCircle className="h-3 w-3" /> Belum lengkap
    </Badge>
  );
}

function ToggleRow({
  label,
  hint,
  checked,
  onChange,
}: {
  label: string;
  hint?: string;
  checked: boolean;
  onChange: (c: boolean) => void;
}) {
  return (
    <label className="flex items-center justify-between gap-3 text-sm">
      <span>
        {label}
        {hint && <span className="ml-2 text-muted-foreground">({hint})</span>}
      </span>
      <Switch checked={checked} onCheckedChange={onChange} />
    </label>
  );
}

function SecretField({
  label,
  placeholder,
  value,
  onChange,
}: {
  label: string;
  placeholder: string;
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div className="space-y-2">
      <Label>{label}</Label>
      <Input
        type="password"
        autoComplete="off"
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  );
}
