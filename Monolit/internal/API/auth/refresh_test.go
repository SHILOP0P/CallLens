package auth

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

func (s *APISuite) TestRefreshSuccess() {
	userID := uuid.New()

	s.service.On("Refresh", mock.Anything, models.RefreshTokenInput{RefreshToken: "refresh"}).
		Return(models.User{ID: userID, Email: "user@example.com", FullName: "Dmitry", FullSurname: "Mukhachev", NickName: "muxa", Role: models.UserRoleUser, CreatedAt: time.Now().UTC()}, "access", "new-refresh", nil).
		Once()

	rec, req := s.request(http.MethodPost, "/api/v1/auth/refresh", "")
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "refresh", Path: refreshTokenCookiePath})

	s.api.Refresh(rec, req)

	s.Require().Equal(http.StatusOK, rec.Code)
	s.requireAuthCookies(rec, "access", "new-refresh")

	var resp dto.AuthResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp))
	s.Require().Equal(userID.String(), resp.User.ID)
	s.Require().NotContains(rec.Body.String(), "access_token")
	s.Require().NotContains(rec.Body.String(), "refresh_token")
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

	rec, req := s.request(http.MethodPost, "/api/v1/auth/refresh", "")
	req.AddCookie(&http.Cookie{Name: refreshTokenCookieName, Value: "bad", Path: refreshTokenCookiePath})

	s.api.Refresh(rec, req)

	s.Require().Equal(http.StatusUnauthorized, rec.Code)
	s.requireErrorCode(rec, response.CodeInvalidRefreshToken)
	s.requireClearedAuthCookies(rec)
}
