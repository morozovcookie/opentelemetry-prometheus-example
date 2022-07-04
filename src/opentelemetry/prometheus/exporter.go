package prometheus

import (
	"fmt"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

func NewExporter() (*prometheus.Exporter, error) {
	var (
		config = prometheus.Config{}
		ctrl   = controller.New(
			processor.NewFactory(
				selector.NewWithHistogramDistribution(
					histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
				),
				aggregation.CumulativeTemporalitySelector(),
				processor.WithMemory(true),
			),
			controller.WithResource(
				resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("server"),
					semconv.ServiceVersionKey.String("1.0.0"),
					semconv.DeploymentEnvironmentKey.String("production"),
				),
			),
		)
	)

	exporter, err := prometheus.New(config, ctrl)
	if err != nil {
		return nil, fmt.Errorf("init prometheus exporter: %w", err)
	}

	return exporter, nil
}
