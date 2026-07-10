package refresh_session

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

func (r *Repository) GetRefreshSessionByUUID(ctx context.Context, sessionID uuid.UUID) (model.RefreshSession, error) {
	query := `
	SELECT session_uuid,
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
	FROM refresh_sessions
	WHERE session_uuid = $1
	`

	row := r.db.QueryRowContext(ctx, query, sessionID)

	repoSession, err := scaner.ScanRefreshSession(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RefreshSession{}, model.ErrRefreshSessionNotFound
		}

		return model.RefreshSession{}, fmt.Errorf("get refresh session by uuid: %w", err)
	}

	return converter.RepoRefreshSessionToModel(repoSession)
}
