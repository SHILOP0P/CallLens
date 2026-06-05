package auth

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const logoutReason = "logout"

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.refreshSessionRepository.RevokeRefreshSession(ctx, sessionID, logoutReason); err != nil {
		s.log.Error(ctx, "logout failed", zap.String("session_id", sessionID.String()), zap.Error(err))
		return err
	}

	s.log.Info(ctx, "user logged out", zap.String("session_id", sessionID.String()))

	return nil
}
