package auth

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (models.User, error) {
	return s.userRepository.GetUserByUUID(ctx, userID)
}
