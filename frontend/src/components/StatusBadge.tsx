import { Badge } from "@/components/ui/badge";
import type { OrderStatus, VoucherStatus } from "@/types";

const orderMap: Record<
  OrderStatus,
  { label: string; variant: "default" | "success" | "warning" | "destructive" | "secondary" }
> = {
  pending: { label: "Menunggu", variant: "warning" },
  paid: { label: "Lunas", variant: "success" },
  failed: { label: "Gagal", variant: "destructive" },
  expired: { label: "Kadaluarsa", variant: "secondary" },
  cancelled: { label: "Dibatalkan", variant: "secondary" },
};

export function OrderStatusBadge({ status }: { status: OrderStatus }) {
  const s = orderMap[status] ?? { label: status, variant: "secondary" as const };
  return <Badge variant={s.variant}>{s.label}</Badge>;
}

const voucherMap: Record<
  VoucherStatus,
  { label: string; variant: "default" | "success" | "warning" | "destructive" | "secondary" }
> = {
  unused: { label: "Belum dipakai", variant: "secondary" },
  active: { label: "Aktif", variant: "success" },
  used: { label: "Terpakai", variant: "default" },
  expired: { label: "Kadaluarsa", variant: "warning" },
  disabled: { label: "Nonaktif", variant: "destructive" },
};

export function VoucherStatusBadge({ status }: { status: VoucherStatus }) {
  const s = voucherMap[status] ?? {
    label: status,
    variant: "secondary" as const,
  };
  return <Badge variant={s.variant}>{s.label}</Badge>;
}
