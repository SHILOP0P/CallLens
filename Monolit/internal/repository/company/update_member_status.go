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

func (r *Repository) UpdateCompanyMemberStatus(ctx context.Context, companyID uuid.UUID, userID uuid.UUID, status model.MembershipStatus) (model.CompanyMember, error) {
	query := `
	UPDATE company_members
	SET status = $3
	WHERE company_uuid = $1
	  AND user_uuid = $2
	  AND role <> 'company_manager'
	RETURNING company_uuid,
	          user_uuid,
	          role,
	          status,
	          created_at
	`

	row := r.db.QueryRowContext(ctx, query, companyID, userID, string(status))

	var repoMember repoModel.CompanyMember
	repoMember, err := scaner.ScanCompanyMember(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CompanyMember{}, model.ErrCompanyNotFound
		}

		return model.CompanyMember{}, fmt.Errorf("update company member status: %w", err)
	}

	return converter.RepoCompanyMemberToModel(repoMember)
}
