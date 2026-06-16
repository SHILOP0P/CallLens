package company

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

func (s *Service) GetCompanyMembersOverview(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) (models.CompanyMembersOverview, error) {
	if companyID == uuid.Nil || userID == uuid.Nil {
		return models.CompanyMembersOverview{}, models.ErrInvalidCompanyInput
	}

	member, err := s.companyRepository.GetCompanyMember(ctx, companyID, userID)
	if err != nil {
		return models.CompanyMembersOverview{}, err
	}

	if member.Role != models.CompanyMemberRoleManager {
		return models.CompanyMembersOverview{}, models.ErrForbidden
	}

	if err := s.requireActiveCompanySubscription(ctx, companyID); err != nil {
		return models.CompanyMembersOverview{}, err
	}

	return s.companyRepository.GetCompanyMembersOverview(ctx, companyID)
}
