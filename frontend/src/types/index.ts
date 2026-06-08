// Shared types mirroring the backend's consistent response envelope and models.

export interface Meta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface FieldError {
  field: string;
  message: string;
  tag?: string;
}

export interface Envelope<T> {
  success: boolean;
  message: string;
  data?: T;
  meta?: Meta;
  error?: { code: string; details?: FieldError[] };
}

export interface UserResponse {
  id: number;
  name: string;
  username: string;
  email: string;
  role: string;
  is_active: boolean;
}

export interface LoginResponse {
  token: string;
  expires_at: string;
  user: UserResponse;
}

export type ValidityUnit = "minute" | "hour" | "day" | "month";

export interface Package {
  id: number;
  name: string;
  slug: string;
  description: string;
  price: number;
  profile: string;
  rate_down_kbps: number;
  rate_up_kbps: number;
  burst_enabled: boolean;
  validity_value: number;
  validity_unit: ValidityUnit;
  session_timeout_secs: number;
  data_quota_mb: number;
  simultaneous_use: number;
  highlight: boolean;
  badge_text: string;
  color: string;
  icon: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PublicPackage {
  id: number;
  name: string;
  slug: string;
  description: string;
  price: number;
  download_mbps: number;
  upload_mbps: number;
  validity: string;
  validity_value: number;
  validity_unit: ValidityUnit;
  data_quota_mb: number;
  highlight: boolean;
  badge_text: string;
  color: string;
  icon: string;
}

export type VoucherStatus =
  | "unused"
  | "active"
  | "used"
  | "expired"
  | "disabled";

export interface Voucher {
  id: number;
  code: string;
  package_id: number;
  package?: Package;
  batch_id?: number;
  order_id?: number;
  status: VoucherStatus;
  profile: string;
  price: number;
  synced_to_radius: boolean;
  activated_at?: string;
  expires_at?: string;
  used_at?: string;
  note: string;
  created_at: string;
}

export interface VoucherBatch {
  id: number;
  name: string;
  package_id: number;
  package?: Package;
  prefix: string;
  quantity: number;
  code_length: number;
  created_by: number;
  vouchers?: Voucher[];
  created_at: string;
}

export type PaymentMethod = "cash" | "midtrans" | "xendit" | "tripay";
export type OrderStatus =
  | "pending"
  | "paid"
  | "failed"
  | "expired"
  | "cancelled";

export interface Order {
  id: number;
  order_number: string;
  package_id: number;
  package?: Package;
  customer_name: string;
  customer_phone: string;
  customer_email: string;
  amount: number;
  payment_method: PaymentMethod;
  status: OrderStatus;
  reference: string;
  payment_url: string;
  payment_token: string;
  qr_string?: string;
  paid_at?: string;
  expires_at?: string;
  voucher_id?: number;
  voucher?: Voucher;
  created_at: string;
}

export interface DashboardStats {
  total_packages: number;
  active_packages: number;
  total_vouchers: number;
  vouchers_by_status: Record<string, number>;
  total_orders: number;
  paid_orders: number;
  pending_orders: number;
  revenue_total: number;
  revenue_today: number;
  revenue_month: number;
  recent_orders: Order[];
}

export interface PublicSettings {
  site_name: string;
  site_subtitle: string;
  site_description: string;
  contact_whatsapp: string;
  currency: string;
  payment_methods: PaymentMethod[];
}

// ─── NAS / RADIUS clients (routers) ──────────────────────────────────────────

export interface NAS {
  id: number;
  nasname: string;
  shortname: string;
  type: string;
  ports?: number | null;
  secret: string;
  server: string;
  community: string;
  description: string;
  hotspot_config: NASHotspotConfig;
}

export interface NASHotspotConfig {
  radius_api_url: string;
  radius_api_key: string;
  radius_ip: string;
  frontend_host: string;
  coa_port: string;
  wan_interface: string;
  hotspot_interface: string;
  bridge_ports: string;
  hotspot_network: string;
  hotspot_gateway: string;
  hotspot_pool_range: string;
  hotspot_dns: string;
}

// ─── Payment gateway settings (secrets masked on read) ───────────────────────

export interface GatewayMidtrans {
  enabled: boolean;
  configured: boolean;
  production: boolean;
  server_key: string;
  client_key: string;
}

export interface GatewayXendit {
  enabled: boolean;
  configured: boolean;
  secret_key: string;
  callback_token: string;
}

export interface GatewayTripay {
  enabled: boolean;
  configured: boolean;
  production: boolean;
  api_key: string;
  private_key: string;
  merchant_code: string;
}

export interface GatewaySettings {
  default_provider: string;
  enable_cash: boolean;
  midtrans: GatewayMidtrans;
  xendit: GatewayXendit;
  tripay: GatewayTripay;
}

// ─── Revenue reporting ───────────────────────────────────────────────────────

export interface ReportSummary {
  start: string;
  end: string;
  revenue_total: number;
  paid_orders: number;
  total_orders: number;
  vouchers_issued: number;
  avg_order_value: number;
}

export interface RevenuePoint {
  date: string;
  revenue: number;
  orders: number;
}

export interface RevenueByMethod {
  method: string;
  revenue: number;
  orders: number;
}

export interface RevenueByPackage {
  package_id: number;
  package_name: string;
  revenue: number;
  orders: number;
}

export interface Report {
  summary: ReportSummary;
  series: RevenuePoint[];
  by_method: RevenueByMethod[];
  by_package: RevenueByPackage[];
}
