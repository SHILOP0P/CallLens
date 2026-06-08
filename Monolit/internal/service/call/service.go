package call

import (
	"calllens/monolit/internal/logger"
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
)

type Service struct {
	repository              repo.CallRepository
	transcriptionRepository repo.TranscriptionRepository
	companyRepository       repo.CompanyRepository
	departmentRepository    repo.DepartmentRepository
	audioStorage            storage.Storage
	log                     logger.Logger
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
		repository:           repository,
		companyRepository:    companyRepository,
		departmentRepository: departmentRepository,
		audioStorage:         audioStorage,
		log:                  log,
	}
}

func (s *Service) SetTranscriptionRepository(repository repo.TranscriptionRepository) {
	s.transcriptionRepository = repository
}
