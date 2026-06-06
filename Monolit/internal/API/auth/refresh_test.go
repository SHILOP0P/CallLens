package auth

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestRefreshSuccess() {
	userID := uuid.New()
	body := `{"refresh_token":"refresh"}`

	s.service.On("Refresh", mock.Anything, models.RefreshTokenInput{RefreshToken: "refresh"}).
		Return(models.User{ID: userID, Email: "user@example.com", FullName: "Dmitry", FullSurname: "Mukhachev", NickName: "muxa", Role: models.UserRoleUser, CreatedAt: time.Now().UTC()}, "access", "new-refresh", nil).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/auth/refresh", body)

	s.api.Refresh(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
}

func (s *APISuite) TestRefreshRejectsInvalidBody() {
	rec, req := s.request(http.MethodPost, "/api/v1/auth/refresh", `{`)

	s.api.Refresh(rec, req)

	s.Require().Equal(http.StatusBadRequest, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidRequestBody)
}

func (s *APISuite) TestRefreshMapsInvalidRefreshToken() {
	s.service.On("Refresh", mock.Anything, models.RefreshTokenInput{RefreshToken: "bad"}).
		Return(models.User{}, "", "", models.ErrInvalidRefreshToken).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"bad"}`)

	s.api.Refresh(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidRefreshToken)
}
