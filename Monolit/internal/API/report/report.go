package report

import (
	"net/http"

	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/service"

	"github.com/google/uuid"
)

type Handler struct {
	service service.ReportService
}

func NewHandler(service service.ReportService) *Handler {
	return &Handler{service: service}
}

func userIDFromRequest(r *http.Request) (uuid.UUID, bool) {
	return middleware.UserIDFromContext(r.Context())
}
