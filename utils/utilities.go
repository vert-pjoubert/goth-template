package utils

import "strings"

// ConvertRolesToString converts a slice of role IDs to a semicolon-separated string
func ConvertRolesToString(roleIDs []string) string {
	return strings.Join(roleIDs, ";")
}

// ConvertStringToRoles converts a semicolon-separated string to a slice of role IDs
func ConvertStringToRoles(rolesString string) []string {
	if rolesString == "" {
		return []string{}
	}
	return strings.Split(rolesString, ";")
}

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

// paginate slices a list into a specific page with pageSize
func Paginate[T any](items []T, page, pageSize int) []T {
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []T{}
	}

	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}
