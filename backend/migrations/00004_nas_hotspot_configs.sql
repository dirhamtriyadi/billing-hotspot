-- +goose Up
-- Per-router hotspot deployment settings owned by the billing app. The
-- FreeRADIUS `nas` table remains the source of truth for RADIUS clients; this
-- table stores UI/script-generation settings that FreeRADIUS does not need.

CREATE TABLE nas_hotspot_configs (
    id                  BIGSERIAL PRIMARY KEY,
    nasname             VARCHAR(128) NOT NULL UNIQUE,
    radius_ip           VARCHAR(128) NOT NULL DEFAULT '',
    frontend_host       VARCHAR(128) NOT NULL DEFAULT '',
    coa_port            VARCHAR(10)  NOT NULL DEFAULT '3799',
    wan_interface       VARCHAR(60)  NOT NULL DEFAULT 'ether1',
    hotspot_interface   VARCHAR(60)  NOT NULL DEFAULT 'bridge-hotspot',
    bridge_ports        VARCHAR(200) NOT NULL DEFAULT 'wlan1,wlan2',
    hotspot_network     VARCHAR(64)  NOT NULL DEFAULT '10.5.50.0/24',
    hotspot_gateway     VARCHAR(64)  NOT NULL DEFAULT '10.5.50.1',
    hotspot_pool_range  VARCHAR(128) NOT NULL DEFAULT '10.5.50.10-10.5.50.254',
    hotspot_dns         VARCHAR(128) NOT NULL DEFAULT '8.8.8.8,1.1.1.1',
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);
CREATE INDEX idx_nas_hotspot_configs_deleted_at ON nas_hotspot_configs (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS nas_hotspot_configs;
