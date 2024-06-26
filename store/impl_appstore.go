package store

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gorilla/sessions"
	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/repositories"
	"github.com/vert-pjoubert/goth-template/repositories/models"
	"github.com/vert-pjoubert/goth-template/utils"
)

// AppStore struct integrates IAppStore interface methods and repository management.
type AppStore struct {
	Database       *SqlxDbStore
	SessionManager auth.ISessionManager
	UserRepository *repositories.SQLUserRepository
	RoleRepository *repositories.SQLRoleRepository
	Repositories   map[reflect.Type]interface{}
}

// NewAppStore initializes a new AppStore.
func NewAppStore(database *SqlxDbStore, sessionManager auth.ISessionManager) *AppStore {
	store := &AppStore{
		Database:       database,
		SessionManager: sessionManager,
		UserRepository: repositories.NewSQLUserRepository(database.db),
		RoleRepository: repositories.NewSQLRoleRepository(database.db),
		Repositories:   make(map[reflect.Type]interface{}),
	}
	return store
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

// RegisterRepo adds a new repository to the store.
func (store *AppStore) RegisterRepo(repo interface{}) {
	repoType := reflect.TypeOf(repo).Elem()
	store.Repositories[repoType] = repo
}

// GetRepo retrieves a repository by type.
func (store *AppStore) GetRepo(repoType reflect.Type) (interface{}, error) {
	if repo, exists := store.Repositories[repoType]; exists {
		return repo, nil
	}
	return nil, fmt.Errorf("repository not registered: %s", repoType.Name())
}

// SearchReposByDomain searches for repositories by domain.
func (store *AppStore) SearchReposByDomain(domain string) []string {
	var repoIDs []string
	for repoID := range store.Repositories {
		parts := strings.Split(repoID, ".")
		if len(parts) > 1 && parts[1] == domain {
			repoIDs = append(repoIDs, repoID)
		}
	}
	return repoIDs
}

// GetUserRepository retrieves the user repository by its ID and type.
func (store *AppStore) GetUserRepository(repoID, repoType string) (interface{}, error) {
	repoIDWithDomain := fmt.Sprintf("%s.%s", repoID, repoType)
	repo, err := store.GetRepoByID(repoIDWithDomain)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

// GetOrCreateUserRepository gets or creates a user-specific repository.
func (store *AppStore) GetOrCreateUserRepository(user *models.User, repoType string) (interface{}, error) {
	repoID := fmt.Sprintf("%s.%s.Private", user.RepoID, repoType)
	repo, err := store.GetRepoByID(repoID)
	if err == nil {
		return repo, nil
	}

	// Create a new repository if it doesn't exist
	var newRepo interface{}
	switch repoType {
	case "profileRepository":
		newRepo = repositories.NewSQLUserRepository(store.Database.db)
	// Add repository types as needed
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", repoType)
	}

	store.RegisterRepoWithID(repoID, newRepo)
	return newRepo, nil
}

// GetUserRepositories retrieves all repositories associated with a user's repoID.
func (store *AppStore) GetUserRepositories(user *models.User) map[string]interface{} {
	userRepoPrefix := fmt.Sprintf("%s.", user.RepoID)
	userRepos := make(map[string]interface{})
	for repoID, repo := range store.Repositories {
		if strings.HasPrefix(repoID, userRepoPrefix) {
			userRepos[repoID] = repo
		}
	}
	return userRepos
}
