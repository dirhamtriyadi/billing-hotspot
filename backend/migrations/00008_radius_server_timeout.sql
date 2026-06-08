-- +goose Up
ALTER TABLE radius_servers
    ADD COLUMN timeout VARCHAR(20) NOT NULL DEFAULT '10s';

-- +goose Down
ALTER TABLE radius_servers
    DROP COLUMN IF EXISTS timeout;
