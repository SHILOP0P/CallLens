package httpserver

import (
	"calllens/monolit/internal/API"
	"calllens/monolit/internal/API/health"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(callAPI API.API) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(middleware.RequestID)
	r.Use(middleware.URLFormat)

	r.Get("/health", health.Health)
	r.Route("/api/v1", func(r chi.Router) {
		//POST
		r.Post("/calls", callAPI.Create)
		//GET
		r.Get("/calls", callAPI.List)
		r.Get("/calls/{uuid}", callAPI.GetByUUID)
		r.Get("/calls/{uuid}/audio", callAPI.GetAudioByUUID)
		//UPDATE
		r.Patch("/calls/{uuid}", callAPI.UpdateCallTitle)
		//DELETE
		r.Delete("/calls/{uuid}", callAPI.DeleteCall)
	})

	return r
}
