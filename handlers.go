package main

import (
	"net/http"

	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/templates"
)

// Handlers struct uses the authenticator and the renderer
type Handlers struct {
	Auth         auth.Authenticator
	Renderer     *TemplRenderer
	ViewRenderer *ViewRenderer
}

// NewHandlers initializes the Handlers struct
func NewHandlers(auth auth.Authenticator, renderer *TemplRenderer, viewRenderer *ViewRenderer) *Handlers {
	return &Handlers{Auth: auth, Renderer: renderer, ViewRenderer: viewRenderer}
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

func (h *Handlers) ViewHandler(w http.ResponseWriter, r *http.Request) {
	if !h.Auth.IsAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	h.ViewRenderer.RenderView(w, r)
}

// Helper functions
func getTheme(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil {
		return "light" // default theme
	}
	return cookie.Value
}
