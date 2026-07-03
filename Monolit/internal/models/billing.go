package models

import (
	"time"

	"github.com/google/uuid"
)

type PlanType string
type PlanCode string
type SubscriptionStatus string
type AnalysisLevel string

const (
	PlanTypePersonal PlanType = "personal"
	PlanTypeBusiness PlanType = "business"
)

const (
	PlanCodePersonalStart PlanCode = "personal_start"
	PlanCodePersonalPlus  PlanCode = "personal_plus"
	PlanCodePersonalPro   PlanCode = "personal_pro"
	PlanCodeBusinessStart PlanCode = "business_start"
	PlanCodeBusinessPlus  PlanCode = "business_plus"
	PlanCodeBusinessPro   PlanCode = "business_pro"
)

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
)

const (
	AnalysisLevelBasic    AnalysisLevel = "basic"
	AnalysisLevelPlus     AnalysisLevel = "plus"
	AnalysisLevelPro      AnalysisLevel = "pro"
	AnalysisLevelPriority AnalysisLevel = "priority"
)

type Plan struct {
	ID                             uuid.UUID
	Code                           PlanCode
	Type                           PlanType
	Name                           string
	MonthlyMinutesLimit            int
	ActiveInstructionLimit         int
	CompanyLimit                   *int
	DepartmentsPerCompanyLimit     *int
	MembersPerCompanyLimit         *int
	InstructionsPerDepartmentLimit *int
	AnalysisLevel                  AnalysisLevel
	HistoryRetentionDays           int
	ExportEnabled                  bool
	TeamAnalyticsEnabled           bool
	APIAccessEnabled               bool
	CreatedAt                      time.Time
	UpdatedAt                      time.Time
}

type Subscription struct {
	ID          uuid.UUID
	Plan        Plan
	UserUUID    uuid.NullUUID
	CompanyUUID uuid.NullUUID
	Status      SubscriptionStatus
	StartsAt    time.Time
	EndsAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UsageCounter struct {
	ID               uuid.UUID
	SubscriptionUUID uuid.UUID
	PeriodStart      time.Time
	PeriodEnd        time.Time
	UsedMinutes      int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SubscriptionUsage struct {
	Subscription            Subscription
	PeriodStart             time.Time
	PeriodEnd               time.Time
	UsedMinutes             int
	LimitMinutes            int
	RemainingMinutes        int
	Percent                 float64
	MembersLimit            *int
	MembersUsed             *int
	DepartmentsLimit        *int
	DepartmentsUsed         *int
	ActiveInstructionsLimit *int
	ActiveInstructionsUsed  *int
}

type UpsertSubscriptionInput struct {
	ID          uuid.UUID
	PlanCode    PlanCode
	UserUUID    uuid.NullUUID
	CompanyUUID uuid.NullUUID
	Status      SubscriptionStatus
	StartsAt    time.Time
	EndsAt      *time.Time
}

type ActivateCompanySubscriptionInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
	PlanCode    PlanCode
}

type ActivatePersonalSubscriptionInput struct {
	UserUUID uuid.UUID
	PlanCode PlanCode
}

type CancelCompanySubscriptionInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
}

type GetCompanySubscriptionInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
}

type GetPersonalSubscriptionUsageInput struct {
	UserUUID    uuid.UUID
	PeriodStart *time.Time
}

type GetCompanySubscriptionUsageInput struct {
	CompanyUUID uuid.UUID
	RequestUser uuid.UUID
	PeriodStart *time.Time
}
