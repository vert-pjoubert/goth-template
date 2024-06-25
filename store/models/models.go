package models

import (
	"time"
)

// Event represents an event with various attributes and roles
type Event struct {
	ID            int64  `db:"id"`
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
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Permissions string `db:"permissions"` // Changed to string, seperated by a semi-colon ";"
}

// TableName returns the table name for the Role model
func (r *Role) TableName() string {
	return "roles"
}

// Server represents a server with various attributes and roles
type Server struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Type         string    `db:"type"`
	URL          string    `db:"url"`
	Roles        string    `db:"roles"` // Changed to string, seperated by a semi-colon ";"
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
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
	ID        int64     `db:"id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Roles     string    `db:"roles"` //seperated by ;
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// TableName returns the table name for the User model
func (u *User) TableName() string {
	return "users"
}

// Permissions represents the various permissions associated with a role
type Permissions struct {
	Get    bool   `json:"get"`
	Post   bool   `json:"post"`
	Put    bool   `json:"put"`
	Patch  bool   `json:"patch"`
	Delete bool   `json:"delete"`
	Other  string `json:"other,omitempty"`
}

func (p *Permissions) TableName() string {
	return "permissions"
}
