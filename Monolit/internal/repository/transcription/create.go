package transcription

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"fmt"
)

func (r *Repository) Create(ctx context.Context, transcription model.Transcription) (model.Transcription, error) {
	repoTranscription, err := converter.ModelTranscriptionToRepoModel(transcription)
	if err != nil {
		return model.Transcription{}, model.ErrInvalidTranscriptionInput
	}

	query := `
	INSERT INTO call_transcriptions (
		transcription_uuid,
		call_uuid,
		status,
		text,
		language,
		provider,
		error_message,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (call_uuid) DO UPDATE
	SET status = EXCLUDED.status,
	    text = NULL,
	    language = NULL,
	    provider = EXCLUDED.provider,
	    error_message = NULL,
	    updated_at = now()
	RETURNING ` + transcriptionReturningColumns

	row := r.db.QueryRowContext(ctx, query,
		repoTranscription.ID,
		repoTranscription.CallUUID,
		repoTranscription.Status,
		repoTranscription.Text,
		repoTranscription.Language,
		repoTranscription.Provider,
		repoTranscription.ErrorMessage,
		repoTranscription.CreatedAt,
		repoTranscription.UpdatedAt,
	)

	createdTranscription, err := scaner.ScanTranscription(row)
	if err != nil {
		return model.Transcription{}, fmt.Errorf("create transcription: %w", err)
	}

	return converter.RepoTranscriptionToModel(createdTranscription)
}
