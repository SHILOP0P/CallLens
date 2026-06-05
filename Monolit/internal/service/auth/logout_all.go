package auth

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const logoutAllReason = "logout_all"

func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.refreshSessionRepository.RevokeAllUserRefreshSessions(ctx, userID, logoutAllReason); err != nil {
		s.log.Error(ctx, "logout all failed", zap.String("user_id", userID.String()), zap.Error(err))
		return err
	}

	s.log.Info(ctx, "user logged out from all sessions", zap.String("user_id", userID.String()))

	return nil
}
