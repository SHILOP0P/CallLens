package company

import (
	"calllens/monolit/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *ServiceSuite) TestUpdateCompanySuccess() {
	companyID := uuid.New()
	managerID := uuid.New()

	s.repository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, managerID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: managerID, Role: models.CompanyMemberRoleManager}, nil).
		Once()
	s.repository.EXPECT().
		UpdateCompany(mock.Anything, companyID, "New name").
		Return(models.Company{ID: companyID, Name: "New name"}, nil).
		Once()

	got, err := s.service.UpdateCompany(s.ctx, models.UpdateCompanyInput{
		CompanyUUID: companyID,
		RequestUser: managerID,
		Name:        "  New name  ",
	})

	s.Require().NoError(err)
	s.Require().Equal("New name", got.Name)
}

func (s *ServiceSuite) TestUpdateCompanyRejectsEmptyName() {
	_, err := s.service.UpdateCompany(s.ctx, models.UpdateCompanyInput{
		CompanyUUID: uuid.New(),
		RequestUser: uuid.New(),
		Name:        " ",
	})

	s.Require().ErrorIs(err, models.ErrInvalidCompanyInput)
}

func (s *ServiceSuite) TestDeleteCompanySuccess() {
	companyID := uuid.New()
	managerID := uuid.New()

	s.repository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, managerID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: managerID, Role: models.CompanyMemberRoleManager}, nil).
		Once()
	s.repository.EXPECT().
		ArchiveCompany(mock.Anything, companyID).
		Return(nil).
		Once()

	err := s.service.DeleteCompany(s.ctx, models.DeleteCompanyInput{
		CompanyUUID: companyID,
		RequestUser: managerID,
	})

	s.Require().NoError(err)
}
