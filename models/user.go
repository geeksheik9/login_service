package models

// User is the implementation of a user that would log in
// swagger:model
type User struct {
	Username  string   `json:"username" bson:"username"`
	FirstName string   `json:"firstName" bson:"firstName"`
	LastName  string   `json:"lastName" bson:"lastName"`
	Password  string   `json:"password,omitempty" bson:"password"`
	Token     string   `json:"token,omitempty" bson:"token"`
	Roles     []string `json:"roles,omitempty" bson:"roles"`
}
