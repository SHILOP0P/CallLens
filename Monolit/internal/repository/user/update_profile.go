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

func (r *Repository) UpdateProfile(ctx context.Context, input model.UpdateUserProfileInput) (model.User, error) {
	query := `UPDATE users
	SET full_name = COALESCE($2, full_name),
	    full_surname = COALESCE($3, full_surname),
	    post = COALESCE($4, post),
	    phone = COALESCE($5, phone),
	    timezone = COALESCE($6, timezone)
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

	row := r.db.QueryRowContext(ctx, query, input.UserUUID, input.FullName, input.FullSurname, input.Post, input.Phone, input.Timezone)

	repoUser, err := scaner.ScanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.User{}, model.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("update profile: %w", err)
	}

	return converter.RepoUserToModel(repoUser)
}
