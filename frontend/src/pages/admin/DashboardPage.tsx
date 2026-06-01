import { useQuery } from "@tanstack/react-query";
import {
  Banknote,
  CalendarDays,
  Package,
  ShoppingCart,
  Ticket,
  TrendingUp,
  type LucideIcon,
} from "lucide-react";
import { api } from "@/lib/api";
import { formatDateTime, formatIDR } from "@/lib/format";
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
import { OrderStatusBadge } from "@/components/StatusBadge";

export default function DashboardPage() {
  const { data, isLoading } = useQuery({
    queryKey: ["dashboard"],
    queryFn: api.dashboard.stats,
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">Ringkasan bisnis hotspot kamu.</p>
      </div>

      {isLoading || !data ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Skeleton key={i} className="h-28 rounded-xl" />
          ))}
        </div>
      ) : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              icon={Banknote}
              label="Pendapatan Total"
              value={formatIDR(data.revenue_total)}
              accent="text-emerald-500"
            />
            <StatCard
              icon={TrendingUp}
              label="Pendapatan Hari Ini"
              value={formatIDR(data.revenue_today)}
              accent="text-primary"
            />
            <StatCard
              icon={CalendarDays}
              label="Pendapatan Bulan Ini"
              value={formatIDR(data.revenue_month)}
              accent="text-violet-500"
            />
            <StatCard
              icon={ShoppingCart}
              label="Pesanan"
              value={`${data.paid_orders} lunas / ${data.total_orders}`}
              accent="text-amber-500"
            />
          </div>

          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard
              icon={Ticket}
              label="Total Voucher"
              value={String(data.total_vouchers)}
            />
            <StatCard
              icon={Ticket}
              label="Voucher Aktif"
              value={String(data.vouchers_by_status?.active ?? 0)}
              accent="text-emerald-500"
            />
            <StatCard
              icon={Package}
              label="Paket Aktif"
              value={`${data.active_packages} / ${data.total_packages}`}
            />
            <StatCard
              icon={ShoppingCart}
              label="Pesanan Pending"
              value={String(data.pending_orders)}
              accent="text-amber-500"
            />
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Pesanan Terbaru</CardTitle>
            </CardHeader>
            <CardContent>
              {data.recent_orders && data.recent_orders.length > 0 ? (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>No. Pesanan</TableHead>
                      <TableHead>Pelanggan</TableHead>
                      <TableHead>Paket</TableHead>
                      <TableHead>Total</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Waktu</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {data.recent_orders.map((o) => (
                      <TableRow key={o.id}>
                        <TableCell className="font-mono text-xs">
                          {o.order_number}
                        </TableCell>
                        <TableCell>{o.customer_name}</TableCell>
                        <TableCell>{o.package?.name ?? "-"}</TableCell>
                        <TableCell>{formatIDR(o.amount)}</TableCell>
                        <TableCell>
                          <OrderStatusBadge status={o.status} />
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {formatDateTime(o.created_at)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              ) : (
                <p className="py-8 text-center text-muted-foreground">
                  Belum ada pesanan.
                </p>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
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
