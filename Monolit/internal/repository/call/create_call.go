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
)

func (r *Repository) CreateCall(ctx context.Context, call model.Call) (model.Call, error) {
	repoCall, err := converter.ModelCallToRepoCall(call)
	if err != nil {
		return model.Call{}, model.ErrCallConvert
	}
	var repoCallNew repoModel.Call

	create := `
	INSERT INTO calls (
		call_uuid,
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
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
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
	          visibility_scope,
	          created_at
	`

	row := r.db.QueryRowContext(ctx, create,
		repoCall.ID,
		repoCall.Title,
		repoCall.Status,
		repoCall.AudioPath,
		repoCall.OriginalFilename,
		repoCall.MimeType,
		repoCall.SizeBytes,
		repoCall.DurationSeconds,
		repoCall.UploadedByUserUUID,
		repoCall.CompanyUUID,
		repoCall.DepartmentUUID,
		repoCall.VisibilityScope,
		repoCall.CreatedAt,
	)

	repoCallNew, err = scaner.ScanCall(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Call{}, fmt.Errorf("creating call failed: %w", model.ErrCallNotFound)
		}
		return model.Call{}, fmt.Errorf("creating call failed: %w", err)
	}

	return converter.RepoCallToModel(repoCallNew)
}
