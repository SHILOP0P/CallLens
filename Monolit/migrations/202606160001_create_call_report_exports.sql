-- +goose Up
CREATE TABLE call_report_exports (
    report_uuid UUID PRIMARY KEY,
    call_uuid UUID NOT NULL REFERENCES calls(call_uuid) ON DELETE CASCADE,
    analysis_uuid UUID NOT NULL REFERENCES call_analyses(analysis_uuid) ON DELETE CASCADE,
    requested_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    format TEXT NOT NULL,
    status TEXT NOT NULL,
    storage_path TEXT,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_call_report_exports_format
        CHECK (format IN ('pdf', 'docx', 'md', 'xlsx')),
    CONSTRAINT chk_call_report_exports_status
        CHECK (status IN ('pending', 'ready', 'failed')),
    CONSTRAINT chk_call_report_exports_storage_for_ready
        CHECK (
            (status = 'ready' AND storage_path IS NOT NULL AND size_bytes > 0)
            OR status <> 'ready'
        )
);

CREATE INDEX idx_call_report_exports_call
    ON call_report_exports (call_uuid, created_at DESC);

CREATE INDEX idx_call_report_exports_expires_at
    ON call_report_exports (expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_call_report_exports_expires_at;
DROP INDEX IF EXISTS idx_call_report_exports_call;
DROP TABLE IF EXISTS call_report_exports;
