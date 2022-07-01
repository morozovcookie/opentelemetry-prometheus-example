package percona

import (
	"context"
	"database/sql"
)

// Stmt is a prepared statement.
type Stmt interface {
	// ExecContext executes a prepared statement with the given arguments and
	// returns a Result summarizing the effect of the statement.
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)

	// Close closes the statement.
	Close(ctx context.Context) error
}

var _ Stmt = (*stmt)(nil)

// stmt is a prepared statement.
type stmt struct {
	sqlStmt *sql.Stmt
}

// ExecContext executes a prepared statement with the given arguments and
// returns a Result summarizing the effect of the statement.
func (stmt *stmt) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	return stmt.sqlStmt.ExecContext(ctx, args...)
}

func (stmt *stmt) Close(_ context.Context) error {
	return stmt.sqlStmt.Close()
}
