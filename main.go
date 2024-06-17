package main

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
	"github.com/vert-pjoubert/goth-template/templates"
)

func getSeverityClass(severity string) string {
	switch severity {
	case "High":
		return "high"
	default:
		return ""
	}
}

func getTheme(r *http.Request) string {
	cookie, err := r.Cookie("theme")
	if err != nil {
		return "light" // default theme
	}
	return cookie.Value
}

func renderWithLayout(w http.ResponseWriter, content templ.Component, r *http.Request) {
	theme := getTheme(r)
	layout := templates.Layout(content, theme)
	layout.Render(context.Background(), w)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthenticated(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	content := templates.Home()
	renderWithLayout(w, content, r)
}

func viewHandler(renderer ViewRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		renderer.RenderView(w, r)
	}
}

func layoutHandler(w http.ResponseWriter, r *http.Request) {
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

func changeThemeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		theme := r.FormValue("theme")
		if theme != "" {
			http.SetCookie(w, &http.Cookie{
				Name:  "theme",
				Value: theme,
				Path:  "/",
				// Set a long expiry time for persistent preference
				MaxAge: 365 * 24 * 60 * 60, // 1 year
			})
		}
		// Redirect back to the referring page to refresh the content
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
		return
	}
	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	// Redirect to the login page
	http.Redirect(w, r, "/login", http.StatusFound)
}

func main() {
	viewRenderer := &DefaultViewRenderer{}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/view", viewHandler(viewRenderer))
	http.HandleFunc("/layout", layoutHandler)
	http.HandleFunc("/change-theme", changeThemeHandler)
	http.HandleFunc("/login", getLoginHandler)
	http.HandleFunc("/logout", logoutHandler)

	http.ListenAndServe(":8080", nil)
}

var events = []templates.Event{
	{
		EventID:       "1",
		ThumbnailURL:  "event-thumbnail.jpg",
		Source:        "System A",
		Time:          "10:30 AM",
		Severity:      "High",
		SeverityClass: getSeverityClass("High"),
		Description:   "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed ac justo vel urna.",
		URL:           "/static/video.mp4",
		EventType:     "video",
	},
	// Add more events as needed
}

var servers = []templates.Server{
	{
		ServerID: "1",
		URL:      "/static/server1.png",
		Name:     "Server A",
		Type:     "video",
	},
	{
		ServerID: "2",
		URL:      "/static/server2.png",
		Name:     "Server B",
		Type:     "map",
	},
}
