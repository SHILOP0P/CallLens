package analysis_instruction

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) GetByUUID(ctx context.Context, id uuid.UUID) (model.AnalysisInstruction, error) {
	query := `
	SELECT ` + analysisInstructionReturningColumns + `
	FROM analysis_instructions
	WHERE instruction_uuid = $1
	  AND is_active = true
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var repoInstruction repoModel.AnalysisInstruction
	repoInstruction, err := scaner.ScanAnalysisInstruction(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AnalysisInstruction{}, model.ErrAnalysisInstructionNotFound
		}

		return model.AnalysisInstruction{}, fmt.Errorf("get analysis instruction by uuid: %w", err)
	}

	return converter.RepoAnalysisInstructionToModel(repoInstruction)
}
