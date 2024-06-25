package store

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/repositories/models"
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

// GetRoleByName retrieves a role by name
func (s *SqlxDbStore) GetRoleByName(name string) (*models.Role, error) {
	role := new(models.Role)
	err := GetTableByFilter(s.db, "roles", "name", name, role)
	return role, err
}
