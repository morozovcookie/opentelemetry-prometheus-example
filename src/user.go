package otelexample

import (
	"time"
)

// User describes the real person.
type User struct {
	// ID is the user unique identifier.
	ID ID

	// FirstName is the user first name.
	FirstName string

	// LastName is the user last name.
	LastName string

	// CreatedAt is the time when user was created.
	CreatedAt time.Time
}

// Clone creates a deep copy of User.
func (ua *User) Clone() *User {
	return &User{
		ID:        ua.ID,
		FirstName: ua.FirstName,
		LastName:  ua.LastName,
		CreatedAt: ua.CreatedAt,
	}
}
