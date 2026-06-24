package call

import (
	"context"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"

	"github.com/google/uuid"
)

func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]model.Call, error) {
	var calls []repoModel.Call

	qList := fmt.Sprintf(`
	SELECT c.call_uuid,
	       title,
	       status,
	       audio_path,
	       original_filename,
	       mime_type,
	       size_bytes,
	       duration_seconds,
	       uploaded_by_user_uuid,
	       company_uuid,
	       department_uuid,
	       visibility_scope,
	       created_at
	FROM calls c
	WHERE %s
	ORDER BY created_at DESC
	`, visibleToUserCondition("c", "$1"))

	rows, err := r.db.QueryContext(ctx, qList, userID)
	if err != nil {
		return nil, fmt.Errorf("list calls: %w", err)
	}
	defer func() { _ = rows.Close() }()

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
