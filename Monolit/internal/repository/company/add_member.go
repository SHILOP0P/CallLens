package company

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"fmt"
)

func (r *Repository) AddCompanyMember(ctx context.Context, member model.CompanyMember) (model.CompanyMember, error) {
	repoMember, err := converter.ModelCompanyMemberToRepoCompanyMember(member)
	if err != nil {
		return model.CompanyMember{}, fmt.Errorf("convert company member: %w", err)
	}

	query := `
	INSERT INTO company_members (
		company_uuid,
		user_uuid,
		role,
		status,
		created_at
	)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (company_uuid, user_uuid)
	DO UPDATE SET role = EXCLUDED.role,
	              status = EXCLUDED.status
	RETURNING company_uuid,
	          user_uuid,
	          role,
	          status,
	          created_at
	`

	row := r.db.QueryRowContext(
		ctx,
		query,
		repoMember.CompanyUUID,
		repoMember.UserUUID,
		repoMember.Role,
		repoMember.Status,
		repoMember.CreatedAt,
	)

	var createdMember repoModel.CompanyMember
	createdMember, err = scaner.ScanCompanyMember(row)
	if err != nil {
		return model.CompanyMember{}, fmt.Errorf("add company member: %w", err)
	}

	return converter.RepoCompanyMemberToModel(createdMember)
}
