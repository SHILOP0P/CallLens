-- +goose Up
CREATE TABLE call_transcriptions (
    transcription_uuid UUID PRIMARY KEY,
    call_uuid UUID NOT NULL REFERENCES calls(call_uuid) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'processing',
    text TEXT NULL,
    language TEXT NULL,
    provider TEXT NOT NULL,
    error_message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_call_transcriptions_call_uuid
        UNIQUE (call_uuid),
    CONSTRAINT chk_call_transcriptions_status
        CHECK (status IN ('processing', 'transcribed', 'failed')),
    CONSTRAINT chk_call_transcriptions_status_data
        CHECK (
            (status = 'processing' AND error_message IS NULL)
            OR
            (status = 'transcribed' AND text IS NOT NULL AND error_message IS NULL)
            OR
            (status = 'failed' AND error_message IS NOT NULL)
        )
);

CREATE INDEX idx_call_transcriptions_status_updated_at
    ON call_transcriptions (status, updated_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_call_transcriptions_status_updated_at;

DROP TABLE IF EXISTS call_transcriptions;
