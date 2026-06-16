package report

import (
	"calllens/monolit/internal/models"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
	DELETE FROM call_report_exports
	WHERE report_uuid = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete report export: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete report export: %w", err)
	}
	if rowsAffected == 0 {
		return models.ErrReportNotFound
	}

	return nil
}

func reportNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return models.ErrReportNotFound
	}

	return err
}
