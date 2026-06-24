package analysis

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	"calllens/monolit/internal/repository/scaner"

	"github.com/google/uuid"
)

func (r *Repository) GetByCallUUID(ctx context.Context, callID uuid.UUID) (model.CallAnalysis, error) {
	query := `
	SELECT ` + analysisReturningColumns + `
	FROM call_analyses
	WHERE call_uuid = $1
	`

	row := r.db.QueryRowContext(ctx, query, callID)

	repoAnalysis, err := scaner.ScanCallAnalysis(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CallAnalysis{}, model.ErrAnalysisNotFound
		}
		return model.CallAnalysis{}, fmt.Errorf("get analysis by call uuid: %w", err)
	}

	return converter.RepoCallAnalysisToModel(repoAnalysis)
}
