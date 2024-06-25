package main

import (
	"net/http"
	"reflect"

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/repositories/models"
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
}

// IAppStore interface defines the core methods required for application operations.
type IAppStore interface {
	GetUserByEmail(email string) (*models.User, error)
	GetUserRoles(user *models.User) ([]string, error)
	GetRolePermissions(role *models.Role) ([]string, error)
	GetRoleByName(name string) (*models.Role, error)
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error
	RegisterRepo(repo interface{})
	GetRepo(repoType reflect.Type) (interface{}, error)
}

type ISessionManager interface {
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(r *http.Request, w http.ResponseWriter, session *sessions.Session) error
}
