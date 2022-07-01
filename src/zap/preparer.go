package zap

import (
	"context"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.uber.org/zap"
)

var _ percona.Preparer = (*Preparer)(nil)

// Preparer represents a service that can create a prepared statement.
type Preparer struct {
	wrapped percona.Preparer
	logger  *zap.Logger
}

// NewPreparer returns a new instance of Preparer.
func NewPreparer(svc percona.Preparer, logger *zap.Logger) *Preparer {
	return &Preparer{
		wrapped: svc,
		logger:  logger,
	}
}

// PrepareContext creates a prepared statement for later queries or executions.
func (svc *Preparer) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
	var (
		perconaStmt percona.Stmt
		err         error

		dbName = svc.wrapped.DBName()
		dbUser = svc.wrapped.DBUser()
	)

	start, end, elapsed := trackOfTime(func() {
		perconaStmt, err = svc.wrapped.PrepareContext(ctx, query)
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("dbName", dbName), zap.String("dbUser", dbUser), zap.String("query", query),
		zap.Error(err),
	}

	svc.logger.Debug("prepare", ff...)

	if err != nil {
		svc.logger.Error("prepare", ff...)

		return nil, err
	}

	return &stmt{
		wrapped: perconaStmt,
		logger:  svc.logger.Named("stmt"),

		query:  query,
		dbName: dbName,
		dbUser: dbUser,
	}, nil
}

// DBName returns name of database which client are connected.
func (svc *Preparer) DBName() string {
	return svc.wrapped.DBName()
}

// DBUser returns name of user which connected to the database.
func (svc *Preparer) DBUser() string {
	return svc.wrapped.DBUser()
}
