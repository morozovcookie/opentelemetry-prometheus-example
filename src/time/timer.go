package time

import (
	"context"
	"time"

	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
)

var _ otelexample.Timer = (*Timer)(nil)

// Timer represents a service for getting time value.
type Timer struct{}

// NewTimer returns a new Timer instance.
func NewTimer() *Timer {
	return &Timer{}
}

// Time returns time value.
func (t *Timer) Time(_ context.Context) time.Time {
	return time.Now().UTC()
}
