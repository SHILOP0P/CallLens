package scaner

import repoModel "calllens/monolit/internal/repository/models"

func ScanUser(row rowScanner) (repoModel.User, error) {
	var user repoModel.User

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.FullSurname,
		&user.Username,
		&user.Role,
		&user.Post,
		&user.CreatedAt,
	)
	if err != nil {
		return repoModel.User{}, err
	}

	return user, nil
}
