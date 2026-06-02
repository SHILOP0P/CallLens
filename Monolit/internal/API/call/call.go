package call

import "calllens/monolit/internal/service"

type CallHandler struct {
	service service.Service
}

func NewCallHandler(service service.Service) *CallHandler {
	return &CallHandler{service: service}
}
