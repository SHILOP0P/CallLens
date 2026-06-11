package analysis_instruction

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"fmt"
)

func (r *Repository) List(ctx context.Context, input model.ListAnalysisInstructionsInput) ([]model.AnalysisInstruction, error) {
	query := `
	SELECT ` + analysisInstructionReturningColumns + `
	FROM analysis_instructions
	WHERE is_active = true
	  AND scope = $1
	  AND (
	      ($1 = 'personal' AND user_uuid = $2)
	      OR
	      ($1 = 'company' AND company_uuid = $3 AND department_uuid IS NULL)
	      OR
	      ($1 = 'department' AND company_uuid = $3 AND department_uuid = $4)
	  )
	ORDER BY sort_order ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(
		ctx,
		query,
		input.Scope,
		input.UserUUID,
		input.CompanyUUID,
		input.DepartmentUUID,
	)
	if err != nil {
		return nil, fmt.Errorf("list analysis instructions: %w", err)
	}
	defer rows.Close()

	instructions := make([]repoModel.AnalysisInstruction, 0)
	for rows.Next() {
		instruction, err := scaner.ScanAnalysisInstruction(rows)
		if err != nil {
			return nil, fmt.Errorf("scan analysis instruction: %w", err)
		}
		instructions = append(instructions, instruction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate analysis instructions: %w", err)
	}

	return converter.RepoAnalysisInstructionsToModels(instructions)
}
