package auth

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestMeSuccess() {
	userID := uuid.New()

	s.service.On("Me", mock.Anything, userID).
		Return(models.User{ID: userID, Email: "user@example.com", FullName: "Dmitry", FullSurname: "Mukhachev", NickName: "muxa", Role: models.UserRoleUser, CreatedAt: time.Now().UTC()}, nil).
		Once()

	rec, req := s.requestWithUser(http.MethodGet, "/api/v1/auth/me", "", userID)

	s.api.Me(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestMeRequiresAuth() {
	rec, req := s.request(http.MethodGet, "/api/v1/auth/me", "")

	s.api.Me(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeUnauthorized)
}

func (s *APISuite) TestMeMapsUserNotFound() {
	userID := uuid.New()

	s.service.On("Me", mock.Anything, userID).
		Return(models.User{}, models.ErrUserNotFound).
		Once()

	rec, req := s.requestWithUser(http.MethodGet, "/api/v1/auth/me", "", userID)

	s.api.Me(rec, req)

	s.Require().Equal(http.StatusNotFound, rec.Code)
	s.requireErrorCode(rec, response.CodeUserNotFound)
}
