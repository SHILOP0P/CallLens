package auth

import (
	"errors"

	"github.com/google/uuid"
)

func (s *ServiceSuite) TestLogoutSuccess() {
	sessionID := uuid.New()

	s.refreshSessionRepository.On("RevokeRefreshSession", s.ctx, sessionID, logoutReason).
		Return(nil).
		Once()

	err := s.service.Logout(s.ctx, sessionID)

	s.Require().NoError(err)
}

func (s *ServiceSuite) TestLogoutReturnsRepositoryError() {
	sessionID := uuid.New()
	repoErr := errors.New("db failed")

	s.refreshSessionRepository.On("RevokeRefreshSession", s.ctx, sessionID, logoutReason).
		Return(repoErr).
		Once()

	err := s.service.Logout(s.ctx, sessionID)

	s.Require().ErrorIs(err, repoErr)
}
