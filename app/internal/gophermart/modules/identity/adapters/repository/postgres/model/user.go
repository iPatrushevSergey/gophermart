package model

import "time"

// User is the DB projection of the users table row.
type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
