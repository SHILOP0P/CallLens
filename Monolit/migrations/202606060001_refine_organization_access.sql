-- +goose Up
ALTER TABLE companies
    RENAME COLUMN owner_user_uuid TO manager_user_uuid;

UPDATE company_members
SET role = 'company_manager'
WHERE role IN ('owner', 'company_owner', 'manager', 'admin');

UPDATE company_members
SET role = 'employee'
WHERE role IN ('member', 'user', 'worker', 'employee');

ALTER TABLE company_members
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active',
    ADD CONSTRAINT chk_company_members_role
        CHECK (role IN ('company_manager', 'employee')),
    ADD CONSTRAINT chk_company_members_status
        CHECK (status IN ('active', 'suspended', 'left'));

UPDATE department_members
SET role = 'department_leader'
WHERE role IN ('leader', 'department_leader', 'manager', 'admin');

UPDATE department_members
SET role = 'employee'
WHERE role IN ('member', 'user', 'worker', 'employee');

ALTER TABLE department_members
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active',
    ADD CONSTRAINT chk_department_members_role
        CHECK (role IN ('department_leader', 'employee')),
    ADD CONSTRAINT chk_department_members_status
        CHECK (status IN ('active', 'suspended', 'left'));

CREATE UNIQUE INDEX idx_companies_manager_user_uuid_unique
    ON companies (manager_user_uuid);

CREATE UNIQUE INDEX idx_company_members_active_manager_unique
    ON company_members (user_uuid)
    WHERE role = 'company_manager'
      AND status = 'active';

CREATE INDEX idx_company_members_active_user
    ON company_members (user_uuid)
    WHERE status = 'active';

CREATE INDEX idx_department_members_active_user
    ON department_members (user_uuid)
    WHERE status = 'active';

ALTER TABLE departments
    ADD CONSTRAINT uq_departments_uuid_company
        UNIQUE (department_uuid, company_uuid);

ALTER TABLE calls
    ADD COLUMN visibility_scope TEXT;

UPDATE calls
SET visibility_scope = CASE
    WHEN company_uuid IS NULL THEN 'personal'
    WHEN department_uuid IS NULL THEN 'company'
    ELSE 'department'
END;

ALTER TABLE calls
    ALTER COLUMN visibility_scope SET NOT NULL,
    ALTER COLUMN visibility_scope SET DEFAULT 'personal',
    ADD CONSTRAINT chk_calls_visibility_scope
        CHECK (visibility_scope IN ('personal', 'company', 'department')),
    ADD CONSTRAINT chk_calls_visibility_scope_consistency
        CHECK (
            (visibility_scope = 'personal' AND company_uuid IS NULL AND department_uuid IS NULL)
            OR
            (visibility_scope = 'company' AND company_uuid IS NOT NULL AND department_uuid IS NULL)
            OR
            (visibility_scope = 'department' AND company_uuid IS NOT NULL AND department_uuid IS NOT NULL)
        ),
    ADD CONSTRAINT fk_calls_department_company
        FOREIGN KEY (department_uuid, company_uuid)
        REFERENCES departments (department_uuid, company_uuid)
        ON DELETE RESTRICT;

CREATE INDEX idx_calls_visibility_scope_created_at
    ON calls (visibility_scope, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_calls_visibility_scope_created_at;

ALTER TABLE calls
    DROP CONSTRAINT IF EXISTS fk_calls_department_company,
    DROP CONSTRAINT IF EXISTS chk_calls_visibility_scope_consistency,
    DROP CONSTRAINT IF EXISTS chk_calls_visibility_scope,
    ALTER COLUMN visibility_scope DROP DEFAULT,
    DROP COLUMN IF EXISTS visibility_scope;

ALTER TABLE departments
    DROP CONSTRAINT IF EXISTS uq_departments_uuid_company;

DROP INDEX IF EXISTS idx_department_members_active_user;
DROP INDEX IF EXISTS idx_company_members_active_user;
DROP INDEX IF EXISTS idx_company_members_active_manager_unique;
DROP INDEX IF EXISTS idx_companies_manager_user_uuid_unique;

ALTER TABLE department_members
    DROP CONSTRAINT IF EXISTS chk_department_members_status,
    DROP CONSTRAINT IF EXISTS chk_department_members_role,
    DROP COLUMN IF EXISTS status;

ALTER TABLE company_members
    DROP CONSTRAINT IF EXISTS chk_company_members_status,
    DROP CONSTRAINT IF EXISTS chk_company_members_role,
    DROP COLUMN IF EXISTS status;

ALTER TABLE companies
    RENAME COLUMN manager_user_uuid TO owner_user_uuid;
