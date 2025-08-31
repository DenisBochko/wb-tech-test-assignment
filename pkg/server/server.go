package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultHost            = "0.0.0.0"
	DefaultPort            = 8080
	DefaultReadTimeout     = 5 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
	DefaultIdleTimeout     = 15 * time.Second
	DefaultShutdownTimeout = 30 * time.Second
)

type HTTPServer interface {
	Run() error
	Shutdown() error
}

type Option func(*http.Server)

func WithAddr(host string, port uint16) Option {
	return func(server *http.Server) {
		server.Addr = net.JoinHostPort(host, strconv.Itoa(int(port)))
	}
}

func WithTimeout(readTimeout, writeTimeout, IdleTimeout time.Duration) Option {
	return func(server *http.Server) {
		server.ReadTimeout = readTimeout
		server.WriteTimeout = writeTimeout
		server.IdleTimeout = IdleTimeout
	}
}

func WithHandler(handler http.Handler) Option {
	return func(server *http.Server) {
		server.Handler = handler
	}
}

type httpServer struct {
	srv *http.Server
}

func NewHTTPServer(opts ...Option) HTTPServer {
	srv := &http.Server{
		Addr:         net.JoinHostPort(DefaultHost, strconv.Itoa(DefaultPort)),
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return &httpServer{
		srv: srv,
	}
}

func (s *httpServer) Run() error {
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *httpServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}
