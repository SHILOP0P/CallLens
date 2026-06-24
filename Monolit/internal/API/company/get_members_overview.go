package company

import (
	"errors"
	"net/http"

	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handler) GetCompanyMembersOverview(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	companyID, err := uuid.Parse(chi.URLParam(r, "uuid"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidCompanyInput, "invalid company uuid")
		return
	}

	overview, err := h.service.GetCompanyMembersOverview(r.Context(), companyID, userID)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCompanyInput) {
			response.WriteError(w, http.StatusBadRequest, response.CodeInvalidCompanyInput, "invalid company input")
			return
		}
		if errors.Is(err, models.ErrCompanyNotFound) {
			response.WriteError(w, http.StatusNotFound, response.CodeCompanyNotFound, "company not found")
			return
		}
		if errors.Is(err, models.ErrForbidden) {
			response.WriteError(w, http.StatusForbidden, response.CodeForbidden, "forbidden")
			return
		}
		if errors.Is(err, models.ErrSubscriptionRequired) {
			response.WriteError(w, http.StatusPaymentRequired, response.CodeSubscriptionRequired, "subscription required")
			return
		}

		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToGetCompanyMembers, "failed to get company members")
		return
	}

	resp, err := converter.CompanyMembersOverviewModelToAPI(overview)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToConvertCompany, "failed to convert company members")
		return
	}

	if err := response.WriteJSON(w, http.StatusOK, resp); err != nil {
		return
	}
}
