package call

import (
	"net/http"

	"calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/service"

	"github.com/google/uuid"
)

type CallHandler struct {
	service service.CallService
}

func NewCallHandler(service service.CallService) *CallHandler {
	return &CallHandler{service: service}
}

func userIDFromRequest(r *http.Request) (uuid.UUID, bool) {
	return middleware.UserIDFromContext(r.Context())
}
