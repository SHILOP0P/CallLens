package converter

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/models"
	"time"
)

func ReportModelToAPI(report models.ReportExport) (dto.ReportResponse, error) {
	var downloadURL *string
	if report.Status == models.ReportStatusReady {
		value := "/api/v1/reports/" + report.ID.String() + "/download"
		downloadURL = &value
	}

	return dto.ReportResponse{
		ID:                  report.ID.String(),
		CallUUID:            report.CallUUID.String(),
		AnalysisUUID:        report.AnalysisUUID.String(),
		RequestedByUserUUID: report.RequestedByUserUUID.String(),
		Format:              string(report.Format),
		Status:              string(report.Status),
		FileName:            report.FileName,
		ContentType:         report.ContentType,
		SizeBytes:           report.SizeBytes,
		ErrorMessage:        report.ErrorMessage,
		DownloadURL:         downloadURL,
		CreatedAt:           report.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           report.UpdatedAt.Format(time.RFC3339),
		ExpiresAt:           report.ExpiresAt.Format(time.RFC3339),
	}, nil
}

func ReportsModelToAPI(reports []models.ReportExport) (dto.ReportsResponse, error) {
	resp := dto.ReportsResponse{
		Reports: make([]dto.ReportResponse, 0, len(reports)),
	}

	for _, report := range reports {
		item, err := ReportModelToAPI(report)
		if err != nil {
			return dto.ReportsResponse{}, err
		}
		resp.Reports = append(resp.Reports, item)
	}

	return resp, nil
}
