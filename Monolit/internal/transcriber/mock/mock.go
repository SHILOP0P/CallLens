package mock

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"
)

type Transcriber struct{}

func New() *Transcriber {
	return &Transcriber{}
}

func (t *Transcriber) Provider() string {
	return "mock"
}

func (t *Transcriber) Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error) {
	select {
	case <-ctx.Done():
		return models.TranscriptionResult{}, ctx.Err()
	default:
	}

	language := "ru"

	return models.TranscriptionResult{
		Text:     fmt.Sprintf("Mock transcription for %s", file.OriginalFilename),
		Language: &language,
	}, nil
}
