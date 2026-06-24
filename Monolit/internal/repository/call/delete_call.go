package call

import (
	"context"
	"fmt"

	"calllens/monolit/internal/models"

	"github.com/google/uuid"
)

func (r *Repository) DeleteCall(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	queryDel := fmt.Sprintf(`
	DELETE FROM calls c
	WHERE c.call_uuid = $1
	  AND %s
	`, visibleToUserCondition("c", "$2"))

	result, err := r.db.ExecContext(ctx, queryDel, id, userID)
	if err != nil {
		return fmt.Errorf("delete call: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete call rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrCallNotFound
	}

	return nil
}
