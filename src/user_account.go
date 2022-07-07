package otelexample

import (
	"context"
	"time"
)

// UserAccount is the user account in the system.
type UserAccount struct {
	// ID is the user account unique identifier.
	ID ID

	// Username is the user account name.
	Username string

	// User is the person who owned the account.
	User *User

	// CreatedAt is the time when user account was created.
	CreatedAt time.Time
}

// Clone creates a deep copy of UserAccount.
func (ua *UserAccount) Clone() *UserAccount {
	return &UserAccount{
		ID:        ua.ID,
		Username:  ua.Username,
		User:      ua.User.Clone(),
		CreatedAt: ua.CreatedAt,
	}
}

// UserAccountService represents a service for managing UserAccount data.
type UserAccountService interface {
	// CreateUserAccount creates a new user account.
	CreateUserAccount(ctx context.Context, ua *UserAccount) error

	// FindUserAccounts returns a list of user accounts.
	FindUserAccounts(ctx context.Context, opts FindOptions) (*FindUserAccountsResult, error)

	// FindUserAccountByID returns user account by unique identifier.
	FindUserAccountByID(ctx context.Context, id ID) (*UserAccount, error)
}

// FindUserAccountsResult is the result of searching user accounts.
type FindUserAccountsResult struct {
	// HasNext is the flag of existent the next page of data.
	HasNext bool

	// Total is the count of records that matches of request
	// without pagination parameters.
	Total uint64

	// Options is the restrictions which would apply to the search.
	Options FindOptions

	// Data is the search result.
	Data []*UserAccount
}
