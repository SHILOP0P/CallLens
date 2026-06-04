package call

import (
	"calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) UpdateCallTitle(ctx context.Context, id uuid.UUID, title string) (models.Call, error) {
	var repoCall repoModel.Call

	queryUpdate := `
	UPDATE calls
	SET title = $2
	WHERE call_uuid = $1
	RETURNING call_uuid,
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
	          created_at
	`

	row := r.db.QueryRowContext(ctx, queryUpdate, id, title)

	repoCall, err := scaner.ScanCall(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Call{}, models.ErrCallNotFound
		}
		return models.Call{}, fmt.Errorf("update call title failed: %w", err)
	}

	return converter.RepoCallToModel(repoCall)
}
