package call

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/converter"
	"net/http"
)

func (h *CallHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	calls, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToListCalls, "failed to list calls")
		return
	}

	resp := make([]dto.CallResponse, len(calls))

	for i, call := range calls {
		callResponse, err := converter.CallModelToAPI(call)
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToConvertCall, "failed to convert call")
			return
		}
		resp[i] = callResponse
	}

	if err := response.WriteJSON(w, http.StatusOK, resp); err != nil {
		return
	}
}
