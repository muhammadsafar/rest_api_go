package utils

import "errors"

func AuthorizeUser(userRole string, allowedRoles ...string) (bool, error) {
	for _, allowedRole := range allowedRoles {
		if userRole == allowedRole {
			return true, nil
		}
	}

	return false, errors.New("User Not Authorize")

}
