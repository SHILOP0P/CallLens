package call

import (
	"context"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]models.Call, error) {
	return s.repository.List(ctx, userID)
}
