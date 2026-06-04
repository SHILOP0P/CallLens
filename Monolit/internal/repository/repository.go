package repository

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

type CallRepository interface {
	//POST
	CreateCall(ctx context.Context, call models.Call) (models.Call, error)
	//GET
	List(ctx context.Context) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (models.Call, error)
	//UPDATE
	UpdateCallTitle(ctx context.Context, id uuid.UUID, title string) (models.Call, error)
	//DELETE
	DeleteCall(ctx context.Context, id uuid.UUID) error
}

type UserRepository interface {
	//GET
	GetUserByUUID(ctx context.Context, id uuid.UUID) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	//POST
	CreateUser(ctx context.Context, user models.User) (models.User, error)
}
