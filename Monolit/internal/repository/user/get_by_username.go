package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	"calllens/monolit/internal/repository/scaner"
)

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	query := `SELECT user_uuid,
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
	       created_at
	FROM users
	WHERE lower(username) = lower($1)`

	row := r.db.QueryRowContext(ctx, query, username)

	repoUser, err := scaner.ScanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("get user by username: %w", err)
	}
	return converter.RepoUserToModel(repoUser)
}
