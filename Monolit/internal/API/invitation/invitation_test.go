package invitation

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

func TestCreateCompanyInvitationSuccess(t *testing.T) {
	companyID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()
	fake := &fakeInvitationService{}
	fake.createCompany = func(ctx context.Context, input models.CreateCompanyInvitationInput) (models.MembershipInvitation, error) {
		require.Equal(t, companyID, input.CompanyUUID)
		require.Equal(t, requestUserID, input.RequestUser)
		require.Equal(t, userID, input.UserUUID)
		return testAPIInvitation(companyID, userID, requestUserID), nil
	}
	api := NewHandler(fake)

	rec, req := request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/invitations", `{"user_uuid":"`+userID.String()+`","role":"employee"}`, requestUserID, map[string]string{"uuid": companyID.String()})
	api.CreateCompanyInvitation(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestCreateDepartmentInvitationSuccess(t *testing.T) {
	companyID := uuid.New()
	departmentID := uuid.New()
	requestUserID := uuid.New()
	userID := uuid.New()
	fake := &fakeInvitationService{}
	fake.createDepartment = func(ctx context.Context, input models.CreateDepartmentInvitationInput) (models.MembershipInvitation, error) {
		require.Equal(t, departmentID, input.DepartmentUUID)
		role := models.DepartmentMemberRoleEmployee
		invitation := testAPIInvitation(companyID, userID, requestUserID)
		invitation.DepartmentUUID = uuid.NullUUID{UUID: departmentID, Valid: true}
		invitation.DepartmentRole = &role
		return invitation, nil
	}
	api := NewHandler(fake)

	rec, req := request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/departments/"+departmentID.String()+"/invitations", `{"user_uuid":"`+userID.String()+`","role":"employee"}`, requestUserID, map[string]string{"uuid": companyID.String(), "department_uuid": departmentID.String()})
	api.CreateDepartmentInvitation(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestListAcceptDeclineAndErrors(t *testing.T) {
	companyID := uuid.New()
	userID := uuid.New()
	invitationID := uuid.New()
	fake := &fakeInvitationService{}
	fake.list = func(ctx context.Context, input models.ListUserInvitationsInput) ([]models.MembershipInvitation, error) {
		require.Equal(t, userID, input.UserUUID)
		return []models.MembershipInvitation{testAPIInvitation(companyID, userID, uuid.New())}, nil
	}
	fake.accept = func(ctx context.Context, input models.AcceptInvitationInput) (models.MembershipInvitation, error) {
		require.Equal(t, invitationID, input.InvitationUUID)
		return testAPIInvitation(companyID, userID, uuid.New()), nil
	}
	fake.decline = func(ctx context.Context, input models.DeclineInvitationInput) (models.MembershipInvitation, error) {
		return testAPIInvitation(companyID, userID, uuid.New()), nil
	}
	api := NewHandler(fake)

	rec, req := request(http.MethodGet, "/api/v1/invitations", "", userID, nil)
	api.ListUserInvitations(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	rec, req = request(http.MethodPost, "/api/v1/invitations/"+invitationID.String()+"/accept", "", userID, map[string]string{"invitation_uuid": invitationID.String()})
	api.AcceptInvitation(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	rec, req = request(http.MethodPost, "/api/v1/invitations/"+invitationID.String()+"/decline", "", userID, map[string]string{"invitation_uuid": invitationID.String()})
	api.DeclineInvitation(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	fake.accept = func(ctx context.Context, input models.AcceptInvitationInput) (models.MembershipInvitation, error) {
		return models.MembershipInvitation{}, models.ErrForbidden
	}
	rec, req = request(http.MethodPost, "/api/v1/invitations/"+invitationID.String()+"/accept", "", userID, map[string]string{"invitation_uuid": invitationID.String()})
	api.AcceptInvitation(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
	requireErrorCode(t, rec, response.CodeForbidden)

	fake.createCompany = func(ctx context.Context, input models.CreateCompanyInvitationInput) (models.MembershipInvitation, error) {
		return models.MembershipInvitation{}, models.ErrInvitationAlreadyExists
	}
	rec, req = request(http.MethodPost, "/api/v1/companies/"+companyID.String()+"/invitations", `{"user_uuid":"`+userID.String()+`","role":"employee"}`, userID, map[string]string{"uuid": companyID.String()})
	api.CreateCompanyInvitation(rec, req)
	require.Equal(t, http.StatusConflict, rec.Code)
	requireErrorCode(t, rec, response.CodeInvitationAlreadyExists)
}

func testAPIInvitation(companyID uuid.UUID, invitedUserID uuid.UUID, invitedByUserID uuid.UUID) models.MembershipInvitation {
	now := time.Now().UTC()
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

func request(method string, path string, body string, userID uuid.UUID, params map[string]string) (*httptest.ResponseRecorder, *http.Request) {
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

func requireErrorCode(t *testing.T, rec *httptest.ResponseRecorder, expectedCode string) {
	var resp response.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, expectedCode, resp.Error.Code)
}

type fakeInvitationService struct {
	createCompany    func(context.Context, models.CreateCompanyInvitationInput) (models.MembershipInvitation, error)
	createDepartment func(context.Context, models.CreateDepartmentInvitationInput) (models.MembershipInvitation, error)
	list             func(context.Context, models.ListUserInvitationsInput) ([]models.MembershipInvitation, error)
	accept           func(context.Context, models.AcceptInvitationInput) (models.MembershipInvitation, error)
	decline          func(context.Context, models.DeclineInvitationInput) (models.MembershipInvitation, error)
	cancel           func(context.Context, models.CancelInvitationInput) (models.MembershipInvitation, error)
}

func (f *fakeInvitationService) CreateCompanyInvitation(ctx context.Context, input models.CreateCompanyInvitationInput) (models.MembershipInvitation, error) {
	return f.createCompany(ctx, input)
}

func (f *fakeInvitationService) CreateDepartmentInvitation(ctx context.Context, input models.CreateDepartmentInvitationInput) (models.MembershipInvitation, error) {
	return f.createDepartment(ctx, input)
}

func (f *fakeInvitationService) ListUserInvitations(ctx context.Context, input models.ListUserInvitationsInput) ([]models.MembershipInvitation, error) {
	return f.list(ctx, input)
}

func (f *fakeInvitationService) AcceptInvitation(ctx context.Context, input models.AcceptInvitationInput) (models.MembershipInvitation, error) {
	return f.accept(ctx, input)
}

func (f *fakeInvitationService) DeclineInvitation(ctx context.Context, input models.DeclineInvitationInput) (models.MembershipInvitation, error) {
	return f.decline(ctx, input)
}

func (f *fakeInvitationService) CancelInvitation(ctx context.Context, input models.CancelInvitationInput) (models.MembershipInvitation, error) {
	return f.cancel(ctx, input)
}
