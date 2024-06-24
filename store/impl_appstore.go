package store

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/store/models"
)

type CachedAppStore struct {
	dbStore   DbStore
	usercahce map[string]*models.User
	session   auth.ISessionManager
}

func NewCachedAppStore(dbStore DbStore, session auth.ISessionManager) *CachedAppStore {
	return &CachedAppStore{
		dbStore:   dbStore,
		usercahce: make(map[string]*models.User),
		session:   session,
	}
}

func (s *CachedAppStore) GetUserWithRoleByEmail(email string) (*models.User, error) {
	if user, ok := s.usercahce[email]; ok {
		return user, nil
	}

	user, err := s.dbStore.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	role, err := s.dbStore.GetRoleByID(user.RoleID)
	if err != nil {
		return nil, err
	}

	user.Role = *role
	s.usercahce[email] = user
	return user, nil
}

func (s *CachedAppStore) CreateUserWithRole(user *models.User, role *models.Role) error {
	err := s.dbStore.CreateRole(role)
	if err != nil {
		return err
	}

	user.RoleID = role.ID
	err = s.dbStore.CreateUser(user)
	if err != nil {
		return err
	}

	s.usercahce[user.Email] = user
	return nil
}

func (s *CachedAppStore) GetRoleByName(name string) (*models.Role, error) {
	return s.dbStore.GetRoleByName(name)
}

func (s *CachedAppStore) GetSession(r *http.Request) (*sessions.Session, error) {
	return s.session.GetSession(r)
}

func (s *CachedAppStore) SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error {
	return s.session.SaveSession(r, w, session)
}

func (s *CachedAppStore) GetServers() ([]models.Server, error) {
	var servers []models.Server
	err := s.dbStore.GetServers(&servers)
	return servers, err
}

func (s *CachedAppStore) GetEvents() ([]models.Event, error) {
	var events []models.Event
	err := s.dbStore.GetEvents(&events)
	return events, err
}

// FilterByUserRoles filters items by user roles
func FilterByUserRoles[T any](items []T, user *models.User, getRolesFunc func(T) string) []T {
	var accessibleItems []T
	for _, item := range items {
		roles := getRolesFunc(item)
		if auth.HasRequiredRoles(user, auth.ConvertStringToRoles(roles)) {
			accessibleItems = append(accessibleItems, item)
		}
	}
	return accessibleItems
}
