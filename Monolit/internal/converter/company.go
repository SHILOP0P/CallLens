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
