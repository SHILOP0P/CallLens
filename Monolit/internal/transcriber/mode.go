package transcriber

import (
	"context"

	"calllens/monolit/internal/models"
)

type ModeAware interface {
	ProviderForMode(mode models.TranscriptionMode) string
	TranscribeForMode(ctx context.Context, file models.File, mode models.TranscriptionMode) (models.TranscriptionResult, error)
}
