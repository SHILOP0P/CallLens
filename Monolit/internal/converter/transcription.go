package converter

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/models"
	"time"
)

func TranscriptionModelToAPI(transcription models.Transcription) (dto.TranscriptionResponse, error) {
	return dto.TranscriptionResponse{
		ID:           transcription.ID.String(),
		CallUUID:     transcription.CallUUID.String(),
		Status:       string(transcription.Status),
		Text:         transcription.Text,
		Language:     transcription.Language,
		Provider:     transcription.Provider,
		ErrorMessage: transcription.ErrorMessage,
		CreatedAt:    transcription.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    transcription.UpdatedAt.Format(time.RFC3339),
	}, nil
}
