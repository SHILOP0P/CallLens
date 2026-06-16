package converter

import (
	model "calllens/monolit/internal/models"
	repoModel "calllens/monolit/internal/repository/models"
	"database/sql"
)

func RepoUserToModel(repoUser repoModel.User) (model.User, error) {
	return model.User{
		ID:           repoUser.ID,
		Email:        repoUser.Email,
		PasswordHash: repoUser.PasswordHash,
		FullName:     repoUser.FullName,
		FullSurname:  repoUser.FullSurname,
		Username:     repoUser.Username,
		Role:         model.UserRole(repoUser.Role),
		Post:         nullStringToStringPtr(repoUser.Post),
		CreatedAt:    repoUser.CreatedAt,
	}, nil
}

func ModelUserToRepoModel(user model.User) (repoModel.User, error) {
	return repoModel.User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		FullName:     user.FullName,
		FullSurname:  user.FullSurname,
		Username:     user.Username,
		Role:         string(user.Role),
		Post:         stringPtrToNullString(user.Post),
		CreatedAt:    user.CreatedAt,
	}, nil
}

func nullStringToStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}

func stringPtrToNullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}
