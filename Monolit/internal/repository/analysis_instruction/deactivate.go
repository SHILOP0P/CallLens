package analysis_instruction

import (
	model "calllens/monolit/internal/models"
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) Deactivate(ctx context.Context, id uuid.UUID) error {
	query := `
	UPDATE analysis_instructions
	SET is_active = false,
	    updated_at = now()
	WHERE instruction_uuid = $1
	  AND is_active = true
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deactivate analysis instruction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("deactivate analysis instruction rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%w: %w", model.ErrAnalysisInstructionNotFound, sql.ErrNoRows)
	}

	return nil
}
