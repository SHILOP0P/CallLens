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

func (r *Repository) GetUserByUUID(ctx context.Context, id uuid.UUID) (model.User, error) {
	query := `
	SELECT user_uuid,
	       email,
	       password_hash,
	       full_name,
	       full_surname,
	       username,
	       role,
	       post,
	       created_at
	FROM users
	WHERE user_uuid = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	repoUser, err := scaner.ScanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("get user by uuid: %w", err)
	}

	return converter.RepoUserToModel(repoUser)
}
