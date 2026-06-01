import {
  Calendar,
  Crown,
  Gauge,
  Rocket,
  Star,
  Sun,
  Wifi,
  Zap,
  type LucideIcon,
} from "lucide-react";

const icons: Record<string, LucideIcon> = {
  wifi: Wifi,
  zap: Zap,
  sun: Sun,
  rocket: Rocket,
  calendar: Calendar,
  crown: Crown,
  gauge: Gauge,
  star: Star,
};

export function PackageIcon({
  name,
  className,
}: {
  name: string;
  className?: string;
}) {
  const Icon = icons[name] ?? Wifi;
  return <Icon className={className} />;
}
