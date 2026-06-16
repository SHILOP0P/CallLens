package user

import (
	"calllens/monolit/internal/models"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func normalizeUserError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return models.ErrUserAlreadyExists
	}

	return err
}
