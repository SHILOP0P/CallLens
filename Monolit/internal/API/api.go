package API

import "net/http"

type API interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetByUUID(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}
