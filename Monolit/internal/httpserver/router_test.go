package httpserver

import (
	apiMocks "calllens/monolit/internal/API/mocks"
	"calllens/monolit/internal/logger"
	repositoryMocks "calllens/monolit/internal/repository/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRouterRegistersPublicAndProtectedRoutes(t *testing.T) {
	router := NewRouter(
		apiMocks.NewCallAPI(t),
		apiMocks.NewAuthAPI(t),
		apiMocks.NewCompanyAPI(t),
		apiMocks.NewDepartmentAPI(t),
		apiMocks.NewAnalysisInstructionAPI(t),
		apiMocks.NewAnalysisAPI(t),
		apiMocks.NewReportAPI(t),
		apiMocks.NewBillingAPI(t),
		apiMocks.NewInvitationAPI(t),
		"test-secret",
		repositoryMocks.NewRefreshSessionRepository(t),
		logger.NewNop(),
	)

	healthRecorder := httptest.NewRecorder()
	router.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusOK, healthRecorder.Code)

	protectedRecorder := httptest.NewRecorder()
	router.ServeHTTP(protectedRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/calls", nil))
	require.Equal(t, http.StatusUnauthorized, protectedRecorder.Code)

	notFoundRecorder := httptest.NewRecorder()
	router.ServeHTTP(notFoundRecorder, httptest.NewRequest(http.MethodGet, "/missing", nil))
	require.Equal(t, http.StatusNotFound, notFoundRecorder.Code)
}
