-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'active_instructions_limit'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'active_instruction_limit'
    ) THEN
        ALTER TABLE plans
            RENAME COLUMN active_instructions_limit TO active_instruction_limit;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'companies_limit'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'company_limit'
    ) THEN
        ALTER TABLE plans
            RENAME COLUMN companies_limit TO company_limit;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'departments_limit'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'departments_per_company_limit'
    ) THEN
        ALTER TABLE plans
            RENAME COLUMN departments_limit TO departments_per_company_limit;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'members_limit'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'members_per_company_limit'
    ) THEN
        ALTER TABLE plans
            RENAME COLUMN members_limit TO members_per_company_limit;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'call_history_days'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'history_retention_days'
    ) THEN
        ALTER TABLE plans
            RENAME COLUMN call_history_days TO history_retention_days;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'export_reports_enabled'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'plans'
          AND column_name = 'export_enabled'
    ) THEN
        ALTER TABLE plans
            RENAME COLUMN export_reports_enabled TO export_enabled;
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE plans
    ADD COLUMN IF NOT EXISTS instructions_per_department_limit INTEGER NULL
        CHECK (instructions_per_department_limit IS NULL OR instructions_per_department_limit >= 0);

ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS type TEXT NULL;

UPDATE subscriptions s
SET type = p.type
FROM plans p
WHERE p.plan_uuid = s.plan_uuid
  AND s.type IS NULL;

ALTER TABLE subscriptions
    ALTER COLUMN type SET NOT NULL;

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_subscriptions_type'
    ) THEN
        ALTER TABLE subscriptions
            ADD CONSTRAINT chk_subscriptions_type
                CHECK (type IN ('personal', 'business'));
    END IF;
END $$;
-- +goose StatementEnd

DROP INDEX IF EXISTS idx_subscriptions_active_user;

CREATE UNIQUE INDEX idx_subscriptions_active_user
    ON subscriptions (type, user_uuid)
    WHERE status = 'active'
      AND user_uuid IS NOT NULL;

UPDATE plans
SET monthly_minutes_limit = v.monthly_minutes_limit,
    active_instruction_limit = v.active_instruction_limit,
    company_limit = v.company_limit,
    departments_per_company_limit = v.departments_per_company_limit,
    members_per_company_limit = v.members_per_company_limit,
    instructions_per_department_limit = v.instructions_per_department_limit,
    analysis_level = v.analysis_level,
    history_retention_days = v.history_retention_days,
    export_enabled = v.export_enabled,
    team_analytics_enabled = v.team_analytics_enabled,
    api_access_enabled = v.api_access_enabled,
    updated_at = now()
FROM (
    VALUES
        ('personal_start', 120, 2, NULL::INTEGER, NULL::INTEGER, NULL::INTEGER, NULL::INTEGER, 'basic', 30, false, false, false),
        ('personal_plus', 600, 5, NULL::INTEGER, NULL::INTEGER, NULL::INTEGER, NULL::INTEGER, 'plus', 180, false, false, false),
        ('personal_pro', 2000, 20, NULL::INTEGER, NULL::INTEGER, NULL::INTEGER, NULL::INTEGER, 'pro', 365, true, false, false),
        ('business_start', 1000, 0, 1, 5, 25, 5, 'plus', 180, false, false, false),
        ('business_plus', 5000, 0, 1, 10, 100, 7, 'pro', 365, true, true, false),
        ('business_pro', 20000, 0, 3, 15, 500, 10, 'priority', 730, true, true, true)
) AS v(
    code,
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
WHERE plans.code = v.code;

-- +goose Down
DROP INDEX IF EXISTS idx_subscriptions_active_user;

CREATE UNIQUE INDEX idx_subscriptions_active_user
    ON subscriptions (user_uuid)
    WHERE status = 'active'
      AND user_uuid IS NOT NULL;
