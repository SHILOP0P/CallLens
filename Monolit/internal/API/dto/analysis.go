package dto

import "encoding/json"

type AnalysisResponse struct {
	ID           string          `json:"id"`
	CallUUID     string          `json:"call_uuid"`
	Status       string          `json:"status"`
	Provider     string          `json:"provider"`
	Model        *string         `json:"model"`
	ResultJSON   json.RawMessage `json:"result_json"`
	ResultText   *string         `json:"result_text"`
	ErrorMessage *string         `json:"error_message"`
	CreatedAt    string          `json:"created_at"`
	UpdatedAt    string          `json:"updated_at"`
}
