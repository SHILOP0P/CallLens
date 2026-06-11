package analysis_instruction

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

func (s *Service) List(ctx context.Context, input models.ListAnalysisInstructionsInput) ([]models.AnalysisInstruction, error) {
	if err := validateListInput(input); err != nil {
		return nil, err
	}

	if err := s.authorizeList(ctx, input); err != nil {
		return nil, err
	}

	return s.repository.List(ctx, input)
}

func validateListInput(input models.ListAnalysisInstructionsInput) error {
	if input.UserUUID == uuid.Nil {
		return models.ErrInvalidAnalysisInstructionInput
	}

	switch input.Scope {
	case models.AnalysisInstructionScopePersonal:
		if input.CompanyUUID.Valid || input.DepartmentUUID.Valid {
			return models.ErrInvalidAnalysisInstructionInput
		}
	case models.AnalysisInstructionScopeCompany:
		if !input.CompanyUUID.Valid || input.DepartmentUUID.Valid {
			return models.ErrInvalidAnalysisInstructionInput
		}
	case models.AnalysisInstructionScopeDepartment:
		if !input.CompanyUUID.Valid || !input.DepartmentUUID.Valid {
			return models.ErrInvalidAnalysisInstructionInput
		}
	default:
		return models.ErrInvalidAnalysisInstructionInput
	}

	return nil
}
