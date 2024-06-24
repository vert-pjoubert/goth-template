package auth

import "strings"

// ConvertPermissionsToString converts a slice of permissions to a semicolon-separated string
func ConvertPermissionsToString(permissions []string) string {
	return strings.Join(permissions, ";")
}

// ConvertStringToPermissions converts a semicolon-separated string to a slice of permissions
func ConvertStringToPermissions(permissionsString string) []string {
	if permissionsString == "" {
		return []string{}
	}
	return strings.Split(permissionsString, ";")
}

// HasPermission checks if a permission exists in the list
func HasPermission(permissions []string, requiredPermission string) bool {
	requiredPermission = strings.ToLower(requiredPermission)
	for _, permission := range permissions {
		if strings.ToLower(permission) == requiredPermission {
			return true
		}
	}
	return false
}
