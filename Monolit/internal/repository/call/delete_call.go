package call

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) DeleteCall(ctx context.Context, id uuid.UUID) error {
	queryDel := `
	DELETE FROM calls
	where call_uuid = $1
	`

	result, err := r.db.ExecContext(ctx, queryDel, id)
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
