package main

import (
	"context"
	"net/http"

	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/templates"
)

// Public Servers and Events for use in other packages
var Servers = []templates.Server{
	{ServerID: "1", Type: "video", Name: "Server A"},
	{ServerID: "2", Type: "map", Name: "Server B"},
}

var Events = []templates.Event{
	{EventID: "1", EventType: "video", Source: "Camera 1"},
	{EventID: "2", EventType: "map", Source: "Sensor A"},
}

// Handlers struct uses the authenticator and the renderer
type Handlers struct {
	Auth     auth.Authenticator
	Renderer *TemplRenderer
}

// NewHandlers initializes the Handlers struct
func NewHandlers(auth auth.Authenticator, renderer *TemplRenderer) *Handlers {
	return &Handlers{Auth: auth, Renderer: renderer}
}

func (h *Handlers) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	content := templates.Home()
	h.Renderer.RenderWithLayout(w, content, r)
}

func (h *Handlers) LayoutHandler(w http.ResponseWriter, r *http.Request) {
	part := r.URL.Query().Get("part")
	ctx := context.Background()
	switch part {
	case "header":
		templates.Header().Render(ctx, w)
	case "footer":
		templates.Footer().Render(ctx, w)
	case "sidebar":
		templates.Sidebar().Render(ctx, w)
	default:
		http.Error(w, "Invalid part", http.StatusBadRequest)
	}
}

func (h *Handlers) ChangeThemeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		theme := r.FormValue("theme")
		if theme != "" {
			http.SetCookie(w, &http.Cookie{
				Name:   "theme",
				Value:  theme,
				Path:   "/",
				MaxAge: 365 * 24 * 60 * 60, // 1 year
			})
		}
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func (h *Handlers) ViewHandler(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	view := r.URL.Query().Get("view")
	switch view {
	case "settings":
		content := templates.Settings()
		h.Renderer.RenderWithLayout(w, content, r)
	case "servers":
		content := templates.ServersList(Servers)
		h.Renderer.RenderWithLayout(w, content, r)
	case "events":
		content := templates.EventsList(Events)
		h.Renderer.RenderWithLayout(w, content, r)
	default:
		http.Error(w, "Invalid view", http.StatusBadRequest)
	}
}

// Helper functions
func getTheme(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil {
		return "light" // default theme
	}
	return cookie.Value
}
