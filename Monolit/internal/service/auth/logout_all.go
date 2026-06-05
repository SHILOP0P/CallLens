package auth

import (
	"context"

	"github.com/google/uuid"
)

const logoutAllReason = "logout_all"

func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.refreshSessionRepository.RevokeAllUserRefreshSessions(ctx, userID, logoutAllReason)
}
