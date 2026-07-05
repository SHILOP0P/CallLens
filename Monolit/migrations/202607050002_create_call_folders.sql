-- +goose Up
CREATE TABLE call_folders (
    folder_uuid UUID PRIMARY KEY,
    scope TEXT NOT NULL,
    user_uuid UUID NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    company_uuid UUID NULL REFERENCES companies(company_uuid) ON DELETE CASCADE,
    department_uuid UUID NULL REFERENCES departments(department_uuid) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NULL,
    color TEXT NULL,
    created_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT chk_call_folders_scope
        CHECK (scope IN ('personal', 'company', 'department')),
    CONSTRAINT chk_call_folders_scope_placement
        CHECK (
            (scope = 'personal' AND user_uuid IS NOT NULL AND company_uuid IS NULL AND department_uuid IS NULL)
            OR (scope = 'company' AND user_uuid IS NULL AND company_uuid IS NOT NULL AND department_uuid IS NULL)
            OR (scope = 'department' AND user_uuid IS NULL AND company_uuid IS NOT NULL AND department_uuid IS NOT NULL)
        )
);

CREATE INDEX idx_call_folders_personal_active
    ON call_folders (user_uuid, created_at DESC)
    WHERE deleted_at IS NULL AND scope = 'personal';

CREATE INDEX idx_call_folders_company_active
    ON call_folders (company_uuid, created_at DESC)
    WHERE deleted_at IS NULL AND scope = 'company';

CREATE INDEX idx_call_folders_department_active
    ON call_folders (department_uuid, created_at DESC)
    WHERE deleted_at IS NULL AND scope = 'department';

CREATE TABLE call_folder_assignments (
    folder_uuid UUID NOT NULL REFERENCES call_folders(folder_uuid) ON DELETE CASCADE,
    call_uuid UUID NOT NULL REFERENCES calls(call_uuid) ON DELETE CASCADE,
    assigned_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (folder_uuid, call_uuid)
);

CREATE INDEX idx_call_folder_assignments_call_uuid
    ON call_folder_assignments (call_uuid);

CREATE INDEX idx_call_folder_assignments_folder_created_at
    ON call_folder_assignments (folder_uuid, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS call_folder_assignments;
DROP TABLE IF EXISTS call_folders;
