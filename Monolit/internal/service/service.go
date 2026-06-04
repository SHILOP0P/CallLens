package service

import (
	"calllens/monolit/internal/models"
	"context"

	"github.com/google/uuid"
)

type CallService interface {
	//POST
	CreateCall(ctx context.Context, input models.CreateCallInput) (models.Call, error)

	//GET
	List(ctx context.Context) ([]models.Call, error)
	GetByUUID(ctx context.Context, id uuid.UUID) (models.Call, error)
	GetAudioByUUID(ctx context.Context, uuid uuid.UUID) (models.File, error)

	//UPDATE
	UpdateCallTitle(ctx context.Context, id uuid.UUID, title string) (models.Call, error)
	//DELETE
	DeleteCall(ctx context.Context, id uuid.UUID) error
}

type AuthService interface {
	Register(ctx context.Context, input models.CreateUserInput) (models.User, error)
	//Login(ctx context.Context, input models.LoginInput) (models.User, string, error)
	//Me(ctx context.Context, userID uuid.UUID) (models.User, error)
}
