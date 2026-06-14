package converter

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/models"
)

func PlanModelToAPI(plan models.Plan) (dto.PlanResponse, error) {
	return dto.PlanResponse{
		ID:                             plan.ID.String(),
		Code:                           string(plan.Code),
		Type:                           string(plan.Type),
		Name:                           plan.Name,
		MonthlyMinutesLimit:            plan.MonthlyMinutesLimit,
		ActiveInstructionLimit:         plan.ActiveInstructionLimit,
		CompanyLimit:                   plan.CompanyLimit,
		DepartmentsPerCompanyLimit:     plan.DepartmentsPerCompanyLimit,
		MembersPerCompanyLimit:         plan.MembersPerCompanyLimit,
		InstructionsPerDepartmentLimit: plan.InstructionsPerDepartmentLimit,
		AnalysisLevel:                  string(plan.AnalysisLevel),
		HistoryRetentionDays:           plan.HistoryRetentionDays,
		ExportEnabled:                  plan.ExportEnabled,
		TeamAnalyticsEnabled:           plan.TeamAnalyticsEnabled,
		APIAccessEnabled:               plan.APIAccessEnabled,
	}, nil
}
