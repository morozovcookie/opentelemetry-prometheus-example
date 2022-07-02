package nanoid

import (
	"context"

	gonanoid "github.com/matoous/go-nanoid/v2"
	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz1234567890"
	size     = 64
)

var _ otelexample.IdentifierGenerator = (*IdentifierGenerator)(nil)

// IdentifierGenerator represents a service for generate unique identifier values.
type IdentifierGenerator struct{}

// NewIdentifierGenerator returns a new IdentifierGenerator instance.
func NewIdentifierGenerator() *IdentifierGenerator {
	return &IdentifierGenerator{}
}

// GenerateIdentifier returns a new unique identifier.
func (svc *IdentifierGenerator) GenerateIdentifier(_ context.Context) otelexample.ID {
	return otelexample.ID(gonanoid.MustGenerate(alphabet, size))
}
