-- +goose Up
-- +goose StatementBegin
ALTER TABLE companies
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

ALTER TABLE departments
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_companies_not_deleted
    ON companies (company_uuid)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_departments_company_not_deleted
    ON departments (company_uuid, department_uuid)
    WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_departments_company_not_deleted;
DROP INDEX IF EXISTS idx_companies_not_deleted;

ALTER TABLE departments
    DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE companies
    DROP COLUMN IF EXISTS deleted_at;
-- +goose StatementEnd
