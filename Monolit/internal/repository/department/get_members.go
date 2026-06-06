package department

import (
	model "calllens/monolit/internal/models"
	"calllens/monolit/internal/repository/converter"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) ListDepartmentMembers(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID) ([]model.DepartmentMember, error) {
	exists, err := r.departmentExists(ctx, companyID, departmentID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, model.ErrDepartmentNotFound
	}

	query := `
	SELECT dm.department_uuid,
	       dm.user_uuid,
	       dm.role,
	       dm.status,
	       dm.created_at
	FROM department_members dm
	JOIN departments d ON d.department_uuid = dm.department_uuid
	WHERE d.company_uuid = $1
	  AND dm.department_uuid = $2
	  AND dm.status = 'active'
	ORDER BY dm.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID, departmentID)
	if err != nil {
		return nil, fmt.Errorf("list department members: %w", err)
	}
	defer rows.Close()

	var repoMembers []repoModel.DepartmentMember
	for rows.Next() {
		member, err := scaner.ScanDepartmentMember(rows)
		if err != nil {
			return nil, fmt.Errorf("list department members: %w", err)
		}

		repoMembers = append(repoMembers, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list department members: %w", err)
	}

	result := make([]model.DepartmentMember, len(repoMembers))
	for i, member := range repoMembers {
		result[i], _ = converter.RepoDepartmentMemberToModel(member)
	}

	return result, nil
}

func (r *Repository) departmentExists(ctx context.Context, companyID uuid.UUID, departmentID uuid.UUID) (bool, error) {
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM departments
		WHERE company_uuid = $1
		  AND department_uuid = $2
	)
	`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, companyID, departmentID).Scan(&exists); err != nil {
		return false, fmt.Errorf("check department exists: %w", err)
	}

	return exists, nil
}
