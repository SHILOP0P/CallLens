package call

import (
	"calllens/monolit/internal/logger"
	repo "calllens/monolit/internal/repository"
	"calllens/monolit/internal/storage"
)

type Service struct {
	repository   repo.CallRepository
	audioStorage storage.Storage
	log          logger.Logger
}

func NewService(repository repo.CallRepository, audioStorage storage.Storage, log logger.Logger) *Service {
	if log == nil {
		log = logger.NewNop()
	}

	return &Service{
		repository:   repository,
		audioStorage: audioStorage,
		log:          log,
	}
}
