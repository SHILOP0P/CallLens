package call

import (
	"calllens/monolit/internal/models"
	"context"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) UpdateCallTitle(ctx context.Context, id uuid.UUID, title string) (models.Call, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return models.Call{}, models.ErrInvalidCallTitle
	}

	return s.repository.UpdateCallTitle(ctx, id, title)
}
