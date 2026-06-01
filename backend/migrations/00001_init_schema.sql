-- +goose Up
-- Billing database schema (PostgreSQL). Mirrors internal/models.

CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(120) NOT NULL,
    username    VARCHAR(60)  NOT NULL,
    email       VARCHAR(160) NOT NULL DEFAULT '',
    password    VARCHAR(255) NOT NULL,
    role        VARCHAR(20)  NOT NULL DEFAULT 'operator',
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_users_username ON users (username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uq_users_email ON users (email) WHERE deleted_at IS NULL AND email <> '';
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

CREATE TABLE packages (
    id                   BIGSERIAL PRIMARY KEY,
    name                 VARCHAR(120) NOT NULL,
    slug                 VARCHAR(140) NOT NULL,
    description          TEXT NOT NULL DEFAULT '',
    price                BIGINT NOT NULL,
    profile              VARCHAR(80) NOT NULL,
    rate_down_kbps       INTEGER NOT NULL,
    rate_up_kbps         INTEGER NOT NULL,
    burst_enabled        BOOLEAN NOT NULL DEFAULT FALSE,
    validity_value       INTEGER NOT NULL DEFAULT 1,
    validity_unit        VARCHAR(10) NOT NULL DEFAULT 'day',
    session_timeout_secs INTEGER NOT NULL DEFAULT 0,
    data_quota_mb        BIGINT NOT NULL DEFAULT 0,
    simultaneous_use     INTEGER NOT NULL DEFAULT 1,
    highlight            BOOLEAN NOT NULL DEFAULT FALSE,
    badge_text           VARCHAR(40) NOT NULL DEFAULT '',
    color                VARCHAR(20) NOT NULL DEFAULT '#2563eb',
    icon                 VARCHAR(40) NOT NULL DEFAULT 'wifi',
    sort_order           INTEGER NOT NULL DEFAULT 0,
    is_active            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at           TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_packages_slug ON packages (slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_packages_deleted_at ON packages (deleted_at);
CREATE INDEX idx_packages_active_sort ON packages (is_active, sort_order);

CREATE TABLE voucher_batches (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(120) NOT NULL,
    package_id  BIGINT NOT NULL REFERENCES packages (id) ON DELETE RESTRICT,
    prefix      VARCHAR(12) NOT NULL DEFAULT '',
    quantity    INTEGER NOT NULL,
    code_length INTEGER NOT NULL DEFAULT 8,
    created_by  BIGINT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_voucher_batches_package_id ON voucher_batches (package_id);
CREATE INDEX idx_voucher_batches_deleted_at ON voucher_batches (deleted_at);

CREATE TABLE orders (
    id             BIGSERIAL PRIMARY KEY,
    order_number   VARCHAR(40) NOT NULL,
    package_id     BIGINT NOT NULL REFERENCES packages (id) ON DELETE RESTRICT,
    customer_name  VARCHAR(120) NOT NULL DEFAULT '',
    customer_phone VARCHAR(30) NOT NULL DEFAULT '',
    customer_email VARCHAR(160) NOT NULL DEFAULT '',
    amount         BIGINT NOT NULL,
    payment_method VARCHAR(20) NOT NULL,
    status         VARCHAR(20) NOT NULL DEFAULT 'pending',
    reference      VARCHAR(120) NOT NULL DEFAULT '',
    payment_url    TEXT NOT NULL DEFAULT '',
    payment_token  VARCHAR(255) NOT NULL DEFAULT '',
    qr_string      TEXT NOT NULL DEFAULT '',
    raw_response   TEXT NOT NULL DEFAULT '',
    paid_at        TIMESTAMPTZ,
    expires_at     TIMESTAMPTZ,
    voucher_id     BIGINT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_orders_order_number ON orders (order_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_orders_package_id ON orders (package_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_payment_method ON orders (payment_method);
CREATE INDEX idx_orders_reference ON orders (reference);
CREATE INDEX idx_orders_voucher_id ON orders (voucher_id);
CREATE INDEX idx_orders_deleted_at ON orders (deleted_at);

CREATE TABLE vouchers (
    id               BIGSERIAL PRIMARY KEY,
    code             VARCHAR(40) NOT NULL,
    package_id       BIGINT NOT NULL REFERENCES packages (id) ON DELETE RESTRICT,
    batch_id         BIGINT REFERENCES voucher_batches (id) ON DELETE CASCADE,
    order_id         BIGINT REFERENCES orders (id) ON DELETE SET NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'unused',
    profile          VARCHAR(80) NOT NULL,
    price            BIGINT NOT NULL,
    synced_to_radius BOOLEAN NOT NULL DEFAULT FALSE,
    activated_at     TIMESTAMPTZ,
    expires_at       TIMESTAMPTZ,
    used_at          TIMESTAMPTZ,
    note             VARCHAR(255) NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);
CREATE UNIQUE INDEX uq_vouchers_code ON vouchers (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_vouchers_package_id ON vouchers (package_id);
CREATE INDEX idx_vouchers_batch_id ON vouchers (batch_id);
CREATE INDEX idx_vouchers_order_id ON vouchers (order_id);
CREATE INDEX idx_vouchers_status ON vouchers (status);
CREATE INDEX idx_vouchers_deleted_at ON vouchers (deleted_at);

-- Close the orders <-> vouchers cycle now that both tables exist.
ALTER TABLE orders
    ADD CONSTRAINT fk_orders_voucher
    FOREIGN KEY (voucher_id) REFERENCES vouchers (id) ON DELETE SET NULL;

CREATE TABLE payment_logs (
    id         BIGSERIAL PRIMARY KEY,
    order_id   BIGINT,
    provider   VARCHAR(20) NOT NULL DEFAULT '',
    event      VARCHAR(60) NOT NULL DEFAULT '',
    reference  VARCHAR(120) NOT NULL DEFAULT '',
    status     VARCHAR(30) NOT NULL DEFAULT '',
    signature  VARCHAR(255) NOT NULL DEFAULT '',
    valid      BOOLEAN NOT NULL DEFAULT FALSE,
    payload    TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_payment_logs_order_id ON payment_logs (order_id);
CREATE INDEX idx_payment_logs_provider ON payment_logs (provider);
CREATE INDEX idx_payment_logs_reference ON payment_logs (reference);
CREATE INDEX idx_payment_logs_deleted_at ON payment_logs (deleted_at);

CREATE TABLE settings (
    key   VARCHAR(80) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);

-- +goose Down
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS payment_logs;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_voucher;
DROP TABLE IF EXISTS vouchers;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS voucher_batches;
DROP TABLE IF EXISTS packages;
DROP TABLE IF EXISTS users;
