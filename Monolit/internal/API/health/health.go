package health

import (
	"calllens/monolit/internal/API/response"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status: "ok",
	}

	if err := response.WriteJSON(w, http.StatusOK, resp); err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToEncodeResponse, "failed to encode response")
	}
}
