package call

import repo "calllens/monolit/internal/repository"

type Service struct {
	repository repo.Repository
}

func NewService(repository repo.Repository) *Service {
	return &Service{
		repository: repository,
	}
}
