package company

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

func (r *Repository) GetManagedCompanyByUserUUID(ctx context.Context, userID uuid.UUID) (model.Company, error) {
	query := `
	SELECT company_uuid,
	       name,
	       manager_user_uuid,
	       member_limit,
	       created_at
	FROM companies
	WHERE manager_user_uuid = $1
	`

	row := r.db.QueryRowContext(ctx, query, userID)

	var repoCompany repoModel.Company
	repoCompany, err := scaner.ScanCompany(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Company{}, model.ErrCompanyNotFound
		}

		return model.Company{}, fmt.Errorf("get managed company by user uuid: %w", err)
	}

	return converter.RepoCompanyToModel(repoCompany)
}
