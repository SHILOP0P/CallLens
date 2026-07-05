package dto

import "encoding/json"

type CreateDeepAnalysisRequest struct {
	Scope          string  `json:"scope"`
	CompanyUUID    *string `json:"company_uuid"`
	DepartmentUUID *string `json:"department_uuid"`
	FolderUUID     *string `json:"folder_uuid"`
	PeriodFrom     string  `json:"period_from"`
	PeriodTo       string  `json:"period_to"`
	Force          bool    `json:"force"`
}

type AggregateAnalysisResponse struct {
	ID                string          `json:"id"`
	Scope             string          `json:"scope"`
	UserUUID          *string         `json:"user_uuid,omitempty"`
	CompanyUUID       *string         `json:"company_uuid"`
	DepartmentUUID    *string         `json:"department_uuid"`
	FolderUUID        *string         `json:"folder_uuid"`
	PeriodFrom        string          `json:"period_from"`
	PeriodTo          string          `json:"period_to"`
	Status            string          `json:"status"`
	Provider          string          `json:"provider"`
	Model             *string         `json:"model"`
	SourceCallsCount  int             `json:"source_calls_count"`
	ResultJSON        json.RawMessage `json:"result_json"`
	ResultText        *string         `json:"result_text"`
	ErrorMessage      *string         `json:"error_message"`
	CreatedByUserUUID string          `json:"created_by_user_uuid"`
	CreatedAt         string          `json:"created_at"`
	UpdatedAt         string          `json:"updated_at"`
}

type ListAggregateAnalysesResponse struct {
	Items  []AggregateAnalysisResponse `json:"items"`
	Total  int                         `json:"total"`
	Limit  int                         `json:"limit"`
	Offset int                         `json:"offset"`
}

type CreateAggregateReportRequest struct {
	Format string `json:"format"`
}

type AggregateReportResponse struct {
	ID                    string  `json:"id"`
	AggregateAnalysisUUID string  `json:"aggregate_analysis_uuid"`
	RequestedByUserUUID   string  `json:"requested_by_user_uuid"`
	Format                string  `json:"format"`
	Status                string  `json:"status"`
	FileName              string  `json:"file_name"`
	ContentType           string  `json:"content_type"`
	SizeBytes             int64   `json:"size_bytes"`
	ErrorMessage          *string `json:"error_message"`
	DownloadURL           *string `json:"download_url"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
	ExpiresAt             string  `json:"expires_at"`
}

type ListAggregateReportsResponse struct {
	Reports []AggregateReportResponse `json:"reports"`
}
