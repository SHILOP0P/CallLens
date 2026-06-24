package refresh_session

import (
	"context"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
)

func (r *Repository) CreateRefreshSession(ctx context.Context, session model.RefreshSession) (model.RefreshSession, error) {
	repoSession, err := converter.ModelRefreshSessionToRepoModel(session)
	if err != nil {
		return model.RefreshSession{}, fmt.Errorf("convert refresh session to repo model: %w", err)
	}

	query := `
	INSERT INTO refresh_sessions (
	    session_uuid,
	    user_uuid,
	    refresh_token_hash,
	    user_agent,
	    ip_address,
	    created_at,
	    last_used_at,
	    expires_at,
	    revoked_at,
	    revoked_reason
	)
	VALUES ($1, $2, $3, $4, $5::INET, $6, $7, $8, $9, $10)
	RETURNING session_uuid,
	          user_uuid,
	          refresh_token_hash,
	          user_agent,
	          ip_address::TEXT,
	          created_at,
	          last_used_at,
	          expires_at,
	          revoked_at,
	          revoked_reason
	`

	var createdRepoSession repoModel.RefreshSession

	row := r.db.QueryRowContext(ctx, query,
		repoSession.ID,
		repoSession.UserID,
		repoSession.RefreshTokenHash,
		repoSession.UserAgent,
		repoSession.IPAddress,
		repoSession.CreatedAt,
		repoSession.LastUsedAt,
		repoSession.ExpiresAt,
		repoSession.RevokedAt,
		repoSession.RevokedReason,
	)

	createdRepoSession, err = scaner.ScanRefreshSession(row)
	if err != nil {
		return model.RefreshSession{}, fmt.Errorf("create refresh session: %w", err)
	}

	return converter.RepoRefreshSessionToModel(createdRepoSession)
}
