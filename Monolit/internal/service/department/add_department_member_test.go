package department

import (
	"calllens/monolit/internal/models"
	"errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *ServiceSuite) TestAddDepartmentMemberSuccess() {
	companyID := uuid.New()
	departmentID := uuid.New()
	managerID := uuid.New()
	userID := uuid.New()

	s.companyRepository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, managerID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: managerID, Role: models.CompanyMemberRoleManager}, nil).
		Once()
	s.companyRepository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, userID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: userID, Role: models.CompanyMemberRoleEmployee, Status: models.MembershipStatusActive}, nil).
		Once()
	s.departmentRepository.EXPECT().
		AddDepartmentMember(mock.Anything, companyID, mock.MatchedBy(func(member models.DepartmentMember) bool {
			return member.DepartmentUUID == departmentID &&
				member.UserUUID == userID &&
				member.Role == models.DepartmentMemberRoleEmployee &&
				member.Status == models.MembershipStatusActive
		})).
		Return(models.DepartmentMember{DepartmentUUID: departmentID, UserUUID: userID, Role: models.DepartmentMemberRoleEmployee, Status: models.MembershipStatusActive}, nil).
		Once()

	got, err := s.service.AddDepartmentMember(s.ctx, models.AddDepartmentMemberInput{
		CompanyUUID:    companyID,
		DepartmentUUID: departmentID,
		RequestUser:    managerID,
		UserUUID:       userID,
		Role:           models.DepartmentMemberRoleEmployee,
	})

	s.Require().NoError(err)
	s.Require().Equal(userID, got.UserUUID)
	s.Require().Equal(models.DepartmentMemberRoleEmployee, got.Role)
}

func (s *ServiceSuite) TestAddDepartmentMemberRejectsInvalidRole() {
	_, err := s.service.AddDepartmentMember(s.ctx, models.AddDepartmentMemberInput{
		CompanyUUID:    uuid.New(),
		DepartmentUUID: uuid.New(),
		RequestUser:    uuid.New(),
		UserUUID:       uuid.New(),
		Role:           models.DepartmentMemberRole("company_manager"),
	})

	s.Require().ErrorIs(err, models.ErrInvalidDepartmentInput)
}

func (s *ServiceSuite) TestAddDepartmentMemberRejectsNonManager() {
	companyID := uuid.New()
	requestUserID := uuid.New()

	s.companyRepository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, requestUserID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: requestUserID, Role: models.CompanyMemberRoleEmployee}, nil).
		Once()

	_, err := s.service.AddDepartmentMember(s.ctx, models.AddDepartmentMemberInput{
		CompanyUUID:    companyID,
		DepartmentUUID: uuid.New(),
		RequestUser:    requestUserID,
		UserUUID:       uuid.New(),
		Role:           models.DepartmentMemberRoleEmployee,
	})

	s.Require().ErrorIs(err, models.ErrForbidden)
}

func (s *ServiceSuite) TestAddDepartmentMemberRequiresCompanyMemberTarget() {
	companyID := uuid.New()
	departmentID := uuid.New()
	managerID := uuid.New()
	userID := uuid.New()
	repoErr := errors.New("member not found")

	s.companyRepository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, managerID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: managerID, Role: models.CompanyMemberRoleManager}, nil).
		Once()
	s.companyRepository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, userID).
		Return(models.CompanyMember{}, repoErr).
		Once()

	_, err := s.service.AddDepartmentMember(s.ctx, models.AddDepartmentMemberInput{
		CompanyUUID:    companyID,
		DepartmentUUID: departmentID,
		RequestUser:    managerID,
		UserUUID:       userID,
		Role:           models.DepartmentMemberRoleEmployee,
	})

	s.Require().ErrorIs(err, repoErr)
}

func (s *ServiceSuite) TestAddDepartmentMemberReturnsRepositoryCreateError() {
	companyID := uuid.New()
	departmentID := uuid.New()
	managerID := uuid.New()
	userID := uuid.New()
	repoErr := errors.New("create failed")

	s.companyRepository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, managerID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: managerID, Role: models.CompanyMemberRoleManager}, nil).
		Once()
	s.companyRepository.EXPECT().
		GetCompanyMember(mock.Anything, companyID, userID).
		Return(models.CompanyMember{CompanyUUID: companyID, UserUUID: userID, Role: models.CompanyMemberRoleEmployee}, nil).
		Once()
	s.departmentRepository.EXPECT().
		AddDepartmentMember(mock.Anything, companyID, mock.Anything).
		Return(models.DepartmentMember{}, repoErr).
		Once()

	_, err := s.service.AddDepartmentMember(s.ctx, models.AddDepartmentMemberInput{
		CompanyUUID:    companyID,
		DepartmentUUID: departmentID,
		RequestUser:    managerID,
		UserUUID:       userID,
		Role:           models.DepartmentMemberRoleEmployee,
	})

	s.Require().ErrorIs(err, repoErr)
}
