package analysis_instruction

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Service) GetFile(ctx context.Context, id uuid.UUID, userID uuid.UUID) (models.File, error) {
	if id == uuid.Nil || userID == uuid.Nil {
		return models.File{}, models.ErrInvalidAnalysisInstructionInput
	}

	instruction, err := s.repository.GetByUUID(ctx, id)
	if err != nil {
		return models.File{}, err
	}

	if err := s.authorizeRead(ctx, instruction, userID); err != nil {
		return models.File{}, err
	}

	content, err := s.instructionStorage.Open(ctx, instruction.FilePath)
	if err != nil {
		return models.File{}, fmt.Errorf("open instruction storage: %w", err)
	}

	return models.File{
		Content:          content,
		Path:             instruction.FilePath,
		OriginalFilename: instruction.OriginalFilename,
		MimeType:         instruction.MimeType,
		SizeBytes:        instruction.SizeBytes,
	}, nil
}
