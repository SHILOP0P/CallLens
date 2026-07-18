package storage

import (
	"context"
	"io"

	"calllens/monolit/internal/models"
)

type AudioStorage interface {
	Save(ctx context.Context, input models.SaveInput) (models.SavedFile, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	OpenReadSeeker(ctx context.Context, path string) (models.ReadSeekCloser, error)
	Delete(ctx context.Context, path string) error
}

type InstructionStorage interface {
	Save(ctx context.Context, input models.SaveInstructionInput) (models.SavedInstructionFile, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

type ReportStorage interface {
	Save(ctx context.Context, input models.SaveReportInput) (models.SavedReportFile, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}

type AvatarStorage interface {
	Save(ctx context.Context, input models.SaveUserAvatarInput) (models.SavedUserAvatar, error)
	Open(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
}
