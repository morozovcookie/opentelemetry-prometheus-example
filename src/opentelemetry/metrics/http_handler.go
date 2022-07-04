package metrics

import (
	"bytes"
	"net"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

var _ http.ResponseWriter = (*response)(nil)

type response struct {
	wrapped http.ResponseWriter

	statusCode int
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
func (resp *response) Header() http.Header {
	return resp.wrapped.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (resp *response) Write(bb []byte) (int, error) {
	return resp.wrapped.Write(bb)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (resp *response) WriteHeader(statusCode int) {
	resp.statusCode = statusCode

	resp.wrapped.WriteHeader(resp.statusCode)
}

func HTTPHandler(meter metric.Meter, attrs ...attribute.KeyValue) func(next http.Handler) http.Handler {
	requestCount, err := meter.SyncInt64().Counter("active_requests",
		instrument.WithDescription("measures the number of concurrent HTTP requests that are currently in-flight"),
		instrument.WithUnit(unit.Dimensionless))
	if err != nil {
		panic(err)
	}

	requestDuration, err := meter.SyncInt64().Histogram("duration",
		instrument.WithDescription("measures the duration of the inbound HTTP request"),
		instrument.WithUnit(unit.Milliseconds))
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			resp := &response{
				wrapped: writer,

				statusCode: http.StatusOK,
			}

			userInfo := request.URL.User
			request.URL.User = nil

			attrs = append(attrs, semconv.HTTPURLKey.String(request.URL.String()),
				semconv.HTTPTargetKey.String(request.URL.Path))

			request.URL.User = userInfo

			_, _, elapsed := trackOfTime(func() {
				next.ServeHTTP(resp, request)
			})

			rattrs := attrs

			if ua := request.UserAgent(); ua != "" {
				rattrs = append(rattrs, semconv.HTTPUserAgentKey.String(ua))
			}

			if request.Host != "" {
				rattrs = append(rattrs, semconv.HTTPHostKey.String(request.Host))
			}

			flavor := new(bytes.Buffer)
			_, _ = flavor.WriteString(strconv.Itoa(request.ProtoMajor))

			if request.ProtoMajor == 1 {
				_, _ = flavor.WriteRune('.')
				_, _ = flavor.WriteString(strconv.Itoa(request.ProtoMinor))
			}

			if val := flavor.String(); val != "" {
				attrs = append(attrs, semconv.HTTPFlavorKey.String(val))
			}

			schema := semconv.HTTPSchemeHTTP
			if request.TLS != nil {
				schema = semconv.HTTPSchemeHTTPS
			}

			clientIP, _, _ := net.SplitHostPort(request.RemoteAddr)

			rattrs = append(rattrs, semconv.HTTPMethodKey.String(request.Method), schema,
				semconv.HTTPClientIPKey.String(clientIP), semconv.HTTPStatusCodeKey.Int(resp.statusCode))

			ctx := request.Context()

			requestCount.Add(ctx, 1, rattrs...)
			requestDuration.Record(ctx, elapsed.Milliseconds(), rattrs...)
		})
	}
}
