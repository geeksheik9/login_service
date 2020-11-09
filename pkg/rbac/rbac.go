package rbac

import "github.com/geeksheik9/login-service/models"

// PerformRBACCheck is a work in progress
func PerformRBACCheck(user models.User, requiredRoles []models.Role) bool {
	var matches []bool
	for _, required := range requiredRoles {
		for _, user := range user.Roles {
			if required == user {
				matches = append(matches, true)
				break
			}
		}
	}

	match := true
	if matches == nil {
		match = false
	} else {
		for _, val := range matches {
			if val == false {
				match = false
				break
			}
		}
	}

	return match
}
