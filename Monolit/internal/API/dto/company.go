package dto

type CreateCompanyRequest struct {
	Name string `json:"name"`
}

type CompanyResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ManagerUserUUID string `json:"manager_user_uuid"`
	MemberLimit     int    `json:"member_limit"`
	CreatedAt       string `json:"created_at"`
}

type CreateDepartmentRequest struct {
	Name string `json:"name"`
}

type DepartmentResponse struct {
	ID          string `json:"id"`
	CompanyUUID string `json:"company_uuid"`
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at"`
}

type AddCompanyMemberRequest struct {
	UserUUID string `json:"user_uuid"`
	Role     string `json:"role"`
}

type CompanyMemberResponse struct {
	CompanyUUID string `json:"company_uuid"`
	UserUUID    string `json:"user_uuid"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type AddDepartmentMemberRequest struct {
	UserUUID string `json:"user_uuid"`
	Role     string `json:"role"`
}

type DepartmentMemberResponse struct {
	DepartmentUUID string `json:"department_uuid"`
	UserUUID       string `json:"user_uuid"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
}
