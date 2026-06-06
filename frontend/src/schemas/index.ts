import { z } from "zod";

export const loginSchema = z.object({
  username: z.string().min(1, "Username wajib diisi"),
  password: z.string().min(1, "Password wajib diisi"),
});
export type LoginValues = z.infer<typeof loginSchema>;

export const packageSchema = z.object({
  name: z.string().min(1, "Nama paket wajib diisi").max(120),
  description: z.string().max(1000).optional(),
  price: z.coerce.number().int("Harga harus bilangan bulat").min(0),
  rate_down_kbps: z.coerce.number().int().min(64, "Minimal 64 kbps"),
  rate_up_kbps: z.coerce.number().int().min(64, "Minimal 64 kbps"),
  burst_enabled: z.boolean(),
  validity_value: z.coerce.number().int().min(1, "Minimal 1"),
  validity_unit: z.enum(["minute", "hour", "day", "month"]),
  session_timeout_secs: z.coerce.number().int().min(0),
  data_quota_mb: z.coerce.number().int().min(0),
  simultaneous_use: z.coerce.number().int().min(1).max(100),
  highlight: z.boolean(),
  badge_text: z.string().max(40).optional(),
  color: z.string().max(20).optional(),
  icon: z.string().max(40).optional(),
  sort_order: z.coerce.number().int(),
  is_active: z.boolean(),
});
export type PackageValues = z.infer<typeof packageSchema>;

export const checkoutSchema = z.object({
  customer_name: z.string().min(1, "Nama wajib diisi").max(120),
  customer_phone: z
    .string()
    .min(8, "Nomor HP tidak valid")
    .max(30, "Nomor HP terlalu panjang"),
  customer_email: z
    .string()
    .email("Email tidak valid")
    .max(160)
    .optional()
    .or(z.literal("")),
  payment_method: z.enum(["cash", "midtrans", "xendit", "tripay"], {
    message: "Pilih metode pembayaran",
  }),
});
export type CheckoutValues = z.infer<typeof checkoutSchema>;

export const batchSchema = z.object({
  name: z.string().max(120).optional(),
  package_id: z.coerce.number().int().min(1, "Pilih paket terlebih dahulu"),
  quantity: z.coerce
    .number()
    .int()
    .min(1, "Minimal 1 voucher")
    .max(2000, "Maksimal 2000 voucher"),
  prefix: z
    .string()
    .max(12)
    .regex(/^[a-zA-Z0-9]*$/, "Hanya huruf dan angka")
    .optional(),
  code_length: z.coerce.number().int().min(4).max(20),
});
export type BatchValues = z.infer<typeof batchSchema>;

const nasHotspotConfigSchema = z.object({
  radius_ip: z.string().max(128, "Maksimal 128 karakter").optional(),
  frontend_host: z.string().max(128, "Maksimal 128 karakter").optional(),
  coa_port: z.string().max(10, "Maksimal 10 karakter").optional(),
  wan_interface: z.string().max(60, "Maksimal 60 karakter").optional(),
  hotspot_interface: z.string().max(60, "Maksimal 60 karakter").optional(),
  bridge_ports: z.string().max(200, "Maksimal 200 karakter").optional(),
  hotspot_network: z.string().max(64, "Maksimal 64 karakter").optional(),
  hotspot_gateway: z.string().max(64, "Maksimal 64 karakter").optional(),
  hotspot_pool_range: z.string().max(128, "Maksimal 128 karakter").optional(),
  hotspot_dns: z.string().max(128, "Maksimal 128 karakter").optional(),
});

export const nasSchema = z.object({
  nasname: z
    .string()
    .min(1, "Alamat / identitas router wajib diisi")
    .max(128, "Maksimal 128 karakter"),
  shortname: z.string().max(32, "Maksimal 32 karakter").optional(),
  secret: z
    .string()
    .min(1, "RADIUS secret wajib diisi")
    .max(60, "Maksimal 60 karakter"),
  type: z.string().max(30).optional(),
  description: z.string().max(200, "Maksimal 200 karakter").optional(),
  ports: z
    .union([
      z.literal(""),
      z.coerce
        .number()
        .int("Harus bilangan bulat")
        .min(1, "Minimal 1")
        .max(65535, "Maksimal 65535"),
    ])
    .optional(),
  hotspot_config: nasHotspotConfigSchema,
});
export type NasValues = z.infer<typeof nasSchema>;

export const changePasswordSchema = z
  .object({
    old_password: z.string().min(1, "Password lama wajib diisi"),
    new_password: z.string().min(6, "Minimal 6 karakter").max(72),
    confirm_password: z.string().min(1, "Konfirmasi password wajib diisi"),
  })
  .refine((d) => d.new_password === d.confirm_password, {
    message: "Konfirmasi password tidak cocok",
    path: ["confirm_password"],
  });
export type ChangePasswordValues = z.infer<typeof changePasswordSchema>;
