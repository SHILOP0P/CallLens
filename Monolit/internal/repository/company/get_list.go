package company

import (
	"context"
	"fmt"

	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"

	"github.com/google/uuid"
)

func (r *Repository) ListUserCompanies(ctx context.Context, userID uuid.UUID) ([]model.Company, error) {
	query := `
	SELECT c.company_uuid,
	       c.name,
	       c.manager_user_uuid,
	       c.member_limit,
	       c.created_at
	FROM companies c
	JOIN company_members cm ON cm.company_uuid = c.company_uuid
	WHERE cm.user_uuid = $1
	  AND cm.status = 'active'
	ORDER BY c.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list user companies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var companies []repoModel.Company
	for rows.Next() {
		company, err := scaner.ScanCompany(rows)
		if err != nil {
			return nil, fmt.Errorf("list user companies: %w", err)
		}

		companies = append(companies, company)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user companies: %w", err)
	}

	return converter.RepoCompaniesToModels(companies)
}
