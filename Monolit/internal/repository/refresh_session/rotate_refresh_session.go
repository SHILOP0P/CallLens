package refresh_session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	"calllens/monolit/internal/repository/scaner"
)

func (r *Repository) RotateRefreshSession(ctx context.Context, oldRefreshTokenHash string, newRefreshTokenHash string, expiresAt time.Time) (model.RefreshSession, error) {
	query := `
	UPDATE refresh_sessions
	SET refresh_token_hash = $2,
	    last_used_at = now(),
	    expires_at = $3
	WHERE refresh_token_hash = $1
	  AND revoked_at IS NULL
	  AND expires_at > now()
	RETURNING session_uuid,
	          user_uuid,
	          refresh_token_hash,
	          access_version,
	          user_agent,
	          ip_address::TEXT,
	          created_at,
	          last_used_at,
	          expires_at,
	          revoked_at,
	          revoked_reason
	`

	row := r.db.QueryRowContext(ctx, query, oldRefreshTokenHash, newRefreshTokenHash, expiresAt)

	repoSession, err := scaner.ScanRefreshSession(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RefreshSession{}, model.ErrRefreshSessionNotFound
		}

		return model.RefreshSession{}, fmt.Errorf("rotate refresh session: %w", err)
	}

	return converter.RepoRefreshSessionToModel(repoSession)
}
