-- +goose Up
ALTER TABLE users
    ADD COLUMN phone TEXT NULL,
    ADD COLUMN timezone TEXT NULL,
    ADD COLUMN avatar_path TEXT NULL,
    ADD COLUMN avatar_mime_type TEXT NULL,
    ADD COLUMN avatar_size_bytes BIGINT NULL,
    ADD COLUMN avatar_updated_at TIMESTAMPTZ NULL;

CREATE TABLE user_preferences (
    user_uuid UUID PRIMARY KEY REFERENCES users(user_uuid) ON DELETE CASCADE,
    active_company_uuid UUID NULL REFERENCES companies(company_uuid) ON DELETE SET NULL,
    theme TEXT NOT NULL DEFAULT 'system',
    date_range_from DATE NULL,
    date_range_to DATE NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_user_preferences_theme CHECK (theme IN ('system', 'light', 'dark')),
    CONSTRAINT chk_user_preferences_date_range CHECK (
        date_range_from IS NULL OR date_range_to IS NULL OR date_range_from <= date_range_to
    )
);

-- +goose Down
DROP TABLE IF EXISTS user_preferences;

ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_updated_at,
    DROP COLUMN IF EXISTS avatar_size_bytes,
    DROP COLUMN IF EXISTS avatar_mime_type,
    DROP COLUMN IF EXISTS avatar_path,
    DROP COLUMN IF EXISTS timezone,
    DROP COLUMN IF EXISTS phone;
