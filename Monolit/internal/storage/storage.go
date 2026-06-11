package storage

import (
	"calllens/monolit/internal/models"
	"context"
	"io"
)

type AudioStorage interface {
	Save(ctx context.Context, input models.SaveInput) (models.SavedFile, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

type InstructionStorage interface {
	Save(ctx context.Context, input models.SaveInstructionInput) (models.SavedInstructionFile, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}
