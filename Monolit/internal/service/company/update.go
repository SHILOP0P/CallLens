package company

import (
	"context"
	"strings"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (s *Service) UpdateCompany(ctx context.Context, input models.UpdateCompanyInput) (models.Company, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.CompanyUUID == uuid.Nil || input.RequestUser == uuid.Nil || input.Name == "" {
		return models.Company{}, models.ErrInvalidCompanyInput
	}

	if err := s.requireCompanyManager(ctx, input.CompanyUUID, input.RequestUser); err != nil {
		return models.Company{}, err
	}

	if err := s.requireActiveCompanySubscription(ctx, input.CompanyUUID); err != nil {
		return models.Company{}, err
	}

	return s.companyRepository.UpdateCompany(ctx, input.CompanyUUID, input.Name)
}

func (s *Service) DeleteCompany(ctx context.Context, input models.DeleteCompanyInput) error {
	if input.CompanyUUID == uuid.Nil || input.RequestUser == uuid.Nil {
		return models.ErrInvalidCompanyInput
	}

	if err := s.requireCompanyManager(ctx, input.CompanyUUID, input.RequestUser); err != nil {
		return err
	}

	return s.companyRepository.ArchiveCompany(ctx, input.CompanyUUID)
}
