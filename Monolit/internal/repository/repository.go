package repository

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	//POST
	CreateCall(ctx context.Context, call models.Call) (models.Call, error)
	//GET
	List(ctx context.Context) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (models.Call, error)
	//UPDATE

	//DELETE
	DeleteCall(ctx context.Context, id uuid.UUID) error

}
