package transcriber

import (
	"calllens/monolit/internal/models"
	"context"
)

type Transcriber interface {
	Provider() string
	Transcribe(ctx context.Context, file models.File) (models.TranscriptionResult, error)
}
