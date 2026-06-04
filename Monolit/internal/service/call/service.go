package call

import (
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
)

type Service struct {
	repository   repo.CallRepository
	audioStorage storage.Storage
}

func NewService(repository repo.CallRepository, audioStorage storage.Storage) *Service {
	return &Service{
		repository:   repository,
		audioStorage: audioStorage,
	}
}
