import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  Banknote,
  Download,
  Receipt,
  ShoppingCart,
  Ticket,
  Loader2,
} from "lucide-react";
import { api } from "@/lib/api";
import { errorMessage } from "@/lib/form";
import { formatIDR } from "@/lib/format";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { LucideIcon } from "lucide-react";

function isoDaysAgo(days: number): string {
  const d = new Date();
  d.setDate(d.getDate() - days);
  return d.toISOString().slice(0, 10);
}

const METHOD_LABELS: Record<string, string> = {
  cash: "Tunai",
  midtrans: "Midtrans",
  xendit: "Xendit",
  tripay: "Tripay",
};

export default function ReportsPage() {
  const [start, setStart] = useState(isoDaysAgo(29));
  const [end, setEnd] = useState(isoDaysAgo(0));
  const [downloading, setDownloading] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["report", start, end],
    queryFn: () => api.reports.revenue({ start, end }),
  });

  const handleExport = async () => {
    setDownloading(true);
    try {
      const blob = await api.reports.download({ start, end });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `laporan-${start}-${end}.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (e) {
      toast.error(errorMessage(e));
    } finally {
      setDownloading(false);
    }
  };

  const s = data?.summary;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Laporan Pendapatan</h1>
          <p className="text-muted-foreground">
            Analisis pendapatan dari pesanan yang sudah lunas.
          </p>
        </div>
        <div className="flex flex-wrap items-end gap-2">
          <div className="space-y-1">
            <Label className="text-xs">Dari</Label>
            <Input
              type="date"
              value={start}
              max={end}
              onChange={(e) => setStart(e.target.value)}
              className="w-40"
            />
          </div>
          <div className="space-y-1">
            <Label className="text-xs">Sampai</Label>
            <Input
              type="date"
              value={end}
              min={start}
              onChange={(e) => setEnd(e.target.value)}
              className="w-40"
            />
          </div>
          <Button variant="outline" onClick={handleExport} disabled={downloading}>
            {downloading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Download className="h-4 w-4" />
            )}
            Ekspor CSV
          </Button>
        </div>
      </div>

      {isLoading || !data || !s ? (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-28 rounded-xl" />
            ))}
          </div>
          <Skeleton className="h-80 rounded-xl" />
        </>
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              icon={Banknote}
              label="Total Pendapatan"
              value={formatIDR(s.revenue_total)}
              accent="text-emerald-500"
            />
            <StatCard
              icon={ShoppingCart}
              label="Pesanan Lunas"
              value={`${s.paid_orders} / ${s.total_orders}`}
              accent="text-primary"
            />
            <StatCard
              icon={Receipt}
              label="Rata-rata / Pesanan"
              value={formatIDR(s.avg_order_value)}
              accent="text-violet-500"
            />
            <StatCard
              icon={Ticket}
              label="Voucher Terbit"
              value={String(s.vouchers_issued)}
              accent="text-amber-500"
            />
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Tren Pendapatan Harian</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-72 w-full">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart
                    data={data.series}
                    margin={{ top: 8, right: 12, left: 4, bottom: 0 }}
                  >
                    <defs>
                      <linearGradient id="rev" x1="0" y1="0" x2="0" y2="1">
                        <stop
                          offset="5%"
                          stopColor="hsl(var(--primary))"
                          stopOpacity={0.35}
                        />
                        <stop
                          offset="95%"
                          stopColor="hsl(var(--primary))"
                          stopOpacity={0}
                        />
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                    <XAxis
                      dataKey="date"
                      tickFormatter={(d: string) => d.slice(5)}
                      fontSize={12}
                      tickMargin={8}
                    />
                    <YAxis
                      width={70}
                      fontSize={12}
                      tickFormatter={(v: number) =>
                        v >= 1000 ? `${Math.round(v / 1000)}rb` : String(v)
                      }
                    />
                    <Tooltip
                      formatter={(v: number) => formatIDR(v)}
                      labelClassName="text-foreground"
                      contentStyle={{
                        borderRadius: 8,
                        border: "1px solid hsl(var(--border))",
                        background: "hsl(var(--background))",
                      }}
                    />
                    <Area
                      type="monotone"
                      dataKey="revenue"
                      name="Pendapatan"
                      stroke="hsl(var(--primary))"
                      fill="url(#rev)"
                      strokeWidth={2}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          <div className="grid gap-4 lg:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">
                  Pendapatan per Metode
                </CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <Breakdown
                  rows={data.by_method.map((m) => ({
                    label: METHOD_LABELS[m.method] ?? m.method,
                    orders: m.orders,
                    revenue: m.revenue,
                  }))}
                  total={s.revenue_total}
                />
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">Pendapatan per Paket</CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <Breakdown
                  rows={data.by_package.map((p) => ({
                    label: p.package_name,
                    orders: p.orders,
                    revenue: p.revenue,
                  }))}
                  total={s.revenue_total}
                />
              </CardContent>
            </Card>
          </div>
        </>
      )}
    </div>
  );
}

function Breakdown({
  rows,
  total,
}: {
  rows: { label: string; orders: number; revenue: number }[];
  total: number;
}) {
  if (rows.length === 0) {
    return (
      <p className="py-8 text-center text-muted-foreground">
        Belum ada data pada rentang ini.
      </p>
    );
  }
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Nama</TableHead>
          <TableHead className="text-right">Pesanan</TableHead>
          <TableHead className="text-right">Pendapatan</TableHead>
          <TableHead className="text-right">%</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((r) => (
          <TableRow key={r.label}>
            <TableCell className="font-medium">{r.label}</TableCell>
            <TableCell className="text-right">{r.orders}</TableCell>
            <TableCell className="text-right">{formatIDR(r.revenue)}</TableCell>
            <TableCell className="text-right text-muted-foreground">
              {total > 0 ? Math.round((r.revenue / total) * 100) : 0}%
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
  accent = "text-foreground",
}: {
  icon: LucideIcon;
  label: string;
  value: string;
  accent?: string;
}) {
  return (
    <Card>
      <CardContent className="flex items-center gap-4 p-5">
        <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-muted">
          <Icon className={`h-6 w-6 ${accent}`} />
        </div>
        <div className="min-w-0">
          <p className="truncate text-sm text-muted-foreground">{label}</p>
          <p className="truncate text-xl font-bold">{value}</p>
        </div>
      </CardContent>
    </Card>
  );
}
