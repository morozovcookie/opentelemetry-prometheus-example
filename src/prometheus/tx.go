package prometheus

import (
	"context"
	"strings"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"github.com/prometheus/client_golang/prometheus"
)

var _ percona.Tx = (*tx)(nil)

// tx is an in-progress database transaction.
type tx struct {
	wrapped percona.Tx

	errorsCounterVec    *prometheus.CounterVec
	rollbacksCounterVec *prometheus.CounterVec
	queryDurationVec    *prometheus.HistogramVec
}

// PrepareContext creates a prepared statement for later queries or executions.
func (tx *tx) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
	perconaStmt, err := tx.wrapped.PrepareContext(ctx, query)
	if err != nil {
		tx.errorsCounterVec.
			With(prometheus.Labels{
				"operation": "PREPARE",
			}).
			Inc()

		return nil, err
	}

	return &stmt{
		wrapped: perconaStmt,

		errorsCounterVec: tx.errorsCounterVec,
		queryDurationVec: tx.queryDurationVec,

		operation: strings.ToUpper(query[:strings.IndexByte(query, ' ')]),
	}, nil
}

// Commit commits the transaction.
func (tx *tx) Commit() error {
	if err := tx.wrapped.Commit(); err != nil {
		tx.errorsCounterVec.
			With(prometheus.Labels{
				"operation": "COMMIT",
			}).
			Inc()

		return err
	}

	return nil
}

// Rollback aborts the transaction.
func (tx *tx) Rollback() error {
	tx.rollbacksCounterVec.
		With(nil).
		Inc()

	if err := tx.wrapped.Rollback(); err != nil {
		tx.errorsCounterVec.
			With(prometheus.Labels{
				"operation": "ROLLBACK",
			}).
			Inc()

		return err
	}

	return nil
}
