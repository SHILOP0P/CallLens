package user

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (r *Repository) UpdateUsername(ctx context.Context, input model.UpdateUsernameInput) (model.User, error) {
	query := `UPDATE users
	SET username = $2
	WHERE user_uuid = $1
	RETURNING user_uuid,
	          email,
	          password_hash,
	          full_name,
	          full_surname,
	          username,
	          role,
	          post,
	          created_at`

	row := r.db.QueryRowContext(ctx, query, input.UserUUID, input.Username)

	repoUser, err := scaner.ScanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("update username: %w", normalizeUserError(err))
	}

	return converter.RepoUserToModel(repoUser)
}
