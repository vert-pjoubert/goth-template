package auth

import (
	"strings"

	"github.com/vert-pjoubert/goth-template/store/models"
)

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

func HasRequiredRoles(user *models.User, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}

	userRole := user.Role.Name
	for _, role := range requiredRoles {
		if role == userRole {
			return true
		}
	}

	return false
}

func HasRequiredPermissions(user *models.User, requiredPermissions []string) bool {
	if len(requiredPermissions) == 0 {
		return true
	}

	userPermissions := ConvertStringToPermissions(user.Role.Permissions)
	for _, perm := range requiredPermissions {
		if !contains(userPermissions, perm) {
			return false
		}
	}

	return true
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ConvertStringToRoles converts a semicolon-separated string to a slice of roles
func ConvertStringToRoles(rolesString string) []string {
	if rolesString == "" {
		return []string{}
	}
	return strings.Split(rolesString, ";")
}

// HasRequiredRolesMap checks if the user has any of the roles required to access an item.
func HasRequiredRolesMap(user *models.User, requiredRoles map[string]bool) bool {
    userRoles := ConvertStringToRolesMap(user.Roles)
    for role := range requiredRoles {
        if userRoles[role] {
            return true
        }
    }
    return false
}

// ConvertStringToRolesMap converts semicolon-separated roles string to a map for quicker access.
func ConvertStringToRolesMap(rolesStr string) map[string]bool {
    rolesMap := make(map[string]bool)
    roles := strings.Split(rolesStr, ";")
    for _, role := range roles {
        rolesMap[role] = true
    }
    return rolesMap
}
