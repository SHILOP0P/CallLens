package dto

type ActivateSubscriptionRequest struct {
	PlanCode string `json:"plan_code"`
}

type PlanResponse struct {
	ID                             string `json:"id"`
	Code                           string `json:"code"`
	Type                           string `json:"type"`
	Name                           string `json:"name"`
	MonthlyMinutesLimit            int    `json:"monthly_minutes_limit"`
	ActiveInstructionLimit         int    `json:"active_instruction_limit"`
	CompanyLimit                   *int   `json:"company_limit"`
	DepartmentsPerCompanyLimit     *int   `json:"departments_per_company_limit"`
	MembersPerCompanyLimit         *int   `json:"members_per_company_limit"`
	InstructionsPerDepartmentLimit *int   `json:"instructions_per_department_limit"`
	AnalysisLevel                  string `json:"analysis_level"`
	HistoryRetentionDays           int    `json:"history_retention_days"`
	ExportEnabled                  bool   `json:"export_enabled"`
	TeamAnalyticsEnabled           bool   `json:"team_analytics_enabled"`
	APIAccessEnabled               bool   `json:"api_access_enabled"`
}

type PlansResponse struct {
	Plans []PlanResponse `json:"plans"`
}

type SubscriptionResponse struct {
	ID          string       `json:"id"`
	Plan        PlanResponse `json:"plan"`
	UserUUID    *string      `json:"user_uuid"`
	CompanyUUID *string      `json:"company_uuid"`
	Status      string       `json:"status"`
	StartsAt    string       `json:"starts_at"`
	EndsAt      *string      `json:"ends_at"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}
