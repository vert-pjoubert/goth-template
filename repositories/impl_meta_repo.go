package repositories

import (
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/repositories/models"
)

type SQLRepositoryMetadataRepository struct {
	db *sqlx.DB
}

func NewSQLRepositoryMetadataRepository(db *sqlx.DB) *SQLRepositoryMetadataRepository {
	return &SQLRepositoryMetadataRepository{db: db}
}

func (repo *SQLRepositoryMetadataRepository) GetRepositoryMetadataByID(id string) (*models.RepositoryMetadata, error) {
	repositoryMetadata := &models.RepositoryMetadata{}
	err := repo.db.Get(repositoryMetadata, "SELECT * FROM repository_metadata WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return repositoryMetadata, nil
}

func (repo *SQLRepositoryMetadataRepository) CreateRepositoryMetadata(repositoryMetadata *models.RepositoryMetadata) error {
	query := `INSERT INTO repository_metadata (id, location, created_at, updated_at) VALUES (:id, :location, :created_at, :updated_at)`
	_, err := repo.db.NamedExec(query, repositoryMetadata)
	return err
}

func (repo *SQLRepositoryMetadataRepository) GetAllRepositoryMetadata() ([]*models.RepositoryMetadata, error) {
	var repositoryMetadata []*models.RepositoryMetadata
	err := repo.db.Select(&repositoryMetadata, "SELECT * FROM repository_metadata")
	if err != nil {
		return nil, err
	}
	return repositoryMetadata, nil
}
