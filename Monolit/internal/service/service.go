package service

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

type Service interface {
	CreateCall(ctx context.Context, call models.Call) (models.Call, error)
	List(ctx context.Context) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (models.Call, error)
}
