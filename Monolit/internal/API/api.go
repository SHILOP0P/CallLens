package API

import "net/http"

type CallAPI interface {
	//POST
	Create(w http.ResponseWriter, r *http.Request)

	//GET
	GetByUUID(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	GetAudioByUUID(w http.ResponseWriter, r *http.Request)

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
	List(w http.ResponseWriter, r *http.Request)
	GetByUUID(w http.ResponseWriter, r *http.Request)
	CreateDepartment(w http.ResponseWriter, r *http.Request)
	AddDepartmentMember(w http.ResponseWriter, r *http.Request)
	ListDepartments(w http.ResponseWriter, r *http.Request)
}
