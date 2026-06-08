import { Navigate, Route, Routes } from "react-router-dom";
import ProtectedRoute from "@/components/ProtectedRoute";
import "@/lib/hotspot"; // capture ?gw= (hotspot gateway) as early as possible
import AdminLayout from "@/layouts/AdminLayout";
import LandingPage from "@/pages/public/LandingPage";
import CheckoutPage from "@/pages/public/CheckoutPage";
import PaymentStatusPage from "@/pages/public/PaymentStatusPage";
import LoginPage from "@/pages/admin/LoginPage";
import DashboardPage from "@/pages/admin/DashboardPage";
import PackagesPage from "@/pages/admin/PackagesPage";
import VouchersPage from "@/pages/admin/VouchersPage";
import BatchesPage from "@/pages/admin/BatchesPage";
import OrdersPage from "@/pages/admin/OrdersPage";
import SettingsPage from "@/pages/admin/SettingsPage";
import ReportsPage from "@/pages/admin/ReportsPage";
import NasPage from "@/pages/admin/NasPage";
import RadiusServersPage from "@/pages/admin/RadiusServersPage";
import PaymentGatewaysPage from "@/pages/admin/PaymentGatewaysPage";

export default function App() {
  return (
    <Routes>
      {/* Public storefront */}
      <Route path="/" element={<LandingPage />} />
      <Route path="/checkout/:slug" element={<CheckoutPage />} />
      <Route path="/payment/:orderNumber" element={<PaymentStatusPage />} />

      {/* Admin */}
      <Route path="/admin/login" element={<LoginPage />} />
      <Route
        path="/admin"
        element={
          <ProtectedRoute>
            <AdminLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<Navigate to="/admin/dashboard" replace />} />
        <Route path="dashboard" element={<DashboardPage />} />
        <Route path="reports" element={<ReportsPage />} />
        <Route path="packages" element={<PackagesPage />} />
        <Route path="vouchers" element={<VouchersPage />} />
        <Route path="batches" element={<BatchesPage />} />
        <Route path="orders" element={<OrdersPage />} />
        <Route path="nas" element={<NasPage />} />
        <Route path="radius-servers" element={<RadiusServersPage />} />
        <Route path="payment-gateways" element={<PaymentGatewaysPage />} />
        <Route path="settings" element={<SettingsPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
