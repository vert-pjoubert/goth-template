package main

import (
	"fmt"
	"net/http"
	"github.com/vert-pjoubert/goth-template/templates"
)

type ViewRenderer interface {
	RenderView(w http.ResponseWriter, r *http.Request)
}

type DefaultViewRenderer struct{}

func (vr *DefaultViewRenderer) RenderView(w http.ResponseWriter, r *http.Request) {
	view := r.URL.Query().Get("view")
	switch view {
	case "settings":
		content := templates.Settings()
		renderWithLayout(w, content, r)
	case "servers":
		content := templates.ServersList(servers)
		renderWithLayout(w, content, r)
	case "events":
		content := templates.EventsList(events)
		renderWithLayout(w, content, r)
	case "server":
		serverID := r.URL.Query().Get("id")
		var server templates.Server
		for _, s := range servers {
			if s.ServerID == serverID {
				server = s
				break
			}
		}
		switch server.Type {
		case "video":
			content := templates.VideoServer(server)
			renderWithLayout(w, content, r)
		case "map":
			content := templates.MapServer(server)
			renderWithLayout(w, content, r)
		default:
			fmt.Fprintf(w, "<div>Unsupported server type</div>")
		}
	case "event":
		eventID := r.URL.Query().Get("id")
		var event templates.Event
		for _, e := range events {
			if e.EventID == eventID {
				event = e
				break
			}
		}
		switch event.EventType {
		case "video":
			content := templates.VideoEvent(event)
			renderWithLayout(w, content, r)
		default:
			fmt.Fprintf(w, "<div>Unsupported event type</div>")
		}
	default:
		http.Error(w, "Invalid view", http.StatusBadRequest)
	}
}
