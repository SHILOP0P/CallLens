package dto

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
