package zap

import (
	"context"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.uber.org/zap"
)

var _ percona.Tx = (*tx)(nil)

// tx is an in-progress database transaction.
type tx struct {
	wrapped percona.Tx
	logger  *zap.Logger

	dbName string
	dbUser string
}

// PrepareContext creates a prepared statement for later queries or executions.
func (tx *tx) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
	var (
		perconaStmt percona.Stmt
		err         error
	)

	start, end, elapsed := trackOfTime(func() {
		perconaStmt, err = tx.wrapped.PrepareContext(ctx, query)
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("dbName", tx.dbName), zap.String("dbUser", tx.dbUser), zap.String("query", query),
		zap.Error(err),
	}

	tx.logger.Debug("prepare", ff...)

	if err != nil {
		tx.logger.Error("prepare", ff...)

		return nil, err // nolint:wrapcheck
	}

	return &stmt{
		wrapped: perconaStmt,
		logger:  tx.logger.Named("stmt"),

		query:  query,
		dbName: tx.dbName,
		dbUser: tx.dbUser,
	}, nil
}

// Commit commits the transaction.
func (tx *tx) Commit() error {
	var err error

	start, end, elapsed := trackOfTime(func() {
		err = tx.wrapped.Commit()
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("dbName", tx.dbName), zap.String("dbUser", tx.dbUser), zap.Error(err),
	}

	tx.logger.Debug("commit", ff...)

	if err != nil {
		tx.logger.Error("commit", ff...)

		return err // nolint:wrapcheck
	}

	return nil
}

// Rollback aborts the transaction.
func (tx *tx) Rollback() error {
	var err error

	start, end, elapsed := trackOfTime(func() {
		err = tx.wrapped.Rollback()
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("dbName", tx.dbName), zap.String("dbUser", tx.dbUser), zap.Error(err),
	}

	tx.logger.Debug("rollback", ff...)

	if err != nil {
		tx.logger.Error("rollback", ff...)

		return err // nolint:wrapcheck
	}

	return nil
}
