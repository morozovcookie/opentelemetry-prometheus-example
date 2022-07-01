package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// A Server defines parameters for running an HTTP server.
type Server struct {
	addr string

	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int

	instance *http.Server
}

// NewServer returns a new Server instance.
func NewServer(addr string, handler http.Handler, opts ...ServerOptionFunc) *Server {
	srv := &Server{
		addr: addr,

		readTimeout:       DefaultReadTimeout,
		readHeaderTimeout: DefaultReadHeaderTimeout,
		writeTimeout:      DefaultWriteTimeout,
		idleTimeout:       DefaultIdleTimeout,
		maxHeaderBytes:    DefaultMaxHeaderBytes,

		instance: nil,
	}

	for _, opt := range opts {
		opt(srv)
	}

	srv.instance = &http.Server{
		Addr:              addr,
		Handler:           handler,
		TLSConfig:         nil,
		ReadTimeout:       srv.readTimeout,
		ReadHeaderTimeout: srv.readHeaderTimeout,
		WriteTimeout:      srv.writeTimeout,
		IdleTimeout:       srv.idleTimeout,
		MaxHeaderBytes:    srv.maxHeaderBytes,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
		BaseContext:       nil,
		ConnContext:       nil,
	}

	return srv
}

// Address returns a served server address.
func (srv *Server) Address() string {
	return srv.addr
}

// ListenAndServe listens on the TCP network address srv.Addr and then calls Serve to handle requests on incoming
// connections. Accepted connections are configured to enable TCP keep-alives.
func (srv *Server) ListenAndServe() error {
	if err := srv.instance.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start http server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (srv *Server) Shutdown(ctx context.Context) error {
	if err := srv.instance.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	return nil
}
