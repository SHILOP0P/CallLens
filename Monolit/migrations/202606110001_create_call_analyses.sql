-- +goose Up
CREATE TABLE call_analyses (
    analysis_uuid UUID PRIMARY KEY,
    call_uuid UUID NOT NULL REFERENCES calls(call_uuid) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    provider TEXT NOT NULL,
    model TEXT NULL,
    result_json JSONB NULL,
    result_text TEXT NULL,
    error_message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_call_analyses_call_uuid
        UNIQUE (call_uuid),
    CONSTRAINT chk_call_analyses_status
        CHECK (status IN ('pending', 'processing', 'done', 'failed')),
    CONSTRAINT chk_call_analyses_status_data
        CHECK (
            (status IN ('pending', 'processing') AND error_message IS NULL)
            OR
            (status = 'done' AND result_json IS NOT NULL AND error_message IS NULL)
            OR
            (status = 'failed' AND error_message IS NOT NULL)
        )
);

CREATE INDEX idx_call_analyses_status_updated_at
    ON call_analyses (status, updated_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_call_analyses_status_updated_at;

DROP TABLE IF EXISTS call_analyses;
