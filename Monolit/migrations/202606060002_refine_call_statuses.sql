-- +goose Up
UPDATE calls
SET status = 'analyzed'
WHERE status = 'done';

ALTER TABLE calls
    ADD CONSTRAINT chk_calls_status
        CHECK (status IN ('new', 'processing', 'transcribed', 'analyzed', 'failed'));

-- +goose Down
ALTER TABLE calls
    DROP CONSTRAINT IF EXISTS chk_calls_status;

UPDATE calls
SET status = 'done'
WHERE status IN ('transcribed', 'analyzed');
