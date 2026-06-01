import { Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowRight,
  Check,
  Clock,
  Database,
  Gauge,
  ShieldCheck,
  Sparkles,
  Wifi,
  Zap,
} from "lucide-react";
import { api } from "@/lib/api";
import { formatIDR, formatQuota } from "@/lib/format";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { PackageIcon } from "@/components/PackageIcon";
import type { PublicPackage } from "@/types";

export default function LandingPage() {
  const navigate = useNavigate();
  const { data: packages, isLoading } = useQuery({
    queryKey: ["public-packages"],
    queryFn: api.public.packages,
  });
  const { data: settings } = useQuery({
    queryKey: ["public-settings"],
    queryFn: api.public.settings,
  });

  const siteName = settings?.site_name || "WiFi Hotspot";
  const subtitle =
    settings?.site_subtitle || "Internet cepat, bayar sesuai kebutuhan";

  return (
    <div className="min-h-screen bg-background">
      {/* Nav */}
      <header className="sticky top-0 z-30 border-b bg-background/80 backdrop-blur">
        <div className="container flex h-16 items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <Wifi className="h-5 w-5" />
            </div>
            <span className="text-lg font-bold">{siteName}</span>
          </div>
          <Link
            to="/admin/login"
            className="text-sm font-medium text-muted-foreground hover:text-foreground"
          >
            Login Admin
          </Link>
        </div>
      </header>

      {/* Hero */}
      <section className="hero-gradient relative overflow-hidden">
        <div className="container py-20 text-center">
          <div className="mx-auto mb-5 inline-flex items-center gap-2 rounded-full border bg-background/60 px-4 py-1.5 text-sm font-medium text-primary">
            <Sparkles className="h-4 w-4" />
            Aktif instan setelah pembayaran
          </div>
          <h1 className="mx-auto max-w-3xl text-4xl font-extrabold tracking-tight sm:text-5xl md:text-6xl">
            {subtitle}
          </h1>
          <p className="mx-auto mt-5 max-w-xl text-lg text-muted-foreground">
            {settings?.site_description ||
              "Pilih paket internet sesuai kebutuhanmu, bayar, dan langsung online."}
          </p>
          <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
            <Button size="lg" asChild>
              <a href="#paket">
                Lihat Paket <ArrowRight className="h-4 w-4" />
              </a>
            </Button>
          </div>
          <div className="mt-10 flex flex-wrap items-center justify-center gap-6 text-sm text-muted-foreground">
            <span className="inline-flex items-center gap-2">
              <ShieldCheck className="h-4 w-4 text-emerald-500" /> Pembayaran aman
            </span>
            <span className="inline-flex items-center gap-2">
              <Zap className="h-4 w-4 text-amber-500" /> Voucher otomatis
            </span>
            <span className="inline-flex items-center gap-2">
              <Gauge className="h-4 w-4 text-primary" /> Kecepatan stabil
            </span>
          </div>
        </div>
      </section>

      {/* Packages */}
      <section id="paket" className="container scroll-mt-20 py-16">
        <div className="mb-10 text-center">
          <h2 className="text-3xl font-bold">Pilih Paket Internet</h2>
          <p className="mt-2 text-muted-foreground">
            Semua paket aktif otomatis dan bisa langsung dipakai.
          </p>
        </div>

        {isLoading ? (
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-80 w-full rounded-2xl" />
            ))}
          </div>
        ) : packages && packages.length > 0 ? (
          <div className="grid items-stretch gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {packages.map((pkg, i) => (
              <PackageCard
                key={pkg.id}
                pkg={pkg}
                index={i}
                onSelect={() => navigate(`/checkout/${pkg.slug}`)}
              />
            ))}
          </div>
        ) : (
          <div className="rounded-2xl border border-dashed py-16 text-center text-muted-foreground">
            Belum ada paket tersedia saat ini.
          </div>
        )}
      </section>

      {/* Footer */}
      <footer className="border-t">
        <div className="container flex flex-col items-center justify-between gap-3 py-8 text-sm text-muted-foreground sm:flex-row">
          <span>
            © {new Date().getFullYear()} {siteName}. Semua hak dilindungi.
          </span>
          {settings?.contact_whatsapp && (
            <a
              href={`https://wa.me/${settings.contact_whatsapp}`}
              target="_blank"
              rel="noreferrer"
              className="hover:text-foreground"
            >
              Butuh bantuan? Hubungi kami
            </a>
          )}
        </div>
      </footer>
    </div>
  );
}

function PackageCard({
  pkg,
  index,
  onSelect,
}: {
  pkg: PublicPackage;
  index: number;
  onSelect: () => void;
}) {
  const accent = pkg.color || "#2563eb";
  return (
    <div
      className={cn(
        "group relative flex animate-fade-in-up flex-col rounded-2xl border bg-card p-6 shadow-sm transition-all hover:-translate-y-1 hover:shadow-xl",
        pkg.highlight && "border-primary/40 ring-2 ring-primary/30",
      )}
      style={{ animationDelay: `${index * 60}ms` }}
    >
      {pkg.highlight && pkg.badge_text && (
        <span className="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary px-3 py-1 text-xs font-bold text-primary-foreground shadow">
          {pkg.badge_text}
        </span>
      )}
      {!pkg.highlight && pkg.badge_text && (
        <span
          className="absolute right-4 top-4 rounded-full px-2.5 py-0.5 text-xs font-semibold text-white"
          style={{ backgroundColor: accent }}
        >
          {pkg.badge_text}
        </span>
      )}

      <div
        className="mb-4 flex h-12 w-12 items-center justify-center rounded-xl"
        style={{ backgroundColor: `${accent}1a`, color: accent }}
      >
        <PackageIcon name={pkg.icon} className="h-6 w-6" />
      </div>

      <h3 className="text-xl font-bold">{pkg.name}</h3>
      <p className="mt-1 line-clamp-2 min-h-[2.5rem] text-sm text-muted-foreground">
        {pkg.description}
      </p>

      <div className="my-5">
        <span className="text-3xl font-extrabold">{formatIDR(pkg.price)}</span>
      </div>

      <ul className="mb-6 space-y-2.5 text-sm">
        <li className="flex items-center gap-2">
          <Gauge className="h-4 w-4 shrink-0 text-primary" />
          Hingga <strong>{pkg.download_mbps} Mbps</strong> download
        </li>
        <li className="flex items-center gap-2">
          <Clock className="h-4 w-4 shrink-0 text-primary" />
          Berlaku <strong>{pkg.validity}</strong>
        </li>
        <li className="flex items-center gap-2">
          <Database className="h-4 w-4 shrink-0 text-primary" />
          Kuota <strong>{formatQuota(pkg.data_quota_mb)}</strong>
        </li>
        <li className="flex items-center gap-2">
          <Check className="h-4 w-4 shrink-0 text-emerald-500" />
          Aktif otomatis
        </li>
      </ul>

      <Button
        className="mt-auto w-full"
        size="lg"
        variant={pkg.highlight ? "default" : "outline"}
        onClick={onSelect}
      >
        Pilih Paket
      </Button>
    </div>
  );
}
