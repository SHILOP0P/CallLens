package refresh_session

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) RevokeOtherUserRefreshSessions(ctx context.Context, userID uuid.UUID, keepSessionID uuid.UUID, reason string) error {
	query := `
	UPDATE refresh_sessions
	SET revoked_at = now(),
	    revoked_reason = $3
	WHERE user_uuid = $1
	  AND session_uuid <> $2
	  AND revoked_at IS NULL
	`

	if _, err := r.db.ExecContext(ctx, query, userID, keepSessionID, reason); err != nil {
		return fmt.Errorf("revoke other user refresh sessions: %w", err)
	}

	return nil
}
