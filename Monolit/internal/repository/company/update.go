package company

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"

	"github.com/google/uuid"
)

func (r *Repository) UpdateCompany(ctx context.Context, companyID uuid.UUID, name string) (model.Company, error) {
	query := `
	UPDATE companies
	SET name = $2
	WHERE company_uuid = $1
	  AND deleted_at IS NULL
	RETURNING company_uuid,
	          name,
	          manager_user_uuid,
	          member_limit,
	          created_at,
	          deleted_at
	`

	row := r.db.QueryRowContext(ctx, query, companyID, name)

	var repoCompany repoModel.Company
	repoCompany, err := scaner.ScanCompany(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Company{}, model.ErrCompanyNotFound
		}

		return model.Company{}, fmt.Errorf("update company: %w", err)
	}

	return converter.RepoCompanyToModel(repoCompany)
}

func (r *Repository) ArchiveCompany(ctx context.Context, companyID uuid.UUID) error {
	query := `
	UPDATE companies
	SET deleted_at = now()
	WHERE company_uuid = $1
	  AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, companyID)
	if err != nil {
		return fmt.Errorf("archive company: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("archive company rows affected: %w", err)
	}
	if rows == 0 {
		return model.ErrCompanyNotFound
	}

	return nil
}
