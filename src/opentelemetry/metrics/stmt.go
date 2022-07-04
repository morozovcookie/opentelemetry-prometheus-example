package metrics

import (
	"context"
	"database/sql"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

var _ percona.Stmt = (*stmt)(nil)

// stmt is a prepared statement.
type stmt struct {
	wrapped percona.Stmt

	errorCounter  syncint64.Counter
	queryDuration syncint64.Histogram
	attrs         []attribute.KeyValue
}

// ExecContext executes a prepared statement with the given arguments and
// returns a Result summarizing the effect of the statement.
func (stmt *stmt) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	var (
		result sql.Result
		err    error
	)

	_, _, elapsed := trackOfTime(func() {
		result, err = stmt.wrapped.ExecContext(ctx, args...)
	})

	stmt.queryDuration.Record(ctx, elapsed.Milliseconds(), stmt.attrs...)

	if err != nil {
		stmt.errorCounter.Add(ctx, 1, stmt.attrs...)

		return nil, err
	}

	return result, nil
}

// QueryRowContext executes a prepared query statement with the given arguments.
func (stmt *stmt) QueryRowContext(ctx context.Context, args ...any) *sql.Row {
	var row *sql.Row

	_, _, elapsed := trackOfTime(func() {
		row = stmt.wrapped.QueryRowContext(ctx, args...)
	})

	stmt.queryDuration.Record(ctx, elapsed.Milliseconds(), stmt.attrs...)

	if err := row.Err(); err != nil {
		stmt.errorCounter.Add(ctx, 1, stmt.attrs...)
	}

	return stmt.wrapped.QueryRowContext(ctx, args...)
}

// QueryContext executes a prepared query statement with the given arguments
// and returns the query results as a *Rows.
func (stmt *stmt) QueryContext(ctx context.Context, args ...any) (*sql.Rows, error) {
	var (
		rows *sql.Rows
		err  error
	)

	_, _, elapsed := trackOfTime(func() {
		rows, err = stmt.wrapped.QueryContext(ctx, args...)
	})

	stmt.queryDuration.Record(ctx, elapsed.Milliseconds(), stmt.attrs...)

	if err != nil {
		stmt.errorCounter.Add(ctx, 1, stmt.attrs...)

		return nil, err
	}

	return rows, nil
}

// Close closes the statement.
func (stmt *stmt) Close(ctx context.Context) error {
	return stmt.wrapped.Close(ctx)
}
