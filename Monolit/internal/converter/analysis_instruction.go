package converter

import (
	"time"

	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/models"
)

func AnalysisInstructionModelToAPI(instruction models.AnalysisInstruction) (dto.AnalysisInstruction, error) {
	return dto.AnalysisInstruction{
		ID:                instruction.ID.String(),
		Scope:             string(instruction.Scope),
		UserUUID:          nullUUIDToStringPtr(instruction.UserUUID),
		CompanyUUID:       nullUUIDToStringPtr(instruction.CompanyUUID),
		DepartmentUUID:    nullUUIDToStringPtr(instruction.DepartmentUUID),
		Title:             instruction.Title,
		OriginalFilename:  instruction.OriginalFilename,
		FilePath:          instruction.FilePath,
		MimeType:          instruction.MimeType,
		SizeBytes:         instruction.SizeBytes,
		ContentSHA256:     instruction.ContentSHA256,
		SortOrder:         instruction.SortOrder,
		IsActive:          instruction.IsActive,
		CreatedByUserUUID: instruction.CreatedByUserUUID.String(),
		CreatedAt:         instruction.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         instruction.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func AnalysisInstructionModelsToAPI(instructions []models.AnalysisInstruction) ([]dto.AnalysisInstruction, error) {
	result := make([]dto.AnalysisInstruction, len(instructions))
	for i, instruction := range instructions {
		result[i], _ = AnalysisInstructionModelToAPI(instruction)
	}

	return result, nil
}
