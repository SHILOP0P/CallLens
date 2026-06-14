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

func TestBusinessProCanHaveUpToThreeCompanies(t *testing.T) {
	limit := 3
	repository := &fakeRepository{
		businessOwnerSubscription: models.Subscription{
			ID: uuid.New(),
			Plan: models.Plan{
				CompanyLimit: &limit,
			},
		},
		companiesCount: 2,
	}

	service := NewService(repository)

	err := service.CanCreateCompany(context.Background(), uuid.New())

	require.NoError(t, err)
}

func TestBusinessProRejectsFourthCompany(t *testing.T) {
	limit := 3
	repository := &fakeRepository{
		businessOwnerSubscription: models.Subscription{
			ID: uuid.New(),
			Plan: models.Plan{
				CompanyLimit: &limit,
			},
		},
		companiesCount: 3,
	}

	service := NewService(repository)

	err := service.CanCreateCompany(context.Background(), uuid.New())

	require.ErrorIs(t, err, models.ErrCompanyLimitExceeded)
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

type fakeRepository struct {
	personalSubscription      models.Subscription
	businessSubscription      models.Subscription
	businessOwnerSubscription models.Subscription
	personalErr               error
	businessErr               error
	businessOwnerErr          error
	usedMinutes               int
	companiesCount            int
	departmentsCount          int
	membersCount              int
	instructionsCount         int
}

func (f *fakeRepository) GetPlanByCode(context.Context, models.PlanCode) (models.Plan, error) {
	return models.Plan{}, nil
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

func (f *fakeRepository) GetActiveBusinessSubscriptionForOwner(context.Context, uuid.UUID) (models.Subscription, error) {
	if f.businessOwnerErr != nil {
		return models.Subscription{}, f.businessOwnerErr
	}
	return f.businessOwnerSubscription, nil
}

func (f *fakeRepository) UpsertSubscription(context.Context, models.UpsertSubscriptionInput) (models.Subscription, error) {
	return models.Subscription{}, nil
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
