package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/repositories/models"
)

type SQLRoleRepository struct {
	db *sqlx.DB
}

func NewSQLRoleRepository(db *sqlx.DB) *SQLRoleRepository {
	return &SQLRoleRepository{db: db}
}

func (repo *SQLRoleRepository) GetRoleByID(id int64) (*models.Role, error) {
	role := &models.Role{}
	err := repo.db.Get(role, "SELECT * FROM roles WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (repo *SQLRoleRepository) GetRoleByName(name string) (*models.Role, error) {
	role := &models.Role{}
	err := repo.db.Get(role, "SELECT * FROM roles WHERE name = $1", name)
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (repo *SQLRoleRepository) CreateRole(role *models.Role) error {
	query := `INSERT INTO roles (name, description, permissions) VALUES (:name, :description, :permissions)`
	_, err := repo.db.NamedExec(query, role)
	return err
}
