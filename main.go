package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
)

type Event struct {
	EventID       string
	ThumbnailURL  string
	Source        string
	Time          string
	Severity      string
	SeverityClass string
	Description   string
	URL           string
	EventType     string
}

type Server struct {
	ServerID string
	URL      string
	Name     string
	Type     string
}

var events = []Event{
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

var servers = []Server{
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
	layout := Layout(content, theme)
	layout.Render(context.Background(), w)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	switch view {
	case "settings":
		content := Settings()
		renderWithLayout(w, content, r)
	case "servers":
		content := ServersList(servers)
		renderWithLayout(w, content, r)
	case "events":
		content := EventsList(events)
		renderWithLayout(w, content, r)
	case "server":
		serverID := r.URL.Query().Get("id")
		var server Server
		for _, s := range servers {
			if s.ServerID == serverID {
				server = s
				break
			}
		}
		switch server.Type {
		case "video":
			content := VideoServer(server)
			renderWithLayout(w, content, r)
		case "map":
			content := MapServer(server)
			renderWithLayout(w, content, r)
		default:
			fmt.Fprintf(w, "<div>Unsupported server type</div>")
		}
	case "event":
		eventID := r.URL.Query().Get("id")
		var event Event
		for _, e := range events {
			if e.EventID == eventID {
				event = e
				break
			}
		}
		switch event.EventType {
		case "video":
			content := VideoEvent(event)
			renderWithLayout(w, content, r)
		default:
			fmt.Fprintf(w, "<div>Unsupported event type</div>")
		}
	default:
		http.Error(w, "Invalid view", http.StatusBadRequest)
	}
}

func layoutHandler(w http.ResponseWriter, r *http.Request) {
	part := r.URL.Query().Get("part")
	switch part {
	case "header":
		header := Header()
		header.Render(context.Background(), w)
	case "footer":
		footer := Footer()
		footer.Render(context.Background(), w)
	case "sidebar":
		sidebar := Sidebar()
		sidebar.Render(context.Background(), w)
	default:
		http.Error(w, "Invalid part", http.StatusBadRequest)
	}
}

func changeThemeHandler(w http.ResponseWriter, r *http.Request) {
	theme := r.URL.Query().Get("theme")
	if theme != "" {
		http.SetCookie(w, &http.Cookie{
			Name:  "theme",
			Value: theme,
			Path:  "/",
			// Set a long expiry time for persistent preference
			MaxAge: 365 * 24 * 60 * 60, // 1 year
		})
	}
	// Load the CSS for the new theme
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<link rel="stylesheet" href="/static/styles-%s.css">`, theme)
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/view", viewHandler)
	http.HandleFunc("/layout", layoutHandler)
	http.HandleFunc("/change-theme", changeThemeHandler)
	http.HandleFunc("/login", getLoginHandler)

	http.ListenAndServe(":8080", nil)
}
