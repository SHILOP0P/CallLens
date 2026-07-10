package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/models"

	"github.com/stretchr/testify/require"
)

func TestGetCapabilities(t *testing.T) {
	adminService := &adminServiceStub{
		capabilities: models.AdminCapabilities{
			Role: models.UserRoleHelper,
			Permissions: []models.AdminPermission{
				models.AdminPermissionPanelAccess,
				models.AdminPermissionUsersRead,
			},
		},
	}
	handler := NewHandler(adminService)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/capabilities", nil)
	req = req.WithContext(middleware.ContextWithUserRole(req.Context(), string(models.UserRoleHelper)))
	rec := httptest.NewRecorder()

	handler.GetCapabilities(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var body dto.AdminCapabilitiesResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, string(models.UserRoleHelper), body.Role)
	require.Equal(t, []string{"admin.panel.access", "admin.users.read"}, body.Permissions)
	require.Equal(t, models.UserRoleHelper, adminService.role)
}

func TestGetCapabilitiesRequiresRole(t *testing.T) {
	handler := NewHandler(&adminServiceStub{})
	rec := httptest.NewRecorder()

	handler.GetCapabilities(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/capabilities", nil))

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	requireErrorCode(t, rec, response.CodeUnauthorized)
}

func TestGetCapabilitiesMapsErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "forbidden", err: models.ErrForbidden, status: http.StatusForbidden, code: response.CodeForbidden},
		{name: "internal", err: errors.New("failed"), status: http.StatusInternalServerError, code: response.CodeFailedToGetAdminCapabilities},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(&adminServiceStub{err: tt.err})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/capabilities", nil)
			req = req.WithContext(middleware.ContextWithUserRole(req.Context(), string(models.UserRoleUser)))
			rec := httptest.NewRecorder()

			handler.GetCapabilities(rec, req)

			require.Equal(t, tt.status, rec.Code)
			requireErrorCode(t, rec, tt.code)
		})
	}
}

type adminServiceStub struct {
	capabilities models.AdminCapabilities
	role         models.UserRole
	err          error
}

func (s *adminServiceStub) GetCapabilities(_ context.Context, role models.UserRole) (models.AdminCapabilities, error) {
	s.role = role
	return s.capabilities, s.err
}

func (s *adminServiceStub) RecordAudit(_ context.Context, _ models.CreateAdminAuditLogInput) (models.AdminAuditLog, error) {
	return models.AdminAuditLog{}, nil
}

func requireErrorCode(t *testing.T, rec *httptest.ResponseRecorder, code string) {
	t.Helper()
	var body response.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, code, body.Error.Code)
}
