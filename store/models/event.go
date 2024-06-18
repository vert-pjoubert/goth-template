package models

type Event struct {
	ID            int64 `xorm:"pk autoincr"`
	Name          string
	EventType     string
	ThumbnailURL  string
	Source        string
	SourceURL     string
	Time          string
	Severity      string
	SeverityClass string
	Description   string
	Roles         []string `xorm:"json"`
}

func (e *Event) TableName() string {
	return "events"
}
