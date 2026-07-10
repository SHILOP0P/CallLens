-- +goose Up
ALTER TABLE refresh_sessions
    ADD COLUMN access_version BIGINT NOT NULL DEFAULT 1
        CHECK (access_version > 0);

-- +goose Down
ALTER TABLE refresh_sessions
    DROP COLUMN IF EXISTS access_version;
