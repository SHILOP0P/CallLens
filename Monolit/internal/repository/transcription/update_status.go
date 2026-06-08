package transcription

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) MarkTranscribed(ctx context.Context, id uuid.UUID, text string, language *string) (model.Transcription, error) {
	query := `
	UPDATE call_transcriptions
	SET status = $2,
	    text = $3,
	    language = $4,
	    error_message = NULL,
	    updated_at = now()
	WHERE transcription_uuid = $1
	RETURNING ` + transcriptionReturningColumns

	row := r.db.QueryRowContext(ctx, query, id, string(model.TranscriptionStatusTranscribed), text, language)

	return scanUpdatedTranscription(row, "mark transcription transcribed")
}

func (r *Repository) MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) (model.Transcription, error) {
	query := `
	UPDATE call_transcriptions
	SET status = $2,
	    text = NULL,
	    language = NULL,
	    error_message = $3,
	    updated_at = now()
	WHERE transcription_uuid = $1
	RETURNING ` + transcriptionReturningColumns

	row := r.db.QueryRowContext(ctx, query, id, string(model.TranscriptionStatusFailed), errorMessage)

	return scanUpdatedTranscription(row, "mark transcription failed")
}

func scanUpdatedTranscription(row interface {
	Scan(dest ...any) error
}, operation string) (model.Transcription, error) {
	repoTranscription, err := scaner.ScanTranscription(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Transcription{}, model.ErrTranscriptionNotFound
		}
		return model.Transcription{}, fmt.Errorf("%s: %w", operation, err)
	}

	return converter.RepoTranscriptionToModel(repoTranscription)
}
