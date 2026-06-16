package billing

import (
	"calllens/monolit/internal/models"
	"context"
)

func (s *Service) ListPlans(ctx context.Context) ([]models.Plan, error) {
	return s.repository.ListPlans(ctx)
}
