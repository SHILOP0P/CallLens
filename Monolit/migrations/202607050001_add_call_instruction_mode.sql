-- +goose Up
ALTER TABLE calls
    ADD COLUMN IF NOT EXISTS skip_custom_instructions BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE calls
    DROP COLUMN IF EXISTS skip_custom_instructions;
