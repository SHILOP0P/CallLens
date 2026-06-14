-- +goose Up
DROP INDEX IF EXISTS idx_company_members_active_manager_unique;

-- +goose Down
CREATE UNIQUE INDEX IF NOT EXISTS idx_company_members_active_manager_unique
    ON company_members (user_uuid)
    WHERE role = 'company_manager'
      AND status = 'active';
