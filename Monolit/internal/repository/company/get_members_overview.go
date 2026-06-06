package company

import (
	model "calllens/monolit/internal/models"
	repoModel "calllens/monolit/internal/repository/models"
	"calllens/monolit/internal/repository/scaner"
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *Repository) GetCompanyMembersOverview(ctx context.Context, companyID uuid.UUID) (model.CompanyMembersOverview, error) {
	companyMembers, err := r.listActiveCompanyMembers(ctx, companyID)
	if err != nil {
		return model.CompanyMembersOverview{}, err
	}

	departments, err := r.listCompanyDepartments(ctx, companyID)
	if err != nil {
		return model.CompanyMembersOverview{}, err
	}

	departmentMembers, err := r.listActiveDepartmentMembers(ctx, companyID)
	if err != nil {
		return model.CompanyMembersOverview{}, err
	}

	membersByDepartment := make(map[uuid.UUID][]model.DepartmentMember, len(departments))
	for _, member := range departmentMembers {
		membersByDepartment[member.DepartmentUUID] = append(membersByDepartment[member.DepartmentUUID], member)
	}

	overview := model.CompanyMembersOverview{
		CompanyUUID: companyID,
		Departments: make([]model.DepartmentMembersOverview, 0, len(departments)),
	}

	for _, member := range companyMembers {
		switch member.Role {
		case model.CompanyMemberRoleManager:
			memberCopy := member
			overview.Manager = &memberCopy
		case model.CompanyMemberRoleEmployee:
			overview.CompanyEmployees = append(overview.CompanyEmployees, member)
		}
	}

	for _, department := range departments {
		overview.Departments = append(overview.Departments, model.DepartmentMembersOverview{
			Department: department,
			Members:    membersByDepartment[department.ID],
		})
	}

	return overview, nil
}

func (r *Repository) listActiveCompanyMembers(ctx context.Context, companyID uuid.UUID) ([]model.CompanyMember, error) {
	query := `
	SELECT company_uuid,
	       user_uuid,
	       role,
	       status,
	       created_at
	FROM company_members
	WHERE company_uuid = $1
	  AND status = 'active'
	ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("list active company members: %w", err)
	}
	defer rows.Close()

	var members []model.CompanyMember
	for rows.Next() {
		var repoMember repoModel.CompanyMember
		repoMember, err = scaner.ScanCompanyMember(rows)
		if err != nil {
			return nil, fmt.Errorf("list active company members: %w", err)
		}

		members = append(members, model.CompanyMember{
			CompanyUUID: repoMember.CompanyUUID,
			UserUUID:    repoMember.UserUUID,
			Role:        model.CompanyMemberRole(repoMember.Role),
			Status:      model.MembershipStatus(repoMember.Status),
			CreatedAt:   repoMember.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list active company members: %w", err)
	}

	return members, nil
}

func (r *Repository) listCompanyDepartments(ctx context.Context, companyID uuid.UUID) ([]model.Department, error) {
	query := `
	SELECT department_uuid,
	       company_uuid,
	       name,
	       created_at
	FROM departments
	WHERE company_uuid = $1
	ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("list company departments: %w", err)
	}
	defer rows.Close()

	var departments []model.Department
	for rows.Next() {
		var repoDepartment repoModel.Department
		repoDepartment, err = scaner.ScanDepartment(rows)
		if err != nil {
			return nil, fmt.Errorf("list company departments: %w", err)
		}

		departments = append(departments, model.Department{
			ID:          repoDepartment.ID,
			CompanyUUID: repoDepartment.CompanyUUID,
			Name:        repoDepartment.Name,
			CreatedAt:   repoDepartment.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list company departments: %w", err)
	}

	return departments, nil
}

func (r *Repository) listActiveDepartmentMembers(ctx context.Context, companyID uuid.UUID) ([]model.DepartmentMember, error) {
	query := `
	SELECT dm.department_uuid,
	       dm.user_uuid,
	       dm.role,
	       dm.status,
	       dm.created_at
	FROM department_members dm
	JOIN departments d ON d.department_uuid = dm.department_uuid
	WHERE d.company_uuid = $1
	  AND dm.status = 'active'
	ORDER BY dm.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, companyID)
	if err != nil {
		return nil, fmt.Errorf("list active department members: %w", err)
	}
	defer rows.Close()

	var members []model.DepartmentMember
	for rows.Next() {
		var repoMember repoModel.DepartmentMember
		repoMember, err = scaner.ScanDepartmentMember(rows)
		if err != nil {
			return nil, fmt.Errorf("list active department members: %w", err)
		}

		members = append(members, model.DepartmentMember{
			DepartmentUUID: repoMember.DepartmentUUID,
			UserUUID:       repoMember.UserUUID,
			Role:           model.DepartmentMemberRole(repoMember.Role),
			Status:         model.MembershipStatus(repoMember.Status),
			CreatedAt:      repoMember.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list active department members: %w", err)
	}

	return members, nil
}
