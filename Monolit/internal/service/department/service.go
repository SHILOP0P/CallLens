package department

import (
	"calllens/monolit/internal/logger"
	repo "calllens/monolit/internal/repository"
)

type Service struct {
	companyRepository    repo.CompanyRepository
	departmentRepository repo.DepartmentRepository
	log                  logger.Logger
}

func NewService(companyRepository repo.CompanyRepository, departmentRepository repo.DepartmentRepository, log logger.Logger) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		companyRepository:    companyRepository,
		departmentRepository: departmentRepository,
		log:                  log,
	}
}
