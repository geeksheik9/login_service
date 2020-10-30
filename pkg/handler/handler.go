package handler

import (
	"github.com/geeksheik9/login-service/models"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//LoginDatabase is the interface setup for the login service
type LoginDatabase interface {
	CreateUser(models.User) error
	LoginUser(models.User) error
	UpdateUser(models.User) error
	FindUser(mongoID primitive.ObjectID) models.User
	DeleteUser(mongoID primitive.ObjectID) error
}

//LoginService is the implementation of a service to login to an application
type LoginService struct {
	Version  string
	Database LoginDatabase
}

//Routes sets up the routes for the RESTful interface
func (s *LoginService) Routes(r *mux.Router) *mux.Router {

	return r
}
