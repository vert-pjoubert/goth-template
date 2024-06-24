package store

import (
	"github.com/vert-pjoubert/goth-template/store/models"
)

type DbStore interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(user *models.User) error
	CreateRole(role *models.Role) error
	GetRoleByID(id int64) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error) // New method
	UpdateRole(role *models.Role) error
	DeleteRole(role *models.Role) error
	GetServers(servers *[]models.Server) error
	GetEvents(events *[]models.Event) error
}
