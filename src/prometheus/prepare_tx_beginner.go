package prometheus

import (
	"context"
	"database/sql"
	"strings"

	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"github.com/prometheus/client_golang/prometheus"
)

var _ percona.PrepareTxBeginner = (*PrepareTxBeginner)(nil)

type PrepareTxBeginner struct {
	wrapped percona.PrepareTxBeginner

	errorsCounterVec    *prometheus.CounterVec
	rollbacksCounterVec *prometheus.CounterVec
	queryDurationVec    *prometheus.HistogramVec
}

// NewPrepareTxBeginner returns a new instance of PrepareTxBeginner.
func NewPrepareTxBeginner(svc percona.PrepareTxBeginner, registerer prometheus.Registerer) *PrepareTxBeginner {
	wrapper := &PrepareTxBeginner{
		wrapped: svc,

		errorsCounterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "query_errors_total",
			Help:        "measures the number of query errors",
			ConstLabels: nil,
		}, []string{"operation"}),
		rollbacksCounterVec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "tx_rollbacks_total",
			Help:        "measures the number of transaction rollbacks",
			ConstLabels: nil,
		}, nil),
		queryDurationVec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "query_duration_seconds",
			Help:        "measures the duration of the SQL query execution",
			ConstLabels: nil,
			Buckets:     prometheus.DefBuckets,
		}, []string{"operation"}),
	}

	registerer.MustRegister(wrapper.errorsCounterVec, wrapper.rollbacksCounterVec, wrapper.queryDurationVec)

	return wrapper
}

// PrepareContext creates a prepared statement for later queries or executions.
func (svc *PrepareTxBeginner) PrepareContext(ctx context.Context, query string) (percona.Stmt, error) {
	perconaStmt, err := svc.wrapped.PrepareContext(ctx, query)
	if err != nil {
		svc.errorsCounterVec.
			With(prometheus.Labels{
				"operation": "PREPARE",
			}).
			Inc()

		return nil, err
	}

	return &stmt{
		wrapped: perconaStmt,

		errorsCounterVec: svc.errorsCounterVec,
		queryDurationVec: svc.queryDurationVec,

		operation: strings.ToUpper(query[:strings.IndexByte(query, ' ')]),
	}, nil
}

// BeginTx starts a transaction.
func (svc *PrepareTxBeginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (percona.Tx, error) {
	perconaTx, err := svc.wrapped.BeginTx(ctx, opts)
	if err != nil {
		svc.errorsCounterVec.
			With(prometheus.Labels{
				"operation": "BEGIN",
			}).
			Inc()

		return nil, err
	}

	return &tx{
		wrapped: perconaTx,

		errorsCounterVec:    svc.errorsCounterVec,
		rollbacksCounterVec: svc.rollbacksCounterVec,
		queryDurationVec:    svc.queryDurationVec,
	}, nil
}
