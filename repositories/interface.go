package repositories

import "github.com/vert-pjoubert/goth-template/repositories/models"

// UserRepository interface defines the methods required for user operations.
type UserRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
}

// RoleRepository interface defines the methods required for role operations.
type RoleRepository interface {
	GetRoleByID(id int64) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	CreateRole(role *models.Role) error
}
