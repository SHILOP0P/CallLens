-- +goose Up
CREATE TABLE processing_jobs (
    job_uuid UUID PRIMARY KEY,
    job_type TEXT NOT NULL,
    entity_uuid UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_at TIMESTAMPTZ NULL,
    locked_by TEXT NULL,
    last_error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_processing_jobs_type
        CHECK (job_type IN ('transcribe_call')),
    CONSTRAINT chk_processing_jobs_status
        CHECK (status IN ('pending', 'running', 'done', 'failed')),
    CONSTRAINT chk_processing_jobs_attempts
        CHECK (attempts >= 0),
    CONSTRAINT chk_processing_jobs_max_attempts
        CHECK (max_attempts > 0),
    CONSTRAINT uq_processing_jobs_type_entity
        UNIQUE (job_type, entity_uuid)
);

CREATE INDEX idx_processing_jobs_take_next
    ON processing_jobs (status, available_at, created_at);

CREATE INDEX idx_processing_jobs_entity
    ON processing_jobs (entity_uuid);

-- +goose Down
DROP INDEX IF EXISTS idx_processing_jobs_entity;
DROP INDEX IF EXISTS idx_processing_jobs_take_next;

DROP TABLE IF EXISTS processing_jobs;
