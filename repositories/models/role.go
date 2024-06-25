package models

// Role represents a role with permissions
type Role struct {
	ID          int64  `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
	Permissions string `db:"permissions" json:"permissions"` // Changed to string, separated by a semi-colon ";"
}
