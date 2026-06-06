package auth

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestLoginSuccess() {
	userID := uuid.New()
	body := `{"email":"user@example.com","password":"password123"}`

	s.service.On("Login", mock.Anything, mock.MatchedBy(func(input models.LoginInput) bool {
		return input.Email == "user@example.com" &&
			input.Password == "password123" &&
			input.IPAddress != nil
	})).
		Return(models.User{ID: userID, Email: "user@example.com", FullName: "Dmitry", FullSurname: "Mukhachev", NickName: "muxa", Role: models.UserRoleUser, CreatedAt: time.Now().UTC()}, "access", "refresh", nil).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/auth/login", body)

	s.api.Login(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestLoginRejectsInvalidBody() {
	rec, req := s.request(http.MethodPost, "/api/v1/auth/login", `{`)

	s.api.Login(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidRequestBody)
}

func (s *APISuite) TestLoginMapsInvalidCredentials() {
	body := `{"email":"user@example.com","password":"bad"}`

	s.service.On("Login", mock.Anything, mock.MatchedBy(func(input models.LoginInput) bool {
		return input.Email == "user@example.com" && input.Password == "bad"
	})).
		Return(models.User{}, "", "", models.ErrInvalidCredentials).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/auth/login", body)

	s.api.Login(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidCredentials)
}
