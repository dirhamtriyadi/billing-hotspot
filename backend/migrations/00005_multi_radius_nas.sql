-- +goose Up
-- Keep a local copy of NAS metadata and the branch-local radius-api endpoint so
-- the billing backend can provision vouchers/profiles to multiple FreeRADIUS
-- servers. Existing rows keep using the environment RADIUS_API_URL/API_KEY when
-- these endpoint columns are empty.

ALTER TABLE nas_hotspot_configs
    ADD COLUMN shortname      VARCHAR(32)  NOT NULL DEFAULT '',
    ADD COLUMN type           VARCHAR(30)  NOT NULL DEFAULT 'mikrotik',
    ADD COLUMN ports          INTEGER,
    ADD COLUMN secret         VARCHAR(60)  NOT NULL DEFAULT '',
    ADD COLUMN description    VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN radius_api_url VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN radius_api_key VARCHAR(255) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE nas_hotspot_configs
    DROP COLUMN IF EXISTS radius_api_key,
    DROP COLUMN IF EXISTS radius_api_url,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS secret,
    DROP COLUMN IF EXISTS ports,
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS shortname;
