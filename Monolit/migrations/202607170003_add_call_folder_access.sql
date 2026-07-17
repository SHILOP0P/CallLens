-- +goose Up
CREATE TABLE call_folder_accesses (
    folder_uuid UUID NOT NULL REFERENCES call_folders(folder_uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    granted_by_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (folder_uuid, user_uuid)
);

CREATE INDEX idx_call_folder_accesses_user_folder
    ON call_folder_accesses (user_uuid, folder_uuid);

-- +goose Down
DROP TABLE IF EXISTS call_folder_accesses;
