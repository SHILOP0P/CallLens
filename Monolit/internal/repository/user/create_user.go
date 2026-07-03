package user

import (
	"context"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
)

func (r *Repository) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	repoUser, err := converter.ModelUserToRepoModel(user)
	if err != nil {
		return model.User{}, fmt.Errorf("convert model to repo model: %w", err)
	}

	query := `
	INSERT INTO users (
					user_uuid,
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
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
	          created_at
	`
	var createdRepoUser repoModel.User

	row := r.db.QueryRowContext(ctx, query,
		repoUser.ID,
		repoUser.Email,
		repoUser.PasswordHash,
		repoUser.FullName,
		repoUser.FullSurname,
		repoUser.Username,
		repoUser.Role,
		repoUser.Post,
		repoUser.Phone,
		repoUser.Timezone,
		repoUser.AvatarPath,
		repoUser.AvatarMime,
		repoUser.AvatarSize,
		repoUser.AvatarUpdatedAt,
		repoUser.CreatedAt,
	)

	createdRepoUser, err = scaner.ScanUser(row)
	if err != nil {
		return model.User{}, fmt.Errorf("create user: %w", normalizeUserError(err))
	}

	return converter.RepoUserToModel(createdRepoUser)
}
