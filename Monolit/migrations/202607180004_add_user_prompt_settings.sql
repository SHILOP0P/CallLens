-- +goose Up
ALTER TABLE prompt_industries ADD COLUMN IF NOT EXISTS base_prompt TEXT NOT NULL DEFAULT '';

CREATE TABLE prompt_user_settings (
    user_uuid UUID PRIMARY KEY REFERENCES users(user_uuid) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE prompt_user_industries (
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    industry_key TEXT NOT NULL REFERENCES prompt_industries(industry_key) ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_uuid, industry_key)
);

CREATE TABLE prompt_user_topics (
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    topic_key TEXT NOT NULL REFERENCES prompt_topics(topic_key) ON DELETE CASCADE,
    source TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'recommended', 'auto')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_uuid, topic_key)
);

-- Existing temporary named profiles are intentionally retained for compatibility,
-- but the application no longer exposes them as a user-facing configuration model.

-- +goose Down
DROP TABLE IF EXISTS prompt_user_topics;
DROP TABLE IF EXISTS prompt_user_industries;
DROP TABLE IF EXISTS prompt_user_settings;
ALTER TABLE prompt_industries DROP COLUMN IF EXISTS base_prompt;
