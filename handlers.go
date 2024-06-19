package main

import (
	"net/http"

	"github.com/vert-pjoubert/goth-template/auth"
	"github.com/vert-pjoubert/goth-template/templates"
)

// Handlers struct uses the authenticator and the renderer
type Handlers struct {
	Auth         IAuthenticator
	Renderer     *TemplRenderer
	ViewRenderer *ViewRenderer
	Session      auth.ISessionManager
}

// NewHandlers initializes the Handlers struct
func NewHandlers(auth IAuthenticator, renderer *TemplRenderer, viewRenderer *ViewRenderer, session auth.ISessionManager) *Handlers {
	return &Handlers{Auth: auth, Renderer: renderer, ViewRenderer: viewRenderer, Session: session}
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

func (h *Handlers) ViewHandler(w http.ResponseWriter, r *http.Request) {
	authenticated, err := h.Auth.IsAuthenticated(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	session, _ := h.Session.GetSession(r)
	token, ok := session.Values["token"].(string)
	if !ok || token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
