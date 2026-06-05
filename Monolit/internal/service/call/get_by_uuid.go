package call

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

func (s *Service) GetByUUID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.Call, error) {
	return s.repository.GetByUUID(ctx, id, userID)
}
