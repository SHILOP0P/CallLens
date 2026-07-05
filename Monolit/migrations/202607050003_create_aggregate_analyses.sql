-- +goose Up
CREATE TABLE aggregate_analyses (
    aggregate_analysis_uuid UUID PRIMARY KEY,
    scope TEXT NOT NULL,
    user_uuid UUID NULL REFERENCES users(user_uuid),
    company_uuid UUID NULL REFERENCES companies(company_uuid),
    department_uuid UUID NULL,
    folder_uuid UUID NULL,
    period_from TIMESTAMPTZ NOT NULL,
    period_to TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NULL,
    source_calls_count INT NOT NULL DEFAULT 0,
    result_json JSONB NULL,
    result_text TEXT NULL,
    error_message TEXT NULL,
    created_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_aggregate_analyses_scope
        CHECK (scope IN ('personal', 'company', 'department', 'folder')),
    CONSTRAINT chk_aggregate_analyses_status
        CHECK (status IN ('pending', 'processing', 'done', 'failed')),
    CONSTRAINT chk_aggregate_analyses_scope_placement
        CHECK (
            (scope = 'personal' AND user_uuid IS NOT NULL AND company_uuid IS NULL AND department_uuid IS NULL AND folder_uuid IS NULL)
            OR (scope = 'company' AND user_uuid IS NULL AND company_uuid IS NOT NULL AND department_uuid IS NULL AND folder_uuid IS NULL)
            OR (scope = 'department' AND user_uuid IS NULL AND company_uuid IS NOT NULL AND department_uuid IS NOT NULL AND folder_uuid IS NULL)
            OR (scope = 'folder' AND folder_uuid IS NOT NULL)
        ),
    CONSTRAINT chk_aggregate_analyses_status_data
        CHECK (
            (status IN ('pending', 'processing') AND error_message IS NULL)
            OR (status = 'done' AND result_json IS NOT NULL AND error_message IS NULL)
            OR (status = 'failed' AND error_message IS NOT NULL)
        )
);

CREATE INDEX idx_aggregate_analyses_created_by
    ON aggregate_analyses (created_by_user_uuid, created_at DESC);

CREATE INDEX idx_aggregate_analyses_scope_subject_period
    ON aggregate_analyses (scope, company_uuid, department_uuid, folder_uuid, period_from, period_to);

CREATE INDEX idx_aggregate_analyses_status_updated
    ON aggregate_analyses (status, updated_at DESC);

CREATE TABLE deep_analysis_usage_counters (
    counter_uuid UUID PRIMARY KEY,
    subject_type TEXT NOT NULL,
    subject_uuid UUID NOT NULL,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    used_count INT NOT NULL DEFAULT 0,
    limit_count INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (subject_type, subject_uuid, period_start),
    CONSTRAINT chk_deep_analysis_usage_subject_type
        CHECK (subject_type IN ('user', 'company')),
    CONSTRAINT chk_deep_analysis_usage_limit
        CHECK (limit_count = 2),
    CONSTRAINT chk_deep_analysis_usage_count
        CHECK (used_count >= 0 AND used_count <= limit_count)
);

-- +goose Down
DROP TABLE IF EXISTS deep_analysis_usage_counters;
DROP TABLE IF EXISTS aggregate_analyses;
