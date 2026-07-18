-- +goose Up
UPDATE companies
SET tag = '@' || company_uuid::text
WHERE tag IS NULL OR BTRIM(tag) = '';

-- +goose Down
-- Existing tags are public identifiers; do not erase them during rollback.
