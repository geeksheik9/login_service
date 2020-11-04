package models

// User is the implementation of a user that would log in
// swagger:model
type User struct {
	Username  string   `json:"username"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Password  string   `json:"password,omitempty"`
	Token     string   `json:"token,omitempty"`
	Roles     []string `json:"roles,omitempty"`
}
