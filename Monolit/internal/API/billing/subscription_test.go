package billing

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestActivateCompanySubscriptionSuccess(t *testing.T) {
	companyID := uuid.New()
	userID := uuid.New()
	subscription := testSubscription(companyID, models.PlanCodeBusinessPlus, models.SubscriptionStatusActive)
	fake := &fakeBillingService{}
	fake.activate = func(ctx context.Context, input models.ActivateCompanySubscriptionInput) (models.Subscription, error) {
		require.Equal(t, companyID, input.CompanyUUID)
		require.Equal(t, userID, input.RequestUser)
		require.Equal(t, models.PlanCodeBusinessPlus, input.PlanCode)
		return subscription, nil
	}
	api := NewHandler(fake)

	rec, req := billingRequest(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/subscription/activate", `{"plan_code":"business_plus"}`, userID, map[string]string{"uuid": companyID.String()})
	api.ActivateCompanySubscription(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var respBody struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Plan   struct {
			Code string `json:"code"`
		} `json:"plan"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
	require.Equal(t, subscription.ID.String(), respBody.ID)
	require.Equal(t, string(models.SubscriptionStatusActive), respBody.Status)
	require.Equal(t, string(models.PlanCodeBusinessPlus), respBody.Plan.Code)
}

func TestActivatePersonalSubscriptionSuccess(t *testing.T) {
	userID := uuid.New()
	subscription := testPersonalSubscription(userID, models.PlanCodePersonalPlus, models.SubscriptionStatusActive)
	fake := &fakeBillingService{}
	fake.activatePersonal = func(ctx context.Context, input models.ActivatePersonalSubscriptionInput) (models.Subscription, error) {
		require.Equal(t, userID, input.UserUUID)
		require.Equal(t, models.PlanCodePersonalPlus, input.PlanCode)
		return subscription, nil
	}
	api := NewHandler(fake)

	rec, req := billingRequest(http.MethodPost, "/api/v1/subscription/activate", `{"plan_code":"personal_plus"}`, userID, nil)
	api.ActivatePersonalSubscription(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var respBody struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Plan   struct {
			Code string `json:"code"`
		} `json:"plan"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
	require.Equal(t, subscription.ID.String(), respBody.ID)
	require.Equal(t, string(models.SubscriptionStatusActive), respBody.Status)
	require.Equal(t, string(models.PlanCodePersonalPlus), respBody.Plan.Code)
}

func TestGetPersonalSubscriptionSuccess(t *testing.T) {
	userID := uuid.New()
	subscription := testPersonalSubscription(userID, models.PlanCodePersonalPro, models.SubscriptionStatusActive)
	fake := &fakeBillingService{}
	fake.getPersonal = func(ctx context.Context, id uuid.UUID) (models.Subscription, error) {
		require.Equal(t, userID, id)
		return subscription, nil
	}
	api := NewHandler(fake)

	rec, req := billingRequest(http.MethodGet, "/api/v1/subscription", "", userID, nil)
	api.GetPersonalSubscription(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCompanySubscriptionMapsNotFound(t *testing.T) {
	companyID := uuid.New()
	userID := uuid.New()
	fake := &fakeBillingService{}
	fake.getCompany = func(ctx context.Context, input models.GetCompanySubscriptionInput) (models.Subscription, error) {
		require.Equal(t, companyID, input.CompanyUUID)
		require.Equal(t, userID, input.RequestUser)
		return models.Subscription{}, models.ErrSubscriptionNotFound
	}
	api := NewHandler(fake)

	rec, req := billingRequest(http.MethodGet, "/api/v1/companies/"+companyID.String()+"/subscription", "", userID, map[string]string{"uuid": companyID.String()})
	api.GetCompanySubscription(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	requireBillingErrorCode(t, rec, response.CodeSubscriptionNotFound)
}

func TestActivateCompanySubscriptionRequiresAuth(t *testing.T) {
	companyID := uuid.New()
	api := NewHandler(&fakeBillingService{})

	rec, req := billingRequest(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/subscription/activate", "", uuid.Nil, map[string]string{"uuid": companyID.String()})
	api.ActivateCompanySubscription(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	requireBillingErrorCode(t, rec, response.CodeUnauthorized)
}

func TestActivateCompanySubscriptionRejectsInvalidBody(t *testing.T) {
	companyID := uuid.New()
	api := NewHandler(&fakeBillingService{})

	rec, req := billingRequest(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/subscription/activate", "{", uuid.New(), map[string]string{"uuid": companyID.String()})
	api.ActivateCompanySubscription(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	requireBillingErrorCode(t, rec, response.CodeInvalidRequestBody)
}

func TestCancelCompanySubscriptionMapsNotFound(t *testing.T) {
	companyID := uuid.New()
	userID := uuid.New()
	fake := &fakeBillingService{}
	fake.cancel = func(ctx context.Context, input models.CancelCompanySubscriptionInput) (models.Subscription, error) {
		require.Equal(t, companyID, input.CompanyUUID)
		require.Equal(t, userID, input.RequestUser)
		return models.Subscription{}, models.ErrSubscriptionNotFound
	}
	api := NewHandler(fake)

	rec, req := billingRequest(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/subscription/cancel", "", userID, map[string]string{"uuid": companyID.String()})
	api.CancelCompanySubscription(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	requireBillingErrorCode(t, rec, response.CodeSubscriptionNotFound)
}

func billingRequest(method string, path string, body string, userID uuid.UUID, params map[string]string) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != uuid.Nil {
		req = req.WithContext(middleware.ContextWithUserID(req.Context(), userID))
	}
	if len(params) > 0 {
		routeCtx := chi.NewRouteContext()
		for key, value := range params {
			routeCtx.URLParams.Add(key, value)
		}
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	}
	return httptest.NewRecorder(), req
}

func requireBillingErrorCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode string) {
	var resp response.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, expectedCode, resp.Error.Code)
}

func testSubscription(companyID uuid.UUID, planCode models.PlanCode, status models.SubscriptionStatus) models.Subscription {
	now := time.Now().UTC()
	return models.Subscription{
		ID:          uuid.New(),
		CompanyUUID: uuid.NullUUID{UUID: companyID, Valid: true},
		Status:      status,
		StartsAt:    now,
		CreatedAt:   now,
		UpdatedAt:   now,
		Plan: models.Plan{
			ID:                     uuid.New(),
			Code:                   planCode,
			Type:                   models.PlanTypeBusiness,
			Name:                   "Business",
			MonthlyMinutesLimit:    1000,
			ActiveInstructionLimit: 0,
			AnalysisLevel:          models.AnalysisLevelPlus,
			CreatedAt:              now,
			UpdatedAt:              now,
		},
	}
}

func testPersonalSubscription(userID uuid.UUID, planCode models.PlanCode, status models.SubscriptionStatus) models.Subscription {
	subscription := testSubscription(uuid.New(), planCode, status)
	subscription.UserUUID = uuid.NullUUID{UUID: userID, Valid: true}
	subscription.CompanyUUID = uuid.NullUUID{}
	subscription.Plan.Type = models.PlanTypePersonal
	subscription.Plan.Name = "Personal"
	return subscription
}

type fakeBillingService struct {
	list             func(context.Context) ([]models.Plan, error)
	getPersonal      func(context.Context, uuid.UUID) (models.Subscription, error)
	getCompany       func(context.Context, models.GetCompanySubscriptionInput) (models.Subscription, error)
	activatePersonal func(context.Context, models.ActivatePersonalSubscriptionInput) (models.Subscription, error)
	activate         func(context.Context, models.ActivateCompanySubscriptionInput) (models.Subscription, error)
	cancel           func(context.Context, models.CancelCompanySubscriptionInput) (models.Subscription, error)
}

func (f *fakeBillingService) ListPlans(ctx context.Context) ([]models.Plan, error) {
	return f.list(ctx)
}

func (f *fakeBillingService) GetPersonalSubscription(ctx context.Context, userID uuid.UUID) (models.Subscription, error) {
	return f.getPersonal(ctx, userID)
}

func (f *fakeBillingService) GetCompanySubscription(ctx context.Context, input models.GetCompanySubscriptionInput) (models.Subscription, error) {
	return f.getCompany(ctx, input)
}

func (f *fakeBillingService) ActivateCompanySubscription(ctx context.Context, input models.ActivateCompanySubscriptionInput) (models.Subscription, error) {
	return f.activate(ctx, input)
}

func (f *fakeBillingService) ActivatePersonalSubscription(ctx context.Context, input models.ActivatePersonalSubscriptionInput) (models.Subscription, error) {
	return f.activatePersonal(ctx, input)
}

func (f *fakeBillingService) CancelCompanySubscription(ctx context.Context, input models.CancelCompanySubscriptionInput) (models.Subscription, error) {
	return f.cancel(ctx, input)
}
