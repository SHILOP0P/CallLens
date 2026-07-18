-- +goose Up
-- Duplicate prevention is enforced transactionally by the repository with a per-call/format advisory lock.
-- A schema constraint is intentionally avoided here because historic exports may already contain duplicates.
SELECT 1;

-- +goose Down
SELECT 1;
