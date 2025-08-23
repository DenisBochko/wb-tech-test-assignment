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

type Server interface {
	Run() error
	Shutdown(ctx context.Context) error
	Info() <-chan string
	Error() <-chan error
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

type HTTPServer struct {
	srv      *http.Server
	infoChan chan string
	errChan  chan error
}

func NewHTTPServer(opts ...Option) Server {
	srv := &http.Server{
		Addr:         net.JoinHostPort(DefaultHost, strconv.Itoa(DefaultPort)),
		ReadTimeout:  DefaultReadTimeout,
		WriteTimeout: DefaultWriteTimeout,
		IdleTimeout:  DefaultIdleTimeout,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return &HTTPServer{
		srv:      srv,
		infoChan: make(chan string, 1),
		errChan:  make(chan error, 1),
	}
}

func (s *HTTPServer) Run() error {
	s.sendInfo(fmt.Sprintf("Starting HTTP server at: %s", s.srv.Addr))

	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.errChan <- fmt.Errorf("failed to start HTTP server: %w", err)
		}
	}()

	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	return nil
}

func (s *HTTPServer) Info() <-chan string {
	return s.infoChan
}

func (s *HTTPServer) Error() <-chan error {
	return s.errChan
}

func (s *HTTPServer) sendInfo(msg string) {
	select {
	case s.infoChan <- msg:
	default:
	}
}
