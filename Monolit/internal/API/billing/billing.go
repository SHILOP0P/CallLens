package billing

import (
	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/service"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	service service.BillingService
}

func NewHandler(service service.BillingService) *Handler {
	return &Handler{service: service}
}

func userIDFromRequest(r *http.Request) (uuid.UUID, bool) {
	return middleware.UserIDFromContext(r.Context())
}
