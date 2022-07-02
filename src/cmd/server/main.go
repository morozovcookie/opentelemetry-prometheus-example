package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	stdtime "time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/morozovcookie/opentelemetry-prometheus-example/http"
	v1 "github.com/morozovcookie/opentelemetry-prometheus-example/http/v1"
	uberzap "go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	config := NewConfig()
	if err := config.Parse(); err != nil {
		log.Fatalln(err)
	}

	logger, err := initLogger(config)
	if err != nil {
		log.Fatalln(err)
	}

	defer func(logger *uberzap.Logger) {
		if err := logger.Sync(); err != nil {
			log.Fatalln(err)
		}
	}(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	be := newBackend(config, logger)
	if err := be.init(ctx); err != nil {
		logger.Fatal("failed to init backend", uberzap.Error(err))
	}

	var (
		httpServer    = initHTTPServer(be)
		monitorServer = initMonitorServer(be)
	)

	logger.Info("starting application")

	group.Go(startServer(monitorServer, "monitor", logger))
	group.Go(startServer(httpServer, "http", logger))

	logger.Info("application is started")

	<-ctx.Done()

	logger.Info("stopping application")

	const timeout = stdtime.Second * 5

	ctx, cancel = context.WithDeadline(ctx, stdtime.Now().Add(timeout))
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("shutdown http server", uberzap.Error(err))
	}

	if err := monitorServer.Shutdown(ctx); err != nil {
		logger.Error("shutdown monitor server", uberzap.Error(err))
	}

	if err := group.Wait(); err != nil {
		logger.Error("waiting for application be stopped", uberzap.Error(err))
	}

	logger.Info("application is stopped")
}

func initLogger(config *Config) (*uberzap.Logger, error) {
	loggerConfig := uberzap.NewProductionConfig()
	loggerConfig.Level = config.ZapLevel

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger.Named("server"), nil
}

func initHTTPServer(be *backend) *http.Server {
	router := chi.NewRouter()
	router.Use(middleware.RealIP, middleware.Logger, middleware.Recoverer)

	router.Mount(v1.UserAccountHandlerPathPrefix, v1.NewUserAccountHandler(be.userAccountService))

	return http.NewServer(be.config.HTTPConfig.Address, router)
}

func initMonitorServer(be *backend) *http.Server {
	router := chi.NewRouter()
	router.Use(middleware.RealIP, middleware.Logger, middleware.Recoverer)

	return http.NewServer(be.config.MonitorConfig.Address, router)
}

func startServer(server *http.Server, name string, logger *uberzap.Logger) func() error {
	return func() error {
		logger.Info(fmt.Sprintf("starting %s server", name), uberzap.String("address", server.Address()))

		if err := server.ListenAndServe(); err != nil {
			return fmt.Errorf("starting %s server: %w", name, err)
		}

		return nil
	}
}
