-- +goose Up
DROP INDEX IF EXISTS idx_companies_manager_user_uuid_unique;
DROP INDEX IF EXISTS idx_company_members_active_manager_unique;

-- +goose Down
CREATE UNIQUE INDEX IF NOT EXISTS idx_company_members_active_manager_unique
    ON company_members (user_uuid)
    WHERE role = 'company_manager'
      AND status = 'active';

CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_manager_user_uuid_unique
    ON companies (manager_user_uuid);
