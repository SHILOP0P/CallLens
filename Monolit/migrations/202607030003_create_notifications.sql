-- +goose Up
CREATE TABLE notifications (
    notification_uuid UUID PRIMARY KEY,
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    entity_type TEXT NULL,
    entity_uuid UUID NULL,
    read_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT notifications_type_check
        CHECK (type IN ('invitation', 'report_ready', 'subscription', 'processing_failed')),
    CONSTRAINT notifications_entity_pair_check
        CHECK (
            (entity_type IS NULL AND entity_uuid IS NULL)
            OR
            (entity_type IS NOT NULL AND entity_uuid IS NOT NULL)
        )
);

CREATE INDEX notifications_user_created_at_idx
    ON notifications (user_uuid, created_at DESC);

CREATE INDEX notifications_user_unread_idx
    ON notifications (user_uuid, created_at DESC)
    WHERE read_at IS NULL;

-- +goose Down
DROP TABLE notifications;
