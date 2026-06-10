package call

import (
	"calllens/monolit/internal/logger"
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
	"context"
)

const defaultProcessingJobMaxAttempts = 3

type DurationDetector interface {
	DetectDuration(ctx context.Context, path string) (int, error)
}

type Service struct {
	repository               repo.CallRepository
	transcriptionRepository  repo.TranscriptionRepository
	processingJobRepository  repo.ProcessingJobRepository
	companyRepository        repo.CompanyRepository
	departmentRepository     repo.DepartmentRepository
	audioStorage             storage.Storage
	durationDetector         DurationDetector
	processingJobMaxAttempts int
	log                      logger.Logger
}

func NewService(
	repository repo.CallRepository,
	companyRepository repo.CompanyRepository,
	departmentRepository repo.DepartmentRepository,
	audioStorage storage.Storage,
	log logger.Logger,
) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		repository:               repository,
		companyRepository:        companyRepository,
		departmentRepository:     departmentRepository,
		audioStorage:             audioStorage,
		processingJobMaxAttempts: defaultProcessingJobMaxAttempts,
		log:                      log,
	}
}

func (s *Service) SetTranscriptionRepository(repository repo.TranscriptionRepository) {
	s.transcriptionRepository = repository
}

func (s *Service) SetProcessingJobRepository(repository repo.ProcessingJobRepository) {
	s.processingJobRepository = repository
}

func (s *Service) SetProcessingJobMaxAttempts(maxAttempts int) {
	if maxAttempts <= 0 {
		maxAttempts = defaultProcessingJobMaxAttempts
	}

	s.processingJobMaxAttempts = maxAttempts
}

func (s *Service) SetDurationDetector(detector DurationDetector) {
	s.durationDetector = detector
}
