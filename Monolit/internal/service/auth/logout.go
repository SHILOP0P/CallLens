package auth

import (
	"context"

	"github.com/google/uuid"
)

const logoutReason = "logout"

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.refreshSessionRepository.RevokeRefreshSession(ctx, sessionID, logoutReason)
}
