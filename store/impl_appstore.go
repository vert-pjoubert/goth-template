package store

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/store/models"
)

type CachedAppStore struct {
	dbStore DbStore
	cache   map[string]*models.User
	session auth.ISessionManager
}

func NewCachedAppStore(dbStore DbStore, session auth.ISessionManager) *CachedAppStore {
	return &CachedAppStore{
		dbStore: dbStore,
		cache:   make(map[string]*models.User),
		session: session,
	}
}

func (s *CachedAppStore) GetUserWithRoleByEmail(email string) (*models.User, error) {
	if user, ok := s.cache[email]; ok {
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
	s.cache[email] = user
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

	s.cache[user.Email] = user
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

func (s *CachedAppStore) GetServers(servers *[]models.Server) error {
	return s.dbStore.GetServers(servers)
}

func (s *CachedAppStore) GetEvents(events *[]models.Event) error {
	return s.dbStore.GetEvents(events)
}
