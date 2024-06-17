package main

import (
	"context"
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
		content.Render(context.Background(), w)
	case "servers":
		content := templates.ServersList(servers)
		content.Render(context.Background(), w)
	case "events":
		content := templates.EventsList(events)
		content.Render(context.Background(), w)
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
			content.Render(context.Background(), w)
		case "map":
			content := templates.MapServer(server)
			content.Render(context.Background(), w)
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
			content.Render(context.Background(), w)
		default:
			fmt.Fprintf(w, "<div>Unsupported event type</div>")
		}
	default:
		http.Error(w, "Invalid view", http.StatusBadRequest)
	}
}
