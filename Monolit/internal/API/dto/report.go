package dto

type CreateReportRequest struct {
	Format string `json:"format"`
}

type ReportResponse struct {
	ID                  string  `json:"id"`
	CallUUID            string  `json:"call_uuid"`
	AnalysisUUID        string  `json:"analysis_uuid"`
	RequestedByUserUUID string  `json:"requested_by_user_uuid"`
	Format              string  `json:"format"`
	Status              string  `json:"status"`
	FileName            string  `json:"file_name"`
	ContentType         string  `json:"content_type"`
	SizeBytes           int64   `json:"size_bytes"`
	ErrorMessage        *string `json:"error_message"`
	DownloadURL         *string `json:"download_url"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	ExpiresAt           string  `json:"expires_at"`
}

type ReportsResponse struct {
	Reports []ReportResponse `json:"reports"`
}
