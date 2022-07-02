package main

import (
	"context"
	"fmt"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
	"github.com/morozovcookie/opentelemetry-prometheus-example/nanoid"
	"github.com/morozovcookie/opentelemetry-prometheus-example/percona"
	"github.com/morozovcookie/opentelemetry-prometheus-example/time"
	"github.com/morozovcookie/opentelemetry-prometheus-example/zap"
	uberzap "go.uber.org/zap"
)

type backend struct {
	config *Config
	logger *uberzap.Logger

	identifierGenerator otelexample.IdentifierGenerator
	timer               otelexample.Timer

	txBeginner percona.TxBeginner
	preparer   percona.Preparer

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
	var (
		perconaClient = percona.NewClient(be.config.PerconaConfig.Dsn)
		perconaLogger = be.logger.Named("percona")
	)

	if err := perconaClient.Connect(ctx); err != nil {
		return fmt.Errorf("init backend: %w", err)
	}

	be.initPreparer(perconaClient, perconaLogger)
	be.initTxBeginner(perconaClient, perconaLogger)

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
	be.preparer = zap.NewPreparer(preparer, logger)
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
