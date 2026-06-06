package company

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"fmt"
)

func (r *Repository) CreateCompany(ctx context.Context, company model.Company, member model.CompanyMember) (model.Company, error) {
	repoCompany, err := converter.ModelCompanyToRepoCompany(company)
	if err != nil {
		return model.Company{}, fmt.Errorf("convert company: %w", err)
	}

	repoMember, err := converter.ModelCompanyMemberToRepoCompanyMember(member)
	if err != nil {
		return model.Company{}, fmt.Errorf("convert company member: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Company{}, fmt.Errorf("begin create company transaction: %w", err)
	}
	defer tx.Rollback()

	createCompanyQuery := `
	INSERT INTO companies (
		company_uuid,
		name,
		manager_user_uuid,
		member_limit,
		created_at
	)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING company_uuid,
	          name,
	          manager_user_uuid,
	          member_limit,
	          created_at
	`

	row := tx.QueryRowContext(
		ctx,
		createCompanyQuery,
		repoCompany.ID,
		repoCompany.Name,
		repoCompany.ManagerUserUUID,
		repoCompany.MemberLimit,
		repoCompany.CreatedAt,
	)

	var createdCompany repoModel.Company
	createdCompany, err = scaner.ScanCompany(row)
	if err != nil {
		return model.Company{}, fmt.Errorf("create company: %w", err)
	}

	createMemberQuery := `
	INSERT INTO company_members (
		company_uuid,
		user_uuid,
		role,
		status,
		created_at
	)
	VALUES ($1, $2, $3, $4, $5)
	`

	if _, err := tx.ExecContext(
		ctx,
		createMemberQuery,
		repoMember.CompanyUUID,
		repoMember.UserUUID,
		repoMember.Role,
		repoMember.Status,
		repoMember.CreatedAt,
	); err != nil {
		return model.Company{}, fmt.Errorf("create company member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return model.Company{}, fmt.Errorf("commit create company transaction: %w", err)
	}

	return converter.RepoCompanyToModel(createdCompany)
}
