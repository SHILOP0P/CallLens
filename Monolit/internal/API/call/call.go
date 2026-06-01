package call

import "calllens/monolit/internal/service"

type CallHandler struct {
	service   service.Service
	uploadDir string
}

func NewCallHandler(service service.Service, uploadDir string) *CallHandler {
	return &CallHandler{service: service, uploadDir: uploadDir}
}
