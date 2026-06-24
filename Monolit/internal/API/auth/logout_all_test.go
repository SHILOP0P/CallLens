package auth

import (
	"errors"
	"net/http"

	"calllens/monolit/internal/API/response"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestLogoutAllSuccess() {
	userID := uuid.New()

	s.service.On("LogoutAll", mock.Anything, userID).Return(nil).Once()

	rec, req := s.requestWithUser(http.MethodPost, "/api/v1/auth/logout-all", "", userID)

	s.api.LogoutAll(rec, req)

	s.Require().Equal(http.StatusNoContent, rec.Code)
	s.requireClearedAuthCookies(rec)
}

func (s *APISuite) TestLogoutAllRequiresAuth() {
	rec, req := s.request(http.MethodPost, "/api/v1/auth/logout-all", "")

	s.api.LogoutAll(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeUnauthorized)
}

func (s *APISuite) TestLogoutAllMapsServiceError() {
	userID := uuid.New()

	s.service.On("LogoutAll", mock.Anything, userID).Return(errors.New("db failed")).Once()

	rec, req := s.requestWithUser(http.MethodPost, "/api/v1/auth/logout-all", "", userID)

	s.api.LogoutAll(rec, req)

	s.Require().Equal(http.StatusInternalServerError, rec.Code)
	s.requireErrorCode(rec, response.CodeFailedToLogoutAll)
}
