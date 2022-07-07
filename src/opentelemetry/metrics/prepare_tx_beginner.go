package metrics

import (
	"context"
	"database/sql"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
)

var _ percona.PrepareTxBeginner = (*PrepareTxBeginner)(nil)

type PrepareTxBeginner struct {
	wrapped percona.PrepareTxBeginner

	errorCounter  syncint64.Counter
	queryDuration syncint64.Histogram

	attrs []attribute.KeyValue
}

// NewPrepareTxBeginner returns a new instance of PrepareTxBeginner.
func NewPrepareTxBeginner(
	svc percona.PrepareTxBeginner,
	meter metric.Meter,
	attrs ...attribute.KeyValue,
) *PrepareTxBeginner {
	var (
		wrapper = &PrepareTxBeginner{
			wrapped: svc,

			errorCounter:  nil,
			queryDuration: nil,

			attrs: attrs,
		}

		err error
	)

	wrapper.errorCounter, err = meter.SyncInt64().Counter("errors_total",
		instrument.WithDescription("measures the number of SQL query errors"),
		instrument.WithUnit(unit.Dimensionless))
	if err != nil {
		panic(err)
	}

	wrapper.queryDuration, err = meter.SyncInt64().Histogram("duration",
		instrument.WithDescription("measures the duration of the SQL query"),
		instrument.WithUnit(unit.Milliseconds))
	if err != nil {
		panic(err)
	}

	return wrapper
}

// BeginTx starts a transaction.
func (svc *PrepareTxBeginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (percona.Tx, error) {
	perconaTx, err := svc.wrapped.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &tx{
		wrapped: perconaTx,

		errorCounter:  svc.errorCounter,
		queryDuration: svc.queryDuration,
		attrs:         svc.attrs,
	}, nil
}

// PrepareContext creates a prepared statement for later queries or executions.
func (svc *PrepareTxBeginner) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
	perconaStmt, err := svc.wrapped.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &stmt{
		wrapped: perconaStmt,

		errorCounter:  svc.errorCounter,
		queryDuration: svc.queryDuration,
		attrs:         svc.attrs,
	}, nil
}
