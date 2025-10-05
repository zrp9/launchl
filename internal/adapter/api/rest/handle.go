package rest

import "net/http"

type Handler interface {
	RegisterRoutes(mr *http.ServeMux)
}

type APIHandler func(w http.ResponseWriter, r *http.Request) error

type APIErr struct {
	Status int
	Err    error
}

func (a APIErr) Error() string {
	return a.Err.Error()
}

func ReturnErr(status int, err error) APIErr {
	return APIErr{
		Status: status,
		Err:    err,
	}
}
