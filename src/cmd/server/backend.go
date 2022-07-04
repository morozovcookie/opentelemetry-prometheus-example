package main

import (
	"context"
	"fmt"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
	"github.com/morozovcookie/opentelemetry-prometheus-example/nanoid"
	"github.com/morozovcookie/opentelemetry-prometheus-example/opentelemetry/metrics"
	"github.com/morozovcookie/opentelemetry-prometheus-example/opentelemetry/prometheus"
	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"github.com/morozovcookie/opentelemetry-prometheus-example/time"
	"github.com/morozovcookie/opentelemetry-prometheus-example/zap"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	uberzap "go.uber.org/zap"
)

type backend struct {
	config *Config
	logger *uberzap.Logger

	identifierGenerator otelexample.IdentifierGenerator
	timer               otelexample.Timer

	txBeginner percona.TxBeginner
	preparer   percona.Preparer

	prometheusExporter *otelprom.Exporter
	meterProvider      metric.MeterProvider

	userAccountService otelexample.UserAccountService
}

func newBackend(config *Config, logger *uberzap.Logger) *backend {
	be := new(backend)

	be.config, be.logger = config, logger

	be.initIdentifierGenerator()
	be.initTimer()

	return be
}

func (be *backend) init(ctx context.Context) error {
	var err error

	if be.prometheusExporter, err = prometheus.NewExporter(); err != nil {
		return fmt.Errorf("init backend: %w", err)
	}

	be.meterProvider = be.prometheusExporter.MeterProvider()

	var prepareTxBeginner percona.PrepareTxBeginner
	{
		perconaClient := percona.NewClient(be.config.PerconaConfig.Dsn)
		if err := perconaClient.Connect(ctx); err != nil {
			return fmt.Errorf("init backend: %w", err)
		}

		prepareTxBeginner = metrics.NewPrepareTxBeginner(perconaClient, be.meterProvider.Meter("sql"),
			semconv.DBSystemMySQL, semconv.DBNameKey.String(perconaClient.DBName()),
			semconv.DBUserKey.String(perconaClient.DBUser()))
	}

	perconaLogger := be.logger.Named("percona")

	be.initPreparer(prepareTxBeginner, perconaLogger)
	be.initTxBeginner(prepareTxBeginner, perconaLogger)

	be.initUserAccountService(perconaLogger)

	return nil
}

func (be *backend) initUserAccountService(logger *uberzap.Logger) {
	be.userAccountService = percona.NewUserAccountService(be.txBeginner, be.preparer, be.identifierGenerator, be.timer)
	be.userAccountService = zap.NewUserAccountService(be.userAccountService, logger.Named("user_account_svc"))
}

func (be *backend) initTxBeginner(beginner percona.TxBeginner, logger *uberzap.Logger) {
	be.txBeginner = zap.NewTxBeginner(beginner, logger)
}

func (be *backend) initPreparer(preparer percona.Preparer, logger *uberzap.Logger) {
	be.preparer = preparer
	be.preparer = zap.NewPreparer(be.preparer, logger)
}

func (be *backend) initIdentifierGenerator() {
	be.identifierGenerator = nanoid.NewIdentifierGenerator()
	be.identifierGenerator = zap.NewIdentifierGenerator(be.identifierGenerator,
		be.logger.Named("identifier_generator"))
}

func (be *backend) initTimer() {
	be.timer = time.NewTimer()
	be.timer = zap.NewTimer(be.timer, be.logger.Named("timer"))
}
