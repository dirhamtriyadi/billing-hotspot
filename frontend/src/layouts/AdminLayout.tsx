import { useState } from "react";
import { Link, NavLink, Outlet, useNavigate } from "react-router-dom";
import {
  LayoutDashboard,
  Package,
  Ticket,
  Layers,
  ShoppingCart,
  Settings,
  LogOut,
  Menu,
  Wifi,
  Router,
  Server,
  BarChart3,
  CreditCard,
  X,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/context/auth";
import { Button } from "@/components/ui/button";

const nav = [
  { to: "/admin/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/admin/reports", label: "Laporan", icon: BarChart3 },
  { to: "/admin/packages", label: "Paket", icon: Package },
  { to: "/admin/vouchers", label: "Voucher", icon: Ticket },
  { to: "/admin/batches", label: "Batch Voucher", icon: Layers },
  { to: "/admin/orders", label: "Pesanan", icon: ShoppingCart },
  { to: "/admin/nas", label: "Router (NAS)", icon: Router },
  { to: "/admin/radius-servers", label: "Radius Server", icon: Server },
  { to: "/admin/payment-gateways", label: "Payment Gateway", icon: CreditCard },
  { to: "/admin/settings", label: "Pengaturan", icon: Settings },
];

export default function AdminLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);

  const handleLogout = () => {
    logout();
    navigate("/admin/login");
  };

  return (
    <div className="min-h-screen bg-muted/30">
      {/* Sidebar */}
      <aside
        className={cn(
          "fixed inset-y-0 left-0 z-40 w-64 transform border-r bg-background transition-transform lg:translate-x-0",
          open ? "translate-x-0" : "-translate-x-full",
        )}
      >
        <div className="flex h-16 items-center gap-2 border-b px-6">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Wifi className="h-5 w-5" />
          </div>
          <span className="text-lg font-bold">Hotspot Admin</span>
          <button
            className="ml-auto lg:hidden"
            onClick={() => setOpen(false)}
            aria-label="Tutup menu"
          >
            <X className="h-5 w-5" />
          </button>
        </div>
        <nav className="space-y-1 p-4">
          {nav.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              onClick={() => setOpen(false)}
              className={({ isActive }) =>
                cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                )
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>

      {open && (
        <div
          className="fixed inset-0 z-30 bg-black/40 lg:hidden"
          onClick={() => setOpen(false)}
        />
      )}

      {/* Main */}
      <div className="lg:pl-64">
        <header className="sticky top-0 z-20 flex h-16 items-center gap-4 border-b bg-background/80 px-4 backdrop-blur lg:px-8">
          <button
            className="lg:hidden"
            onClick={() => setOpen(true)}
            aria-label="Buka menu"
          >
            <Menu className="h-5 w-5" />
          </button>
          <div className="ml-auto flex items-center gap-3">
            <Link
              to="/"
              target="_blank"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Lihat storefront →
            </Link>
            <div className="text-right">
              <div className="text-sm font-medium">{user?.name}</div>
              <div className="text-xs capitalize text-muted-foreground">
                {user?.role}
              </div>
            </div>
            <Button variant="outline" size="icon" onClick={handleLogout}>
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        </header>

        <main className="p-4 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
