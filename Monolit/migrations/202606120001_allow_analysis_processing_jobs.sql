-- +goose Up
ALTER TABLE processing_jobs
    DROP CONSTRAINT chk_processing_jobs_type;

ALTER TABLE processing_jobs
    ADD CONSTRAINT chk_processing_jobs_type
        CHECK (job_type IN ('transcribe_call', 'analyze_call'));

-- +goose Down
ALTER TABLE processing_jobs
    DROP CONSTRAINT chk_processing_jobs_type;

ALTER TABLE processing_jobs
    ADD CONSTRAINT chk_processing_jobs_type
        CHECK (job_type IN ('transcribe_call'));
