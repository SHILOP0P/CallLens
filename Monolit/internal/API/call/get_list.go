package call

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/converter"
	"encoding/json"
	"net/http"
)

func (h *CallHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	calls, err := h.service.List(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]dto.CallResponse, len(calls))

	for i, call := range calls {
		callResponse, err := converter.CallModelToAPI(call)
		if err != nil {
			http.Error(w, "failed to convert call", http.StatusInternalServerError)
			return
		}
		response[i] = callResponse
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
