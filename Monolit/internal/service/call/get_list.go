package call

import (
	"context"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]models.Call, error) {
	return s.repository.List(ctx, userID)
}

func (s *Service) ListFiltered(ctx context.Context, input models.ListCallsInput) (models.ListCallsResult, error) {
	return s.repository.ListFiltered(ctx, input)
}

func (s *Service) GetFilterOptions(ctx context.Context, input models.CallFilterOptionsInput) (models.CallFilterOptions, error) {
	return s.repository.GetFilterOptions(ctx, input)
}
