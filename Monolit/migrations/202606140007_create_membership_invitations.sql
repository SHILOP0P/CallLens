-- +goose Up
CREATE TABLE membership_invitations (
    invitation_uuid UUID PRIMARY KEY,
    company_uuid UUID NOT NULL REFERENCES companies(company_uuid) ON DELETE CASCADE,
    department_uuid UUID NULL REFERENCES departments(department_uuid) ON DELETE CASCADE,
    invited_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    invited_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
    company_role TEXT NOT NULL DEFAULT 'employee',
    department_role TEXT NULL,
    status TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    responded_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT membership_invitations_status_check
        CHECK (status IN ('pending', 'accepted', 'declined', 'canceled', 'expired')),
    CONSTRAINT membership_invitations_company_role_check
        CHECK (company_role IN ('employee')),
    CONSTRAINT membership_invitations_department_role_check
        CHECK (
            (department_uuid IS NULL AND department_role IS NULL)
            OR
            (department_uuid IS NOT NULL AND department_role IN ('department_leader', 'employee'))
        )
);

CREATE INDEX membership_invitations_invited_user_status_idx
    ON membership_invitations (invited_user_uuid, status);

CREATE INDEX membership_invitations_company_status_idx
    ON membership_invitations (company_uuid, status);

CREATE INDEX membership_invitations_department_status_idx
    ON membership_invitations (department_uuid, status)
    WHERE department_uuid IS NOT NULL;

CREATE UNIQUE INDEX membership_invitations_pending_company_unique_idx
    ON membership_invitations (company_uuid, invited_user_uuid)
    WHERE status = 'pending' AND department_uuid IS NULL;

CREATE UNIQUE INDEX membership_invitations_pending_department_unique_idx
    ON membership_invitations (department_uuid, invited_user_uuid)
    WHERE status = 'pending' AND department_uuid IS NOT NULL;

-- +goose Down
DROP TABLE membership_invitations;
