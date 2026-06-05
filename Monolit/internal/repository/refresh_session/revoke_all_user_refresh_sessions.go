package refresh_session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) RevokeAllUserRefreshSessions(ctx context.Context, userID uuid.UUID, reason string) error {
	query := `
	UPDATE refresh_sessions
	SET revoked_at = now(),
	    revoked_reason = $2
	WHERE user_uuid = $1
	  AND revoked_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, query, userID, reason); err != nil {
		return fmt.Errorf("revoke all user refresh sessions: %w", err)
	}

	return nil
}
