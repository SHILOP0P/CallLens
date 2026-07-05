-- +goose Up
CREATE TABLE aggregate_report_exports (
    report_uuid UUID PRIMARY KEY,
    aggregate_analysis_uuid UUID NOT NULL REFERENCES aggregate_analyses(aggregate_analysis_uuid) ON DELETE CASCADE,
    requested_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    format TEXT NOT NULL,
    status TEXT NOT NULL,
    storage_path TEXT NULL,
    file_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    error_message TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT chk_aggregate_report_exports_format
        CHECK (format IN ('pdf', 'docx', 'md', 'xlsx')),
    CONSTRAINT chk_aggregate_report_exports_status
        CHECK (status IN ('pending', 'ready', 'failed')),
    CONSTRAINT chk_aggregate_report_exports_storage_for_ready
        CHECK (status <> 'ready' OR storage_path IS NOT NULL)
);

CREATE INDEX idx_aggregate_report_exports_analysis
    ON aggregate_report_exports (aggregate_analysis_uuid, created_at DESC);

CREATE INDEX idx_aggregate_report_exports_requested_by
    ON aggregate_report_exports (requested_by_user_uuid, created_at DESC);

CREATE INDEX idx_aggregate_report_exports_expires_at
    ON aggregate_report_exports (expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_aggregate_report_exports_expires_at;
DROP INDEX IF EXISTS idx_aggregate_report_exports_requested_by;
DROP INDEX IF EXISTS idx_aggregate_report_exports_analysis;
DROP TABLE IF EXISTS aggregate_report_exports;
