package otelexample

import (
	"context"
	"time"
)

// Timer represents a service for getting time value.
type Timer interface {
	// Time returns time value.
	Time(ctx context.Context) time.Time
}
