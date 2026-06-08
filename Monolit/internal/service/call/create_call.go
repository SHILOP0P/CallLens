package call

import (
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Service) CreateCall(ctx context.Context, input models.CreateCallInput) (models.Call, error) {
	if err := validateAudioInput(input); err != nil {
		s.log.Warn(ctx, "create call failed", zap.String("reason", "invalid_audio_input"), zap.String("user_id", input.UploadedByUserUUID.String()), zap.Error(err))
		return models.Call{}, err
	}

	if err := s.authorizeUpload(ctx, input); err != nil {
		s.log.Warn(ctx, "create call failed", zap.String("reason", "upload_forbidden"), zap.String("user_id", input.UploadedByUserUUID.String()), zap.String("visibility_scope", string(input.VisibilityScope)), zap.Error(err))
		return models.Call{}, err
	}

	callUUID, err := uuid.NewV7()
	if err != nil {
		s.log.Error(ctx, "failed to generate call uuid", zap.String("user_id", input.UploadedByUserUUID.String()), zap.Error(err))
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
		s.log.Error(ctx, "failed to save audio file", zap.String("user_id", input.UploadedByUserUUID.String()), zap.String("call_id", callUUID.String()), zap.Error(err))
		return models.Call{}, err
	}

	now := time.Now().UTC()
	call, err := converter.SavedFileToModel(savedFile, callUUID, input, now)
	if err != nil {
		_ = s.audioStorage.Delete(context.Background(), savedFile.Path)
		s.log.Error(ctx, "failed to build call model", zap.String("user_id", input.UploadedByUserUUID.String()), zap.String("call_id", callUUID.String()), zap.Error(err))
		return models.Call{}, err
	}

	createdCall, err := s.createCallRecord(ctx, call, now)
	if err != nil {
		_ = s.audioStorage.Delete(context.Background(), savedFile.Path)
		s.log.Error(ctx, "failed to create call record", zap.String("user_id", input.UploadedByUserUUID.String()), zap.String("call_id", callUUID.String()), zap.Error(err))
		return models.Call{}, err
	}

	s.log.Info(
		ctx,
		"call created",
		zap.String("user_id", input.UploadedByUserUUID.String()),
		zap.String("call_id", createdCall.ID.String()),
		zap.String("mime_type", createdCall.MimeType),
		zap.Int64("size_bytes", createdCall.SizeBytes),
	)

	return createdCall, nil
}

func (s *Service) createCallRecord(ctx context.Context, call models.Call, now time.Time) (models.Call, error) {
	if s.processingJobRepository == nil {
		return s.repository.CreateCall(ctx, call)
	}

	jobID, err := uuid.NewV7()
	if err != nil {
		return models.Call{}, err
	}

	job := models.ProcessingJob{
		ID:          jobID,
		Type:        models.ProcessingJobTypeTranscribeCall,
		EntityUUID:  call.ID,
		Status:      models.ProcessingJobStatusPending,
		Attempts:    0,
		MaxAttempts: 3,
		AvailableAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return s.repository.CreateCallWithProcessingJob(ctx, call, job)
}
