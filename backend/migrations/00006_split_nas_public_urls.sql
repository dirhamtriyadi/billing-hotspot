-- +goose Up
-- Split the previous ambiguous frontend_host into explicit public URLs:
-- frontend_url is the storefront opened by hotspot users, backend_url is the
-- API host used by the router to fetch login.html. The old column is retained
-- for backward-compatible fallback.

ALTER TABLE nas_hotspot_configs
    ADD COLUMN frontend_url VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN backend_url  VARCHAR(255) NOT NULL DEFAULT '';

UPDATE nas_hotspot_configs
SET
    frontend_url = frontend_host,
    backend_url = frontend_host
WHERE frontend_host <> '';

-- +goose Down
ALTER TABLE nas_hotspot_configs
    DROP COLUMN IF EXISTS backend_url,
    DROP COLUMN IF EXISTS frontend_url;
