-- +goose Up
ALTER TABLE processing_jobs
    ADD COLUMN transcription_mode TEXT NOT NULL DEFAULT 'standard',
    ADD CONSTRAINT chk_processing_jobs_transcription_mode
        CHECK (transcription_mode IN ('standard', 'diarized'));

-- +goose Down
ALTER TABLE processing_jobs
    DROP CONSTRAINT chk_processing_jobs_transcription_mode,
    DROP COLUMN transcription_mode;
