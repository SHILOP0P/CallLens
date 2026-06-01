package call

import (
	"calllens/monolit/internal/models"
	"context"
)

func (s *Service) List(ctx context.Context) ([]models.Call, error) {
	return s.repository.List(ctx)
}
