package analysis

import (
	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/service"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service service.AnalysisService
}

func NewHandler(service service.AnalysisService) *Handler {
	return &Handler{service: service}
}

func userIDFromRequest(r *http.Request) (uuid.UUID, bool) {
	return middleware.UserIDFromContext(r.Context())
}
