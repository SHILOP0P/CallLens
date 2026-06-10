package processing

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) ProcessCall(ctx context.Context, callID uuid.UUID) {
	if callID == uuid.Nil {
		return
	}

	if err := s.ProcessTranscribeCall(ctx, callID); err != nil {
		s.log.Error(ctx, "call processing failed", zap.String("call_id", callID.String()), zap.Error(err))
	}
}

func (s *Service) ProcessClaimedCall(ctx context.Context, call models.Call) {
	if call.ID == uuid.Nil {
		return
	}

	if err := s.processTranscribeCall(ctx, call); err != nil {
		s.log.Error(ctx, "claimed call processing failed", zap.String("call_id", call.ID.String()), zap.Error(err))
	}
}

func (s *Service) ProcessJob(ctx context.Context, job models.ProcessingJob) error {
	switch job.Type {
	case models.ProcessingJobTypeTranscribeCall:
		return s.ProcessTranscribeCall(ctx, job.EntityUUID)
	default:
		return models.ErrInvalidProcessingJobType
	}
}

func (s *Service) ProcessTranscribeCall(ctx context.Context, callID uuid.UUID) error {
	if callID == uuid.Nil {
		return models.ErrCallNotFound
	}

	if s.transcriber == nil {
		return models.ErrTranscriberNotConfigured
	}

	call, err := s.callRepository.GetByUUIDForProcessing(ctx, callID)
	if err != nil {
		return fmt.Errorf("get call for processing: %w", err)
	}

	return s.processTranscribeCall(ctx, call)
}

func (s *Service) processTranscribeCall(ctx context.Context, call models.Call) error {
	if s.transcriber == nil {
		return models.ErrTranscriberNotConfigured
	}

	if call.Status == models.CallStatusTranscribed {
		s.log.Info(ctx, "call already transcribed", zap.String("call_id", call.ID.String()))
		return nil
	}

	if call.Status == models.CallStatusNew {
		updatedCall, err := s.callRepository.UpdateCallStatus(ctx, call.ID, models.CallStatusProcessing)
		if err != nil {
			return fmt.Errorf("mark call processing: %w", err)
		}

		call = updatedCall
	}

	if call.Status != models.CallStatusProcessing {
		return fmt.Errorf("%w: %s", models.ErrInvalidCallStatusTransition, call.Status)
	}

	transcription, err := s.createProcessingTranscription(ctx, call.ID)
	if err != nil {
		return fmt.Errorf("create transcription record: %w", err)
	}

	audioFile, err := s.openAudio(ctx, call)
	if err != nil {
		return fmt.Errorf("open audio: %w", err)
	}
	defer audioFile.Content.Close()

	result, err := s.transcriber.Transcribe(ctx, audioFile)
	if err != nil {
		return fmt.Errorf("transcribe audio: %w", err)
	}

	if _, err = s.transcriptionRepository.MarkTranscribed(ctx, transcription.ID, result.Text, result.Language); err != nil {
		return fmt.Errorf("mark transcription transcribed: %w", err)
	}

	if _, err = s.callRepository.UpdateCallStatus(ctx, call.ID, models.CallStatusTranscribed); err != nil {
		return fmt.Errorf("mark call transcribed: %w", err)
	}

	s.log.Info(ctx, "call transcribed", zap.String("call_id", call.ID.String()), zap.String("provider", s.transcriber.Provider()))

	return nil
}

func (s *Service) createProcessingTranscription(ctx context.Context, callID uuid.UUID) (models.Transcription, error) {
	transcriptionID, err := uuid.NewV7()
	if err != nil {
		return models.Transcription{}, err
	}

	now := time.Now().UTC()

	return s.transcriptionRepository.Create(ctx, models.Transcription{
		ID:        transcriptionID,
		CallUUID:  callID,
		Status:    models.TranscriptionStatusProcessing,
		Provider:  s.transcriber.Provider(),
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) openAudio(ctx context.Context, call models.Call) (models.File, error) {
	content, err := s.audioStorage.Open(ctx, call.AudioPath)
	if err != nil {
		return models.File{}, err
	}

	return models.File{
		Content:          content,
		Path:             call.AudioPath,
		OriginalFilename: call.OriginalFilename,
		MimeType:         call.MimeType,
		SizeBytes:        call.SizeBytes,
	}, nil
}

func (s *Service) MarkJobFailed(ctx context.Context, job models.ProcessingJob, cause error) {
	if job.Type != models.ProcessingJobTypeTranscribeCall {
		return
	}

	transcription, err := s.transcriptionRepository.GetByCallUUID(ctx, job.EntityUUID)
	if err == nil && transcription.ID != uuid.Nil {
		_, _ = s.transcriptionRepository.MarkFailed(context.Background(), transcription.ID, cause.Error())
	}

	_, _ = s.callRepository.UpdateCallStatus(context.Background(), job.EntityUUID, models.CallStatusFailed)

	s.log.Error(ctx, "processing job permanently failed", zap.String("call_id", job.EntityUUID.String()), zap.String("job_id", job.ID.String()), zap.Error(cause))
}
