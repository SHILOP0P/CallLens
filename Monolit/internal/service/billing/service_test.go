package billing

import (
	"calllens/monolit/internal/models"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCanUploadPersonalCallWhenMinutesExceeded(t *testing.T) {
	userID := uuid.New()
	subscriptionID := uuid.New()
	repository := &fakeRepository{
		personalSubscription: models.Subscription{
			ID: subscriptionID,
			Plan: models.Plan{
				MonthlyMinutesLimit: 10,
			},
		},
		usedMinutes: 9,
	}

	service := NewService(repository)

	err := service.CanUploadPersonalCall(context.Background(), userID, 61)

	require.ErrorIs(t, err, models.ErrMonthlyMinutesLimitExceeded)
}

func TestPersonalStartCannotCreateMoreThanTwoActivePersonalInstructions(t *testing.T) {
	repository := &fakeRepository{
		personalSubscription: models.Subscription{
			ID: uuid.New(),
			Plan: models.Plan{
				ActiveInstructionLimit: 2,
			},
		},
		instructionsCount: 2,
	}

	service := NewService(repository)

	err := service.CanCreatePersonalInstruction(context.Background(), uuid.New())

	require.ErrorIs(t, err, models.ErrInstructionLimitExceeded)
}

func TestCanCreateDepartmentWhenLimitExceeded(t *testing.T) {
	companyID := uuid.New()
	limit := 5
	repository := &fakeRepository{
		businessSubscription: models.Subscription{
			ID: uuid.New(),
			Plan: models.Plan{
				DepartmentsPerCompanyLimit: &limit,
			},
		},
		departmentsCount: 5,
	}

	service := NewService(repository)

	err := service.CanCreateDepartment(context.Background(), companyID)

	require.ErrorIs(t, err, models.ErrDepartmentLimitExceeded)
}

func TestBusinessPlusCannotCreateMoreThanSevenActiveDepartmentInstructions(t *testing.T) {
	limit := 7
	repository := &fakeRepository{
		businessSubscription: models.Subscription{
			ID: uuid.New(),
			Plan: models.Plan{
				InstructionsPerDepartmentLimit: &limit,
			},
		},
		instructionsCount: 7,
	}

	service := NewService(repository)

	err := service.CanCreateDepartmentInstruction(context.Background(), uuid.New(), uuid.New())

	require.ErrorIs(t, err, models.ErrInstructionLimitExceeded)
}

func TestCanCreateCompanyDoesNotRequireBusinessSubscription(t *testing.T) {
	service := NewService(&fakeRepository{
		businessErr: models.ErrSubscriptionNotFound,
	})

	err := service.CanCreateCompany(context.Background(), uuid.New())

	require.NoError(t, err)
}

func TestAPIAccessAvailableOnlyOnBusinessPro(t *testing.T) {
	repository := &fakeRepository{
		businessSubscription: models.Subscription{
			ID: uuid.New(),
			Plan: models.Plan{
				APIAccessEnabled: false,
			},
		},
	}

	service := NewService(repository)

	err := service.CanAccessAPI(context.Background(), uuid.New())

	require.ErrorIs(t, err, models.ErrAPIAccessDenied)

	repository.businessSubscription.Plan.APIAccessEnabled = true

	err = service.CanAccessAPI(context.Background(), uuid.New())

	require.NoError(t, err)
}

func TestSubscriptionRequiredWhenActiveSubscriptionIsMissing(t *testing.T) {
	repository := &fakeRepository{
		personalErr: models.ErrSubscriptionNotFound,
	}

	service := NewService(repository)

	err := service.CanCreatePersonalInstruction(context.Background(), uuid.New())

	require.ErrorIs(t, err, models.ErrSubscriptionRequired)
}

func TestManagerBusinessSubscriptionRaisesPersonalLevel(t *testing.T) {
	userID := uuid.New()
	subscriptionID := uuid.New()
	repository := &fakeRepository{
		personalSubscription: models.Subscription{
			ID: subscriptionID,
			Plan: models.Plan{
				Code:                models.PlanCodePersonalStart,
				Type:                models.PlanTypePersonal,
				MonthlyMinutesLimit: 120,
			},
		},
		managerBusinessSubscription: models.Subscription{
			ID: uuid.New(),
			Plan: models.Plan{
				Code: models.PlanCodeBusinessStart,
				Type: models.PlanTypeBusiness,
			},
		},
		plans: map[models.PlanCode]models.Plan{
			models.PlanCodePersonalPlus: {
				Code:                models.PlanCodePersonalPlus,
				Type:                models.PlanTypePersonal,
				MonthlyMinutesLimit: 600,
			},
		},
		usedMinutes: 120,
	}

	service := NewService(repository)

	err := service.CanUploadPersonalCall(context.Background(), userID, 60)

	require.NoError(t, err)
}

func TestActivatePersonalSubscriptionRejectsBusinessPlan(t *testing.T) {
	repository := &fakeRepository{
		plan: models.Plan{
			Code: models.PlanCodeBusinessStart,
			Type: models.PlanTypeBusiness,
		},
	}
	service := NewService(repository)

	_, err := service.ActivatePersonalSubscription(context.Background(), models.ActivatePersonalSubscriptionInput{
		UserUUID: uuid.New(),
		PlanCode: models.PlanCodeBusinessStart,
	})

	require.ErrorIs(t, err, models.ErrInvalidBillingInput)
}

func TestActivatePersonalSubscriptionDefaultsPersonalStart(t *testing.T) {
	userID := uuid.New()
	repository := &fakeRepository{
		plan: models.Plan{
			Code: models.PlanCodePersonalStart,
			Type: models.PlanTypePersonal,
		},
		activatePersonalSubscription: models.Subscription{
			ID:       uuid.New(),
			UserUUID: uuid.NullUUID{UUID: userID, Valid: true},
			Status:   models.SubscriptionStatusActive,
			Plan: models.Plan{
				Code: models.PlanCodePersonalStart,
				Type: models.PlanTypePersonal,
			},
		},
	}
	service := NewService(repository)

	_, err := service.ActivatePersonalSubscription(context.Background(), models.ActivatePersonalSubscriptionInput{
		UserUUID: userID,
	})

	require.NoError(t, err)
	require.Equal(t, models.PlanCodePersonalStart, repository.activatePersonalInput.PlanCode)
	require.Equal(t, userID, repository.activatePersonalInput.UserUUID)
}

func TestActivateCompanySubscriptionDefaultsBusinessPlanForManager(t *testing.T) {
	companyID := uuid.New()
	managerID := uuid.New()
	repository := &fakeRepository{
		plan: models.Plan{
			Code: models.PlanCodeBusinessStart,
			Type: models.PlanTypeBusiness,
		},
		activateSubscription: models.Subscription{
			ID:          uuid.New(),
			CompanyUUID: uuid.NullUUID{UUID: companyID, Valid: true},
			Status:      models.SubscriptionStatusActive,
		},
	}
	service := NewService(repository)
	service.SetCompanyRepository(&fakeCompanyRepository{
		member: models.CompanyMember{
			CompanyUUID: companyID,
			UserUUID:    managerID,
			Role:        models.CompanyMemberRoleManager,
			Status:      models.MembershipStatusActive,
		},
	})

	_, err := service.ActivateCompanySubscription(context.Background(), models.ActivateCompanySubscriptionInput{
		CompanyUUID: companyID,
		RequestUser: managerID,
	})

	require.NoError(t, err)
	require.Equal(t, models.PlanCodeBusinessStart, repository.activateInput.PlanCode)
	require.Equal(t, companyID, repository.activateInput.CompanyUUID)
	require.Equal(t, managerID, repository.activateInput.RequestUser)
}

func TestActivateCompanySubscriptionRejectsNonManager(t *testing.T) {
	companyID := uuid.New()
	userID := uuid.New()
	service := NewService(&fakeRepository{})
	service.SetCompanyRepository(&fakeCompanyRepository{
		member: models.CompanyMember{
			CompanyUUID: companyID,
			UserUUID:    userID,
			Role:        models.CompanyMemberRoleEmployee,
			Status:      models.MembershipStatusActive,
		},
	})

	_, err := service.ActivateCompanySubscription(context.Background(), models.ActivateCompanySubscriptionInput{
		CompanyUUID: companyID,
		RequestUser: userID,
	})

	require.ErrorIs(t, err, models.ErrForbidden)
}

func TestActivateCompanySubscriptionRejectsPersonalPlan(t *testing.T) {
	companyID := uuid.New()
	managerID := uuid.New()
	repository := &fakeRepository{
		plan: models.Plan{
			Code: models.PlanCodePersonalStart,
			Type: models.PlanTypePersonal,
		},
	}
	service := NewService(repository)
	service.SetCompanyRepository(&fakeCompanyRepository{
		member: models.CompanyMember{
			CompanyUUID: companyID,
			UserUUID:    managerID,
			Role:        models.CompanyMemberRoleManager,
			Status:      models.MembershipStatusActive,
		},
	})

	_, err := service.ActivateCompanySubscription(context.Background(), models.ActivateCompanySubscriptionInput{
		CompanyUUID: companyID,
		RequestUser: managerID,
		PlanCode:    models.PlanCodePersonalStart,
	})

	require.ErrorIs(t, err, models.ErrInvalidBillingInput)
}

type fakeRepository struct {
	personalSubscription         models.Subscription
	businessSubscription         models.Subscription
	managerBusinessSubscription  models.Subscription
	plan                         models.Plan
	plans                        map[models.PlanCode]models.Plan
	personalErr                  error
	businessErr                  error
	managerBusinessErr           error
	planErr                      error
	usedMinutes                  int
	companiesCount               int
	departmentsCount             int
	membersCount                 int
	instructionsCount            int
	activatePersonalInput        models.ActivatePersonalSubscriptionInput
	activatePersonalSubscription models.Subscription
	activatePersonalErr          error
	activateInput                models.ActivateCompanySubscriptionInput
	activateSubscription         models.Subscription
	activateErr                  error
	cancelCompanyID              uuid.UUID
	cancelSubscription           models.Subscription
	cancelErr                    error
}

func (f *fakeRepository) GetPlanByCode(_ context.Context, code models.PlanCode) (models.Plan, error) {
	if f.planErr != nil {
		return models.Plan{}, f.planErr
	}
	if f.plans != nil {
		if plan, ok := f.plans[code]; ok {
			return plan, nil
		}
	}
	if f.plan.Code == "" {
		switch code {
		case models.PlanCodePersonalStart, models.PlanCodePersonalPlus, models.PlanCodePersonalPro:
			return models.Plan{Code: code, Type: models.PlanTypePersonal}, nil
		default:
			return models.Plan{Code: code, Type: models.PlanTypeBusiness}, nil
		}
	}
	return f.plan, nil
}

func (f *fakeRepository) ListPlans(context.Context) ([]models.Plan, error) {
	return nil, nil
}

func (f *fakeRepository) GetActivePersonalSubscription(context.Context, uuid.UUID) (models.Subscription, error) {
	if f.personalErr != nil {
		return models.Subscription{}, f.personalErr
	}
	return f.personalSubscription, nil
}

func (f *fakeRepository) GetActiveBusinessSubscription(context.Context, uuid.UUID) (models.Subscription, error) {
	if f.businessErr != nil {
		return models.Subscription{}, f.businessErr
	}
	return f.businessSubscription, nil
}

func (f *fakeRepository) GetBestActiveBusinessSubscriptionForManager(context.Context, uuid.UUID) (models.Subscription, error) {
	if f.managerBusinessErr != nil {
		return models.Subscription{}, f.managerBusinessErr
	}
	if f.managerBusinessSubscription.ID == uuid.Nil {
		return models.Subscription{}, models.ErrSubscriptionNotFound
	}
	return f.managerBusinessSubscription, nil
}

func (f *fakeRepository) UpsertSubscription(context.Context, models.UpsertSubscriptionInput) (models.Subscription, error) {
	return models.Subscription{}, nil
}

func (f *fakeRepository) ActivatePersonalSubscription(_ context.Context, input models.ActivatePersonalSubscriptionInput, _ time.Time) (models.Subscription, error) {
	f.activatePersonalInput = input
	if f.activatePersonalErr != nil {
		return models.Subscription{}, f.activatePersonalErr
	}
	return f.activatePersonalSubscription, nil
}

func (f *fakeRepository) ActivateCompanySubscription(_ context.Context, input models.ActivateCompanySubscriptionInput, _ time.Time) (models.Subscription, error) {
	f.activateInput = input
	if f.activateErr != nil {
		return models.Subscription{}, f.activateErr
	}
	return f.activateSubscription, nil
}

func (f *fakeRepository) CancelCompanySubscription(_ context.Context, companyID uuid.UUID, _ time.Time) (models.Subscription, error) {
	f.cancelCompanyID = companyID
	if f.cancelErr != nil {
		return models.Subscription{}, f.cancelErr
	}
	return f.cancelSubscription, nil
}

func (f *fakeRepository) CountUsedMinutes(context.Context, uuid.UUID, time.Time) (int, error) {
	return f.usedMinutes, nil
}

func (f *fakeRepository) AddUsageMinutes(context.Context, uuid.UUID, time.Time, int) (models.UsageCounter, error) {
	return models.UsageCounter{}, nil
}

func (f *fakeRepository) CountOwnerCompanies(context.Context, uuid.UUID) (int, error) {
	return f.companiesCount, nil
}

func (f *fakeRepository) CountCompanyDepartments(context.Context, uuid.UUID) (int, error) {
	return f.departmentsCount, nil
}

func (f *fakeRepository) CountCompanyMembers(context.Context, uuid.UUID) (int, error) {
	return f.membersCount, nil
}

func (f *fakeRepository) CountActiveInstructions(context.Context, models.ListAnalysisInstructionsInput) (int, error) {
	return f.instructionsCount, nil
}

type fakeCompanyRepository struct {
	member models.CompanyMember
	err    error
}

func (f *fakeCompanyRepository) GetCompanyMember(context.Context, uuid.UUID, uuid.UUID) (models.CompanyMember, error) {
	if f.err != nil {
		return models.CompanyMember{}, f.err
	}
	return f.member, nil
}
