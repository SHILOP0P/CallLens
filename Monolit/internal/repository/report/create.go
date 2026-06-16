package report

import (
	"calllens/monolit/internal/models"
	"context"
	"fmt"
)

func (r *Repository) Create(ctx context.Context, report models.ReportExport) (models.ReportExport, error) {
	query := `
	INSERT INTO call_report_exports (
		report_uuid,
		call_uuid,
		analysis_uuid,
		requested_by_user_uuid,
		format,
		status,
		storage_path,
		file_name,
		content_type,
		size_bytes,
		error_message,
		created_at,
		updated_at,
		expires_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
	)
	RETURNING ` + reportColumns

	row := r.db.QueryRowContext(
		ctx,
		query,
		report.ID,
		report.CallUUID,
		report.AnalysisUUID,
		report.RequestedByUserUUID,
		report.Format,
		report.Status,
		report.StoragePath,
		report.FileName,
		report.ContentType,
		report.SizeBytes,
		report.ErrorMessage,
		report.CreatedAt,
		report.UpdatedAt,
		report.ExpiresAt,
	)

	created, err := scanReport(row)
	if err != nil {
		return models.ReportExport{}, fmt.Errorf("create report export: %w", err)
	}

	return created, nil
}
