package company

import (
	"calllens/monolit/internal/models"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *ServiceSuite) TestCreateCompanySuccess() {
	userID := uuid.New()

	s.repository.EXPECT().
		GetManagedCompanyByUserUUID(mock.Anything, userID).
		Return(models.Company{}, models.ErrCompanyNotFound).
		Once()
	s.repository.EXPECT().
		CreateCompany(mock.Anything, mock.MatchedBy(func(company models.Company) bool {
			return company.Name == "CallLens" &&
				company.ManagerUserUUID == userID &&
				company.MemberLimit == defaultMemberLimit
		}), mock.MatchedBy(func(member models.CompanyMember) bool {
			return member.UserUUID == userID &&
				member.Role == models.CompanyMemberRoleManager &&
				member.Status == models.MembershipStatusActive
		})).
		Return(models.Company{Name: "CallLens", ManagerUserUUID: userID, MemberLimit: defaultMemberLimit}, nil).
		Once()

	got, err := s.service.CreateCompany(s.ctx, models.CreateCompanyInput{
		Name:          "  CallLens  ",
		ManagerUserID: userID,
	})

	s.Require().NoError(err)
	s.Require().Equal("CallLens", got.Name)
	s.Require().Equal(userID, got.ManagerUserUUID)
}

func (s *ServiceSuite) TestCreateCompanyRejectsInvalidInput() {
	_, err := s.service.CreateCompany(s.ctx, models.CreateCompanyInput{
		Name:          " ",
		ManagerUserID: uuid.New(),
	})
	s.Require().ErrorIs(err, models.ErrInvalidCompanyInput)

	_, err = s.service.CreateCompany(s.ctx, models.CreateCompanyInput{
		Name:          "CallLens",
		ManagerUserID: uuid.Nil,
	})
	s.Require().ErrorIs(err, models.ErrInvalidCompanyInput)
}

func (s *ServiceSuite) TestCreateCompanyRejectsAlreadyManagedCompany() {
	userID := uuid.New()

	s.repository.EXPECT().
		GetManagedCompanyByUserUUID(mock.Anything, userID).
		Return(models.Company{ManagerUserUUID: userID}, nil).
		Once()

	_, err := s.service.CreateCompany(s.ctx, models.CreateCompanyInput{
		Name:          "CallLens",
		ManagerUserID: userID,
	})

	s.Require().ErrorIs(err, models.ErrUserAlreadyManagesCompany)
}

func (s *ServiceSuite) TestCreateCompanyReturnsManagedCompanyLookupError() {
	userID := uuid.New()
	repoErr := errors.New("lookup failed")

	s.repository.EXPECT().
		GetManagedCompanyByUserUUID(mock.Anything, userID).
		Return(models.Company{}, repoErr).
		Once()

	_, err := s.service.CreateCompany(s.ctx, models.CreateCompanyInput{
		Name:          "CallLens",
		ManagerUserID: userID,
	})

	s.Require().ErrorIs(err, repoErr)
}

func (s *ServiceSuite) TestCreateCompanyReturnsCreateError() {
	userID := uuid.New()
	repoErr := errors.New("create failed")

	s.repository.EXPECT().
		GetManagedCompanyByUserUUID(mock.Anything, userID).
		Return(models.Company{}, models.ErrCompanyNotFound).
		Once()
	s.repository.EXPECT().
		CreateCompany(mock.Anything, mock.Anything, mock.Anything).
		Return(models.Company{}, repoErr).
		Once()

	_, err := s.service.CreateCompany(s.ctx, models.CreateCompanyInput{
		Name:          "CallLens",
		ManagerUserID: userID,
	})

	s.Require().ErrorIs(err, repoErr)
}
