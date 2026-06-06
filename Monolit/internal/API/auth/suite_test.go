package auth

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/httpserver/middleware"
	serviceMocks "calllens/monolit/internal/service/mocks"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type APISuite struct {
	suite.Suite
	ctx     context.Context
	service *serviceMocks.AuthService
	api     *AuthHandler
}

func (s *APISuite) SetupTest() {
	s.ctx = context.Background()
	s.service = serviceMocks.NewAuthService(s.T())
	s.api = NewAuthHandler(s.service)
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}

func (s *APISuite) request(method string, path string, body string) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(s.ctx)
	return httptest.NewRecorder(), req
}

func (s *APISuite) requestWithUser(method string, path string, body string, userID uuid.UUID) (*httptest.ResponseRecorder, *http.Request) {
	rec, req := s.request(method, path, body)
	req = req.WithContext(middleware.ContextWithUserID(req.Context(), userID))
	return rec, req
}

func (s *APISuite) requestWithSession(method string, path string, body string, sessionID uuid.UUID) (*httptest.ResponseRecorder, *http.Request) {
	rec, req := s.request(method, path, body)
	req = req.WithContext(middleware.ContextWithSessionID(req.Context(), sessionID))
	return rec, req
}

func (s *APISuite) requireErrorCode(rec *httptest.ResponseRecorder, expectedCode string) {
	var resp response.ErrorResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp))
	s.Require().Equal(expectedCode, resp.Error.Code)
}
