package dto

import "mime/multipart"

type CreateCallRequest struct {
	Title string
	Audio *multipart.FileHeader
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
	CreatedAt          string  `json:"created_at"`
}

type UpdateCallTitleRequest struct {
	Title string `json:"title"`
}
