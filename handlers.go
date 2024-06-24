package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/vert-pjoubert/goth-template/store/models"
	"github.com/vert-pjoubert/goth-template/templates"
)

// Handlers struct uses the authenticator and the renderer
type Handlers struct {
	Auth         IAuthenticator
	Renderer     *TemplRenderer
	ViewRenderer *ViewRenderer
	Session      ISessionManager
	baseURL      string
}

func NewHandlers(auth IAuthenticator, renderer *TemplRenderer, viewRenderer *ViewRenderer, session ISessionManager) *Handlers {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &Handlers{Auth: auth, Renderer: renderer, ViewRenderer: viewRenderer, Session: session, baseURL: baseURL}
}

func (h *Handlers) IndexHandler(w http.ResponseWriter, r *http.Request) {
	authenticated, err := h.Auth.IsAuthenticated(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	content := templates.Home()
	h.Renderer.RenderWithLayout(w, content, r)
}

func (h *Handlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	h.Auth.LoginHandler(w, r)
}

func (h *Handlers) LayoutHandler(w http.ResponseWriter, r *http.Request) {
	part := r.URL.Query().Get("part")
	switch part {
	case "header":
		templates.Header().Render(r.Context(), w)
	case "footer":
		templates.Footer().Render(r.Context(), w)
	case "sidebar":
		templates.Sidebar().Render(r.Context(), w)
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

func (h *Handlers) SettingsViewHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	content := templates.Settings()
	content.Render(context.Background(), w)
}

func (h *Handlers) ServersViewHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	h.ViewRenderer.RenderAccessibleServers(w, r, user)
}

func (h *Handlers) EventsViewHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	h.ViewRenderer.RenderAccessibleEvents(w, r, user)
}

func (h *Handlers) ProtectedViewHandler(w http.ResponseWriter, r *http.Request, user *models.User) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

func secureFileServer(root http.FileSystem) http.Handler {
	allowedExtensions := map[string]bool{
		".css":  true,
		".svg":  true,
		".png":  true,
		".jpg":  true,
		".jpeg": true,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Prevent directory traversal by cleaning the URL path
		cleanPath := filepath.Clean(r.URL.Path)
		if strings.Contains(cleanPath, "..") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check for allowed file extensions
		ext := strings.ToLower(filepath.Ext(cleanPath))
		if !allowedExtensions[ext] {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Ensure the file is within the static directory
		if !strings.HasPrefix(cleanPath, "/static/") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Serve the file if it exists
		http.FileServer(root).ServeHTTP(w, r)
	})
}
