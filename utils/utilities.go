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
