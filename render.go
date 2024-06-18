package main

import (
	"context"
	"net/http"

	"github.com/a-h/templ"

	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
)

// ###################################################
// layout renderer
type TemplRenderer struct{}

func NewTemplRenderer() *TemplRenderer {
	return &TemplRenderer{}
}

func (r *TemplRenderer) RenderWithLayout(w http.ResponseWriter, content templ.Component, req *http.Request) {
	theme := getTheme(req)
	layout := templates.Layout(content, theme)
	layout.Render(context.Background(), w)
}

// ###################################################
// view renderer
type ViewRenderer struct {
	AppStore AppStore
}

func NewViewRenderer(appStore AppStore) *ViewRenderer {
	return &ViewRenderer{AppStore: appStore}
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
	switch view {
	case "settings":
		content := templates.Settings()
		content.Render(context.Background(), w)
	case "servers":
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
	case "events":
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
	default:
		http.Error(w, "Invalid view", http.StatusBadRequest)
	}
}

func (vr *ViewRenderer) getAccessibleServers(userRole string) ([]models.Server, error) {
	var servers []models.Server
	err := vr.AppStore.GetServers(&servers)
	if err != nil {
		return nil, err
	}

	var accessibleServers []models.Server
	for _, server := range servers {
		if contains(server.Roles, userRole) {
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
		if contains(event.Roles, userRole) {
			accessibleEvents = append(accessibleEvents, event)
		}
	}
	return accessibleEvents, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
