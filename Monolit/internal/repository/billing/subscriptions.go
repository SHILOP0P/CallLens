package billing

import (
	"calllens/monolit/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (r *Repository) GetActivePersonalSubscription(ctx context.Context, userID uuid.UUID) (models.Subscription, error) {
	query := activeSubscriptionQuery("s.type = 'personal' AND s.user_uuid = $1")
	return r.getSubscription(ctx, query, userID)
}

func (r *Repository) GetActiveBusinessSubscription(ctx context.Context, companyID uuid.UUID) (models.Subscription, error) {
	query := activeSubscriptionQuery(`s.type = 'business'
	  AND (
	      s.company_uuid = $1
	      OR s.user_uuid = (
	          SELECT manager_user_uuid
	          FROM companies
	          WHERE company_uuid = $1
	      )
	  )`)
	return r.getSubscription(ctx, query, companyID)
}

func (r *Repository) GetActiveBusinessSubscriptionForOwner(ctx context.Context, ownerID uuid.UUID) (models.Subscription, error) {
	query := activeSubscriptionQuery("s.type = 'business' AND s.user_uuid = $1")
	return r.getSubscription(ctx, query, ownerID)
}

func (r *Repository) UpsertSubscription(ctx context.Context, input models.UpsertSubscriptionInput) (models.Subscription, error) {
	if input.ID == uuid.Nil {
		id, err := uuid.NewV7()
		if err != nil {
			return models.Subscription{}, fmt.Errorf("generate subscription uuid: %w", err)
		}
		input.ID = id
	}

	if input.Status == "" {
		input.Status = models.SubscriptionStatusActive
	}

	if input.StartsAt.IsZero() {
		input.StartsAt = time.Now().UTC()
	}

	query := `
	WITH selected_plan AS (
	    SELECT plan_uuid, type
	    FROM plans
	    WHERE code = $2
	),
	upserted AS (
	    INSERT INTO subscriptions (
	        subscription_uuid,
	        plan_uuid,
	        type,
	        user_uuid,
	        company_uuid,
	        status,
	        starts_at,
	        ends_at
	    )
	    SELECT $1, plan_uuid, type, $3, $4, $5, $6, $7
	    FROM selected_plan
	    ON CONFLICT (subscription_uuid)
	    DO UPDATE SET plan_uuid = EXCLUDED.plan_uuid,
	                  type = EXCLUDED.type,
	                  user_uuid = EXCLUDED.user_uuid,
	                  company_uuid = EXCLUDED.company_uuid,
	                  status = EXCLUDED.status,
	                  starts_at = EXCLUDED.starts_at,
	                  ends_at = EXCLUDED.ends_at,
	                  updated_at = now()
	    RETURNING *
	)
	SELECT ` + subscriptionColumns("u", "p") + `
	FROM upserted u
	JOIN plans p ON p.plan_uuid = u.plan_uuid
	`

	subscription, err := scanSubscription(r.db.QueryRowContext(
		ctx,
		query,
		input.ID,
		input.PlanCode,
		nullableUUID(input.UserUUID),
		nullableUUID(input.CompanyUUID),
		input.Status,
		input.StartsAt,
		input.EndsAt,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Subscription{}, models.ErrPlanNotFound
		}
		return models.Subscription{}, fmt.Errorf("upsert subscription: %w", err)
	}

	return subscription, nil
}

func (r *Repository) getSubscription(ctx context.Context, query string, args ...any) (models.Subscription, error) {
	subscription, err := scanSubscription(r.db.QueryRowContext(ctx, query, args...))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Subscription{}, models.ErrSubscriptionNotFound
		}
		return models.Subscription{}, fmt.Errorf("get subscription: %w", err)
	}

	return subscription, nil
}

func activeSubscriptionQuery(where string) string {
	return `
	SELECT ` + subscriptionColumns("s", "p") + `
	FROM subscriptions s
	JOIN plans p ON p.plan_uuid = s.plan_uuid
	WHERE s.status = 'active'
	  AND s.starts_at <= now()
	  AND (s.ends_at IS NULL OR s.ends_at > now())
	  AND ` + where + `
	`
}

func subscriptionColumns(subscriptionAlias string, planAlias string) string {
	return subscriptionAlias + `.subscription_uuid,
	       ` + subscriptionAlias + `.type,
	       ` + subscriptionAlias + `.user_uuid,
	       ` + subscriptionAlias + `.company_uuid,
	       ` + subscriptionAlias + `.status,
	       ` + subscriptionAlias + `.starts_at,
	       ` + subscriptionAlias + `.ends_at,
	       ` + subscriptionAlias + `.created_at,
	       ` + subscriptionAlias + `.updated_at,
	       ` + planAlias + `.plan_uuid,
	       ` + planAlias + `.code,
	       ` + planAlias + `.type,
	       ` + planAlias + `.name,
	       ` + planAlias + `.monthly_minutes_limit,
	       ` + planAlias + `.active_instruction_limit,
	       ` + planAlias + `.company_limit,
	       ` + planAlias + `.departments_per_company_limit,
	       ` + planAlias + `.members_per_company_limit,
	       ` + planAlias + `.instructions_per_department_limit,
	       ` + planAlias + `.analysis_level,
	       ` + planAlias + `.history_retention_days,
	       ` + planAlias + `.export_enabled,
	       ` + planAlias + `.team_analytics_enabled,
	       ` + planAlias + `.api_access_enabled,
	       ` + planAlias + `.created_at,
	       ` + planAlias + `.updated_at`
}

func scanSubscription(row planScanner) (models.Subscription, error) {
	var subscription models.Subscription
	var subscriptionType string
	var status string
	var endsAt sql.NullTime
	var planCode string
	var planType string
	var companyLimit sql.NullInt64
	var departmentsPerCompanyLimit sql.NullInt64
	var membersPerCompanyLimit sql.NullInt64
	var instructionsPerDepartmentLimit sql.NullInt64
	var analysisLevel string

	if err := row.Scan(
		&subscription.ID,
		&subscriptionType,
		&subscription.UserUUID,
		&subscription.CompanyUUID,
		&status,
		&subscription.StartsAt,
		&endsAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
		&subscription.Plan.ID,
		&planCode,
		&planType,
		&subscription.Plan.Name,
		&subscription.Plan.MonthlyMinutesLimit,
		&subscription.Plan.ActiveInstructionLimit,
		&companyLimit,
		&departmentsPerCompanyLimit,
		&membersPerCompanyLimit,
		&instructionsPerDepartmentLimit,
		&analysisLevel,
		&subscription.Plan.HistoryRetentionDays,
		&subscription.Plan.ExportEnabled,
		&subscription.Plan.TeamAnalyticsEnabled,
		&subscription.Plan.APIAccessEnabled,
		&subscription.Plan.CreatedAt,
		&subscription.Plan.UpdatedAt,
	); err != nil {
		return models.Subscription{}, err
	}

	subscription.Status = models.SubscriptionStatus(status)
	subscription.Plan.Code = models.PlanCode(planCode)
	subscription.Plan.Type = models.PlanType(planType)
	subscription.Plan.CompanyLimit = nullableInt(companyLimit)
	subscription.Plan.DepartmentsPerCompanyLimit = nullableInt(departmentsPerCompanyLimit)
	subscription.Plan.MembersPerCompanyLimit = nullableInt(membersPerCompanyLimit)
	subscription.Plan.InstructionsPerDepartmentLimit = nullableInt(instructionsPerDepartmentLimit)
	subscription.Plan.AnalysisLevel = models.AnalysisLevel(analysisLevel)
	if endsAt.Valid {
		subscription.EndsAt = &endsAt.Time
	}

	return subscription, nil
}
