-- +goose Up
CREATE TABLE refresh_sessions (
    session_uuid UUID PRIMARY KEY,
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL UNIQUE,
    user_agent TEXT NULL,
    ip_address INET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    revoked_reason TEXT NULL,

    CHECK (expires_at > created_at)
);

CREATE INDEX idx_refresh_sessions_user_uuid
    ON refresh_sessions (user_uuid);

CREATE INDEX idx_refresh_sessions_user_active
    ON refresh_sessions (user_uuid, expires_at DESC)
    WHERE revoked_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_refresh_sessions_user_active;
DROP INDEX IF EXISTS idx_refresh_sessions_user_uuid;

DROP TABLE IF EXISTS refresh_sessions;