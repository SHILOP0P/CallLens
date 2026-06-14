package billing

import (
	"calllens/monolit/internal/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetPlanByCode(ctx context.Context, code models.PlanCode) (models.Plan, error)
	ListPlans(ctx context.Context) ([]models.Plan, error)
	GetActivePersonalSubscription(ctx context.Context, userID uuid.UUID) (models.Subscription, error)
	GetActiveBusinessSubscription(ctx context.Context, companyID uuid.UUID) (models.Subscription, error)
	GetActiveBusinessSubscriptionForOwner(ctx context.Context, ownerID uuid.UUID) (models.Subscription, error)
	UpsertSubscription(ctx context.Context, input models.UpsertSubscriptionInput) (models.Subscription, error)
	CountUsedMinutes(ctx context.Context, subscriptionID uuid.UUID, periodStart time.Time) (int, error)
	AddUsageMinutes(ctx context.Context, subscriptionID uuid.UUID, periodStart time.Time, minutes int) (models.UsageCounter, error)
	CountOwnerCompanies(ctx context.Context, ownerID uuid.UUID) (int, error)
	CountCompanyDepartments(ctx context.Context, companyID uuid.UUID) (int, error)
	CountCompanyMembers(ctx context.Context, companyID uuid.UUID) (int, error)
	CountActiveInstructions(ctx context.Context, input models.ListAnalysisInstructionsInput) (int, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) ListPlans(ctx context.Context) ([]models.Plan, error) {
	return s.repository.ListPlans(ctx)
}

func (s *Service) CanCreatePersonalInstruction(ctx context.Context, userID uuid.UUID) error {
	subscription, err := s.activePersonalSubscription(ctx, userID)
	if err != nil {
		return err
	}

	count, err := s.repository.CountActiveInstructions(ctx, models.ListAnalysisInstructionsInput{
		Scope:    models.AnalysisInstructionScopePersonal,
		UserUUID: userID,
	})
	if err != nil {
		return err
	}

	if count >= subscription.Plan.ActiveInstructionLimit {
		return models.ErrInstructionLimitExceeded
	}

	return nil
}

func (s *Service) CanCreateCompanyInstruction(ctx context.Context, companyID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	count, err := s.repository.CountActiveInstructions(ctx, models.ListAnalysisInstructionsInput{
		Scope:       models.AnalysisInstructionScopeCompany,
		CompanyUUID: uuid.NullUUID{UUID: companyID, Valid: true},
	})
	if err != nil {
		return err
	}

	if subscription.Plan.InstructionsPerDepartmentLimit == nil {
		return models.ErrInstructionLimitExceeded
	}

	if count >= *subscription.Plan.InstructionsPerDepartmentLimit {
		return models.ErrInstructionLimitExceeded
	}

	return nil
}

func (s *Service) CanCreateDepartmentInstruction(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	count, err := s.repository.CountActiveInstructions(ctx, models.ListAnalysisInstructionsInput{
		Scope:          models.AnalysisInstructionScopeDepartment,
		CompanyUUID:    uuid.NullUUID{UUID: companyID, Valid: true},
		DepartmentUUID: uuid.NullUUID{UUID: departmentID, Valid: true},
	})
	if err != nil {
		return err
	}

	if subscription.Plan.InstructionsPerDepartmentLimit == nil {
		return models.ErrInstructionLimitExceeded
	}

	if count >= *subscription.Plan.InstructionsPerDepartmentLimit {
		return models.ErrInstructionLimitExceeded
	}

	return nil
}

func (s *Service) CanUploadPersonalCall(ctx context.Context, userID uuid.UUID, durationSeconds int) error {
	subscription, err := s.activePersonalSubscription(ctx, userID)
	if err != nil {
		return err
	}

	return s.canUseMinutes(ctx, subscription, durationSeconds)
}

func (s *Service) CanUploadBusinessCall(ctx context.Context, companyID uuid.UUID, durationSeconds int) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	return s.canUseMinutes(ctx, subscription, durationSeconds)
}

func (s *Service) AddPersonalUsageMinutes(ctx context.Context, userID uuid.UUID, durationSeconds int) error {
	subscription, err := s.activePersonalSubscription(ctx, userID)
	if err != nil {
		return err
	}

	return s.addUsageMinutes(ctx, subscription.ID, durationSeconds)
}

func (s *Service) AddBusinessUsageMinutes(ctx context.Context, companyID uuid.UUID, durationSeconds int) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	return s.addUsageMinutes(ctx, subscription.ID, durationSeconds)
}

func (s *Service) CanCreateCompany(ctx context.Context, ownerID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscriptionForOwner(ctx, ownerID)
	if err != nil {
		return err
	}

	if subscription.Plan.CompanyLimit == nil {
		return models.ErrCompanyLimitExceeded
	}

	count, err := s.repository.CountOwnerCompanies(ctx, ownerID)
	if err != nil {
		return err
	}

	if count >= *subscription.Plan.CompanyLimit {
		return models.ErrCompanyLimitExceeded
	}

	return nil
}

func (s *Service) CanCreateDepartment(ctx context.Context, companyID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	if subscription.Plan.DepartmentsPerCompanyLimit == nil {
		return models.ErrDepartmentLimitExceeded
	}

	count, err := s.repository.CountCompanyDepartments(ctx, companyID)
	if err != nil {
		return err
	}

	if count >= *subscription.Plan.DepartmentsPerCompanyLimit {
		return models.ErrDepartmentLimitExceeded
	}

	return nil
}

func (s *Service) CanAddCompanyMember(ctx context.Context, companyID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	if subscription.Plan.MembersPerCompanyLimit == nil {
		return models.ErrMemberLimitExceeded
	}

	count, err := s.repository.CountCompanyMembers(ctx, companyID)
	if err != nil {
		return err
	}

	if count >= *subscription.Plan.MembersPerCompanyLimit {
		return models.ErrMemberLimitExceeded
	}

	return nil
}

func (s *Service) CanAccessAPI(ctx context.Context, companyID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	if !subscription.Plan.APIAccessEnabled {
		return models.ErrAPIAccessDenied
	}

	return nil
}

func (s *Service) CanExportReports(ctx context.Context, companyID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	if !subscription.Plan.ExportEnabled {
		return models.ErrExportAccessDenied
	}

	return nil
}

func (s *Service) CanAccessTeamAnalytics(ctx context.Context, companyID uuid.UUID) error {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return err
	}

	if !subscription.Plan.TeamAnalyticsEnabled {
		return models.ErrTeamAnalyticsAccessDenied
	}

	return nil
}

func (s *Service) AnalysisLevelForUser(ctx context.Context, userID uuid.UUID) (models.AnalysisLevel, error) {
	subscription, err := s.activePersonalSubscription(ctx, userID)
	if err != nil {
		return "", err
	}

	return subscription.Plan.AnalysisLevel, nil
}

func (s *Service) AnalysisLevelForCompany(ctx context.Context, companyID uuid.UUID) (models.AnalysisLevel, error) {
	subscription, err := s.activeBusinessSubscription(ctx, companyID)
	if err != nil {
		return "", err
	}

	return subscription.Plan.AnalysisLevel, nil
}

func (s *Service) canUseMinutes(ctx context.Context, subscription models.Subscription, durationSeconds int) error {
	minutes := minutesFromSeconds(durationSeconds)
	if minutes == 0 {
		return nil
	}

	usedMinutes, err := s.repository.CountUsedMinutes(ctx, subscription.ID, s.now())
	if err != nil {
		return err
	}

	if usedMinutes+minutes > subscription.Plan.MonthlyMinutesLimit {
		return models.ErrMonthlyMinutesLimitExceeded
	}

	return nil
}

func (s *Service) addUsageMinutes(ctx context.Context, subscriptionID uuid.UUID, durationSeconds int) error {
	minutes := minutesFromSeconds(durationSeconds)
	if minutes == 0 {
		return nil
	}

	_, err := s.repository.AddUsageMinutes(ctx, subscriptionID, s.now(), minutes)
	return err
}

func (s *Service) activePersonalSubscription(ctx context.Context, userID uuid.UUID) (models.Subscription, error) {
	subscription, err := s.repository.GetActivePersonalSubscription(ctx, userID)
	return normalizeSubscriptionError(subscription, err)
}

func (s *Service) activeBusinessSubscription(ctx context.Context, companyID uuid.UUID) (models.Subscription, error) {
	subscription, err := s.repository.GetActiveBusinessSubscription(ctx, companyID)
	return normalizeSubscriptionError(subscription, err)
}

func (s *Service) activeBusinessSubscriptionForOwner(ctx context.Context, ownerID uuid.UUID) (models.Subscription, error) {
	subscription, err := s.repository.GetActiveBusinessSubscriptionForOwner(ctx, ownerID)
	return normalizeSubscriptionError(subscription, err)
}

func normalizeSubscriptionError(subscription models.Subscription, err error) (models.Subscription, error) {
	if err == nil {
		return subscription, nil
	}

	if errors.Is(err, models.ErrSubscriptionNotFound) {
		return models.Subscription{}, models.ErrSubscriptionRequired
	}

	return models.Subscription{}, err
}

func minutesFromSeconds(seconds int) int {
	if seconds <= 0 {
		return 0
	}

	return (seconds + 59) / 60
}
