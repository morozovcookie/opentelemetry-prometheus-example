package percona

import (
	"time"
)

// ClientOption represents an option for configure Client instance.
type ClientOption interface {
	apply(client *Client)
}

type clientOptionFunc func(client *Client)

func (fn clientOptionFunc) apply(client *Client) {
	fn(client)
}

// DefaultConnMaxLifetime is the maximum amount of time a connection may be reused.
const DefaultConnMaxLifetime = time.Minute

// WithConnMaxLifetime sets up the maximum amount of time a connection may be reused.
func WithConnMaxLifetime(connMaxLifetime time.Duration) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.connMaxLifetime = connMaxLifetime
	})
}

// DefaultMaxIdleConns is the maximum number of connections in the idle connection pool.
const DefaultMaxIdleConns = 5

// WithMaxIdleConns sets up the maximum number of connections in the idle connection pool.
func WithMaxIdleConns(maxIdleConns int) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.maxIdleConns = maxIdleConns
	})
}

// DefaultMaxOpenConns is the maximum number of open connections to the database.
const DefaultMaxOpenConns = 5

// WithMaxOpenConns sets up the maximum amount of time a connection may be reused.
func WithMaxOpenConns(maxOpenConns int) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.maxOpenConns = maxOpenConns
	})
}

// DefaultConnMaxIdleTime is the maximum amount of time a connection may be idle.
const DefaultConnMaxIdleTime = 0

// WithConnMaxIdleTime sets up the maximum amount of time a connection may be idle.
func WithConnMaxIdleTime(connMaxIdleTime time.Duration) ClientOption {
	return clientOptionFunc(func(c *Client) {
		c.connMaxIdleTime = connMaxIdleTime
	})
}
