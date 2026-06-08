package converter

import (
	model "calllens/monolit/internal/models"
	repoModel "calllens/monolit/internal/repository/models"
)

func RepoTranscriptionToModel(repoTranscription repoModel.Transcription) (model.Transcription, error) {
	return model.Transcription{
		ID:           repoTranscription.ID,
		CallUUID:     repoTranscription.CallUUID,
		Status:       model.TranscriptionStatus(repoTranscription.Status),
		Text:         nullStringToStringPtr(repoTranscription.Text),
		Language:     nullStringToStringPtr(repoTranscription.Language),
		Provider:     repoTranscription.Provider,
		ErrorMessage: nullStringToStringPtr(repoTranscription.ErrorMessage),
		CreatedAt:    repoTranscription.CreatedAt,
		UpdatedAt:    repoTranscription.UpdatedAt,
	}, nil
}

func ModelTranscriptionToRepoModel(transcription model.Transcription) (repoModel.Transcription, error) {
	return repoModel.Transcription{
		ID:           transcription.ID,
		CallUUID:     transcription.CallUUID,
		Status:       repoModel.TranscriptionStatus(transcription.Status),
		Text:         stringPtrToNullString(transcription.Text),
		Language:     stringPtrToNullString(transcription.Language),
		Provider:     transcription.Provider,
		ErrorMessage: stringPtrToNullString(transcription.ErrorMessage),
		CreatedAt:    transcription.CreatedAt,
		UpdatedAt:    transcription.UpdatedAt,
	}, nil
}
