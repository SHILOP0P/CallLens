package API

import "net/http"

type API interface {
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
