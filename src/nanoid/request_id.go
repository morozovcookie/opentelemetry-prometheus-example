package nanoid

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	otelexample "github.com/morozovcookie/opentelemetry-prometheus-example"
)

func RequestID(generator otelexample.IdentifierGenerator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()

			requestID := request.Header.Get(middleware.RequestIDHeader)
			if requestID == "" {
				requestID = generator.GenerateIdentifier(ctx).String()
			}

			next.ServeHTTP(writer, request.WithContext(context.WithValue(ctx, middleware.RequestIDKey, requestID)))
		})
	}
}
