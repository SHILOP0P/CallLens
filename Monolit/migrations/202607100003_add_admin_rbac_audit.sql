-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    invalid_roles TEXT;
    superadmin_count BIGINT;
BEGIN
    SELECT string_agg(role, ', ' ORDER BY role)
    INTO invalid_roles
    FROM (
        SELECT DISTINCT role
        FROM users
        WHERE role NOT IN ('user', 'helper', 'admin', 'superadmin')
    ) AS invalid;

    IF invalid_roles IS NOT NULL THEN
        RAISE EXCEPTION
            'cannot enforce users role constraint; unsupported roles exist: %',
            invalid_roles;
    END IF;

    SELECT count(*)
    INTO superadmin_count
    FROM users
    WHERE role = 'superadmin';

    IF superadmin_count > 1 THEN
        RAISE EXCEPTION
            'cannot enforce singleton superadmin; found % superadmin users',
            superadmin_count;
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE users
    ADD CONSTRAINT chk_users_role
        CHECK (role IN ('user', 'helper', 'admin', 'superadmin'))
        NOT VALID;

ALTER TABLE users
    VALIDATE CONSTRAINT chk_users_role;

CREATE UNIQUE INDEX uq_users_single_superadmin
    ON users (role)
    WHERE role = 'superadmin';

CREATE TABLE admin_audit_logs (
    audit_uuid UUID PRIMARY KEY,
    actor_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
    actor_role TEXT NOT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_uuid UUID NULL,
    before_data JSONB NULL,
    after_data JSONB NULL,
    reason TEXT NULL,
    request_id TEXT NULL,
    ip_address INET NULL,
    user_agent TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_admin_audit_actor_role
        CHECK (actor_role IN ('helper', 'admin', 'superadmin')),
    CONSTRAINT chk_admin_audit_action_not_blank
        CHECK (length(btrim(action)) > 0),
    CONSTRAINT chk_admin_audit_target_type_not_blank
        CHECK (length(btrim(target_type)) > 0),
    CONSTRAINT chk_admin_audit_before_object
        CHECK (before_data IS NULL OR jsonb_typeof(before_data) = 'object'),
    CONSTRAINT chk_admin_audit_after_object
        CHECK (after_data IS NULL OR jsonb_typeof(after_data) = 'object')
);

CREATE INDEX idx_admin_audit_logs_created_at
    ON admin_audit_logs (created_at DESC, audit_uuid DESC);

CREATE INDEX idx_admin_audit_logs_actor_created_at
    ON admin_audit_logs (actor_user_uuid, created_at DESC);

CREATE INDEX idx_admin_audit_logs_target_created_at
    ON admin_audit_logs (target_type, target_uuid, created_at DESC);

CREATE INDEX idx_admin_audit_logs_action_created_at
    ON admin_audit_logs (action, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS admin_audit_logs;

DROP INDEX IF EXISTS uq_users_single_superadmin;

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS chk_users_role;
