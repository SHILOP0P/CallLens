package dto

import "mime/multipart"

type CreateCallRequest struct {
	Title          string
	Audio          *multipart.FileHeader
	CompanyUUID    string
	DepartmentUUID string
}

type CallResponse struct {
	ID                 string  `json:"id"`
	Title              string  `json:"title"`
	Status             string  `json:"status"`
	OriginalFilename   string  `json:"original_filename"`
	MimeType           string  `json:"mime_type"`
	SizeBytes          int64   `json:"size_bytes"`
	DurationSeconds    int     `json:"duration_seconds"`
	UploadedByUserUUID *string `json:"uploaded_by_user_uuid"`
	CompanyUUID        *string `json:"company_uuid"`
	DepartmentUUID     *string `json:"department_uuid"`
	VisibilityScope    string  `json:"visibility_scope"`
	CreatedAt          string  `json:"created_at"`
}

type CallStatusEvent struct {
	CallID    string `json:"call_id"`
	Status    string `json:"status"`
	Terminal  bool   `json:"terminal"`
	Timestamp string `json:"timestamp"`
}

type UpdateCallTitleRequest struct {
	Title string `json:"title"`
}
