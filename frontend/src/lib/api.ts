import axios, {
  AxiosError,
  type AxiosRequestConfig,
  type AxiosInstance,
} from "axios";
import type {
  DashboardStats,
  Envelope,
  FieldError,
  GatewaySettings,
  LoginResponse,
  Meta,
  NAS,
  Order,
  Package,
  PublicPackage,
  PublicSettings,
  RadiusServer,
  Report,
  UserResponse,
  Voucher,
  VoucherBatch,
} from "@/types";

const TOKEN_KEY = "bh_auth_token";

export const tokenStore = {
  get: () => localStorage.getItem(TOKEN_KEY),
  set: (t: string) => localStorage.setItem(TOKEN_KEY, t),
  clear: () => localStorage.removeItem(TOKEN_KEY),
};

// API version prefix. Kept in code (not in the env var) so VITE_API_BASE_URL
// stays a pure host ("http://host:8080") and the version can be bumped here —
// or run v1/v2 side by side — without touching deploy config.
const API_VERSION = "/api/v1";

// VITE_API_BASE_URL is the backend HOST only (no version path). A trailing
// slash is trimmed so concatenation stays clean.
const apiHost = (
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080"
).replace(/\/+$/, "");

const baseURL = `${apiHost}${API_VERSION}`;

const client: AxiosInstance = axios.create({ baseURL });

client.interceptors.request.use((config) => {
  const token = tokenStore.get();
  if (token) {
    config.headers = config.headers ?? {};
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

client.interceptors.response.use(
  (res) => res,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      tokenStore.clear();
      if (
        typeof window !== "undefined" &&
        window.location.pathname.startsWith("/admin") &&
        !window.location.pathname.includes("/login")
      ) {
        window.location.assign("/admin/login");
      }
    }
    return Promise.reject(error);
  },
);

/** Normalised API error carrying the backend's error code and field details. */
export class ApiError extends Error {
  code: string;
  status: number;
  details?: FieldError[];
  constructor(
    message: string,
    code: string,
    status: number,
    details?: FieldError[],
  ) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
    this.details = details;
  }
}

function toApiError(error: unknown): ApiError {
  if (axios.isAxiosError(error)) {
    const env = error.response?.data as Envelope<unknown> | undefined;
    return new ApiError(
      env?.message || error.message || "Network error",
      env?.error?.code || "NETWORK_ERROR",
      error.response?.status || 0,
      env?.error?.details,
    );
  }
  return new ApiError("Unexpected error", "UNKNOWN", 0);
}

async function request<T>(
  config: AxiosRequestConfig,
): Promise<{ data: T; meta?: Meta }> {
  try {
    const res = await client.request<Envelope<T>>(config);
    return { data: res.data.data as T, meta: res.data.meta };
  } catch (error) {
    throw toApiError(error);
  }
}

const unwrap = async <T>(config: AxiosRequestConfig): Promise<T> =>
  (await request<T>(config)).data;

export interface ListResult<T> {
  data: T[];
  meta?: Meta;
}

async function listRequest<T>(
  url: string,
  params?: Record<string, unknown>,
): Promise<ListResult<T>> {
  const { data, meta } = await request<T[]>({ url, method: "GET", params });
  return { data: data ?? [], meta };
}

// ─── Endpoint groups ────────────────────────────────────────────────────────

export const api = {
  auth: {
    login: (body: { username: string; password: string }) =>
      unwrap<LoginResponse>({ url: "/auth/login", method: "POST", data: body }),
    me: () => unwrap<UserResponse>({ url: "/auth/me", method: "GET" }),
    changePassword: (body: { old_password: string; new_password: string }) =>
      unwrap<null>({
        url: "/auth/change-password",
        method: "POST",
        data: body,
      }),
  },

  public: {
    packages: () =>
      unwrap<PublicPackage[]>({ url: "/public/packages", method: "GET" }),
    settings: () =>
      unwrap<PublicSettings>({ url: "/public/settings", method: "GET" }),
    checkout: (body: unknown) =>
      unwrap<Order>({ url: "/public/checkout", method: "POST", data: body }),
    orderStatus: (orderNumber: string) =>
      unwrap<Order>({ url: `/public/orders/${orderNumber}`, method: "GET" }),
  },

  dashboard: {
    stats: () =>
      unwrap<DashboardStats>({ url: "/dashboard/stats", method: "GET" }),
  },

  packages: {
    list: (params?: Record<string, unknown>) =>
      listRequest<Package>("/packages", params),
    get: (id: number) => unwrap<Package>({ url: `/packages/${id}` }),
    create: (body: unknown) =>
      unwrap<Package>({ url: "/packages", method: "POST", data: body }),
    update: (id: number, body: unknown) =>
      unwrap<Package>({ url: `/packages/${id}`, method: "PUT", data: body }),
    remove: (id: number) =>
      unwrap<null>({ url: `/packages/${id}`, method: "DELETE" }),
  },

  vouchers: {
    list: (params?: Record<string, unknown>) =>
      listRequest<Voucher>("/vouchers", params),
    get: (id: number) => unwrap<Voucher>({ url: `/vouchers/${id}` }),
    setStatus: (id: number, status: "active" | "disabled") =>
      unwrap<Voucher>({
        url: `/vouchers/${id}/status`,
        method: "PATCH",
        data: { status },
      }),
    remove: (id: number) =>
      unwrap<null>({ url: `/vouchers/${id}`, method: "DELETE" }),
  },

  batches: {
    list: (params?: Record<string, unknown>) =>
      listRequest<VoucherBatch>("/batches", params),
    get: (id: number) => unwrap<VoucherBatch>({ url: `/batches/${id}` }),
    create: (body: unknown) =>
      unwrap<VoucherBatch>({ url: "/batches", method: "POST", data: body }),
    remove: (id: number) =>
      unwrap<null>({ url: `/batches/${id}`, method: "DELETE" }),
  },

  orders: {
    list: (params?: Record<string, unknown>) =>
      listRequest<Order>("/orders", params),
    get: (id: number) => unwrap<Order>({ url: `/orders/${id}` }),
    confirmCash: (id: number) =>
      unwrap<Order>({ url: `/orders/${id}/confirm-cash`, method: "POST" }),
    markPaid: (id: number) =>
      unwrap<Order>({ url: `/orders/${id}/mark-paid`, method: "POST" }),
  },

  settings: {
    get: () => unwrap<Record<string, string>>({ url: "/settings" }),
    update: (body: Record<string, string>) =>
      unwrap<Record<string, string>>({
        url: "/settings",
        method: "PUT",
        data: body,
      }),
  },

  paymentGateways: {
    get: () => unwrap<GatewaySettings>({ url: "/payment-gateways" }),
    update: (body: unknown) =>
      unwrap<GatewaySettings>({
        url: "/payment-gateways",
        method: "PUT",
        data: body,
      }),
  },

  nas: {
    list: () => unwrap<NAS[]>({ url: "/nas" }),
    upsert: (body: unknown) =>
      unwrap<NAS>({ url: "/nas", method: "POST", data: body }),
    remove: (id: number) =>
      unwrap<null>({ url: `/nas/${id}`, method: "DELETE" }),
  },

  radiusServers: {
    list: () => unwrap<RadiusServer[]>({ url: "/radius-servers" }),
    create: (body: unknown) =>
      unwrap<RadiusServer>({
        url: "/radius-servers",
        method: "POST",
        data: body,
      }),
    update: (id: number, body: unknown) =>
      unwrap<RadiusServer>({
        url: `/radius-servers/${id}`,
        method: "PUT",
        data: body,
      }),
    remove: (id: number) =>
      unwrap<null>({ url: `/radius-servers/${id}`, method: "DELETE" }),
  },

  reports: {
    revenue: (params?: Record<string, unknown>) =>
      unwrap<Report>({ url: "/reports/revenue", params }),
    // download streams the CSV as a Blob using the authenticated client so the
    // caller can trigger a browser download.
    download: async (params: { start: string; end: string }) => {
      const res = await client.get<Blob>("/reports/export", {
        params,
        responseType: "blob",
      });
      return res.data;
    },
  },
};
