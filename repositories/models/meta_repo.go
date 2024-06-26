package models

import "time"

type RepositoryMetadata struct {
	ID        string    `db:"id"`
	Location  string    `db:"location"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
