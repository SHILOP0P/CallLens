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

func (r *Repository) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, passwordHash string) (model.User, error) {
	query := `UPDATE users
	SET password_hash = $2
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

	row := r.db.QueryRowContext(ctx, query, userID, passwordHash)

	repoUser, err := scaner.ScanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("update password hash: %w", err)
	}

	return converter.RepoUserToModel(repoUser)
}
