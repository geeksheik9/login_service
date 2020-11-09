package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/geeksheik9/login-service/models"
	"github.com/geeksheik9/login-service/pkg/api"
	"github.com/geeksheik9/login-service/pkg/rbac"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

//LoginDatabase is the interface setup for the login service
type LoginDatabase interface {
	RegisterUser(user *models.User) error
	LoginUser(user *models.User) (string, error)
	CreateRole(role *models.Role) error
	DeleteRole(role *models.Role) error
	GetRoles(queryParams url.Values) ([]models.Role, error)
	AddUserRole(user models.User, role *models.Role) error
	RemoveUserRole(user models.User, role *models.Role) error
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

	// swagger:route POST /register RegisterUser
	//
	// Login Service
	//
	// Consumes:
	// - application/json
	// Schemes: http, https
	//
	// responses:
	// 200: description:User Created
	// 400: description:Bad request
	// 500: description:Internal Server Error
	r.HandleFunc("/register", s.RegisterUser).Methods(http.MethodPost)
	// swagger:route POST /login LoginUser
	//
	// Login Service
	//
	// Consumes:
	// - application/json
	// Schemes: http, https
	//
	// responses:
	// 200: description:Success, returns JWT token
	// 400: description:Bad request
	// 404: description:Not Found
	// 500: description:Internal Server Error
	r.HandleFunc("/login", s.LoginUser).Methods(http.MethodPost)
	// swagger:route GET /profile GetUserProfile
	//
	// Login Service
	//
	// Consumes:
	// - application/json
	// Schemes: http, https
	//
	// responses:
	// 200:	User
	// 400: description:Bad request
	// 404: description:NotFound
	// 500: description:Internal Server Error
	r.HandleFunc("/profile", s.GetUserProfile).Methods(http.MethodGet)

	r.HandleFunc("/role", s.CreateRole).Methods(http.MethodPost)

	r.HandleFunc("/role", s.DeleteRole).Methods(http.MethodDelete)

	r.HandleFunc("/role", s.GetRoles).Methods(http.MethodGet)

	r.HandleFunc("/add-role/{role}", s.AddUserRole).Methods(http.MethodPost)

	r.HandleFunc("/remove-role/{role}", s.RemoveUserRole).Methods(http.MethodPost)

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

		response := models.HealthCheckResponse{
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

	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	if tokenString == "" {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	user, err := decodeToken(tokenString)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	role := models.Role{
		Name: "admin",
	}
	requiredRoles := []models.Role{
		role,
	}

	err = s.verifyRoles(requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	authorized, err := rbac.PerformRBACCheck(user, requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}
	if !authorized {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	var register models.User

	err = json.NewDecoder(r.Body).Decode(&register)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	err = s.Database.RegisterUser(&register)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "User Created")
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
}

// GetUserProfile returns all the information for users
func (s *LoginService) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("GetUserProfile invoked with URL: %v", r.URL)
	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	if tokenString == "" {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
	}

	result, err := decodeToken(tokenString)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, result)
}

// CreateRole is the handler func to add a role to the roles collection
func (s *LoginService) CreateRole(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("CreateRole invoked with URL: %v", r.URL)
	defer r.Body.Close()

	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	if tokenString == "" {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	user, err := decodeToken(tokenString)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	role := models.Role{
		Name: "admin",
	}
	requiredRoles := []models.Role{
		role,
	}

	err = s.verifyRoles(requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	authorized, err := rbac.PerformRBACCheck(user, requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}
	if !authorized {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	var create models.Role
	err = json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	err = s.Database.CreateRole(&create)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "Role Created")

}

// DeleteRole is the handler func to remove a role from the roles collection
func (s *LoginService) DeleteRole(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("CreateRole invoked with URL: %v", r.URL)
	defer r.Body.Close()

	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	if tokenString == "" {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	user, err := decodeToken(tokenString)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	role := models.Role{
		Name: "admin",
	}
	requiredRoles := []models.Role{
		role,
	}

	err = s.verifyRoles(requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	authorized, err := rbac.PerformRBACCheck(user, requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}
	if !authorized {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	var delete models.Role
	err = json.NewDecoder(r.Body).Decode(&delete)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	err = s.Database.DeleteRole(&role)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "Role Deleted")
}

// GetRoles is the handler func to return all roles in the role collection
func (s *LoginService) GetRoles(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("CreateRole invoked with URL: %v", r.URL)

	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	if tokenString == "" {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	user, err := decodeToken(tokenString)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	role := models.Role{
		Name: "admin",
	}
	requiredRoles := []models.Role{
		role,
	}

	err = s.verifyRoles(requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	authorized, err := rbac.PerformRBACCheck(user, requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}
	if !authorized {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	roles, err := s.Database.GetRoles(r.URL.Query())
	if err != nil || roles == nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, roles)
}

// AddUserRole is the handler func to add a role to a user
func (s *LoginService) AddUserRole(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("CreateRole invoked with URL: %v", r.URL)
	defer r.Body.Close()

	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	if tokenString == "" {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	admin, err := decodeToken(tokenString)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	role := models.Role{
		Name: "admin",
	}
	requiredRoles := []models.Role{
		role,
	}

	err = s.verifyRoles(requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	authorized, err := rbac.PerformRBACCheck(admin, requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}
	if !authorized {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	vars := mux.Vars(r)
	add := vars["role"]

	query, err := url.ParseQuery("name=" + add)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	roles, err := s.Database.GetRoles(query)
	if err != nil || roles == nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}
	if roles == nil {
		api.RespondWithError(w, http.StatusNotFound, "No role to add to user")
		return
	}
	if len(roles) > 1 {
		api.RespondWithError(w, http.StatusConflict, "More than one rule found, please specify role")
		return
	}

	var user models.User

	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	err = s.Database.AddUserRole(user, &roles[0])
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "Role added to user")
}

// RemoveUserRole is the handler func to remove a role from a user
func (s *LoginService) RemoveUserRole(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("RemovedRole invoked with URL: %v", r.URL)
	defer r.Body.Close()

	tokenString := r.Header.Get("Authorization")
	if strings.Contains(tokenString, "Bearer") {
		tokenString = strings.Trim(tokenString, "Bearer")
		tokenString = strings.Trim(tokenString, " ")
	}
	if tokenString == "" {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	admin, err := decodeToken(tokenString)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	role := models.Role{
		Name: "admin",
	}
	requiredRoles := []models.Role{
		role,
	}

	err = s.verifyRoles(requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	authorized, err := rbac.PerformRBACCheck(admin, requiredRoles)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}
	if !authorized {
		api.RespondWithError(w, http.StatusUnauthorized, "User is not authorized to make this request")
		return
	}

	vars := mux.Vars(r)
	removeRole := vars["role"]

	roles, err := s.Database.GetRoles(nil)
	if err != nil || roles == nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	matches := false
	var roleToRemove models.Role
	for _, role := range roles {
		if removeRole == role.Name {
			matches = true
			roleToRemove = role
		}
	}
	if matches == false {
		api.RespondWithError(w, http.StatusNotFound, "No such role exists")
		return
	}

	var user models.User

	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	err = s.Database.RemoveUserRole(user, &roleToRemove)
	if err != nil {
		api.RespondWithError(w, api.CheckError(err), err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "User Role Removed")
}

/**
 *
 * Helpers
 *
 **/

func decodeToken(tokenString string) (models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})

	if err != nil {
		return models.User{}, err
	}

	var result models.User
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result.Username = claims["username"].(string)
		result.FirstName = claims["firstname"].(string)
		result.LastName = claims["lastname"].(string)
		roles := claims["roles"].([]interface{})
		for _, role := range roles {
			str, _ := json.Marshal(role)
			var roleValue models.Role
			json.Unmarshal(str, &roleValue)
			result.Roles = append(result.Roles, roleValue)
		}
	}
	return result, nil
}

func (s *LoginService) verifyRoles(roles []models.Role) error {
	var exists []bool
	for _, role := range roles {
		query, err := url.ParseQuery("name=" + role.Name)
		if err != nil {
			return err
		}
		returned, err := s.Database.GetRoles(query)
		for _, values := range returned {
			if role == values {
				exists = append(exists, true)
				break
			}
		}
	}
	if len(exists) != len(roles) {
		return errors.New("All required roles not found in database")
	}
	return nil
}
