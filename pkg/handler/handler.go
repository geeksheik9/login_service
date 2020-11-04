package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/geeksheik9/login-service/models"
	"github.com/geeksheik9/login-service/pkg/api"
	model "github.com/geeksheik9/sheet-CRUD/models"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

//LoginDatabase is the interface setup for the login service
type LoginDatabase interface {
	RegisterUser(user *models.User) error
	LoginUser(user *models.User) (string, error)
	Ping() error
}

//LoginService is the implementation of a service to login to an application
type LoginService struct {
	Version  string
	Database LoginDatabase
}

//Routes sets up the routes for the RESTful interface
func (s *LoginService) Routes(r *mux.Router) *mux.Router {
	r.HandleFunc("/ping", s.PingCheck).Methods(http.MethodGet)
	r.Handle("/health", s.healthCheck(s.Database)).Methods(http.MethodGet)

	r.HandleFunc("/register", s.RegisterUser).Methods(http.MethodPost)

	r.HandleFunc("/login", s.LoginUser).Methods(http.MethodPost)

	r.HandleFunc("/profile", s.GetUserProfile).Methods(http.MethodGet)

	return r
}

//PingCheck checks that the app is running and returns 200, OK, version
func (s *LoginService) PingCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK, " + s.Version))
}

func (s *LoginService) healthCheck(database LoginDatabase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbErr := database.Ping()
		var stringDBErr string

		if dbErr != nil {
			stringDBErr = dbErr.Error()
		}

		response := model.HealthCheckResponse{
			APIVersion: s.Version,
			DBError:    stringDBErr,
		}

		if dbErr != nil {
			api.RespondWithJSON(w, http.StatusFailedDependency, response)
			return
		}

		api.RespondWithJSON(w, http.StatusOK, response)
	})
}

// RegisterUser allows a user to register if a username does not already exist
func (s *LoginService) RegisterUser(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("RegisterUser invoked with URL: %v", r.URL)
	defer r.Body.Close()

	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	err = s.Database.RegisterUser(&user)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "User Created")
	return
}

// LoginUser checks the database for a user and compare allowing users to login
func (s *LoginService) LoginUser(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("LoginUser invoked with URL: %v", r.URL)
	defer r.Body.Close()

	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	token, err := s.Database.LoginUser(&user)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, token)
	return

}

// GetUserProfile returns all the information for users
func (s *LoginService) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("GetUserProfile invoked with URL: %v", r.URL)
	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})

	var result models.User
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result.Username = claims["username"].(string)
		result.FirstName = claims["firstname"].(string)
		result.LastName = claims["lastname"].(string)
		if claims["roles"] != nil {
			result.Roles = claims["roles"].([]string)
		}
		api.RespondWithJSON(w, http.StatusOK, result)
		return
	}

	api.RespondWithError(w, api.CheckError(err), err.Error())
	return
}
