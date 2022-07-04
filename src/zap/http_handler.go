package zap

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

var _ http.ResponseWriter = (*response)(nil)

type response struct {
	wrapped http.ResponseWriter

	statusCode int
	buffer     *bytes.Buffer
}

// Header returns the header map that will be sent by
// WriteHeader. The Header map also is the mechanism with which
// Handlers can set HTTP trailers.
func (resp *response) Header() http.Header {
	return resp.wrapped.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (resp *response) Write(bb []byte) (int, error) {
	if _, err := io.Copy(resp.buffer, bytes.NewBuffer(bb)); err != nil {
		return 0, err
	}

	return resp.wrapped.Write(bb)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (resp *response) WriteHeader(statusCode int) {
	resp.statusCode = statusCode

	resp.wrapped.WriteHeader(resp.statusCode)
}

func HTTPHandler(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer recoverRequestPanic(logger, writer, request)

			logRequest(logger, next, writer, request)
		})
	}
}

func recoverRequestPanic(logger *zap.Logger, writer http.ResponseWriter, request *http.Request) {
	panicError := recover()
	if panicError == nil {
		return
	}

	err, ok := panicError.(error)
	if !ok || errors.Is(err, http.ErrAbortHandler) {
		return
	}

	dumpedRequest, dumpErr := httputil.DumpRequest(request, false)
	if dumpErr != nil {
		logger.Error("dump request", zap.Error(dumpErr))

		return
	}

	buf := bytes.NewBuffer(dumpedRequest)

	if isBrokenPipe(err) {
		logger.Error(request.URL.Path, zap.Error(err), zap.Stringer("request", buf))

		return
	}

	logger.Error("[Recovery from panic]", zap.Time("time", time.Now().UTC()), zap.Error(err),
		zap.Stringer("request", buf), zap.String("stack", string(debug.Stack())))

	writer.WriteHeader(http.StatusInternalServerError)
}

func isBrokenPipe(err error) bool {
	var netError *net.OpError
	if !errors.As(err, &netError) {
		return false
	}

	var syscallError *os.SyscallError
	if !errors.As(netError.Err, &syscallError) {
		return false
	}

	return strings.Contains(strings.ToLower(syscallError.Error()), "broken pipe") ||
		strings.Contains(strings.ToLower(syscallError.Error()), "connection reset by peer")
}

func logRequest(logger *zap.Logger, next http.Handler, writer http.ResponseWriter, request *http.Request) {
	resp := &response{
		wrapped: writer,

		statusCode: http.StatusOK,
		buffer:     bytes.NewBuffer(nil),
	}

	start, end, elapsed := trackOfTime(func() {
		next.ServeHTTP(resp, request)
	})

	ff := []zap.Field{
		zap.String("rid", middleware.GetReqID(request.Context())), zap.Int("status", resp.statusCode),
		zap.Stringer("start", start), zap.Stringer("end", end), zap.Stringer("elapsed", elapsed),
		zap.String("http-method", request.Method), zap.String("path", request.URL.Path),
		zap.String("user-agent", request.UserAgent()), zap.String("query", request.URL.RawQuery),
		zap.String("ip", request.RemoteAddr),
	}

	if check := logger.Check(zap.DebugLevel, request.URL.Path); check != nil {
		check.Write(append(ff, zap.Stringer("response", resp.buffer))...)

		return
	}

	logger.Info(request.URL.Path, ff...)
}
