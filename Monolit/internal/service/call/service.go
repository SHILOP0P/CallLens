package call

import (
	"calllens/monolit/internal/logger"
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
	"context"

	"github.com/google/uuid"
)

const defaultProcessingJobMaxAttempts = 3

type DurationDetector interface {
	DetectDuration(ctx context.Context, path string) (int, error)
}

type BillingLimiter interface {
	CanUploadPersonalCall(ctx context.Context, userID uuid.UUID, durationSeconds int) error
	CanUploadBusinessCall(ctx context.Context, companyID uuid.UUID, durationSeconds int) error
	AddPersonalUsageMinutes(ctx context.Context, userID uuid.UUID, durationSeconds int) error
	AddBusinessUsageMinutes(ctx context.Context, companyID uuid.UUID, durationSeconds int) error
}

type Service struct {
	repository               repo.CallRepository
	transcriptionRepository  repo.TranscriptionRepository
	processingJobRepository  repo.ProcessingJobRepository
	companyRepository        repo.CompanyRepository
	departmentRepository     repo.DepartmentRepository
	audioStorage             storage.AudioStorage
	durationDetector         DurationDetector
	billingLimiter           BillingLimiter
	processingJobMaxAttempts int
	log                      logger.Logger
}

func NewService(
	repository repo.CallRepository,
	companyRepository repo.CompanyRepository,
	departmentRepository repo.DepartmentRepository,
	audioStorage storage.AudioStorage,
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

func (s *Service) SetBillingLimiter(limiter BillingLimiter) {
	s.billingLimiter = limiter
}
