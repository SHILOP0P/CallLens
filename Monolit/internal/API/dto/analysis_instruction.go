package dto

type AnalysisInstruction struct {
	ID                string  `json:"id"`
	Scope             string  `json:"scope"`
	UserUUID          *string `json:"user_uuid"`
	CompanyUUID       *string `json:"company_uuid"`
	DepartmentUUID    *string `json:"department_uuid"`
	Title             string  `json:"title"`
	OriginalFilename  string  `json:"original_filename"`
	FilePath          string  `json:"file_path"`
	MimeType          string  `json:"mime_type"`
	SizeBytes         int64   `json:"size_bytes"`
	ContentSHA256     string  `json:"content_sha256"`
	SortOrder         int     `json:"sort_order"`
	IsActive          bool    `json:"is_active"`
	CreatedByUserUUID string  `json:"created_by_user_uuid"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}
