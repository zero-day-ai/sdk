package serve

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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

	// LocalMode enables Unix domain socket listening alongside TCP.
	// When enabled, the server creates a Unix socket at the specified path
	// for local IPC communication. The socket is created with 0600 permissions
	// and cleaned up on server shutdown.
	// If empty, LocalMode is disabled.
	LocalMode string

	// AdvertiseAddr is the address to advertise to the registry for other
	// components to connect to this service. This is useful in containerized
	// environments where the hostname differs from localhost.
	// Format: "hostname:port" or just "hostname" (port will be appended).
	// If empty, defaults to "localhost:{port}".
	// Can be set via GIBSON_ADVERTISE_ADDR environment variable.
	AdvertiseAddr string

	// Registry is the optional registry for service registration.
	// If provided, components will register themselves on startup
	// and deregister on shutdown.
	Registry interface {
		Register(ctx context.Context, info interface{}) error
		Deregister(ctx context.Context, info interface{}) error
		Close() error
	}
}

// DefaultConfig returns default serve configuration.
// These defaults are suitable for local development and testing.
//
// Port resolution order:
//  1. --port CLI flag (if present)
//  2. GIBSON_PORT environment variable
//  3. Default: 50051
func DefaultConfig() *Config {
	port := 50051

	// Check for --port CLI flag first
	if cliPort := getPortFromCLI(); cliPort > 0 {
		port = cliPort
	} else if envPort := os.Getenv("GIBSON_PORT"); envPort != "" {
		// Check GIBSON_PORT env var
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			port = p
		}
	}

	return &Config{
		Port:            port,
		HealthEndpoint:  "/health",
		GracefulTimeout: 30 * time.Second,
	}
}

// getPortFromCLI parses --port flag from command line arguments.
// Returns 0 if not found or invalid.
func getPortFromCLI() int {
	for i, arg := range os.Args[1:] {
		if arg == "--port" && i+1 < len(os.Args[1:]) {
			if p, err := strconv.Atoi(os.Args[i+2]); err == nil && p > 0 {
				return p
			}
		}
		// Handle --port=8080 format
		if len(arg) > 7 && arg[:7] == "--port=" {
			if p, err := strconv.Atoi(arg[7:]); err == nil && p > 0 {
				return p
			}
		}
	}
	return 0
}

// Server wraps a gRPC server with lifecycle management.
// It handles server initialization, startup, graceful shutdown,
// and health check registration.
type Server struct {
	grpcServer     *grpc.Server
	listener       net.Listener
	unixListener   net.Listener // Optional Unix domain socket listener for LocalMode
	config         *Config
	healthServer   *health.Server
	unixSocketPath string // Path to Unix socket for cleanup
}

// NewServer creates a new gRPC server with the provided configuration.
// It sets up the gRPC server with appropriate options (e.g., TLS)
// and registers the health check service.
func NewServer(cfg *Config) (*Server, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create TCP listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", cfg.Port, err)
	}

	// Create Unix socket listener if LocalMode is enabled
	var unixListener net.Listener
	var unixSocketPath string
	if cfg.LocalMode != "" {
		// Create parent directory if it doesn't exist
		socketDir := filepath.Dir(cfg.LocalMode)
		if err := os.MkdirAll(socketDir, 0755); err != nil {
			listener.Close()
			return nil, fmt.Errorf("failed to create socket directory %s: %w", socketDir, err)
		}

		// Remove existing socket if it exists
		if err := os.Remove(cfg.LocalMode); err != nil && !os.IsNotExist(err) {
			listener.Close()
			return nil, fmt.Errorf("failed to remove existing socket %s: %w", cfg.LocalMode, err)
		}

		// Create Unix domain socket with restricted permissions
		unixListener, err = net.Listen("unix", cfg.LocalMode)
		if err != nil {
			listener.Close()
			return nil, fmt.Errorf("failed to create unix socket at %s: %w", cfg.LocalMode, err)
		}

		// Set socket permissions to 0600 (owner read/write only)
		if err := os.Chmod(cfg.LocalMode, 0600); err != nil {
			listener.Close()
			unixListener.Close()
			os.Remove(cfg.LocalMode)
			return nil, fmt.Errorf("failed to set socket permissions: %w", err)
		}

		unixSocketPath = cfg.LocalMode
	}

	// Build gRPC server options
	var opts []grpc.ServerOption

	// Configure TLS if cert and key are provided
	if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			listener.Close()
			if unixListener != nil {
				unixListener.Close()
				os.Remove(unixSocketPath)
			}
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
		grpcServer:     grpcServer,
		listener:       listener,
		unixListener:   unixListener,
		config:         cfg,
		healthServer:   healthServer,
		unixSocketPath: unixSocketPath,
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
// When LocalMode is enabled, the server listens on both TCP and Unix socket.
func (s *Server) Serve(ctx context.Context) error {
	// Create error channel for serve errors (buffer size 2 for TCP and Unix listeners)
	errCh := make(chan error, 2)

	// Start serving on TCP listener
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			errCh <- fmt.Errorf("gRPC TCP server error: %w", err)
		}
	}()

	// Start serving on Unix socket if LocalMode is enabled
	if s.unixListener != nil {
		go func() {
			if err := s.grpcServer.Serve(s.unixListener); err != nil {
				errCh <- fmt.Errorf("gRPC Unix socket server error: %w", err)
			}
		}()
	}

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
		_ = sig // Signal received, shutting down gracefully
		s.GracefulStop()
		return nil
	case err := <-errCh:
		// Server error
		s.cleanup()
		return err
	}
}

// Stop immediately stops the gRPC server.
// Active RPCs will be terminated abruptly.
// This should only be used when graceful shutdown is not required.
func (s *Server) Stop() {
	s.grpcServer.Stop()
	s.cleanup()
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
	case <-ctx.Done():
		// Timeout - force stop
		s.grpcServer.Stop()
	}

	// Clean up Unix socket
	s.cleanup()
}

// cleanup removes the Unix socket file if it exists.
// This is called during server shutdown to prevent stale socket files.
func (s *Server) cleanup() {
	if s.unixSocketPath != "" {
		// Attempt to remove Unix socket, ignore NotExist errors
		_ = os.Remove(s.unixSocketPath)
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
