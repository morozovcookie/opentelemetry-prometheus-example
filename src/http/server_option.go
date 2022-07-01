package http

import (
	"net/http"
	"time"
)

// ServerOption is the option for configure Server object.
type ServerOption interface {
	apply(server *Server)
}

var _ ServerOption = (*serverOptionFunc)(nil)

type serverOptionFunc func(server *Server)

func (fn serverOptionFunc) apply(server *Server) {
	fn(server)
}

// DefaultReadTimeout is the default duration for reading the entire request, including the body.
const DefaultReadTimeout = time.Millisecond * 100

// WithReadTimeout sets up the maximum duration for reading the entire request, including the body.
func WithReadTimeout(timeout time.Duration) ServerOption {
	return serverOptionFunc(func(server *Server) {
		server.readTimeout = timeout
	})
}

// DefaultReadHeaderTimeout is the default amount of time allowed to read request headers.
const DefaultReadHeaderTimeout = time.Millisecond * 100

// WithReadHeaderTimeout sets up the amount of time allowed reading request headers.
func WithReadHeaderTimeout(timeout time.Duration) ServerOption {
	return serverOptionFunc(func(server *Server) {
		server.readHeaderTimeout = timeout
	})
}

// DefaultWriteTimeout is the default duration before timing out writes of the response.
const DefaultWriteTimeout = time.Millisecond * 100

// WithWriteTimeout set the maximum duration before timing out writes of the response.
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return serverOptionFunc(func(server *Server) {
		server.writeTimeout = timeout
	})
}

// DefaultIdleTimeout is the default amount of time to wait for the next request when keep-alives are enabled.
const DefaultIdleTimeout = time.Millisecond * 100

// WithIdleTimeout set the maximum amount of time to wait for the next request when keep-alives are enabled.
func WithIdleTimeout(timeout time.Duration) ServerOption {
	return serverOptionFunc(func(server *Server) {
		server.idleTimeout = timeout
	})
}

// DefaultMaxHeaderBytes controls the default number of bytes the server will read parsing the request header's keys and
// values, including the request line.
const DefaultMaxHeaderBytes = http.DefaultMaxHeaderBytes

// WithMaxHeaderBytes set the maximum number of bytes the server will read parsing the request header's keys and values,
// including the request line.
func WithMaxHeaderBytes(bytes int) ServerOption {
	return serverOptionFunc(func(server *Server) {
		server.maxHeaderBytes = bytes
	})
}
