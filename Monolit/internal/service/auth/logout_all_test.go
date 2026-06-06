package auth

import (
	"errors"

	"github.com/google/uuid"
)

func (s *ServiceSuite) TestLogoutAllSuccess() {
	userID := uuid.New()

	s.refreshSessionRepository.On("RevokeAllUserRefreshSessions", s.ctx, userID, logoutAllReason).
		Return(nil).
		Once()

	err := s.service.LogoutAll(s.ctx, userID)

	s.Require().NoError(err)
}

func (s *ServiceSuite) TestLogoutAllReturnsRepositoryError() {
	userID := uuid.New()
	repoErr := errors.New("db failed")

	s.refreshSessionRepository.On("RevokeAllUserRefreshSessions", s.ctx, userID, logoutAllReason).
		Return(repoErr).
		Once()

	err := s.service.LogoutAll(s.ctx, userID)

	s.Require().ErrorIs(err, repoErr)
}
