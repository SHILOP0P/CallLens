-- +goose Up
CREATE TABLE analysis_instructions (
    instruction_uuid UUID PRIMARY KEY,
    scope TEXT NOT NULL,
    user_uuid UUID NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    company_uuid UUID NULL REFERENCES companies(company_uuid) ON DELETE CASCADE,
    department_uuid UUID NULL,
    title TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    content_sha256 TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_analysis_instructions_scope
        CHECK (scope IN ('personal', 'company', 'department')),
    CONSTRAINT chk_analysis_instructions_scope_consistency
        CHECK (
            (
                scope = 'personal'
                AND user_uuid IS NOT NULL
                AND company_uuid IS NULL
                AND department_uuid IS NULL
            )
            OR
            (
                scope = 'company'
                AND user_uuid IS NULL
                AND company_uuid IS NOT NULL
                AND department_uuid IS NULL
            )
            OR
            (
                scope = 'department'
                AND user_uuid IS NULL
                AND company_uuid IS NOT NULL
                AND department_uuid IS NOT NULL
            )
        ),
    CONSTRAINT chk_analysis_instructions_size
        CHECK (size_bytes > 0),
    CONSTRAINT chk_analysis_instructions_sort_order
        CHECK (sort_order >= 0),
    CONSTRAINT fk_analysis_instructions_department_company
        FOREIGN KEY (department_uuid, company_uuid)
        REFERENCES departments (department_uuid, company_uuid)
        ON DELETE CASCADE
);

CREATE INDEX idx_analysis_instructions_personal_active
    ON analysis_instructions (user_uuid, sort_order, created_at)
    WHERE scope = 'personal'
      AND is_active = true;

CREATE INDEX idx_analysis_instructions_company_active
    ON analysis_instructions (company_uuid, sort_order, created_at)
    WHERE scope = 'company'
      AND is_active = true;

CREATE INDEX idx_analysis_instructions_department_active
    ON analysis_instructions (department_uuid, sort_order, created_at)
    WHERE scope = 'department'
      AND is_active = true;

CREATE INDEX idx_analysis_instructions_created_by_user
    ON analysis_instructions (created_by_user_uuid, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_analysis_instructions_created_by_user;
DROP INDEX IF EXISTS idx_analysis_instructions_department_active;
DROP INDEX IF EXISTS idx_analysis_instructions_company_active;
DROP INDEX IF EXISTS idx_analysis_instructions_personal_active;

DROP TABLE IF EXISTS analysis_instructions;
