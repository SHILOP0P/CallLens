package company

import (
	"calllens/monolit/internal/logger"
	repo "calllens/monolit/internal/repository"
)

type Service struct {
	companyRepository repo.CompanyRepository
	log               logger.Logger
}

func NewService(companyRepository repo.CompanyRepository, log logger.Logger) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		companyRepository: companyRepository,
		log:               log,
	}
}
