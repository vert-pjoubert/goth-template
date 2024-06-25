package main

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/store"
	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
	"github.com/vert-pjoubert/goth-template/utils"
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

	user, err := vr.AppStore.GetUserWithRoleByEmail(userEmail)
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
	if !auth.HasRequiredRoles(user, viewMetadata.RequiredRoles) || !auth.HasRequiredPermissions(user, viewMetadata.RequiredPermissions) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	viewMetadata.Handler(w, r, user)
}

// RenderAccessibleServers renders a list of accessible servers inside the infinite scroll template
func (vr *ViewRenderer) ServersViewRender(w http.ResponseWriter, r *http.Request, user *models.User) {
	page := getPageNumber(r)
	cacheKey := "servers_" + user.Email + "_" + strconv.Itoa(page)

	if cached, found := vr.cache.Get(cacheKey); found {
		templateServers := cached.([]templates.Server)
		nextPage := strconv.Itoa(page + 1)
		content := templates.ServersInfiniteScroll(nextPage, templateServers)
		content.Render(context.Background(), w)
		return
	}

	servers, err := vr.AppStore.GetServers()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accessibleServers := store.FilterByUserRoles(servers, user, func(server models.Server) string {
		return server.Roles
	})

	paginatedServers := utils.Paginate(accessibleServers, page, pageSize)

	templateServers := make([]templates.Server, len(paginatedServers))
	for i, server := range paginatedServers {
		templateServers[i] = templates.NewServer(server)
	}

	vr.cache.Set(cacheKey, templateServers)

	nextPageStr := strconv.Itoa(page)
	content := templates.ServersInfiniteScroll(nextPageStr, templateServers)
	content.Render(context.Background(), w)
}

// RenderAccessibleEvents renders a list of accessible events inside the infinite scroll template
func (vr *ViewRenderer) EventsViewRender(w http.ResponseWriter, r *http.Request, user *models.User) {
	page := getPageNumber(r)
	cacheKey := "events_" + user.Email + "_" + strconv.Itoa(page)

	if cached, found := vr.cache.Get(cacheKey); found {
		templateEvents := cached.([]templates.Event)
		nextPage := strconv.Itoa(page + 1)
		content := templates.EventsInfiniteScroll(nextPage, templateEvents)
		content.Render(context.Background(), w)
		return
	}

	events, err := vr.AppStore.GetEvents()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	accessibleEvents := store.FilterByUserRoles(events, user, func(event models.Event) string {
		return event.Roles
	})

	paginatedEvents := utils.Paginate(accessibleEvents, page, pageSize)

	templateEvents := make([]templates.Event, len(paginatedEvents))
	for i, event := range paginatedEvents {
		templateEvents[i] = templates.NewEvent(event)
	}

	vr.cache.Set(cacheKey, templateEvents)

	nextPageStr := strconv.Itoa(page)
	content := templates.EventsInfiniteScroll(nextPageStr, templateEvents)
	content.Render(context.Background(), w)
}

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
