-- +goose Up
ALTER TABLE call_transcriptions
    ADD COLUMN segments JSONB NULL;

-- +goose Down
ALTER TABLE call_transcriptions
    DROP COLUMN IF EXISTS segments;
