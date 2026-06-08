-- +goose Up
-- Master data for branch-local radius-api endpoints. NAS records can reference
-- one of these rows instead of hand-entering Radius API URL/key repeatedly.

CREATE TABLE radius_servers (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(120) NOT NULL UNIQUE,
    api_url     VARCHAR(255) NOT NULL,
    api_key     VARCHAR(255) NOT NULL DEFAULT '',
    radius_ip   VARCHAR(128) NOT NULL DEFAULT '',
    coa_port    VARCHAR(10)  NOT NULL DEFAULT '3799',
    description VARCHAR(200) NOT NULL DEFAULT '',
    is_default  BOOLEAN      NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_radius_servers_deleted_at ON radius_servers (deleted_at);

ALTER TABLE nas_hotspot_configs
    ADD COLUMN radius_server_id BIGINT REFERENCES radius_servers(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE nas_hotspot_configs
    DROP COLUMN IF EXISTS radius_server_id;
DROP TABLE IF EXISTS radius_servers;
