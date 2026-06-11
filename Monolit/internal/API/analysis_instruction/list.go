package analysis_instruction

import (
	"calllens/monolit/internal/API/response"
	"calllens/monolit/internal/converter"
	"calllens/monolit/internal/models"
	"net/http"

	"github.com/google/uuid"
)

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, response.CodeUnauthorized, "unauthorized")
		return
	}

	scope, _, companyUUID, departmentUUID, err := parseInstructionPlacement(
		r.URL.Query().Get("scope"),
		userID,
		r.URL.Query().Get("company_uuid"),
		r.URL.Query().Get("department_uuid"),
	)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, response.CodeInvalidAnalysisInstructionInput, "invalid instruction placement")
		return
	}

	instructions, err := h.service.List(r.Context(), models.ListAnalysisInstructionsInput{
		Scope:          scope,
		UserUUID:       userID,
		CompanyUUID:    companyUUID,
		DepartmentUUID: departmentUUID,
	})
	if err != nil {
		writeInstructionError(w, err, response.CodeFailedToListInstructions, "failed to list instructions")
		return
	}

	resp, err := converter.AnalysisInstructionModelsToAPI(instructions)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, response.CodeFailedToConvertInstruction, "failed to convert instructions")
		return
	}

	_ = response.WriteJSON(w, http.StatusOK, resp)
}

func parseInstructionUUID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, models.ErrInvalidAnalysisInstructionInput
	}

	return id, nil
}
