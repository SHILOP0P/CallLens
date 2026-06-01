package call

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"fmt"
)

func (r *Repository) List(ctx context.Context) ([]model.Call, error) {
	var calls []repoModel.Call

	qList := `
	SELECT call_uuid,
	       title,
	       status,
	       audio_path,
	       original_filename,
	       mime_type,
	       size_bytes,
	       duration_seconds,
	       created_at
	FROM calllens
	ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, qList)
	if err != nil {
		return nil, fmt.Errorf("list calls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var call repoModel.Call
		call, err = scaner.ScanCall(rows)
		if err != nil {
			return nil, fmt.Errorf("list calls: %w", err)
		}
		calls = append(calls, call)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list calls: %w", err)
	}

	return converter.RepoCallsToModels(calls)
}
