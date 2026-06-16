package invitation

import (
	"calllens/monolit/internal/models"
	"context"
	"errors"

	"github.com/google/uuid"
)

func (s *Service) CreateCompanyInvitation(ctx context.Context, input models.CreateCompanyInvitationInput) (models.MembershipInvitation, error) {
	if input.CompanyUUID == uuid.Nil || input.RequestUser == uuid.Nil || input.UserUUID == uuid.Nil || input.Role != models.CompanyMemberRoleEmployee {
		return models.MembershipInvitation{}, models.ErrInvalidInvitationInput
	}

	if err := s.ensureTargetUser(ctx, input.RequestUser, input.UserUUID); err != nil {
		return models.MembershipInvitation{}, err
	}

	if err := s.requireCompanyManager(ctx, input.CompanyUUID, input.RequestUser); err != nil {
		return models.MembershipInvitation{}, err
	}

	if err := s.requireActiveCompanySubscription(ctx, input.CompanyUUID); err != nil {
		return models.MembershipInvitation{}, err
	}

	active, err := s.isActiveCompanyMember(ctx, input.CompanyUUID, input.UserUUID)
	if err != nil {
		return models.MembershipInvitation{}, err
	}
	if active {
		return models.MembershipInvitation{}, models.ErrInvalidInvitationInput
	}

	now := s.now()
	return s.invitationRepository.CreateInvitation(ctx, models.MembershipInvitation{
		ID:                uuid.New(),
		CompanyUUID:       input.CompanyUUID,
		InvitedUserUUID:   input.UserUUID,
		InvitedByUserUUID: input.RequestUser,
		CompanyRole:       models.CompanyMemberRoleEmployee,
		Status:            models.InvitationStatusPending,
		ExpiresAt:         now.Add(defaultInvitationTTL),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
}

func (s *Service) CreateDepartmentInvitation(ctx context.Context, input models.CreateDepartmentInvitationInput) (models.MembershipInvitation, error) {
	if input.CompanyUUID == uuid.Nil || input.DepartmentUUID == uuid.Nil || input.RequestUser == uuid.Nil || input.UserUUID == uuid.Nil {
		return models.MembershipInvitation{}, models.ErrInvalidInvitationInput
	}
	if input.Role != models.DepartmentMemberRoleEmployee && input.Role != models.DepartmentMemberRoleLeader {
		return models.MembershipInvitation{}, models.ErrInvalidInvitationInput
	}

	if err := s.ensureTargetUser(ctx, input.RequestUser, input.UserUUID); err != nil {
		return models.MembershipInvitation{}, err
	}

	if err := s.requireDepartmentInvitePermission(ctx, input); err != nil {
		return models.MembershipInvitation{}, err
	}

	if err := s.requireActiveCompanySubscription(ctx, input.CompanyUUID); err != nil {
		return models.MembershipInvitation{}, err
	}

	activeCompanyMember, err := s.isActiveCompanyMember(ctx, input.CompanyUUID, input.UserUUID)
	if err != nil {
		return models.MembershipInvitation{}, err
	}
	if !activeCompanyMember {
		return models.MembershipInvitation{}, models.ErrForbidden
	}

	activeDepartmentMember, err := s.isActiveDepartmentMember(ctx, input.CompanyUUID, input.DepartmentUUID, input.UserUUID)
	if err != nil {
		return models.MembershipInvitation{}, err
	}
	if activeDepartmentMember {
		return models.MembershipInvitation{}, models.ErrInvalidInvitationInput
	}

	now := s.now()
	role := input.Role
	return s.invitationRepository.CreateInvitation(ctx, models.MembershipInvitation{
		ID:                uuid.New(),
		CompanyUUID:       input.CompanyUUID,
		DepartmentUUID:    uuid.NullUUID{UUID: input.DepartmentUUID, Valid: true},
		InvitedUserUUID:   input.UserUUID,
		InvitedByUserUUID: input.RequestUser,
		CompanyRole:       models.CompanyMemberRoleEmployee,
		DepartmentRole:    &role,
		Status:            models.InvitationStatusPending,
		ExpiresAt:         now.Add(defaultInvitationTTL),
		CreatedAt:         now,
		UpdatedAt:         now,
	})
}

func (s *Service) requireDepartmentInvitePermission(ctx context.Context, input models.CreateDepartmentInvitationInput) error {
	manager, err := s.companyRepository.GetCompanyMember(ctx, input.CompanyUUID, input.RequestUser)
	if err == nil && manager.Role == models.CompanyMemberRoleManager {
		return nil
	}
	if err != nil && !errors.Is(err, models.ErrCompanyNotFound) {
		return err
	}

	member, err := s.departmentRepository.GetDepartmentMember(ctx, input.CompanyUUID, input.DepartmentUUID, input.RequestUser)
	if err != nil {
		return err
	}
	if member.Role != models.DepartmentMemberRoleLeader {
		return models.ErrForbidden
	}
	if input.Role != models.DepartmentMemberRoleEmployee {
		return models.ErrForbidden
	}

	return nil
}
