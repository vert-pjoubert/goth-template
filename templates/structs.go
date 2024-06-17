package templates
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