package report

import (
	"calllens/monolit/internal/models"
	"database/sql"
)

type rowScanner interface {
	Scan(dest ...any) error
}

func scanReport(row rowScanner) (models.ReportExport, error) {
	var report models.ReportExport
	var format string
	var status string
	var storagePath sql.NullString
	var errorMessage sql.NullString

	if err := row.Scan(
		&report.ID,
		&report.CallUUID,
		&report.AnalysisUUID,
		&report.RequestedByUserUUID,
		&format,
		&status,
		&storagePath,
		&report.FileName,
		&report.ContentType,
		&report.SizeBytes,
		&errorMessage,
		&report.CreatedAt,
		&report.UpdatedAt,
		&report.ExpiresAt,
	); err != nil {
		return models.ReportExport{}, err
	}

	report.Format = models.ReportFormat(format)
	report.Status = models.ReportStatus(status)
	if storagePath.Valid {
		report.StoragePath = &storagePath.String
	}
	if errorMessage.Valid {
		report.ErrorMessage = &errorMessage.String
	}

	return report, nil
}
