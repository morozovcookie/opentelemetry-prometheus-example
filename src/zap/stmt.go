package zap

import (
	"context"
	"database/sql"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.uber.org/zap"
)

var _ percona.Stmt = (*stmt)(nil)

// stmt is a prepared statement.
type stmt struct {
	wrapped percona.Stmt
	logger  *zap.Logger
	fields  []zap.Field

	query string
}

// ExecContext executes a prepared statement with the given arguments and
// returns a Result summarizing the effect of the statement.
func (stmt *stmt) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	var (
		result sql.Result
		err    error
	)

	start, end, elapsed := trackOfTime(func() {
		result, err = stmt.wrapped.ExecContext(ctx, args...)
	})

	ff := append(stmt.fields, zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Any("args", args), zap.String("query", stmt.query), zap.Error(err))

	stmt.logger.Debug("exec", ff...)

	if err != nil {
		stmt.logger.Error("exec", ff...)

		return nil, err // nolint:wrapcheck
	}

	return result, nil
}

// QueryRowContext executes a prepared query statement with the given arguments.
func (stmt *stmt) QueryRowContext(ctx context.Context, args ...any) *sql.Row {
	var row *sql.Row

	start, end, elapsed := trackOfTime(func() {
		row = stmt.wrapped.QueryRowContext(ctx, args...)
	})

	ff := append(stmt.fields, zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Any("args", args), zap.String("query", stmt.query), zap.Error(row.Err()))

	stmt.logger.Debug("query row", ff...)

	return row
}

// QueryContext executes a prepared query statement with the given arguments
// and returns the query results as a *Rows.
func (stmt *stmt) QueryContext(ctx context.Context, args ...any) (*sql.Rows, error) {
	var (
		rows *sql.Rows
		err  error
	)

	start, end, elapsed := trackOfTime(func() {
		rows, err = stmt.wrapped.QueryContext(ctx, args...)
	})

	ff := append(stmt.fields, zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Any("args", args), zap.String("query", stmt.query), zap.Error(err))

	stmt.logger.Debug("query", ff...)

	if err != nil {
		stmt.logger.Error("query", ff...)

		return nil, err
	}

	return rows, nil
}

func (stmt *stmt) Close(ctx context.Context) error {
	var err error

	start, end, elapsed := trackOfTime(func() {
		err = stmt.wrapped.Close(ctx)
	})

	ff := append(stmt.fields, zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Error(err), zap.String("query", stmt.query))

	stmt.logger.Debug("close", ff...)

	if err != nil {
		stmt.logger.Error("close", ff...)

		return err // nolint:wrapcheck
	}

	return nil
}
