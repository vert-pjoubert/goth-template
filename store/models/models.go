package models

import (
	"time"
)

// Event represents an event with various attributes and roles
type Event struct {
	ID            int64  `xorm:"pk autoincr" db:"id"`
	Name          string `db:"name"`
	EventType     string `db:"event_type"`
	ThumbnailURL  string `db:"thumbnail_url"`
	Source        string `db:"source"`
	SourceURL     string `db:"source_url"`
	Time          string `db:"time"`
	Severity      string `db:"severity"`
	SeverityClass string `db:"severity_class"`
	Description   string `db:"description"`
	Roles         string `db:"roles"` // Changed to string, seperated by a semi-colon ";"
}

// TableName returns the table name for the Event model
func (e *Event) TableName() string {
	return "events"
}

// Role represents a role with permissions
type Role struct {
	ID          int64  `xorm:"pk autoincr" db:"id"`
	Name        string `xorm:"unique" db:"name"`
	Description string `db:"description"`
	Permissions string `xorm:"json" db:"permissions"` // Changed to string, seperated by a semi-colon ";"
}

// TableName returns the table name for the Role model
func (r *Role) TableName() string {
	return "roles"
}

// Server represents a server with various attributes and roles
type Server struct {
	ID           int64     `xorm:"pk autoincr" db:"id"`
	Name         string    `db:"name"`
	Type         string    `db:"type"`
	URL          string    `db:"url"`
	Roles        string    `db:"roles"` // Changed to string, seperated by a semi-colon ";"
	CreatedAt    time.Time `xorm:"created" db:"created_at"`
	UpdatedAt    time.Time `xorm:"updated" db:"updated_at"`
	Description  string    `db:"description"`
	Status       string    `db:"status"`
	IPAddress    string    `db:"ip_address"`
	Location     string    `db:"location"`
	PublicKey    string    `db:"public_key"`
	MAC          string    `db:"mac"`
	Model        string    `db:"model"`
	Manufacturer string    `db:"manufacturer"`
}

// TableName returns the table name for the Server model
func (s *Server) TableName() string {
	return "servers"
}

// User represents a user with role and timestamps
type User struct {
	ID        int64     `xorm:"pk autoincr" db:"id"`
	Email     string    `xorm:"unique" db:"email"`
	Name      string    `db:"name"`
	RoleID    int64     `xorm:"index" db:"role_id"`
	Role      Role      `xorm:"role" db:"role"`
	CreatedAt time.Time `xorm:"created" db:"created_at"`
	UpdatedAt time.Time `xorm:"updated" db:"updated_at"`
}

// TableName returns the table name for the User model
func (u *User) TableName() string {
	return "users"
}
