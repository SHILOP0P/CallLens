package processing

import (
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
	"calllens/monolit/internal/transcriber"
)

type Service struct {
	callRepository          repository.CallRepository
	transcriptionRepository repository.TranscriptionRepository
	processingJobRepository repository.ProcessingJobRepository
	audioStorage            storage.Storage
	transcriber             transcriber.Transcriber
	log                     logger.Logger
}

func NewService(
	callRepository repository.CallRepository,
	transcriptionRepository repository.TranscriptionRepository,
	processingJobRepository repository.ProcessingJobRepository,
	audioStorage storage.Storage,
	transcriber transcriber.Transcriber,
	log logger.Logger,
) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		callRepository:          callRepository,
		transcriptionRepository: transcriptionRepository,
		processingJobRepository: processingJobRepository,
		audioStorage:            audioStorage,
		transcriber:             transcriber,
		log:                     log,
	}
}
