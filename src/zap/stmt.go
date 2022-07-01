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

	query  string
	dbName string
	dbUser string
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

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("dbName", stmt.dbName), zap.String("dbUser", stmt.dbUser), zap.Any("args", args),
		zap.String("query", stmt.query), zap.Error(err),
	}

	stmt.logger.Debug("exec", ff...)

	if err != nil {
		stmt.logger.Error("exec", ff...)

		return nil, err // nolint:wrapcheck
	}

	return result, nil
}

func (stmt *stmt) Close(ctx context.Context) error {
	var err error

	start, end, elapsed := trackOfTime(func() {
		err = stmt.wrapped.Close(ctx)
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("dbName", stmt.dbName), zap.String("dbUser", stmt.dbUser), zap.Error(err),
		zap.String("query", stmt.query),
	}

	stmt.logger.Debug("close", ff...)

	if err != nil {
		stmt.logger.Error("close", ff...)

		return err // nolint:wrapcheck
	}

	return nil
}
