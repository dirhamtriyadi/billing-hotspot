import { Link, useNavigate, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import {
  ArrowLeft,
  Banknote,
  Clock,
  CreditCard,
  Database,
  Gauge,
  Loader2,
  QrCode,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { api } from "@/lib/api";
import { formatIDR, formatQuota } from "@/lib/format";
import { cn } from "@/lib/utils";
import { applyApiErrors, errorMessage } from "@/lib/form";
import { checkoutSchema, type CheckoutValues } from "@/schemas";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { PackageIcon } from "@/components/PackageIcon";
import type { PaymentMethod } from "@/types";

const methodMeta: Record<
  PaymentMethod,
  { label: string; desc: string; icon: LucideIcon }
> = {
  cash: {
    label: "Tunai / Cash",
    desc: "Bayar langsung ke operator",
    icon: Banknote,
  },
  midtrans: {
    label: "Midtrans",
    desc: "Kartu, e-wallet, VA, QRIS",
    icon: CreditCard,
  },
  xendit: { label: "Xendit", desc: "VA, e-wallet, QRIS, retail", icon: Wallet },
  tripay: { label: "Tripay", desc: "QRIS, VA, e-wallet", icon: QrCode },
};

export default function CheckoutPage() {
  const { slug } = useParams<{ slug: string }>();
  const navigate = useNavigate();

  const { data: packages, isLoading } = useQuery({
    queryKey: ["public-packages"],
    queryFn: api.public.packages,
  });
  const { data: settings } = useQuery({
    queryKey: ["public-settings"],
    queryFn: api.public.settings,
  });

  const pkg = packages?.find((p) => p.slug === slug);
  const methods = settings?.payment_methods ?? [];

  const form = useForm<CheckoutValues>({
    resolver: zodResolver(checkoutSchema),
    defaultValues: {
      customer_name: "",
      customer_phone: "",
      customer_email: "",
    },
  });

  const selectedMethod = form.watch("payment_method");

  const onSubmit = async (values: CheckoutValues) => {
    if (!pkg) return;
    try {
      const order = await api.public.checkout({
        package_id: pkg.id,
        customer_name: values.customer_name,
        customer_phone: values.customer_phone,
        customer_email: values.customer_email || "",
        payment_method: values.payment_method,
      });
      if (order.payment_url) {
        window.location.href = order.payment_url;
        return;
      }
      navigate(`/payment/${order.order_number}`);
    } catch (e) {
      if (!applyApiErrors(e, form.setError)) toast.error(errorMessage(e));
    }
  };

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!pkg) {
    return (
      <div className="container py-20 text-center">
        <p className="text-lg text-muted-foreground">Paket tidak ditemukan.</p>
        <Button className="mt-4" asChild>
          <Link to="/">Kembali ke beranda</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-muted/30">
      <div className="container max-w-5xl py-8">
        <Link
          to="/"
          className="mb-6 inline-flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Kembali ke daftar paket
        </Link>

        <div className="grid gap-6 lg:grid-cols-[1fr_360px]">
          {/* Form */}
          <Card>
            <CardHeader>
              <CardTitle>Data Pembeli & Pembayaran</CardTitle>
            </CardHeader>
            <CardContent>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="space-y-5"
                noValidate
              >
                <div className="space-y-2">
                  <Label htmlFor="customer_name">Nama Lengkap</Label>
                  <Input
                    id="customer_name"
                    placeholder="Nama kamu"
                    {...form.register("customer_name")}
                  />
                  <FieldError msg={form.formState.errors.customer_name?.message} />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="customer_phone">Nomor WhatsApp / HP</Label>
                  <Input
                    id="customer_phone"
                    placeholder="08xxxxxxxxxx"
                    inputMode="numeric"
                    {...form.register("customer_phone")}
                  />
                  <FieldError
                    msg={form.formState.errors.customer_phone?.message}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="customer_email">Email (opsional)</Label>
                  <Input
                    id="customer_email"
                    type="email"
                    placeholder="email@contoh.com"
                    {...form.register("customer_email")}
                  />
                  <FieldError
                    msg={form.formState.errors.customer_email?.message}
                  />
                </div>

                <div className="space-y-2">
                  <Label>Metode Pembayaran</Label>
                  <div className="grid gap-3 sm:grid-cols-2">
                    {methods.length === 0 && (
                      <p className="text-sm text-muted-foreground">
                        Belum ada metode pembayaran aktif.
                      </p>
                    )}
                    {methods.map((m) => {
                      const meta = methodMeta[m];
                      if (!meta) return null;
                      const active = selectedMethod === m;
                      return (
                        <button
                          type="button"
                          key={m}
                          onClick={() =>
                            form.setValue("payment_method", m, {
                              shouldValidate: true,
                            })
                          }
                          className={cn(
                            "flex items-start gap-3 rounded-xl border p-4 text-left transition-all",
                            active
                              ? "border-primary bg-primary/5 ring-2 ring-primary/30"
                              : "hover:border-primary/40 hover:bg-accent",
                          )}
                        >
                          <meta.icon className="mt-0.5 h-5 w-5 text-primary" />
                          <div>
                            <div className="font-medium">{meta.label}</div>
                            <div className="text-xs text-muted-foreground">
                              {meta.desc}
                            </div>
                          </div>
                        </button>
                      );
                    })}
                  </div>
                  <FieldError
                    msg={form.formState.errors.payment_method?.message}
                  />
                </div>

                <Button
                  type="submit"
                  size="lg"
                  className="w-full"
                  disabled={form.formState.isSubmitting}
                >
                  {form.formState.isSubmitting && (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  )}
                  Bayar {formatIDR(pkg.price)}
                </Button>
              </form>
            </CardContent>
          </Card>

          {/* Summary */}
          <Card className="h-fit lg:sticky lg:top-8">
            <CardHeader>
              <CardTitle className="text-base">Ringkasan Pesanan</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-3">
                <div
                  className="flex h-11 w-11 items-center justify-center rounded-xl"
                  style={{
                    backgroundColor: `${pkg.color}1a`,
                    color: pkg.color,
                  }}
                >
                  <PackageIcon name={pkg.icon} className="h-5 w-5" />
                </div>
                <div>
                  <div className="font-semibold">{pkg.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {pkg.description}
                  </div>
                </div>
              </div>

              <div className="space-y-2 border-t pt-4 text-sm">
                <Row
                  icon={Gauge}
                  label="Kecepatan"
                  value={`${pkg.download_mbps} Mbps`}
                />
                <Row icon={Clock} label="Masa aktif" value={pkg.validity} />
                <Row
                  icon={Database}
                  label="Kuota"
                  value={formatQuota(pkg.data_quota_mb)}
                />
              </div>

              <div className="flex items-center justify-between border-t pt-4">
                <span className="text-muted-foreground">Total</span>
                <span className="text-2xl font-extrabold">
                  {formatIDR(pkg.price)}
                </span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

function FieldError({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <p className="text-sm text-destructive">{msg}</p>;
}

function Row({
  icon: Icon,
  label,
  value,
}: {
  icon: LucideIcon;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="inline-flex items-center gap-2 text-muted-foreground">
        <Icon className="h-4 w-4" /> {label}
      </span>
      <span className="font-medium">{value}</span>
    </div>
  );
}
