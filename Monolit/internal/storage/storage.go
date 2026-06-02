package storage

import (
	"calllens/monolit/internal/models"
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, input models.SaveInput) (models.SavedFile, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}
