-- +goose Up
CREATE TABLE calls (
                       call_uuid UUID PRIMARY KEY,
                       title TEXT NOT NULL,
                       status TEXT NOT NULL,
                       audio_path TEXT NOT NULL,
                       original_filename TEXT NOT NULL,
                       mime_type TEXT NOT NULL,
                       size_bytes BIGINT NOT NULL,
                       duration_seconds INTEGER NOT NULL DEFAULT 0,
                       created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_calls_created_at
    ON calls (created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_calls_created_at;

DROP TABLE IF EXISTS calls;
