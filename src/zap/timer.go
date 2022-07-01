package zap

import (
	"context"
	"time"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
	"go.uber.org/zap"
)

var _ otelexample.Timer = (*Timer)(nil)

// Timer represents a service for getting time value.
type Timer struct {
	wrapped otelexample.Timer
	logger  *zap.Logger
}

// NewTimer returns a new Timer.
func NewTimer(svc otelexample.Timer, logger *zap.Logger) *Timer {
	return &Timer{
		wrapped: svc,
		logger:  logger,
	}
}

// Time returns time value.
func (svc *Timer) Time(ctx context.Context) time.Time {
	var timeValue time.Time

	start, end, elapsed := trackOfTime(func() {
		timeValue = svc.wrapped.Time(ctx)
	})

	svc.logger.Debug("time", zap.Stringer("start", start), zap.Stringer("end", end),
		zap.Stringer("elapsed", elapsed), zap.Stringer("time", timeValue))

	return timeValue
}
