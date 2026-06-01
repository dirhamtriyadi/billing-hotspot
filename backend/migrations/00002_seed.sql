-- +goose Up
-- Default admin (username: admin / password: admin123 — CHANGE IN PRODUCTION),
-- a set of attractive sample packages, and editable business settings.

INSERT INTO users (name, username, email, password, role, is_active)
VALUES (
    'Administrator',
    'admin',
    'admin@hotspot.local',
    '$2a$10$BDJeGx6EbnDCMHdXlx1z/eEWgf8l1aAFHUwONgMGC4NBhsplbrnzG',
    'admin',
    TRUE
);

INSERT INTO packages
    (name, slug, description, price, profile, rate_down_kbps, rate_up_kbps, burst_enabled,
     validity_value, validity_unit, session_timeout_secs, data_quota_mb, simultaneous_use,
     highlight, badge_text, color, icon, sort_order, is_active)
VALUES
    ('Kilat 1 Jam', 'kilat-1-jam', 'Akses cepat untuk kebutuhan sebentar. Cocok untuk meeting online atau download kilat.',
     3000, 'kilat-1-jam', 5000, 2000, TRUE,
     1, 'hour', 0, 0, 1,
     FALSE, '', '#f59e0b', 'zap', 1, TRUE),

    ('Harian Hemat', 'harian-hemat', 'Internet seharian penuh dengan harga bersahabat untuk browsing & sosial media.',
     8000, 'harian-hemat', 10000, 3000, TRUE,
     1, 'day', 0, 0, 1,
     FALSE, '', '#10b981', 'sun', 2, TRUE),

    ('Harian Pro', 'harian-pro', 'Kecepatan ngebut seharian. Streaming 4K, gaming, dan kerja tanpa hambatan.',
     15000, 'harian-pro', 20000, 5000, TRUE,
     1, 'day', 0, 0, 2,
     TRUE, 'TERLARIS', '#2563eb', 'rocket', 3, TRUE),

    ('Mingguan', 'mingguan', 'Hemat untuk pemakaian seminggu penuh. Stabil untuk WFH dan belajar online.',
     50000, 'mingguan', 15000, 5000, TRUE,
     7, 'day', 0, 0, 2,
     FALSE, '', '#0ea5e9', 'calendar', 4, TRUE),

    ('Bulanan Unlimited', 'bulanan-unlimited', 'Paket terbaik! Kecepatan maksimal sebulan penuh tanpa batas kuota.',
     150000, 'bulanan-unlimited', 25000, 10000, TRUE,
     1, 'month', 0, 0, 3,
     TRUE, 'HEMAT', '#7c3aed', 'crown', 5, TRUE);

INSERT INTO settings (key, value) VALUES
    ('site_name', 'WiFi Hotspot'),
    ('site_subtitle', 'Internet cepat, bayar sesuai kebutuhan'),
    ('site_description', 'Pilih paket internet sesuai kebutuhanmu, bayar, dan langsung online.'),
    ('contact_whatsapp', '6281234567890'),
    ('currency', 'IDR'),
    ('enabled_providers', 'midtrans,xendit,tripay'),
    ('enable_cash', 'true'),
    ('tax_percent', '0');

-- +goose Down
DELETE FROM settings WHERE key IN
    ('site_name','site_subtitle','site_description','contact_whatsapp','currency','enabled_providers','enable_cash','tax_percent');
DELETE FROM packages WHERE slug IN
    ('kilat-1-jam','harian-hemat','harian-pro','mingguan','bulanan-unlimited');
DELETE FROM users WHERE username = 'admin';
