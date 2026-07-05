package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apiMocks "calllens/monolit/internal/API/mocks"
	"calllens/monolit/internal/logger"
	repositoryMocks "calllens/monolit/internal/repository/mocks"

	"github.com/stretchr/testify/require"
)

func TestNewRouterRegistersPublicAndProtectedRoutes(t *testing.T) {
	router := NewRouter(
		apiMocks.NewCallAPI(t),
		stubCallFolderAPI{},
		apiMocks.NewAuthAPI(t),
		apiMocks.NewCompanyAPI(t),
		apiMocks.NewDepartmentAPI(t),
		apiMocks.NewAnalysisInstructionAPI(t),
		apiMocks.NewAnalysisAPI(t),
		apiMocks.NewReportAPI(t),
		apiMocks.NewBillingAPI(t),
		apiMocks.NewInvitationAPI(t),
		apiMocks.NewAnalyticsAPI(t),
		apiMocks.NewMonitoringAPI(t),
		stubSearchAPI{},
		stubNotificationAPI{},
		nil,
		"test-secret",
		repositoryMocks.NewRefreshSessionRepository(t),
		logger.NewNop(),
	)

	healthRecorder := httptest.NewRecorder()
	router.ServeHTTP(healthRecorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	require.Equal(t, http.StatusOK, healthRecorder.Code)

	readyRecorder := httptest.NewRecorder()
	router.ServeHTTP(readyRecorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	require.Equal(t, http.StatusOK, readyRecorder.Code)

	protectedRecorder := httptest.NewRecorder()
	router.ServeHTTP(protectedRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/calls", nil))
	require.Equal(t, http.StatusUnauthorized, protectedRecorder.Code)

	notFoundRecorder := httptest.NewRecorder()
	router.ServeHTTP(notFoundRecorder, httptest.NewRequest(http.MethodGet, "/missing", nil))
	require.Equal(t, http.StatusNotFound, notFoundRecorder.Code)
}

type stubSearchAPI struct{}

func (stubSearchAPI) Search(w http.ResponseWriter, r *http.Request) {}

type stubCallFolderAPI struct{}

func (stubCallFolderAPI) Create(w http.ResponseWriter, r *http.Request)     {}
func (stubCallFolderAPI) List(w http.ResponseWriter, r *http.Request)       {}
func (stubCallFolderAPI) Get(w http.ResponseWriter, r *http.Request)        {}
func (stubCallFolderAPI) Update(w http.ResponseWriter, r *http.Request)     {}
func (stubCallFolderAPI) Delete(w http.ResponseWriter, r *http.Request)     {}
func (stubCallFolderAPI) ListCalls(w http.ResponseWriter, r *http.Request)  {}
func (stubCallFolderAPI) AssignCall(w http.ResponseWriter, r *http.Request) {}
func (stubCallFolderAPI) RemoveCall(w http.ResponseWriter, r *http.Request) {}

type stubNotificationAPI struct{}

func (stubNotificationAPI) List(w http.ResponseWriter, r *http.Request)        {}
func (stubNotificationAPI) MarkRead(w http.ResponseWriter, r *http.Request)    {}
func (stubNotificationAPI) MarkAllRead(w http.ResponseWriter, r *http.Request) {}
