package handler

import (
	"github.com/gorilla/mux"
)

//LoginService is the implementation of a service to login to an application
type LoginService struct {
	Version string
}

//Routes sets up the routes for the RESTful interface
func (s *LoginService) Routes(r *mux.Router) *mux.Router {

	return r
}
