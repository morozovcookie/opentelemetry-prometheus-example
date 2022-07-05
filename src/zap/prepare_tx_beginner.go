package zap

import (
	"context"
	"database/sql"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.uber.org/zap"
)

var _ percona.PrepareTxBeginner = (*PrepareTxBeginner)(nil)

type PrepareTxBeginner struct {
	wrapped percona.PrepareTxBeginner
	logger  *zap.Logger
}

// NewPrepareTxBeginner returns a new instance of PrepareTxBeginner.
func NewPrepareTxBeginner(svc percona.PrepareTxBeginner, logger *zap.Logger) *PrepareTxBeginner {
	return &PrepareTxBeginner{
		wrapped: svc,
		logger:  logger,
	}
}

// PrepareContext creates a prepared statement for later queries or executions.
func (svc *PrepareTxBeginner) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
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

// BeginTx starts a transaction.
func (svc *PrepareTxBeginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (percona.Tx, error) {
	var (
		perconaTx percona.Tx
		err       error

		dbName = svc.wrapped.DBName()
		dbUser = svc.wrapped.DBUser()
	)

	start, end, elapsed := trackOfTime(func() {
		perconaTx, err = svc.wrapped.BeginTx(ctx, opts)
	})

	ff := []zap.Field{
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("dbName", dbName), zap.String("dbUser", dbUser), zap.Any("options", opts),
		zap.Error(err),
	}

	svc.logger.Debug("begin tx", ff...)

	if err != nil {
		svc.logger.Error("begin tx", ff...)

		return nil, err // nolint:wrapcheck
	}

	return &tx{
		wrapped: perconaTx,
		logger:  svc.logger.Named("tx"),

		dbName: dbName,
		dbUser: dbUser,
	}, nil
}

// DBName returns name of database which client are connected.
func (svc *PrepareTxBeginner) DBName() string {
	return svc.wrapped.DBName()
}

// DBUser returns name of user which connected to the database.
func (svc *PrepareTxBeginner) DBUser() string {
	return svc.wrapped.DBUser()
}
