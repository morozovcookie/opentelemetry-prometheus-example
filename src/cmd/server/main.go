package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/morozovcookie/opentelemetry-prometheus-example/http"
	uberzap "go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := NewConfig()
	if err := cfg.Parse(); err != nil {
		log.Fatalln(err)
	}

	loggerConfig := uberzap.NewProductionConfig()
	loggerConfig.Level = cfg.ZapLevel

	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalln(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	var (
		httpRouter = chi.NewRouter()
		httpServer = http.NewServer(cfg.HTTPConfig.Address, httpRouter)

		monitorRouter = chi.NewRouter()
		monitorServer = http.NewServer(cfg.MonitorConfig.Address, monitorRouter)
	)

	logger.Info("starting application")

	group.Go(startServer(monitorServer, "monitor", logger))
	group.Go(startServer(httpServer, "http", logger))

	logger.Info("application is started")

	<-ctx.Done()

	logger.Info("stopping application")

	const timeout = time.Second * 5

	ctx, cancel = context.WithDeadline(ctx, time.Now().Add(timeout))
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

func startServer(server *http.Server, name string, logger *uberzap.Logger) func() error {
	return func() error {
		logger.Info(fmt.Sprintf("starting %s server", name), uberzap.String("address", server.Address()))

		if err := server.ListenAndServe(); err != nil {
			return fmt.Errorf("starting %s server: %w", name, err)
		}

		return nil
	}
}
