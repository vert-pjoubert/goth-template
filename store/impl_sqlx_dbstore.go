package store

import (
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/store/models"
)

type SqlxDbStore struct {
	db *sqlx.DB
}

func NewSqlxDbStore(db *sqlx.DB) *SqlxDbStore {
	return &SqlxDbStore{db: db}
}

// ##############################################################
// Generic Utility Functions

// FilterBy retrieves multiple records from a specified table based on a given field and value.
func FilterBy(db *sqlx.DB, tableName string, fieldName string, fieldValue interface{}, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", tableName, fieldName)
	return db.Select(dest, query, fieldValue)
}

// GetTableByFilter retrieves a single record from a specified table based on a given field and value.
func GetTableByFilter(db *sqlx.DB, tableName string, fieldName string, fieldValue interface{}, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", tableName, fieldName)
	return db.Get(dest, query, fieldValue)
}

// ##############################################################
// User Methods

func (s *SqlxDbStore) CreateUser(user *models.User) error {
	query := `INSERT INTO users (name, email, password) VALUES (:name, :email, :password)`
	_, err := s.db.NamedExec(query, user)
	return err
}

func (s *SqlxDbStore) GetUserByEmail(email string) (*models.User, error) {
	user := new(models.User)
	err := GetTableByFilter(s.db, "users", "email", email, user)
	return user, err
}

func (s *SqlxDbStore) UpdateUser(user *models.User) error {
	query := `UPDATE users SET name = :name, email = :email, password = :password WHERE id = :id`
	_, err := s.db.NamedExec(query, user)
	return err
}

func (s *SqlxDbStore) DeleteUser(user *models.User) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := s.db.Exec(query, user.ID)
	return err
}

// ##############################################################
// Role Methods

func (s *SqlxDbStore) CreateRole(role *models.Role) error {
	query := `INSERT INTO roles (name) VALUES (:name)`
	_, err := s.db.NamedExec(query, role)
	return err
}

func (s *SqlxDbStore) GetRoleByID(id int64) (*models.Role, error) {
	role := new(models.Role)
	err := GetTableByFilter(s.db, "roles", "id", id, role)
	return role, err
}

func (s *SqlxDbStore) UpdateRole(role *models.Role) error {
	query := `UPDATE roles SET name = :name WHERE id = :id`
	_, err := s.db.NamedExec(query, role)
	return err
}

func (s *SqlxDbStore) DeleteRole(role *models.Role) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := s.db.Exec(query, role.ID)
	return err
}

// ##############################################################
// Server Methods

func (s *SqlxDbStore) CreateServer(server *models.Server) error {
	query := `INSERT INTO servers (name, ip) VALUES (:name, :ip)`
	_, err := s.db.NamedExec(query, server)
	return err
}

func (s *SqlxDbStore) GetServerByID(id int64) (*models.Server, error) {
	server := new(models.Server)
	err := GetTableByFilter(s.db, "servers", "id", id, server)
	return server, err
}

func (s *SqlxDbStore) UpdateServer(server *models.Server) error {
	query := `UPDATE servers SET name = :name, ip = :ip WHERE id = :id`
	_, err := s.db.NamedExec(query, server)
	return err
}

func (s *SqlxDbStore) DeleteServer(id int64) error {
	query := `DELETE FROM servers WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// ##############################################################
// Event Methods

func (s *SqlxDbStore) CreateEvent(event *models.Event) error {
	query := `INSERT INTO events (name, date) VALUES (:name, :date)`
	_, err := s.db.NamedExec(query, event)
	return err
}

// Helper struct to handle JSON string conversion
type eventWithJSONString struct {
	models.Event
	RolesJSON string `db:"roles"`
}

func (s *SqlxDbStore) GetEventByID(id int64) (*models.Event, error) {
	event := new(eventWithJSONString)
	err := GetTableByFilter(s.db, "events", "id", id, event)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON string in the RolesJSON field into the Roles field
	if err := json.Unmarshal([]byte(event.RolesJSON), &event.Roles); err != nil {
		return nil, fmt.Errorf("error unmarshalling roles: %v", err)
	}

	return &event.Event, nil
}

func (s *SqlxDbStore) UpdateEvent(event *models.Event) error {
	query := `UPDATE events SET name = :name, date = :date WHERE id = :id`
	_, err := s.db.NamedExec(query, event)
	return err
}

func (s *SqlxDbStore) DeleteEvent(id int64) error {
	query := `DELETE FROM events WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

// ##############################################################
// Get Data for Views

func (s *SqlxDbStore) GetServers(servers *[]models.Server) error {
	query := `SELECT * FROM servers`
	return s.db.Select(servers, query)
}

func (s *SqlxDbStore) GetEvents(events *[]models.Event) error {
	query := `SELECT * FROM events`
	return s.db.Select(events, query)
}

// ##############################################################
// Example of FilterBy Usage

func (s *SqlxDbStore) GetEventsByRole(role string) ([]models.Event, error) {
	var events []models.Event
	err := FilterBy(s.db, "events", "role", role, &events)
	return events, err
}

// GetRoleByName retrieves a role by name
func (s *SqlxDbStore) GetRoleByName(name string) (*models.Role, error) {
	role := new(models.Role)
	err := GetTableByFilter(s.db, "roles", "name", name, role)
	return role, err
}
