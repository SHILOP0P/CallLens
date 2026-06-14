-- +goose Up
DROP INDEX IF EXISTS idx_companies_manager_user_uuid_unique;

CREATE TABLE plans (
    plan_uuid UUID PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    monthly_minutes_limit INTEGER NOT NULL CHECK (monthly_minutes_limit >= 0),
    active_instruction_limit INTEGER NOT NULL CHECK (active_instruction_limit >= 0),
    company_limit INTEGER NULL CHECK (company_limit IS NULL OR company_limit >= 0),
    departments_per_company_limit INTEGER NULL CHECK (departments_per_company_limit IS NULL OR departments_per_company_limit >= 0),
    members_per_company_limit INTEGER NULL CHECK (members_per_company_limit IS NULL OR members_per_company_limit >= 0),
    instructions_per_department_limit INTEGER NULL CHECK (instructions_per_department_limit IS NULL OR instructions_per_department_limit >= 0),
    analysis_level TEXT NOT NULL,
    history_retention_days INTEGER NOT NULL CHECK (history_retention_days >= 0),
    export_enabled BOOLEAN NOT NULL DEFAULT false,
    team_analytics_enabled BOOLEAN NOT NULL DEFAULT false,
    api_access_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_plans_type
        CHECK (type IN ('personal', 'business')),
    CONSTRAINT chk_plans_code
        CHECK (code IN (
            'personal_start',
            'personal_plus',
            'personal_pro',
            'business_start',
            'business_plus',
            'business_pro'
        ))
);

CREATE TABLE subscriptions (
    subscription_uuid UUID PRIMARY KEY,
    plan_uuid UUID NOT NULL REFERENCES plans(plan_uuid) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    user_uuid UUID NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    company_uuid UUID NULL REFERENCES companies(company_uuid) ON DELETE CASCADE,
    status TEXT NOT NULL,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_subscriptions_status
        CHECK (status IN ('active', 'canceled', 'expired')),
    CONSTRAINT chk_subscriptions_type
        CHECK (type IN ('personal', 'business')),
    CONSTRAINT chk_subscriptions_owner
        CHECK (
            (user_uuid IS NOT NULL AND company_uuid IS NULL)
            OR (user_uuid IS NULL AND company_uuid IS NOT NULL)
        ),
    CONSTRAINT chk_subscriptions_period
        CHECK (ends_at IS NULL OR ends_at > starts_at)
);

CREATE UNIQUE INDEX idx_subscriptions_active_user
    ON subscriptions (type, user_uuid)
    WHERE status = 'active'
      AND user_uuid IS NOT NULL;

CREATE UNIQUE INDEX idx_subscriptions_active_company
    ON subscriptions (company_uuid)
    WHERE status = 'active'
      AND company_uuid IS NOT NULL;

CREATE INDEX idx_subscriptions_plan
    ON subscriptions (plan_uuid);

CREATE TABLE usage_counters (
    usage_counter_uuid UUID PRIMARY KEY,
    subscription_uuid UUID NOT NULL REFERENCES subscriptions(subscription_uuid) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    used_minutes INTEGER NOT NULL DEFAULT 0 CHECK (used_minutes >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_usage_counters_period
        CHECK (period_end > period_start),
    CONSTRAINT uq_usage_counters_subscription_period
        UNIQUE (subscription_uuid, period_start)
);

CREATE INDEX idx_usage_counters_subscription_period
    ON usage_counters (subscription_uuid, period_start, period_end);

INSERT INTO plans (
    plan_uuid,
    code,
    type,
    name,
    monthly_minutes_limit,
    active_instruction_limit,
    company_limit,
    departments_per_company_limit,
    members_per_company_limit,
    instructions_per_department_limit,
    analysis_level,
    history_retention_days,
    export_enabled,
    team_analytics_enabled,
    api_access_enabled
)
VALUES
    ('11111111-1111-7111-8111-111111111111', 'personal_start', 'personal', 'Personal Start', 120, 2, NULL, NULL, NULL, NULL, 'basic', 30, false, false, false),
    ('11111111-1111-7111-8111-111111111112', 'personal_plus', 'personal', 'Personal Plus', 600, 5, NULL, NULL, NULL, NULL, 'plus', 180, false, false, false),
    ('11111111-1111-7111-8111-111111111113', 'personal_pro', 'personal', 'Personal Pro', 2000, 20, NULL, NULL, NULL, NULL, 'pro', 365, true, false, false),
    ('22222222-2222-7222-8222-222222222221', 'business_start', 'business', 'Business Start', 1000, 0, 1, 5, 25, 5, 'plus', 180, false, false, false),
    ('22222222-2222-7222-8222-222222222222', 'business_plus', 'business', 'Business Plus', 5000, 0, 1, 10, 100, 7, 'pro', 365, true, true, false),
    ('22222222-2222-7222-8222-222222222223', 'business_pro', 'business', 'Business Pro', 20000, 0, 3, 15, 500, 10, 'priority', 730, true, true, true)
ON CONFLICT (code) DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS idx_usage_counters_subscription_period;
DROP TABLE IF EXISTS usage_counters;

DROP INDEX IF EXISTS idx_subscriptions_plan;
DROP INDEX IF EXISTS idx_subscriptions_active_company;
DROP INDEX IF EXISTS idx_subscriptions_active_user;
DROP TABLE IF EXISTS subscriptions;

DROP TABLE IF EXISTS plans;

CREATE UNIQUE INDEX IF NOT EXISTS idx_companies_manager_user_uuid_unique
    ON companies (manager_user_uuid);
