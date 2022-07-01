package zap

import (
	"context"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
	"go.uber.org/zap"
)

var _ otelexample.IdentifierGenerator = (*IdentifierGenerator)(nil)

// IdentifierGenerator represents a service for generate unique identifier values.
type IdentifierGenerator struct {
	wrapped otelexample.IdentifierGenerator
	logger  *zap.Logger
}

// NewIdentifierGenerator returns a new instance of IdentifierGenerator.
func NewIdentifierGenerator(svc otelexample.IdentifierGenerator, logger *zap.Logger) *IdentifierGenerator {
	return &IdentifierGenerator{
		wrapped: svc,
		logger:  logger,
	}
}

// GenerateIdentifier returns a new unique identifier.
func (svc *IdentifierGenerator) GenerateIdentifier(ctx context.Context) otelexample.ID {
	var id otelexample.ID

	start, end, elapsed := trackOfTime(func() {
		id = svc.wrapped.GenerateIdentifier(ctx)
	})

	svc.logger.Debug("generate identifier", zap.Stringer("start", start), zap.Stringer("end", end),
		zap.Stringer("elapsed", elapsed), zap.Stringer("id", id))

	return id
}
