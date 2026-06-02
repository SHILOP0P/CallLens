package call

import (
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Service) CreateCall(ctx context.Context, input models.CreateCallInput) (models.Call, error) {
	if err := validateAudioInput(input); err != nil {
		return models.Call{}, err
	}

	callUUID, err := uuid.NewV7()
	if err != nil {
		return models.Call{}, err
	}

	savedFile, err := s.audioStorage.Save(ctx, models.SaveInput{
		CallID:           callUUID,
		OriginalFilename: input.OriginalFilename,
		Content:          input.Content,
		SizeBytes:        input.SizeBytes,
		MimeType:         input.MimeType,
	})

	if err != nil {
		return models.Call{}, err
	}

	now := time.Now().UTC()
	call, err := converter.SavedFileToModel(savedFile, callUUID, input, now)
	if err != nil {
		_ = s.audioStorage.Delete(context.Background(), savedFile.Path)
		return models.Call{}, err
	}

	createdCall, err := s.repository.CreateCall(ctx, call)
	if err != nil {
		_ = s.audioStorage.Delete(context.Background(), savedFile.Path)
		return models.Call{}, err
	}

	return createdCall, nil
}
