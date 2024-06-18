package templates

import (
	"strconv"

	"github.com/vert-pjoubert/goth-template/store/models"
)

type Server struct {
	ServerID string
	Name     string
	Type     string
	Roles    []string
}

func NewServer(server models.Server) Server {
	return Server{
		ServerID: strconv.FormatInt(server.ID, 10),
		Name:     server.Name,
		Type:     server.Type,
		Roles:    server.Roles,
	}
}

type Event struct {
	EventID       string
	Name          string
	EventType     string
	ThumbnailURL  string
	Source        string
	SourceURL     string
	Time          string
	Severity      string
	SeverityClass string
	Description   string
	Roles         []string
}

func NewEvent(event models.Event) Event {
	return Event{
		EventID:       strconv.FormatInt(event.ID, 10),
		Name:          event.Name,
		EventType:     event.EventType,
		ThumbnailURL:  event.ThumbnailURL,
		Source:        event.Source,
		SourceURL:     event.SourceURL,
		Time:          event.Time,
		Severity:      event.Severity,
		SeverityClass: event.SeverityClass,
		Description:   event.Description,
		Roles:         event.Roles,
	}
}
