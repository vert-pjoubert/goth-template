package main

import (
	"context"
	"net/http"

	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
	"github.com/vert-pjoubert/goth-template/utils"
)

type ViewHandler func(http.ResponseWriter, *http.Request, *models.User)

type ViewMetadata struct {
	Handler             ViewHandler
	RequiredRoles       []string
	RequiredPermissions []string
}

type ViewRenderer struct {
	AppStore IAppStore
	Views    map[string]ViewMetadata
}

func NewViewRenderer(appStore IAppStore) *ViewRenderer {
	return &ViewRenderer{
		AppStore: appStore,
		Views:    make(map[string]ViewMetadata),
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
	if !vr.userHasRequiredRoles(user, viewMetadata.RequiredRoles) || !vr.userHasRequiredPermissions(user, viewMetadata.RequiredPermissions) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	viewMetadata.Handler(w, r, user)
}

func (vr *ViewRenderer) userHasRequiredRoles(user *models.User, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}

	userRole := user.Role.Name
	for _, role := range requiredRoles {
		if role == userRole {
			return true
		}
	}

	return false
}

func (vr *ViewRenderer) userHasRequiredPermissions(user *models.User, requiredPermissions []string) bool {
	if len(requiredPermissions) == 0 {
		return true
	}

	userPermissions := auth.ConvertStringToPermissions(user.Role.Permissions)
	for _, perm := range requiredPermissions {
		if !contains(userPermissions, perm) {
			return false
		}
	}

	return true
}

func (vr *ViewRenderer) RenderAccessibleServers(w http.ResponseWriter, r *http.Request, user *models.User) {
	servers, err := vr.getAccessibleServers(user.Role.Name)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	templateServers := make([]templates.Server, len(servers))
	for i, server := range servers {
		templateServers[i] = templates.NewServer(server)
	}
	content := templates.ServersList(templateServers)
	content.Render(context.Background(), w)
}

func (vr *ViewRenderer) RenderAccessibleEvents(w http.ResponseWriter, r *http.Request, user *models.User) {
	events, err := vr.getAccessibleEvents(user.Role.Name)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	templateEvents := make([]templates.Event, len(events))
	for i, event := range events {
		templateEvents[i] = templates.NewEvent(event)
	}
	content := templates.EventsList(templateEvents)
	content.Render(context.Background(), w)
}

func (vr *ViewRenderer) getAccessibleServers(userRole string) ([]models.Server, error) {
	var servers []models.Server
	err := vr.AppStore.GetServers(&servers)
	if err != nil {
		return nil, err
	}

	var accessibleServers []models.Server
	for _, server := range servers {
		roles := utils.ConvertStringToRoles(server.Roles) // Parse roles string into a list
		if contains(roles, userRole) {
			accessibleServers = append(accessibleServers, server)
		}
	}
	return accessibleServers, nil
}

func (vr *ViewRenderer) getAccessibleEvents(userRole string) ([]models.Event, error) {
	var events []models.Event
	err := vr.AppStore.GetEvents(&events)
	if err != nil {
		return nil, err
	}

	var accessibleEvents []models.Event
	for _, event := range events {
		roles := utils.ConvertStringToRoles(event.Roles) // Parse roles string into a list
		if contains(roles, userRole) {
			accessibleEvents = append(accessibleEvents, event)
		}
	}
	return accessibleEvents, nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
