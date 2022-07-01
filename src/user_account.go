package otelexample

import (
	"context"
	"time"
)

type UserAccount struct {
	// ID is the user account unique identifier.
	ID ID

	// Username is the username.
	Username string

	// User is the person who owned the account.
	User *User

	// CreatedAt is the time when user account was created.
	CreatedAt time.Time
}

// UserAccountService represents a service for managing UserAccount data.
type UserAccountService interface {
	// CreateUserAccount creates a new user account.
	CreateUserAccount(ctx context.Context, ua *UserAccount) error

	// FindUserAccounts returns a list of user accounts.
	FindUserAccounts(ctx context.Context, opts FindOptions) ([]*UserAccount, error)
}
