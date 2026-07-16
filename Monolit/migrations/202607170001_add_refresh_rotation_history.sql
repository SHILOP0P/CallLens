-- +goose Up
ALTER TABLE refresh_sessions
    ADD COLUMN previous_refresh_token_hash TEXT NULL,
    ADD COLUMN rotated_at TIMESTAMPTZ NULL;

CREATE INDEX idx_refresh_sessions_previous_token
    ON refresh_sessions (previous_refresh_token_hash)
    WHERE previous_refresh_token_hash IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_refresh_sessions_previous_token;

ALTER TABLE refresh_sessions
    DROP COLUMN IF EXISTS rotated_at,
    DROP COLUMN IF EXISTS previous_refresh_token_hash;
