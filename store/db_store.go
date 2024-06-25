package store

import (
	"github.com/vert-pjoubert/goth-template/repositories/models"
)

type DbStore interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(user *models.User) error
	CreateRole(role *models.Role) error
	GetRoleByID(id int64) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	UpdateRole(role *models.Role) error
	DeleteRole(role *models.Role) error
}
