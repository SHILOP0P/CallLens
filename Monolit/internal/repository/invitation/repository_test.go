package invitation

import (
	"calllens/monolit/internal/models"
	"time"

	"github.com/google/uuid"
)

func (s *RepositorySuite) createUser(email string) models.User {
	user := models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "hash",
		FullName:     "Dmitry",
		FullSurname:  "Mukhachev",
		NickName:     "muxa",
		Role:         models.UserRoleUser,
		CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
	}

	created, err := s.userRepository.CreateUser(s.ctx, user)
	s.Require().NoError(err)

	return created
}

func (s *RepositorySuite) createCompanyWithManager() (models.Company, models.User) {
	manager := s.createUser(uuid.NewString() + "@example.com")
	company := models.Company{
		ID:              uuid.New(),
		Name:            "CallLens",
		ManagerUserUUID: manager.ID,
		MemberLimit:     5,
		CreatedAt:       time.Now().UTC().Truncate(time.Microsecond),
	}
	member := models.CompanyMember{
		CompanyUUID: company.ID,
		UserUUID:    manager.ID,
		Role:        models.CompanyMemberRoleManager,
		Status:      models.MembershipStatusActive,
		CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
	}

	created, err := s.companyRepository.CreateCompany(s.ctx, company, member)
	s.Require().NoError(err)

	return created, manager
}

func (s *RepositorySuite) createDepartment(companyID uuid.UUID) models.Department {
	department := models.Department{
		ID:          uuid.New(),
		CompanyUUID: companyID,
		Name:        "Sales",
		CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
	}
	created, err := s.departmentRepository.CreateDepartment(s.ctx, department)
	s.Require().NoError(err)

	return created
}

func (s *RepositorySuite) addCompanyMember(companyID uuid.UUID, userID uuid.UUID, status models.MembershipStatus) {
	_, err := s.companyRepository.AddCompanyMember(s.ctx, models.CompanyMember{
		CompanyUUID: companyID,
		UserUUID:    userID,
		Role:        models.CompanyMemberRoleEmployee,
		Status:      status,
		CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
	})
	s.Require().NoError(err)
}

func testInvitation(companyID uuid.UUID, invitedUserID uuid.UUID, invitedByUserID uuid.UUID) models.MembershipInvitation {
	now := time.Now().UTC().Truncate(time.Microsecond)
	return models.MembershipInvitation{
		ID:                uuid.New(),
		CompanyUUID:       companyID,
		InvitedUserUUID:   invitedUserID,
		InvitedByUserUUID: invitedByUserID,
		CompanyRole:       models.CompanyMemberRoleEmployee,
		Status:            models.InvitationStatusPending,
		ExpiresAt:         now.Add(time.Hour),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (s *RepositorySuite) TestCreateListGetInvitation() {
	company, manager := s.createCompanyWithManager()
	invited := s.createUser(uuid.NewString() + "@example.com")
	invitation := testInvitation(company.ID, invited.ID, manager.ID)

	created, err := s.repository.CreateInvitation(s.ctx, invitation)
	s.Require().NoError(err)
	s.Require().Equal(invitation.ID, created.ID)

	got, err := s.repository.GetInvitationByUUID(s.ctx, invitation.ID)
	s.Require().NoError(err)
	s.Require().Equal(invited.ID, got.InvitedUserUUID)

	list, err := s.repository.ListUserInvitations(s.ctx, models.ListUserInvitationsInput{
		UserUUID: invited.ID,
		Status:   models.InvitationStatusPending,
	})
	s.Require().NoError(err)
	s.Require().Len(list, 1)
	s.Require().Equal(invitation.ID, list[0].ID)
}

func (s *RepositorySuite) TestUniquePendingCompanyInvitation() {
	company, manager := s.createCompanyWithManager()
	invited := s.createUser(uuid.NewString() + "@example.com")
	invitation := testInvitation(company.ID, invited.ID, manager.ID)

	_, err := s.repository.CreateInvitation(s.ctx, invitation)
	s.Require().NoError(err)

	duplicate := testInvitation(company.ID, invited.ID, manager.ID)
	_, err = s.repository.CreateInvitation(s.ctx, duplicate)
	s.Require().ErrorIs(err, models.ErrInvitationAlreadyExists)
}

func (s *RepositorySuite) TestUniquePendingDepartmentInvitation() {
	company, manager := s.createCompanyWithManager()
	department := s.createDepartment(company.ID)
	invited := s.createUser(uuid.NewString() + "@example.com")
	role := models.DepartmentMemberRoleEmployee
	invitation := testInvitation(company.ID, invited.ID, manager.ID)
	invitation.DepartmentUUID = uuid.NullUUID{UUID: department.ID, Valid: true}
	invitation.DepartmentRole = &role

	_, err := s.repository.CreateInvitation(s.ctx, invitation)
	s.Require().NoError(err)

	duplicate := testInvitation(company.ID, invited.ID, manager.ID)
	duplicate.DepartmentUUID = uuid.NullUUID{UUID: department.ID, Valid: true}
	duplicate.DepartmentRole = &role
	_, err = s.repository.CreateInvitation(s.ctx, duplicate)
	s.Require().ErrorIs(err, models.ErrInvitationAlreadyExists)
}

func (s *RepositorySuite) TestAcceptCompanyInvitationCreatesAndReactivatesMember() {
	company, manager := s.createCompanyWithManager()
	invited := s.createUser(uuid.NewString() + "@example.com")
	s.addCompanyMember(company.ID, invited.ID, models.MembershipStatusLeft)

	invitation := testInvitation(company.ID, invited.ID, manager.ID)
	created, err := s.repository.CreateInvitation(s.ctx, invitation)
	s.Require().NoError(err)

	accepted, err := s.repository.AcceptInvitation(s.ctx, created.ID, time.Now().UTC())
	s.Require().NoError(err)
	s.Require().Equal(models.InvitationStatusAccepted, accepted.Status)

	member, err := s.companyRepository.GetCompanyMember(s.ctx, company.ID, invited.ID)
	s.Require().NoError(err)
	s.Require().Equal(models.MembershipStatusActive, member.Status)
}

func (s *RepositorySuite) TestAcceptDepartmentInvitationCreatesDepartmentMemberForCompanyMember() {
	company, manager := s.createCompanyWithManager()
	department := s.createDepartment(company.ID)
	invited := s.createUser(uuid.NewString() + "@example.com")
	s.addCompanyMember(company.ID, invited.ID, models.MembershipStatusActive)
	role := models.DepartmentMemberRoleEmployee
	invitation := testInvitation(company.ID, invited.ID, manager.ID)
	invitation.DepartmentUUID = uuid.NullUUID{UUID: department.ID, Valid: true}
	invitation.DepartmentRole = &role
	created, err := s.repository.CreateInvitation(s.ctx, invitation)
	s.Require().NoError(err)

	accepted, err := s.repository.AcceptInvitation(s.ctx, created.ID, time.Now().UTC())
	s.Require().NoError(err)
	s.Require().Equal(models.InvitationStatusAccepted, accepted.Status)

	member, err := s.departmentRepository.GetDepartmentMember(s.ctx, company.ID, department.ID, invited.ID)
	s.Require().NoError(err)
	s.Require().Equal(models.DepartmentMemberRoleEmployee, member.Role)
}

func (s *RepositorySuite) TestAcceptExpiredInvitationMarksExpired() {
	company, manager := s.createCompanyWithManager()
	invited := s.createUser(uuid.NewString() + "@example.com")
	invitation := testInvitation(company.ID, invited.ID, manager.ID)
	invitation.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	created, err := s.repository.CreateInvitation(s.ctx, invitation)
	s.Require().NoError(err)

	_, err = s.repository.AcceptInvitation(s.ctx, created.ID, time.Now().UTC())
	s.Require().ErrorIs(err, models.ErrInvitationExpired)

	got, err := s.repository.GetInvitationByUUID(s.ctx, created.ID)
	s.Require().NoError(err)
	s.Require().Equal(models.InvitationStatusExpired, got.Status)
}
