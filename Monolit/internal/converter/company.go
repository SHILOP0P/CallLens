package converter

import (
	"calllens/monolit/internal/API/dto"
	"calllens/monolit/internal/models"
	"time"
)

func CompanyModelToAPI(company models.Company) (dto.CompanyResponse, error) {
	return dto.CompanyResponse{
		ID:              company.ID.String(),
		Name:            company.Name,
		ManagerUserUUID: company.ManagerUserUUID.String(),
		MemberLimit:     company.MemberLimit,
		CreatedAt:       company.CreatedAt.Format(time.RFC3339),
	}, nil
}

func DepartmentModelToAPI(department models.Department) (dto.DepartmentResponse, error) {
	return dto.DepartmentResponse{
		ID:          department.ID.String(),
		CompanyUUID: department.CompanyUUID.String(),
		Name:        department.Name,
		CreatedAt:   department.CreatedAt.Format(time.RFC3339),
	}, nil
}

func CompanyMemberModelToAPI(member models.CompanyMember) (dto.CompanyMemberResponse, error) {
	return dto.CompanyMemberResponse{
		CompanyUUID: member.CompanyUUID.String(),
		UserUUID:    member.UserUUID.String(),
		Role:        string(member.Role),
		Status:      string(member.Status),
		CreatedAt:   member.CreatedAt.Format(time.RFC3339),
	}, nil
}

func DepartmentMemberModelToAPI(member models.DepartmentMember) (dto.DepartmentMemberResponse, error) {
	return dto.DepartmentMemberResponse{
		DepartmentUUID: member.DepartmentUUID.String(),
		UserUUID:       member.UserUUID.String(),
		Role:           string(member.Role),
		Status:         string(member.Status),
		CreatedAt:      member.CreatedAt.Format(time.RFC3339),
	}, nil
}

func CompanyMembersOverviewModelToAPI(overview models.CompanyMembersOverview) (dto.CompanyMembersOverviewResponse, error) {
	resp := dto.CompanyMembersOverviewResponse{
		CompanyUUID:      overview.CompanyUUID.String(),
		CompanyEmployees: make([]dto.CompanyMemberResponse, 0, len(overview.CompanyEmployees)),
		Departments:      make([]dto.DepartmentMembersOverviewResponse, 0, len(overview.Departments)),
	}

	if overview.Manager != nil {
		manager, err := CompanyMemberModelToAPI(*overview.Manager)
		if err != nil {
			return dto.CompanyMembersOverviewResponse{}, err
		}
		resp.Manager = &manager
	}

	for _, member := range overview.CompanyEmployees {
		memberResponse, err := CompanyMemberModelToAPI(member)
		if err != nil {
			return dto.CompanyMembersOverviewResponse{}, err
		}
		resp.CompanyEmployees = append(resp.CompanyEmployees, memberResponse)
	}

	for _, department := range overview.Departments {
		departmentResponse, err := DepartmentModelToAPI(department.Department)
		if err != nil {
			return dto.CompanyMembersOverviewResponse{}, err
		}

		departmentOverview := dto.DepartmentMembersOverviewResponse{
			Department: departmentResponse,
			Members:    make([]dto.DepartmentMemberResponse, 0, len(department.Members)),
		}

		for _, member := range department.Members {
			memberResponse, err := DepartmentMemberModelToAPI(member)
			if err != nil {
				return dto.CompanyMembersOverviewResponse{}, err
			}
			departmentOverview.Members = append(departmentOverview.Members, memberResponse)
		}

		resp.Departments = append(resp.Departments, departmentOverview)
	}

	return resp, nil
}
