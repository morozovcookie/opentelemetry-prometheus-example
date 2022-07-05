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
	fields  []zap.Field
}

// NewPrepareTxBeginner returns a new instance of PrepareTxBeginner.
func NewPrepareTxBeginner(svc percona.PrepareTxBeginner, logger *zap.Logger, ff ...zap.Field) *PrepareTxBeginner {
	return &PrepareTxBeginner{
		wrapped: svc,
		logger:  logger,
		fields:  ff,
	}
}

// PrepareContext creates a prepared statement for later queries or executions.
func (svc *PrepareTxBeginner) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
	var (
		perconaStmt percona.Stmt
		err         error
	)

	start, end, elapsed := trackOfTime(func() {
		perconaStmt, err = svc.wrapped.PrepareContext(ctx, query)
	})

	ff := append(svc.fields, zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("query", query), zap.Error(err))

	svc.logger.Debug("prepare", ff...)

	if err != nil {
		svc.logger.Error("prepare", ff...)

		return nil, err
	}

	return &stmt{
		wrapped: perconaStmt,
		logger:  svc.logger.Named("stmt"),
		fields:  svc.fields,

		query: query,
	}, nil
}

// BeginTx starts a transaction.
func (svc *PrepareTxBeginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (percona.Tx, error) {
	var (
		perconaTx percona.Tx
		err       error
	)

	start, end, elapsed := trackOfTime(func() {
		perconaTx, err = svc.wrapped.BeginTx(ctx, opts)
	})

	ff := append(svc.fields, zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.Any("options", opts), zap.Error(err))

	svc.logger.Debug("begin tx", ff...)

	if err != nil {
		svc.logger.Error("begin tx", ff...)

		return nil, err // nolint:wrapcheck
	}

	return &tx{
		wrapped: perconaTx,
		logger:  svc.logger.Named("tx"),
		fields:  svc.fields,
	}, nil
}
