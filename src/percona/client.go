package percona

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	// DriverName is the name of driver which will be used for work with sql data storage.
	DriverName = "mysql"

	// ConnectTimeout is the maximum time for waiting until connect operation will be finished.
	ConnectTimeout = time.Second

	// PingTimeout is the maximum time for waiting until ping operation will be finished.
	PingTimeout = time.Millisecond * 100
)

// DBInfo represents a service for getting information about database.
type DBInfo interface {
	// DBName returns name of database which client are connected.
	DBName() string

	// DBUser returns name of user which connected to the database.
	DBUser() string
}

// Preparer represents a service that can create a prepared statement.
type Preparer interface {
	DBInfo

	// PrepareContext creates a prepared statement for later queries or executions.
	PrepareContext(ctx context.Context, query string) (Stmt, error)
}

// TxBeginner represents a service that can start a transaction.
type TxBeginner interface {
	DBInfo

	// BeginTx starts a transaction.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

var (
	_ TxBeginner = (*Client)(nil)
	_ Preparer   = (*Client)(nil)
)

// Client represents an object for basic manipulation with Percona MySQL Database System.
type Client struct {
	db  *sql.DB
	dsn string

	connMaxLifetime time.Duration
	maxIdleConns    int
	connMaxIdleTime time.Duration
	maxOpenConns    int

	dbName string
	dbUser string
}

// NewClient returns a new Client instance.
func NewClient(dsn string, opts ...ClientOption) *Client {
	client := &Client{
		db:  nil,
		dsn: dsn,

		connMaxLifetime: DefaultConnMaxLifetime,
		maxIdleConns:    DefaultMaxIdleConns,
		connMaxIdleTime: DefaultConnMaxIdleTime,
		maxOpenConns:    DefaultMaxOpenConns,

		dbName: "",
		dbUser: "",
	}

	for _, opt := range opts {
		opt.apply(client)
	}

	return client
}

// DBName returns name of database which client are connected.
func (c *Client) DBName() string {
	return c.dbName
}

// DBUser returns name of user which connected to the database.
func (c *Client) DBUser() string {
	return c.dbUser
}

// Connect connects to a database.
func (c *Client) Connect(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(ConnectTimeout))
	defer cancel()

	config, err := mysql.ParseDSN(c.dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	c.dbName, c.dbUser = config.DBName, config.User

	if c.db, err = sql.Open(DriverName, c.dsn); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	c.db.SetConnMaxLifetime(c.connMaxLifetime)
	c.db.SetMaxIdleConns(c.maxIdleConns)
	c.db.SetConnMaxIdleTime(c.connMaxIdleTime)
	c.db.SetMaxOpenConns(c.maxOpenConns)

	if err = c.Ping(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	return nil
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(PingTimeout))
	defer cancel()

	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}

// BeginTx starts a transaction.
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	sqlTx, err := c.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}

	return &tx{
		sqlTx: sqlTx,
	}, nil
}

// PrepareContext creates a prepared statement for later queries or executions.
func (c *Client) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	sqlStmt, err := c.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}

	return &stmt{
		sqlStmt: sqlStmt,
	}, nil
}
