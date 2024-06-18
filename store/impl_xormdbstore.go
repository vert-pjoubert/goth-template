package store

import (
	"github.com/vert-pjoubert/goth-template/store/models"
	"xorm.io/xorm"
)

type XormDbStore struct {
	engine *xorm.Engine
}

func NewXormDbStore(engine *xorm.Engine) *XormDbStore {
	return &XormDbStore{engine: engine}
}

// ##############################################################
// User Methods
func (s *XormDbStore) CreateUser(user *models.User) error {
	_, err := s.engine.Insert(user)
	return err
}

func (s *XormDbStore) GetUserByEmail(email string) (*models.User, error) {
	user := new(models.User)
	has, err := s.engine.Where("email = ?", email).Get(user)
	if err != nil || !has {
		return nil, err
	}
	return user, nil
}

func (s *XormDbStore) UpdateUser(user *models.User) error {
	_, err := s.engine.ID(user.ID).Update(user)
	return err
}

func (s *XormDbStore) DeleteUser(user *models.User) error {
	_, err := s.engine.ID(user.ID).Delete(user)
	return err
}

// ##############################################################
// Role Methods
func (s *XormDbStore) CreateRole(role *models.Role) error {
	_, err := s.engine.Insert(role)
	return err
}

func (s *XormDbStore) GetRoleByID(id int64) (*models.Role, error) {
	role := new(models.Role)
	has, err := s.engine.ID(id).Get(role)
	if err != nil || !has {
		return nil, err
	}
	return role, nil
}

func (s *XormDbStore) UpdateRole(role *models.Role) error {
	_, err := s.engine.ID(role.ID).Update(role)
	return err
}

func (s *XormDbStore) DeleteRole(role *models.Role) error {
	_, err := s.engine.ID(role.ID).Delete(role)
	return err
}

// ##############################################################
// Server methods
func (s *XormDbStore) CreateServer(server *models.Server) error {
	_, err := s.engine.Insert(server)
	return err
}

func (s *XormDbStore) GetServerByID(id int64) (*models.Server, error) {
	server := new(models.Server)
	has, err := s.engine.ID(id).Get(server)
	if err != nil || !has {
		return nil, err
	}
	return server, nil
}

func (s *XormDbStore) UpdateServer(server *models.Server) error {
	_, err := s.engine.ID(server.ID).Update(server)
	return err
}

func (s *XormDbStore) DeleteServer(id int64) error {
	_, err := s.engine.ID(id).Delete(new(models.Server))
	return err
}

// ##############################################################
// Event methods
func (s *XormDbStore) CreateEvent(event *models.Event) error {
	_, err := s.engine.Insert(event)
	return err
}

func (s *XormDbStore) GetEventByID(id int64) (*models.Event, error) {
	event := new(models.Event)
	has, err := s.engine.ID(id).Get(event)
	if err != nil || !has {
		return nil, err
	}
	return event, nil
}

func (s *XormDbStore) UpdateEvent(event *models.Event) error {
	_, err := s.engine.ID(event.ID).Update(event)
	return err
}

func (s *XormDbStore) DeleteEvent(id int64) error {
	_, err := s.engine.ID(id).Delete(new(models.Event))
	return err
}

// ##############################################################
// Get data for views
func (s *XormDbStore) GetServers(servers *[]models.Server) error {
	return s.engine.Find(servers)
}

func (s *XormDbStore) GetEvents(events *[]models.Event) error {
	return s.engine.Find(events)
}
