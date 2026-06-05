package call

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

func (r *Repository) GetByUUID(ctx context.Context, callUUID uuid.UUID, userID uuid.UUID) (model.Call, error) {
	var repoCall repoModel.Call
	getQuery := `
	SELECT call_uuid, 
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
	       FROM calls
	WHERE call_uuid = $1
	  AND uploaded_by_user_uuid = $2
	`
	row := r.db.QueryRowContext(ctx, getQuery, callUUID, userID)

	repoCall, err := scaner.ScanCall(row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Call{}, fmt.Errorf("selecting call failed: %w", model.ErrCallNotFound)
		}
		return model.Call{}, fmt.Errorf("selecting call failed: %w", err)
	}

	return converter.RepoCallToModel(repoCall)
}
