-- +goose Up
CREATE TABLE users (
                       user_uuid UUID PRIMARY KEY,
                       email TEXT NOT NULL,
                       password_hash TEXT NOT NULL,
                       full_name TEXT NOT NULL,
                       full_surname TEXT NOT NULL,
                       username TEXT NOT NULL,
                       role TEXT NOT NULL,
                       post TEXT NULL,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email_lower
    ON users (lower(email));

CREATE UNIQUE INDEX idx_users_username_lower
    ON users (lower(username));

CREATE TABLE companies (
                           company_uuid UUID PRIMARY KEY,
                           name TEXT NOT NULL,
                           owner_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
                           member_limit INTEGER NOT NULL DEFAULT 1 CHECK (member_limit > 0),
                           created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE company_members (
                                 company_uuid UUID NOT NULL REFERENCES companies(company_uuid) ON DELETE CASCADE,
                                 user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
                                 role TEXT NOT NULL,
                                 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

                                 PRIMARY KEY (company_uuid, user_uuid)
);

CREATE TABLE departments (
                             department_uuid UUID PRIMARY KEY,
                             company_uuid UUID NOT NULL REFERENCES companies(company_uuid) ON DELETE CASCADE,
                             name TEXT NOT NULL,
                             created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE department_members (
    department_uuid UUID NOT NULL REFERENCES departments(department_uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (department_uuid, user_uuid)
    );

ALTER TABLE calls
    ADD COLUMN uploaded_by_user_uuid UUID NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
    ADD COLUMN company_uuid UUID NULL REFERENCES companies(company_uuid) ON DELETE RESTRICT,
    ADD COLUMN department_uuid UUID NULL REFERENCES departments(department_uuid) ON DELETE RESTRICT,
    ADD CONSTRAINT chk_calls_department_requires_company
        CHECK (department_uuid IS NULL OR company_uuid IS NOT NULL);

CREATE INDEX idx_calls_personal_owner_created_at
    ON calls (uploaded_by_user_uuid, created_at DESC)
    WHERE company_uuid IS NULL;

CREATE INDEX idx_calls_company_created_at
    ON calls (company_uuid, created_at DESC)
    WHERE company_uuid IS NOT NULL;

CREATE INDEX idx_calls_department_created_at
    ON calls (department_uuid, created_at DESC)
    WHERE department_uuid IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_calls_department_created_at;
DROP INDEX IF EXISTS idx_calls_company_created_at;
DROP INDEX IF EXISTS idx_calls_personal_owner_created_at;

ALTER TABLE calls
    DROP CONSTRAINT IF EXISTS chk_calls_department_requires_company,
    DROP COLUMN IF EXISTS department_uuid,
    DROP COLUMN IF EXISTS company_uuid,
    DROP COLUMN IF EXISTS uploaded_by_user_uuid;

DROP TABLE IF EXISTS department_members;
DROP TABLE IF EXISTS departments;
DROP TABLE IF EXISTS company_members;
DROP TABLE IF EXISTS companies;
DROP TABLE IF EXISTS users;
