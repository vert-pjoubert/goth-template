package store

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/store/models"
)

type AppStore interface {
	GetUserWithRoleByEmail(email string) (*models.User, error)
	CreateUserWithRole(user *models.User, role *models.Role) error
	GetSession(r *http.Request) (*sessions.Session, error)
	SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error
	GetServers(servers *[]models.Server) error
	GetEvents(events *[]models.Event) error
}
