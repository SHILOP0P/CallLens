package call

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) DeleteCall(ctx context.Context, id uuid.UUID) error {
	call, err := s.GetByUUID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repository.DeleteCall(ctx, id); err != nil {
		return err
	}

	if err := s.audioStorage.Delete(ctx, call.AudioPath); err != nil {
		return fmt.Errorf("delete audio file: %w", err)
	}

	return nil
}
