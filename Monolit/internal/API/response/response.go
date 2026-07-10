package response

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func WriteErrorWithDetails(w http.ResponseWriter, status int, code string, message string, details any) {
	_ = WriteJSON(w, status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func WriteJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")

	if payload == nil {
		w.WriteHeader(status)
		return nil
	}

	var buffer bytes.Buffer
	if err := json.NewEncoder(&buffer).Encode(payload); err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err := w.Write(buffer.Bytes())
	return err
}

func WriteError(w http.ResponseWriter, status int, code string, message string) {
	_ = WriteJSON(w, status, ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
