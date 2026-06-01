package call

import (
	"calllens/monolit/internal/models"
	"context"
)

func (s *Service) CreateCall(ctx context.Context, call models.Call) (models.Call, error) {
	return s.repository.CreateCall(ctx, call)
}
