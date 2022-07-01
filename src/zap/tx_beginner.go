package zap

import (
	"context"
	"database/sql"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.uber.org/zap"
)

var _ percona.TxBeginner = (*TxBeginner)(nil)

// TxBeginner represents a service that can start a transaction.
type TxBeginner struct {
	wrapped percona.TxBeginner
	logger  *zap.Logger
}

// NewTxBeginner returns a new instance of TxBeginner.
func NewTxBeginner(beginner percona.TxBeginner, logger *zap.Logger) *TxBeginner {
	return &TxBeginner{
		wrapped: beginner,
		logger:  logger,
	}
}

// BeginTx starts a transaction.
func (svc *TxBeginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (percona.Tx, error) {
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
func (svc *TxBeginner) DBName() string {
	return svc.wrapped.DBName()
}

// DBUser returns name of user which connected to the database.
func (svc *TxBeginner) DBUser() string {
	return svc.wrapped.DBUser()
}
