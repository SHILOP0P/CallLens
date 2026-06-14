package analysis

import (
	"calllens/monolit/internal/analyzer"
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/models"
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
)

type Service struct {
	callRepository           repo.CallRepository
	transcriptionRepository  repo.TranscriptionRepository
	instructionRepository    repo.AnalysisInstructionRepository
	analysisRepository       repo.AnalysisRepository
	processingJobRepository  repo.ProcessingJobRepository
	instructionStorage       storage.InstructionStorage
	analyzer                 analyzer.Analyzer
	processingJobMaxAttempts int
	log                      logger.Logger
}

func NewService(
	callRepository repo.CallRepository,
	transcriptionRepository repo.TranscriptionRepository,
	instructionRepository repo.AnalysisInstructionRepository,
	analysisRepository repo.AnalysisRepository,
	instructionStorage storage.InstructionStorage,
	analyzer analyzer.Analyzer,
	log logger.Logger,
) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		callRepository:           callRepository,
		transcriptionRepository:  transcriptionRepository,
		instructionRepository:    instructionRepository,
		analysisRepository:       analysisRepository,
		instructionStorage:       instructionStorage,
		analyzer:                 analyzer,
		processingJobMaxAttempts: models.DefaultProcessingJobMaxAttempts,
		log:                      log,
	}
}

func (s *Service) SetProcessingJobRepository(repository repo.ProcessingJobRepository) {
	s.processingJobRepository = repository
}

func (s *Service) SetProcessingJobMaxAttempts(maxAttempts int) {
	if maxAttempts <= 0 {
		maxAttempts = models.DefaultProcessingJobMaxAttempts
	}

	s.processingJobMaxAttempts = maxAttempts
}
