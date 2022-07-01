package percona

import (
	"context"
	"database/sql"
	"fmt"
)

// Tx is an in-progress database transaction.
type Tx interface {
	// PrepareContext creates a prepared statement for later queries or executions.
	PrepareContext(ctx context.Context, query string) (Stmt, error)

	// Commit commits the transaction.
	Commit() error

	// Rollback aborts the transaction.
	Rollback() error
}

var _ Tx = (*tx)(nil)

// tx is an in-progress database transaction.
type tx struct {
	sqlTx *sql.Tx
}

// PrepareContext creates a prepared statement for use within a transaction.
func (tx *tx) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	sqlStmt, err := tx.sqlTx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("prepare: %w", err)
	}

	return &stmt{
		sqlStmt: sqlStmt,
	}, nil
}

// Commit commits the transaction.
func (tx *tx) Commit() error {
	return tx.sqlTx.Commit()
}

// Rollback aborts the transaction.
func (tx *tx) Rollback() error {
	return tx.sqlTx.Rollback()
}
