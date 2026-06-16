package API

import "net/http"

type CallAPI interface {
	//POST
	Create(w http.ResponseWriter, r *http.Request)

	//GET
	GetByUUID(w http.ResponseWriter, r *http.Request)
	Events(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	GetAudioByUUID(w http.ResponseWriter, r *http.Request)
	GetTranscriptionByCallUUID(w http.ResponseWriter, r *http.Request)

	//UPDATE
	UpdateCallTitle(w http.ResponseWriter, r *http.Request)
	//DELETE
	DeleteCall(w http.ResponseWriter, r *http.Request)
}

type AuthAPI interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	LogoutAll(w http.ResponseWriter, r *http.Request)
	Me(w http.ResponseWriter, r *http.Request)
}

type CompanyAPI interface {
	Create(w http.ResponseWriter, r *http.Request)
	AddCompanyMember(w http.ResponseWriter, r *http.Request)
	UpdateCompanyMemberRole(w http.ResponseWriter, r *http.Request)
	UpdateCompanyMemberStatus(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	GetByUUID(w http.ResponseWriter, r *http.Request)
	GetCompanyMembersOverview(w http.ResponseWriter, r *http.Request)
}

type DepartmentAPI interface {
	CreateDepartment(w http.ResponseWriter, r *http.Request)
	AddDepartmentMember(w http.ResponseWriter, r *http.Request)
	ListDepartmentMembers(w http.ResponseWriter, r *http.Request)
	UpdateDepartmentMemberRole(w http.ResponseWriter, r *http.Request)
	UpdateDepartmentMemberStatus(w http.ResponseWriter, r *http.Request)
	ListDepartments(w http.ResponseWriter, r *http.Request)
}

type AnalysisInstructionAPI interface {
	Create(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	GetFile(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type AnalysisAPI interface {
	AnalyzeCall(w http.ResponseWriter, r *http.Request)
	GetByCallUUID(w http.ResponseWriter, r *http.Request)
}

type ReportAPI interface {
	Create(w http.ResponseWriter, r *http.Request)
	ListByCallUUID(w http.ResponseWriter, r *http.Request)
	Download(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type BillingAPI interface {
	ListPlans(w http.ResponseWriter, r *http.Request)
	GetPersonalSubscription(w http.ResponseWriter, r *http.Request)
	GetCompanySubscription(w http.ResponseWriter, r *http.Request)
	ActivatePersonalSubscription(w http.ResponseWriter, r *http.Request)
	ActivateCompanySubscription(w http.ResponseWriter, r *http.Request)
	CancelCompanySubscription(w http.ResponseWriter, r *http.Request)
}

type InvitationAPI interface {
	CreateCompanyInvitation(w http.ResponseWriter, r *http.Request)
	CreateDepartmentInvitation(w http.ResponseWriter, r *http.Request)
	ListUserInvitations(w http.ResponseWriter, r *http.Request)
	AcceptInvitation(w http.ResponseWriter, r *http.Request)
	DeclineInvitation(w http.ResponseWriter, r *http.Request)
	CancelCompanyInvitation(w http.ResponseWriter, r *http.Request)
	CancelDepartmentInvitation(w http.ResponseWriter, r *http.Request)
}
