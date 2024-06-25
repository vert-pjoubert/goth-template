package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/repositories/models"
)

type SQLUserRepository struct {
	db *sqlx.DB
}

func NewSQLUserRepository(db *sqlx.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

func (repo *SQLUserRepository) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := repo.db.Get(user, "SELECT * FROM users WHERE email = $1", email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *SQLUserRepository) CreateUser(user *models.User) error {
	query := `INSERT INTO users (name, email, roles, created_at, updated_at) VALUES (:name, :email, :roles, :created_at, :updated_at)`
	_, err := repo.db.NamedExec(query, user)
	return err
}
