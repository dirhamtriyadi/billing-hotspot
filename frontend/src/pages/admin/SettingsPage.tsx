import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";
import { api } from "@/lib/api";
import { errorMessage, applyApiErrors } from "@/lib/form";
import { normalizeWaNumber, isValidWaNumber } from "@/lib/phone";
import { changePasswordSchema, type ChangePasswordValues } from "@/schemas";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const FIELDS: { key: string; label: string; type?: "textarea" }[] = [
  { key: "site_name", label: "Nama Situs" },
  { key: "site_subtitle", label: "Subjudul (hero)" },
  { key: "site_description", label: "Deskripsi", type: "textarea" },
  { key: "contact_whatsapp", label: "WhatsApp (mis. 6281234567890)" },
  { key: "enabled_providers", label: "Provider Aktif (midtrans,xendit,tripay)" },
  { key: "enable_cash", label: "Aktifkan Cash (true/false)" },
];

export default function SettingsPage() {
  const qc = useQueryClient();
  const [values, setValues] = useState<Record<string, string>>({});
  const [waError, setWaError] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["settings"],
    queryFn: api.settings.get,
  });

  useEffect(() => {
    if (data) setValues(data);
  }, [data]);

  const save = useMutation({
    mutationFn: (payload: Record<string, string>) =>
      api.settings.update(payload),
    onSuccess: () => {
      toast.success("Pengaturan disimpan");
      qc.invalidateQueries({ queryKey: ["settings"] });
      qc.invalidateQueries({ queryKey: ["public-settings"] });
    },
    onError: (e) => toast.error(errorMessage(e)),
  });

  // Validate + normalise the WhatsApp number at save time so a malformed value
  // never reaches the DB. Empty is allowed (button just won't show downstream).
  const handleSave = () => {
    const wa = (values.contact_whatsapp ?? "").trim();
    if (wa && !isValidWaNumber(wa)) {
      setWaError(
        "Nomor WhatsApp tidak valid. Contoh: 081313102678 atau 6281313102678.",
      );
      toast.error("Nomor WhatsApp tidak valid");
      return;
    }
    setWaError("");
    const payload = { ...values };
    if (wa) payload.contact_whatsapp = normalizeWaNumber(wa); // standardise → 62…
    setValues(payload);
    save.mutate(payload);
  };

  const pwForm = useForm<ChangePasswordValues>({
    resolver: zodResolver(changePasswordSchema),
    defaultValues: { old_password: "", new_password: "", confirm_password: "" },
  });

  const changePw = async (v: ChangePasswordValues) => {
    try {
      await api.auth.changePassword({
        old_password: v.old_password,
        new_password: v.new_password,
      });
      toast.success("Password diperbarui");
      pwForm.reset();
    } catch (e) {
      if (!applyApiErrors(e, pwForm.setError)) toast.error(errorMessage(e));
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Pengaturan</h1>
        <p className="text-muted-foreground">
          Konfigurasi storefront dan akun.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Storefront</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {isLoading ? (
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
            ) : (
              <>
                {FIELDS.map((f) => (
                  <div key={f.key} className="space-y-2">
                    <Label>{f.label}</Label>
                    {f.type === "textarea" ? (
                      <Textarea
                        rows={2}
                        value={values[f.key] ?? ""}
                        onChange={(e) =>
                          setValues((v) => ({ ...v, [f.key]: e.target.value }))
                        }
                      />
                    ) : f.key === "contact_whatsapp" ? (
                      <>
                        <Input
                          value={values[f.key] ?? ""}
                          placeholder="081313102678 / 6281313102678"
                          onChange={(e) => {
                            setWaError("");
                            setValues((v) => ({
                              ...v,
                              [f.key]: e.target.value,
                            }));
                          }}
                          onBlur={(e) => {
                            // Rapikan ke standar 62… begitu admin pindah fokus,
                            // jadi yang tampil = yang akan disimpan.
                            const n = normalizeWaNumber(e.target.value);
                            if (n)
                              setValues((v) => ({ ...v, [f.key]: n }));
                          }}
                        />
                        {waError && (
                          <p className="text-sm text-destructive">{waError}</p>
                        )}
                      </>
                    ) : (
                      <Input
                        value={values[f.key] ?? ""}
                        onChange={(e) =>
                          setValues((v) => ({ ...v, [f.key]: e.target.value }))
                        }
                      />
                    )}
                  </div>
                ))}
                <Button onClick={handleSave} disabled={save.isPending}>
                  {save.isPending && (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  )}
                  Simpan Pengaturan
                </Button>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Ganti Password</CardTitle>
          </CardHeader>
          <CardContent>
            <form
              onSubmit={pwForm.handleSubmit(changePw)}
              className="space-y-4"
              noValidate
            >
              <div className="space-y-2">
                <Label>Password Lama</Label>
                <Input type="password" {...pwForm.register("old_password")} />
                {pwForm.formState.errors.old_password && (
                  <p className="text-sm text-destructive">
                    {pwForm.formState.errors.old_password.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label>Password Baru</Label>
                <Input type="password" {...pwForm.register("new_password")} />
                {pwForm.formState.errors.new_password && (
                  <p className="text-sm text-destructive">
                    {pwForm.formState.errors.new_password.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label>Konfirmasi Password Baru</Label>
                <Input
                  type="password"
                  {...pwForm.register("confirm_password")}
                />
                {pwForm.formState.errors.confirm_password && (
                  <p className="text-sm text-destructive">
                    {pwForm.formState.errors.confirm_password.message}
                  </p>
                )}
              </div>
              <Button type="submit" disabled={pwForm.formState.isSubmitting}>
                {pwForm.formState.isSubmitting && (
                  <Loader2 className="h-4 w-4 animate-spin" />
                )}
                Ganti Password
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
