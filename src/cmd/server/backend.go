package main

import (
	"context"
	"fmt"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
	"github.com/morozovcookie/opentelemetry-prometheus-example/nanoid"
	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"github.com/morozovcookie/opentelemetry-prometheus-example/time"
	"github.com/morozovcookie/opentelemetry-prometheus-example/zap"
	"github.com/prometheus/client_golang/prometheus"
	uberzap "go.uber.org/zap"
)

type backend struct {
	config *Config
	logger *uberzap.Logger

	registerer prometheus.Registerer
	gatherer   prometheus.Gatherer

	identifierGenerator otelexample.IdentifierGenerator
	timer               otelexample.Timer

	prepareTxBeginner percona.PrepareTxBeginner

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
	registry := prometheus.NewRegistry()
	be.registerer, be.gatherer = registry, registry

	be.registerer = prometheus.WrapRegistererWithPrefix("server_", be.registerer)

	perconaLogger := be.logger.Named("percona")

	if err := be.initPrepareTxBeginner(ctx, perconaLogger); err != nil {
		return fmt.Errorf("init backend: %w", err)
	}

	be.initUserAccountService(perconaLogger)

	return nil
}

func (be *backend) initPrepareTxBeginner(ctx context.Context, logger *uberzap.Logger) error {
	perconaClient := percona.NewClient(be.config.PerconaConfig.Dsn)
	if err := perconaClient.Connect(ctx); err != nil {
		return err
	}

	be.prepareTxBeginner = perconaClient
	be.prepareTxBeginner = zap.NewPrepareTxBeginner(be.prepareTxBeginner, logger)

	return nil
}

func (be *backend) initUserAccountService(logger *uberzap.Logger) {
	be.userAccountService = percona.NewUserAccountService(be.prepareTxBeginner, be.identifierGenerator, be.timer)
	be.userAccountService = zap.NewUserAccountService(be.userAccountService, logger.Named("user_account_svc"))
}

func (be *backend) initIdentifierGenerator() {
	be.identifierGenerator = nanoid.NewIdentifierGenerator()
	be.identifierGenerator = zap.NewIdentifierGenerator(be.identifierGenerator, be.logger.Named("identifier_generator"))
}

func (be *backend) initTimer() {
	be.timer = time.NewTimer()
	be.timer = zap.NewTimer(be.timer, be.logger.Named("timer"))
}
