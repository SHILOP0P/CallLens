package invitation

import (
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/models"
	repo "calllens/monolit/internal/repository"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const defaultInvitationTTL = 7 * 24 * time.Hour

type BillingLimiter interface {
	CanUseCompany(ctx context.Context, companyID uuid.UUID) error
	CanAddCompanyMember(ctx context.Context, companyID uuid.UUID) error
}

type Service struct {
	invitationRepository repo.InvitationRepository
	userRepository       repo.UserRepository
	companyRepository    repo.CompanyRepository
	departmentRepository repo.DepartmentRepository
	billingLimiter       BillingLimiter
	now                  func() time.Time
	log                  logger.Logger
}

func NewService(invitationRepository repo.InvitationRepository, userRepository repo.UserRepository, companyRepository repo.CompanyRepository, departmentRepository repo.DepartmentRepository, log logger.Logger) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		invitationRepository: invitationRepository,
		userRepository:       userRepository,
		companyRepository:    companyRepository,
		departmentRepository: departmentRepository,
		now:                  func() time.Time { return time.Now().UTC() },
		log:                  log,
	}
}

func (s *Service) SetBillingLimiter(limiter BillingLimiter) {
	s.billingLimiter = limiter
}

func (s *Service) SetNow(now func() time.Time) {
	if now != nil {
		s.now = now
	}
}

func validInvitationStatus(status models.InvitationStatus) bool {
	return status == "" ||
		status == models.InvitationStatusPending ||
		status == models.InvitationStatusAccepted ||
		status == models.InvitationStatusDeclined ||
		status == models.InvitationStatusCanceled ||
		status == models.InvitationStatusExpired
}

func (s *Service) ensureTargetUser(ctx context.Context, requestUser uuid.UUID, targetUser uuid.UUID) error {
	if requestUser == uuid.Nil || targetUser == uuid.Nil || requestUser == targetUser {
		return models.ErrInvalidInvitationInput
	}

	_, err := s.userRepository.GetUserByUUID(ctx, targetUser)
	return err
}

func (s *Service) isActiveCompanyMember(ctx context.Context, companyID uuid.UUID, userID uuid.UUID) (bool, error) {
	_, err := s.companyRepository.GetCompanyMember(ctx, companyID, userID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, models.ErrCompanyNotFound) {
		return false, nil
	}
	return false, err
}

func (s *Service) isActiveDepartmentMember(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID, userID uuid.UUID) (bool, error) {
	_, err := s.departmentRepository.GetDepartmentMember(ctx, companyID, departmentID, userID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, models.ErrDepartmentNotFound) {
		return false, nil
	}
	return false, err
}
