package main

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/repositories/models"
)

const pageSize = 25
const cacheDuration = time.Minute

type ViewHandler func(http.ResponseWriter, *http.Request, *models.User)

type ViewMetadata struct {
	Handler             ViewHandler
	RequiredRoles       []string
	RequiredPermissions []string
}

type ViewRenderer struct {
	AppStore IAppStore
	Views    map[string]ViewMetadata
	cache    *Cache
}

type Cache struct {
	mu      sync.Mutex
	entries map[string]CacheEntry
}

type CacheEntry struct {
	data      interface{}
	timestamp time.Time
}

func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
	}
}

// Set adds an item to the cache with a timestamp
func (c *Cache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = CacheEntry{data, time.Now()}
}

// Get retrieves an item from the cache if it exists and is not expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, found := c.entries[key]
	if !found || time.Since(entry.timestamp) > cacheDuration {
		if found {
			delete(c.entries, key)
		}
		return nil, false
	}
	return entry.data, true
}

func NewViewRenderer(appStore IAppStore) *ViewRenderer {
	return &ViewRenderer{
		AppStore: appStore,
		Views:    make(map[string]ViewMetadata),
		cache:    NewCache(),
	}
}

func (vr *ViewRenderer) RegisterView(name string, handler ViewHandler, requiredRoles []string, requiredPermissions []string) {
	vr.Views[name] = ViewMetadata{
		Handler:             handler,
		RequiredRoles:       requiredRoles,
		RequiredPermissions: requiredPermissions,
	}
}

func (vr *ViewRenderer) RenderView(w http.ResponseWriter, r *http.Request) {
	session, err := vr.AppStore.GetSession(r)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	userEmail, ok := session.Values["user"].(string)
	if !ok {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	user, err := vr.AppStore.GetUserByEmail(userEmail)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	view := r.URL.Query().Get("view")
	viewMetadata, ok := vr.Views[view]
	if !ok {
		http.Error(w, "Invalid view", http.StatusBadRequest)
		return
	}

	// Check if the user has the required roles and permissions
	userRoles, err := vr.AppStore.GetUserRoles(user)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	rolePermissions := make(map[string][]string)
	for _, role := range userRoles {
		roleModel, err := vr.AppStore.GetRoleByName(role)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		rolePermissions[role], err = vr.AppStore.GetRolePermissions(roleModel)
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	if !auth.HasRequiredRoles(user, viewMetadata.RequiredRoles) || !auth.HasRequiredPermissions(user, viewMetadata.RequiredPermissions) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	viewMetadata.Handler(w, r, user)
}

// getPageNumber extracts the page number from the request
func getPageNumber(r *http.Request) int {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		return 1
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 1
	}
	return page
}
