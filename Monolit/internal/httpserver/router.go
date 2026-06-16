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

func NewRouter(callAPI API.CallAPI, authAPI API.AuthAPI, companyAPI API.CompanyAPI, departmentAPI API.DepartmentAPI, instructionAPI API.AnalysisInstructionAPI, analysisAPI API.AnalysisAPI, reportAPI API.ReportAPI, billingAPI API.BillingAPI, invitationAPI API.InvitationAPI, jwtSecret string, refreshSessionRepository repository.RefreshSessionRepository, log logger.Logger) http.Handler {
	r := chi.NewRouter()

	authGuard := authMiddleware.Auth(jwtSecret, refreshSessionRepository)

	r.Use(middleware.RequestID)
	r.Use(authMiddleware.RequestLogger(log))
	r.Use(authMiddleware.Recoverer(log))
	r.Use(middleware.URLFormat)

	r.Get("/health", health.Health)
	r.Route("/api/v1", func(r chi.Router) {
		r.With(authGuard).Get("/calls/{uuid}/events", callAPI.Events)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(10 * time.Second))

			//CALL
			//POST
			r.With(authGuard).Post("/calls", callAPI.Create)
			//GET
			r.With(authGuard).Get("/calls", callAPI.List)
			r.With(authGuard).Get("/calls/{uuid}", callAPI.GetByUUID)
			r.With(authGuard).Get("/calls/{uuid}/audio", callAPI.GetAudioByUUID)
			r.With(authGuard).Get("/calls/{uuid}/transcription", callAPI.GetTranscriptionByCallUUID)
			r.With(authGuard).Post("/calls/{uuid}/analysis", analysisAPI.AnalyzeCall)
			r.With(authGuard).Get("/calls/{uuid}/analysis", analysisAPI.GetByCallUUID)
			r.With(authGuard).Post("/calls/{uuid}/reports", reportAPI.Create)
			r.With(authGuard).Get("/calls/{uuid}/reports", reportAPI.ListByCallUUID)
			r.With(authGuard).Get("/reports/{report_uuid}/download", reportAPI.Download)
			r.With(authGuard).Delete("/reports/{report_uuid}", reportAPI.Delete)
			//UPDATE
			r.With(authGuard).Patch("/calls/{uuid}", callAPI.UpdateCallTitle)
			//DELETE
			r.With(authGuard).Delete("/calls/{uuid}", callAPI.DeleteCall)

			//BILLING
			r.Get("/plans", billingAPI.ListPlans)
			r.With(authGuard).Get("/subscription", billingAPI.GetPersonalSubscription)
			r.With(authGuard).Post("/subscription/activate", billingAPI.ActivatePersonalSubscription)
			r.With(authGuard).Get("/companies/{uuid}/subscription", billingAPI.GetCompanySubscription)
			r.With(authGuard).Post("/companies/{uuid}/subscription/activate", billingAPI.ActivateCompanySubscription)
			r.With(authGuard).Post("/companies/{uuid}/subscription/cancel", billingAPI.CancelCompanySubscription)

			//INVITATIONS
			r.With(authGuard).Get("/invitations", invitationAPI.ListUserInvitations)
			r.With(authGuard).Post("/invitations/{invitation_uuid}/accept", invitationAPI.AcceptInvitation)
			r.With(authGuard).Post("/invitations/{invitation_uuid}/decline", invitationAPI.DeclineInvitation)

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
			r.With(authGuard).Post("/companies/{uuid}/invitations", invitationAPI.CreateCompanyInvitation)
			r.With(authGuard).Post("/companies/{uuid}/invitations/{invitation_uuid}/cancel", invitationAPI.CancelCompanyInvitation)
			r.With(authGuard).Patch("/companies/{uuid}/members/{user_uuid}/role", companyAPI.UpdateCompanyMemberRole)
			r.With(authGuard).Patch("/companies/{uuid}/members/{user_uuid}/status", companyAPI.UpdateCompanyMemberStatus)
			r.With(authGuard).Post("/companies/{uuid}/departments", departmentAPI.CreateDepartment)
			r.With(authGuard).Get("/companies/{uuid}/departments", departmentAPI.ListDepartments)
			r.With(authGuard).Get("/companies/{uuid}/departments/{department_uuid}/members", departmentAPI.ListDepartmentMembers)
			r.With(authGuard).Post("/companies/{uuid}/departments/{department_uuid}/members", departmentAPI.AddDepartmentMember)
			r.With(authGuard).Post("/companies/{uuid}/departments/{department_uuid}/invitations", invitationAPI.CreateDepartmentInvitation)
			r.With(authGuard).Post("/companies/{uuid}/departments/{department_uuid}/invitations/{invitation_uuid}/cancel", invitationAPI.CancelDepartmentInvitation)
			r.With(authGuard).Patch("/companies/{uuid}/departments/{department_uuid}/members/{user_uuid}/role", departmentAPI.UpdateDepartmentMemberRole)
			r.With(authGuard).Patch("/companies/{uuid}/departments/{department_uuid}/members/{user_uuid}/status", departmentAPI.UpdateDepartmentMemberStatus)

			//ANALYSIS INSTRUCTIONS
			r.With(authGuard).Post("/instructions", instructionAPI.Create)
			r.With(authGuard).Get("/instructions", instructionAPI.List)
			r.With(authGuard).Get("/instructions/{uuid}/file", instructionAPI.GetFile)
			r.With(authGuard).Delete("/instructions/{uuid}", instructionAPI.Delete)
		})
	})

	return r
}
