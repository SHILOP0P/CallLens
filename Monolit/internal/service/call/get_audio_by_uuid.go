package call

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) GetAudioByUUID(ctx context.Context, uuid uuid.UUID) (models.File, error) {
	call, err := s.repository.GetByUUID(ctx, uuid)
	if err != nil {
		return models.File{}, err
	}

	content, err := s.audioStorage.Open(ctx, call.AudioPath)
	if err != nil {
		return models.File{}, fmt.Errorf("error opening audio storage: %w", err)
	}

	return models.File{
		Content:          content,
		Path:             call.AudioPath,
		OriginalFilename: call.OriginalFilename,
		MimeType:         call.MimeType,
		SizeBytes:        call.SizeBytes,
	}, nil
}
