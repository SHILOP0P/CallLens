package analysis_instruction

import (
	"calllens/monolit/internal/logger"
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
)

type Service struct {
	repository           repo.AnalysisInstructionRepository
	companyRepository    repo.CompanyRepository
	departmentRepository repo.DepartmentRepository
	instructionStorage   storage.InstructionStorage
	log                  logger.Logger
}

func NewService(
	repository repo.AnalysisInstructionRepository,
	companyRepository repo.CompanyRepository,
	departmentRepository repo.DepartmentRepository,
	instructionStorage storage.InstructionStorage,
	log logger.Logger,
) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		repository:           repository,
		companyRepository:    companyRepository,
		departmentRepository: departmentRepository,
		instructionStorage:   instructionStorage,
		log:                  log,
	}
}
