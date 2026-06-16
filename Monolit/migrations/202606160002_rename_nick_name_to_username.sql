-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
          AND column_name = 'nick_name'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
          AND column_name = 'username'
    ) THEN
        ALTER TABLE users RENAME COLUMN nick_name TO username;
    END IF;
END $$;
-- +goose StatementEnd

UPDATE users
SET username = lower(username);

UPDATE users
SET username = '@' || username
WHERE username !~ '^@';

UPDATE users
SET username = '@user_' || substr(replace(user_uuid::text, '-', ''), 1, 6)
WHERE username !~ '^@[a-z][a-z0-9_]{3,23}$';

WITH duplicates AS (
    SELECT user_uuid,
           row_number() OVER (PARTITION BY lower(username) ORDER BY created_at, user_uuid) AS row_number
    FROM users
)
UPDATE users u
SET username = '@user_' || substr(replace(u.user_uuid::text, '-', ''), 1, 6)
FROM duplicates d
WHERE u.user_uuid = d.user_uuid
  AND d.row_number > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower
    ON users (lower(username));

-- +goose Down
DROP INDEX IF EXISTS idx_users_username_lower;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
          AND column_name = 'username'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'users'
          AND column_name = 'nick_name'
    ) THEN
        ALTER TABLE users RENAME COLUMN username TO nick_name;
    END IF;
END $$;
-- +goose StatementEnd
