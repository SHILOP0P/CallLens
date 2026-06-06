package company

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

func (s *Service) ListCompanyDepartments(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) ([]models.Department, error) {
	if companyID == uuid.Nil || userID == uuid.Nil {
		return nil, models.ErrInvalidDepartmentInput
	}

	return s.departmentRepository.ListVisibleCompanyDepartments(ctx, companyID, userID)
}
