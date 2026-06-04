package call

import "calllens/monolit/internal/service"

type CallHandler struct {
	service service.CallService
}

func NewCallHandler(service service.CallService) *CallHandler {
	return &CallHandler{service: service}
}
