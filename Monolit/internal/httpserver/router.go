package httpserver

import (
	"calllens/monolit/internal/API"
	"calllens/monolit/internal/API/health"
	authMiddleware "calllens/monolit/internal/httpserver/middleware"
	"calllens/monolit/internal/logger"
	"calllens/monolit/internal/repository"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(callAPI API.CallAPI, authAPI API.AuthAPI, companyAPI API.CompanyAPI, departmentAPI API.DepartmentAPI, jwtSecret string, refreshSessionRepository repository.RefreshSessionRepository, log logger.Logger) http.Handler {
	r := chi.NewRouter()

	authGuard := authMiddleware.Auth(jwtSecret, refreshSessionRepository)

	r.Use(middleware.RequestID)
	r.Use(authMiddleware.RequestLogger(log))
	r.Use(authMiddleware.Recoverer(log))
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(middleware.URLFormat)

	r.Get("/health", health.Health)
	r.Route("/api/v1", func(r chi.Router) {
		//CALL
		//POST
		r.With(authGuard).Post("/calls", callAPI.Create)
		//GET
		r.With(authGuard).Get("/calls", callAPI.List)
		r.With(authGuard).Get("/calls/{uuid}", callAPI.GetByUUID)
		r.With(authGuard).Get("/calls/{uuid}/audio", callAPI.GetAudioByUUID)
		//UPDATE
		r.With(authGuard).Patch("/calls/{uuid}", callAPI.UpdateCallTitle)
		//DELETE
		r.With(authGuard).Delete("/calls/{uuid}", callAPI.DeleteCall)

		//AUTH
		r.Post("/auth/register", authAPI.Register)
		r.Post("/auth/login", authAPI.Login)
		r.Post("/auth/refresh", authAPI.Refresh)
		r.With(authGuard).Get("/auth/me", authAPI.Me)
		r.With(authGuard).Post("/auth/logout", authAPI.Logout)
		r.With(authGuard).Post("/auth/logout-all", authAPI.LogoutAll)

		//COMPANY
		r.With(authGuard).Post("/companies", companyAPI.Create)
		r.With(authGuard).Get("/companies", companyAPI.List)
		r.With(authGuard).Get("/companies/{uuid}", companyAPI.GetByUUID)
		r.With(authGuard).Get("/companies/{uuid}/members", companyAPI.GetCompanyMembersOverview)
		r.With(authGuard).Post("/companies/{uuid}/members", companyAPI.AddCompanyMember)
		r.With(authGuard).Patch("/companies/{uuid}/members/{user_uuid}/role", companyAPI.UpdateCompanyMemberRole)
		r.With(authGuard).Patch("/companies/{uuid}/members/{user_uuid}/status", companyAPI.UpdateCompanyMemberStatus)
		r.With(authGuard).Post("/companies/{uuid}/departments", departmentAPI.CreateDepartment)
		r.With(authGuard).Get("/companies/{uuid}/departments", departmentAPI.ListDepartments)
		r.With(authGuard).Get("/companies/{uuid}/departments/{department_uuid}/members", departmentAPI.ListDepartmentMembers)
		r.With(authGuard).Post("/companies/{uuid}/departments/{department_uuid}/members", departmentAPI.AddDepartmentMember)
		r.With(authGuard).Patch("/companies/{uuid}/departments/{department_uuid}/members/{user_uuid}/role", departmentAPI.UpdateDepartmentMemberRole)
		r.With(authGuard).Patch("/companies/{uuid}/departments/{department_uuid}/members/{user_uuid}/status", departmentAPI.UpdateDepartmentMemberStatus)
	})

	return r
}
