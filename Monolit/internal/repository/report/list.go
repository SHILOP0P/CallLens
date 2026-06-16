package report

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (r *Repository) ListByCallUUID(ctx context.Context, callID uuid.UUID, now time.Time) ([]models.ReportExport, error) {
	query := `
	SELECT ` + reportColumns + `
	FROM call_report_exports
	WHERE call_uuid = $1
	  AND expires_at > $2
	ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, callID, now)
	if err != nil {
		return nil, fmt.Errorf("list report exports by call uuid: %w", err)
	}
	defer rows.Close()

	reports := make([]models.ReportExport, 0)
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("list report exports by call uuid: %w", err)
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list report exports by call uuid: %w", err)
	}

	return reports, nil
}

func (r *Repository) ListExpiredReady(ctx context.Context, now time.Time, limit int) ([]models.ReportExport, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
	SELECT ` + reportColumns + `
	FROM call_report_exports
	WHERE status = 'ready'
	  AND expires_at <= $1
	ORDER BY expires_at ASC
	LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list expired report exports: %w", err)
	}
	defer rows.Close()

	reports := make([]models.ReportExport, 0)
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("list expired report exports: %w", err)
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list expired report exports: %w", err)
	}

	return reports, nil
}
