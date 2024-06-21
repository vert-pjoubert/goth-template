package models

import (
	"time"
)

type Event struct {
	ID            int64    `xorm:"pk autoincr" db:"id"`
	Name          string   `db:"name"`
	EventType     string   `db:"event_type"`
	ThumbnailURL  string   `db:"thumbnail_url"`
	Source        string   `db:"source"`
	SourceURL     string   `db:"source_url"`
	Time          string   `db:"time"`
	Severity      string   `db:"severity"`
	SeverityClass string   `db:"severity_class"`
	Description   string   `db:"description"`
	Roles         []string `xorm:"json" db:"roles"`
}

func (e *Event) TableName() string {
	return "events"
}

type Role struct {
	ID          int64    `xorm:"pk autoincr" db:"id"`
	Name        string   `xorm:"unique" db:"name"`
	Description string   `db:"description"`
	Permissions []string `xorm:"json" db:"permissions"`
}

func (r *Role) TableName() string {
	return "roles"
}

type Server struct {
	ID    int64    `xorm:"pk autoincr" db:"id"`
	Name  string   `db:"name"`
	Type  string   `db:"type"`
	URL   string   `db:"url"`
	Roles []string `xorm:"json" db:"roles"`
}

func (s *Server) TableName() string {
	return "servers"
}

type User struct {
	ID        int64     `xorm:"pk autoincr" db:"id"`
	Email     string    `xorm:"unique" db:"email"`
	Name      string    `db:"name"`
	RoleID    int64     `xorm:"index" db:"role_id"`
	Role      Role      `xorm:"-" db:"-"`
	CreatedAt time.Time `xorm:"created" db:"created_at"`
	UpdatedAt time.Time `xorm:"updated" db:"updated_at"`
}

func (u *User) TableName() string {
	return "users"
}
