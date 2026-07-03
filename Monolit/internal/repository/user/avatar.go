package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	"calllens/monolit/internal/repository/scaner"

	"github.com/google/uuid"
)

func (r *Repository) UpdateAvatar(ctx context.Context, input model.UserAvatarUpdate) (model.User, error) {
	query := `UPDATE users
	SET avatar_path = $2,
	    avatar_mime_type = $3,
	    avatar_size_bytes = $4,
	    avatar_updated_at = $5
	WHERE user_uuid = $1
	RETURNING user_uuid,
	          email,
	          password_hash,
	          full_name,
	          full_surname,
	          username,
	          role,
	          post,
	          phone,
	          timezone,
	          avatar_path,
	          avatar_mime_type,
	          avatar_size_bytes,
	          avatar_updated_at,
	          created_at`

	row := r.db.QueryRowContext(ctx, query, input.UserUUID, input.Path, input.MimeType, input.SizeBytes, input.UpdatedAt)
	return scanAvatarUser(row, "update avatar")
}

func (r *Repository) DeleteAvatar(ctx context.Context, userID uuid.UUID) (model.User, error) {
	query := `UPDATE users
	SET avatar_path = NULL,
	    avatar_mime_type = NULL,
	    avatar_size_bytes = NULL,
	    avatar_updated_at = NULL
	WHERE user_uuid = $1
	RETURNING user_uuid,
	          email,
	          password_hash,
	          full_name,
	          full_surname,
	          username,
	          role,
	          post,
	          phone,
	          timezone,
	          avatar_path,
	          avatar_mime_type,
	          avatar_size_bytes,
	          avatar_updated_at,
	          created_at`

	row := r.db.QueryRowContext(ctx, query, userID)
	return scanAvatarUser(row, "delete avatar")
}

func scanAvatarUser(row interface{ Scan(dest ...any) error }, operation string) (model.User, error) {
	repoUser, err := scaner.ScanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("%s: %w", operation, err)
	}

	return converter.RepoUserToModel(repoUser)
}
