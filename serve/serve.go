package serve

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Config holds serve configuration.
// It defines the server's network settings, health check configuration,
// graceful shutdown behavior, and optional TLS settings.
type Config struct {
	// Port is the TCP port on which the gRPC server listens.
	// Default: 50051
	Port int

	// HealthEndpoint is the path for HTTP health checks.
	// This is not currently used but reserved for future HTTP health endpoint.
	// Default: /health
	HealthEndpoint string

	// GracefulTimeout is the maximum duration to wait for active requests
	// to complete during graceful shutdown.
	// Default: 30 seconds
	GracefulTimeout time.Duration

	// TLSCertFile is the path to the TLS certificate file.
	// If empty, TLS is disabled.
	TLSCertFile string

	// TLSKeyFile is the path to the TLS private key file.
	// If empty, TLS is disabled.
	TLSKeyFile string
}

// DefaultConfig returns default serve configuration.
// These defaults are suitable for local development and testing.
func DefaultConfig() *Config {
	return &Config{
		Port:            50051,
		HealthEndpoint:  "/health",
		GracefulTimeout: 30 * time.Second,
	}
}

// Server wraps a gRPC server with lifecycle management.
// It handles server initialization, startup, graceful shutdown,
// and health check registration.
type Server struct {
	grpcServer   *grpc.Server
	listener     net.Listener
	config       *Config
	healthServer *health.Server
}

// NewServer creates a new gRPC server with the provided configuration.
// It sets up the gRPC server with appropriate options (e.g., TLS)
// and registers the health check service.
func NewServer(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Port, err)
	}

	// Build gRPC server options
	var opts []grpc.ServerOption

	// Configure TLS if cert and key are provided
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			listener.Close()
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Create gRPC server
	grpcServer := grpc.NewServer(opts...)

	// Create and register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	return &Server{
		grpcServer:   grpcServer,
		listener:     listener,
		config:       cfg,
		healthServer: healthServer,
	}, nil
}

// GRPCServer returns the underlying gRPC server.
// This allows callers to register additional services.
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

// HealthServer returns the health check server.
// This allows callers to set service health status.
func (s *Server) HealthServer() *health.Server {
	return s.healthServer
}

// Serve starts the gRPC server and blocks until shutdown.
// It handles graceful shutdown on SIGINT/SIGTERM signals.
// The context can be used to initiate shutdown programmatically.
func (s *Server) Serve(ctx context.Context) error {
	// Create error channel for serve errors
	errCh := make(chan error, 1)

	// Start serving in a goroutine
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			errCh <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal, context cancellation, or error
	select {
	case <-ctx.Done():
		// Context cancelled - graceful shutdown
		s.GracefulStop()
		return ctx.Err()
	case sig := <-sigCh:
		// Signal received - graceful shutdown
		fmt.Printf("Received signal %v, shutting down gracefully...\n", sig)
		s.GracefulStop()
		return nil
	case err := <-errCh:
		// Server error
		return err
	}
}

// Stop immediately stops the gRPC server.
// Active RPCs will be terminated abruptly.
// This should only be used when graceful shutdown is not required.
func (s *Server) Stop() {
	s.grpcServer.Stop()
}

// GracefulStop gracefully stops the gRPC server.
// It stops accepting new connections and waits for active RPCs
// to complete within the configured timeout period.
func (s *Server) GracefulStop() {
	// Create a timeout context for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulTimeout)
	defer cancel()

	// Channel to signal graceful stop completion
	done := make(chan struct{})

	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-done:
		// Graceful stop completed
		fmt.Println("Server stopped gracefully")
	case <-ctx.Done():
		// Timeout - force stop
		fmt.Println("Graceful shutdown timeout, forcing stop")
		s.grpcServer.Stop()
	}
}

// Port returns the port the server is listening on.
// This is useful when using port 0 to get an available port.
func (s *Server) Port() int {
	if s.listener != nil {
		if addr, ok := s.listener.Addr().(*net.TCPAddr); ok {
			return addr.Port
		}
	}
	return s.config.Port
}
