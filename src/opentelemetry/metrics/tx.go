package metrics

import (
	"context"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

var _ percona.Tx = (*tx)(nil)

// tx is an in-progress database transaction.
type tx struct {
	wrapped percona.Tx

	errorCounter  syncint64.Counter
	queryDuration syncint64.Histogram
	attrs         []attribute.KeyValue
}

// PrepareContext creates a prepared statement for later queries or executions.
func (tx *tx) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
	perconaStmt, err := tx.wrapped.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &stmt{
		wrapped: perconaStmt,

		errorCounter:  tx.errorCounter,
		queryDuration: tx.queryDuration,
		attrs:         tx.attrs,
	}, nil
}

// Commit commits the transaction.
func (tx *tx) Commit() error {
	return tx.wrapped.Commit()
}

// Rollback aborts the transaction.
func (tx *tx) Rollback() error {
	return tx.wrapped.Rollback()
}
