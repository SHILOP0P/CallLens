package analytics

import (
	"context"

	"calllens/monolit/internal/models"
	"calllens/monolit/internal/repository"
)

type Service struct {
	analyticsRepository repository.AnalyticsRepository
}

func NewService(analyticsRepository repository.AnalyticsRepository) *Service {
	return &Service{analyticsRepository: analyticsRepository}
}

func (s *Service) GetOverview(ctx context.Context, input models.AnalyticsOverviewInput) (models.AnalyticsOverview, error) {
	return s.analyticsRepository.GetAnalyticsOverview(ctx, input)
}
