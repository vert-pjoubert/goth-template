package store

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/repositories"
	"github.com/vert-pjoubert/goth-template/repositories/models"
	"github.com/vert-pjoubert/goth-template/utils"
)

// AppStore struct integrates IAppStore interface methods and repository management.
type AppStore struct {
	mu                           sync.Mutex
	Database                     *SqlxDbStore
	SessionManager               auth.ISessionManager
	UserRepository               *repositories.SQLUserRepository
	RoleRepository               *repositories.SQLRoleRepository
	RepositoryMetadataRepository *repositories.SQLRepositoryMetadataRepository
	Repositories                 map[string]interface{}
	RepoTypeMeta                 map[string]func(*sqlx.DB) interface{}
}

// NewAppStore initializes a new AppStore.
func NewAppStore(database *SqlxDbStore, sessionManager auth.ISessionManager) *AppStore {
	store := &AppStore{
		Database:                     database,
		SessionManager:               sessionManager,
		UserRepository:               repositories.NewSQLUserRepository(database.db),
		RoleRepository:               repositories.NewSQLRoleRepository(database.db),
		RepositoryMetadataRepository: repositories.NewSQLRepositoryMetadataRepository(database.db),
		Repositories:                 make(map[string]interface{}),
		RepoTypeMeta:                 make(map[string]func(*sqlx.DB) interface{}),
	}
	store.Init()
	store.RegisterRepoTypeMeta("UserRepository", func(db *sqlx.DB) interface{} { return repositories.NewSQLUserRepository(db) })
	return store
}

// Init loads repository metadata into the AppStore.
func (store *AppStore) Init() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	repoMetadata, err := store.RepositoryMetadataRepository.GetAllRepositoryMetadata()
	if err != nil {
		return err
	}

	for _, metadata := range repoMetadata {
		// Use the RepoTypeMeta map to create the repository
		if createRepoFunc, exists := store.RepoTypeMeta[metadata.Location]; exists {
			repo := createRepoFunc(store.Database.db)
			store.RegisterRepoWithID(metadata.ID, repo)
		} else {
			return fmt.Errorf("unsupported repository type: %s", metadata.Location)
		}
	}
	return nil
}

// RegisterRepoTypeMeta registers a new repository type with a creation function.
func (store *AppStore) RegisterRepoTypeMeta(repoType string, createFunc func(*sqlx.DB) interface{}) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.RepoTypeMeta[repoType] = createFunc
}

// User and Role Methods
func (store *AppStore) GetUserByEmail(email string) (*models.User, error) {
	return store.UserRepository.GetUserByEmail(email)
}

func (store *AppStore) GetUserRoles(user *models.User) ([]string, error) {
	roleNames := utils.ConvertStringToRoles(user.Roles)
	var expandedRoles []string
	for _, roleName := range roleNames {
		role, err := store.RoleRepository.GetRoleByName(roleName)
		if err != nil {
			return nil, err
		}
		expandedRoles = append(expandedRoles, role.Name)
	}
	return expandedRoles, nil
}

func (store *AppStore) GetRolePermissions(role *models.Role) ([]string, error) {
	return utils.ConvertStringToPermissions(role.Permissions), nil
}

func (store *AppStore) GetRoleByName(name string) (*models.Role, error) {
	return store.RoleRepository.GetRoleByName(name)
}

// Session Methods
func (store *AppStore) GetSession(r *http.Request) (*sessions.Session, error) {
	return store.SessionManager.GetSession(r)
}

func (store *AppStore) SaveSession(session *sessions.Session, r *http.Request, w http.ResponseWriter) error {
	return store.SessionManager.SaveSession(r, w, session)
}

// RegisterRepoWithID registers a new repository with a specific ID.
func (store *AppStore) RegisterRepoWithID(id string, repo interface{}) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.Repositories[id] = repo
}

// SearchReposByDomain searches for repositories by domain.
func (store *AppStore) SearchReposByDomain(domain string) []string {
	store.mu.Lock()
	defer store.mu.Unlock()

	var repoIDs []string
	for repoID := range store.Repositories {
		parts := strings.Split(repoID, ".")
		if len(parts) > 1 && parts[1] == domain {
			repoIDs = append(repoIDs, repoID)
		}
	}
	return repoIDs
}

// GetRepoByID retrieves a repository by its ID.
func (store *AppStore) GetRepoByID(id string) (interface{}, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if repo, exists := store.Repositories[id]; exists {
		return repo, nil
	}
	return nil, fmt.Errorf("repository not registered: %s", id)
}

// GetUserRepository retrieves the user repository by its ID and type.
func (store *AppStore) GetUserRepository(repoID, repoType string, access string) (interface{}, error) {
	repoIDWithDomain := fmt.Sprintf("%s.%s.%s", repoID, repoType, access)
	return store.GetRepoByID(repoIDWithDomain)
}

// GetOrCreateUserRepository gets or creates a user-specific repository.
func (store *AppStore) GetOrCreateUserRepository(user *models.User, repoType string) (interface{}, error) {
	repoID := fmt.Sprintf("%s.%s.Private", user.RepoID, repoType)
	repo, err := store.GetRepoByID(repoID)
	if err == nil {
		return repo, nil
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// Create a new repository if it doesn't exist
	if createRepoFunc, exists := store.RepoTypeMeta[repoType]; exists {
		newRepo := createRepoFunc(store.Database.db)
		store.RegisterRepoWithID(repoID, newRepo)
		return newRepo, nil
	}
	return nil, fmt.Errorf("unsupported repository type: %s", repoType)
}

// GetUserRepositories retrieves all repositories associated with a user's repoID.
func (store *AppStore) GetUserRepositories(user *models.User) map[string]interface{} {
	store.mu.Lock()
	defer store.mu.Unlock()

	userRepoPrefix := fmt.Sprintf("%s.", user.RepoID)
	userRepos := make(map[string]interface{})
	for repoID, repo := range store.Repositories {
		if strings.HasPrefix(repoID, userRepoPrefix) {
			userRepos[repoID] = repo
		}
	}
	return userRepos
}
