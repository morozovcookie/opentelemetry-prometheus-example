package prometheus

import (
	"net"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
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

func HTTPHandler(registry prometheus.Registerer) func(next http.Handler) http.Handler {
	var (
		requestCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "requests_total",
			Help:        "measures the number of concurrent HTTP requests that are currently in-flight",
			ConstLabels: nil,
		},
			[]string{"host", "method", "status_code", "target", "client_ip", "scheme", "user_agent"})
		requestDurationVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   "",
			Subsystem:   "",
			Name:        "request_duration_seconds",
			Help:        "measures the duration of the inbound HTTP request",
			ConstLabels: nil,
			Buckets:     prometheus.DefBuckets,
		},
			[]string{"host", "method", "status_code", "target"})
	)

	registry.MustRegister(requestCounterVec, requestDurationVec)

	return httpHandler(requestCounterVec, requestDurationVec)
}

func httpHandler(
	counter *prometheus.CounterVec,
	histogram *prometheus.HistogramVec,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			resp := &response{
				wrapped: writer,

				statusCode: http.StatusOK,
			}

			userInfo := request.URL.User
			request.URL.User = nil

			target := request.URL.Path

			request.URL.User = userInfo

			clientIP, _, _ := net.SplitHostPort(request.RemoteAddr)

			_, _, elapsed := trackOfTime(func() {
				next.ServeHTTP(resp, request)
			})

			host, method, scheme, statusCode, ua := request.Host, request.Method, takeHTTPScheme(request),
				strconv.Itoa(resp.statusCode), request.UserAgent()

			labels := prometheus.Labels{
				"host":        host,
				"method":      method,
				"status_code": statusCode,
				"target":      target,
			}

			histogram.
				With(labels).
				Observe(elapsed.Seconds())

			labels["client_ip"], labels["scheme"], labels["user_agent"] = clientIP, scheme, ua

			counter.
				With(labels).
				Inc()
		})
	}
}

func takeHTTPScheme(request *http.Request) string {
	if request.TLS != nil {
		return "https"
	}

	return "http"
}
