-- +goose Up
ALTER TABLE companies ADD COLUMN tag TEXT;

UPDATE companies
SET tag = '@' || company_uuid::TEXT
WHERE tag IS NULL;

ALTER TABLE companies ALTER COLUMN tag SET NOT NULL;

CREATE UNIQUE INDEX idx_companies_active_tag_unique
    ON companies (LOWER(tag))
    WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_companies_active_tag_unique;
ALTER TABLE companies DROP COLUMN IF EXISTS tag;
