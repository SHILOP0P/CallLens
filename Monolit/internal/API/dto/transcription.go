package dto

type TranscriptionResponse struct {
	ID           string  `json:"id"`
	CallUUID     string  `json:"call_uuid"`
	Status       string  `json:"status"`
	Text         *string `json:"text"`
	Language     *string `json:"language"`
	Provider     string  `json:"provider"`
	ErrorMessage *string `json:"error_message"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}
