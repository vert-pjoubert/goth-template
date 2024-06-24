package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/store/models"
)

// IAuthenticator interface, Do not alter. Main HARD depends on this interface.
type IAuthenticator interface {
	LoginHandler(http.ResponseWriter, *http.Request)
	CallbackHandler(http.ResponseWriter, *http.Request)
	LogoutHandler(http.ResponseWriter, *http.Request)
	IsAuthenticated(w http.ResponseWriter, r *http.Request) (bool, error)
	HasPermission(userRole string, requiredPermission string) (bool, error)
}

type DbStore interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(user *models.User) error
	CreateRole(role *models.Role) error
	GetRoleByID(id int64) (*models.Role, error)
	UpdateRole(role *models.Role) error
	DeleteRole(role *models.Role) error
	GetServers(servers *[]models.Server) error
	GetEvents(events *[]models.Event) error
}

// Update IAppStore interface to include GetRoleByName
type IAppStore interface {
	GetUserWithRoleByEmail(email string) (*models.User, error)
	CreateUserWithRole(user *models.User, role *models.Role) error
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error
	GetServers(servers *[]models.Server) error
	GetEvents(events *[]models.Event) error
	GetRoleByName(name string) (*models.Role, error) // New method
}

type ISessionManager interface {
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error
}
