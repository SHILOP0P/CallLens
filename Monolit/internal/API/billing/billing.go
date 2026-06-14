package billing

import "calllens/monolit/internal/service"

type Handler struct {
	service service.BillingService
}

func NewHandler(service service.BillingService) *Handler {
	return &Handler{service: service}
}
