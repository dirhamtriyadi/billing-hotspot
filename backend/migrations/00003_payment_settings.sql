-- +goose Up
-- Seed default keys for the admin-managed payment gateway credentials. Values
-- are intentionally blank so the environment variables remain the fallback
-- (see payment.OverlayConfig). ON CONFLICT keeps any value already set.
INSERT INTO settings (key, value) VALUES
    ('payment_default_provider', 'midtrans'),
    ('midtrans_server_key', ''),
    ('midtrans_client_key', ''),
    ('midtrans_is_production', 'false'),
    ('xendit_secret_key', ''),
    ('xendit_callback_token', ''),
    ('tripay_api_key', ''),
    ('tripay_private_key', ''),
    ('tripay_merchant_code', ''),
    ('tripay_is_production', 'false')
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM settings WHERE key IN
    ('payment_default_provider','midtrans_server_key','midtrans_client_key','midtrans_is_production',
     'xendit_secret_key','xendit_callback_token','tripay_api_key','tripay_private_key',
     'tripay_merchant_code','tripay_is_production');
