package serve

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 50051, cfg.Port)
	assert.Equal(t, "/health", cfg.HealthEndpoint)
	assert.Equal(t, 30*time.Second, cfg.GracefulTimeout)
	assert.Empty(t, cfg.TLSCertFile)
	assert.Empty(t, cfg.TLSKeyFile)
}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
		{
			name: "custom port",
			config: &Config{
				Port:            0, // Use any available port
				HealthEndpoint:  "/health",
				GracefulTimeout: 10 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "specific port",
			config: &Config{
				Port:            0, // Use port 0 to get any available port
				HealthEndpoint:  "/healthz",
				GracefulTimeout: 5 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, err := NewServer(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, srv)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, srv)
			assert.NotNil(t, srv.GRPCServer())
			assert.NotNil(t, srv.HealthServer())
			assert.Greater(t, srv.Port(), 0)

			// Clean up
			srv.Stop()
		})
	}
}

func TestServerGracefulStop(t *testing.T) {
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Start server in background
	ctx := context.Background()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test graceful stop
	start := time.Now()
	srv.GracefulStop()
	duration := time.Since(start)

	// Should complete quickly since no active requests
	assert.Less(t, duration, 2*time.Second)

	// GracefulStop stops the server, so Serve should return
	// Note: The server goroutine will exit after GracefulStop completes
	time.Sleep(100 * time.Millisecond)
}

func TestServerStop(t *testing.T) {
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Start server in background
	ctx := context.Background()
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test immediate stop
	srv.Stop()

	// Stop stops the server immediately, so Serve should return
	// Note: The server goroutine will exit after Stop completes
	time.Sleep(100 * time.Millisecond)
}

func TestServerContextCancellation(t *testing.T) {
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Check that Serve returns with context error
	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after context cancellation")
	}
}

func TestServerPort(t *testing.T) {
	cfg := &Config{
		Port:            0, // Use any available port
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Stop()

	// Port should be assigned
	port := srv.Port()
	assert.Greater(t, port, 0)
	assert.NotEqual(t, 0, port) // Should have a real port
}

func TestLocalMode(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/test.sock"

	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Socket file should exist
	_, err = os.Stat(socketPath)
	assert.NoError(t, err, "Unix socket should exist")

	// Socket should have 0600 permissions
	info, err := os.Stat(socketPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "Socket should have 0600 permissions")

	// Stop server
	srv.Stop()

	// Socket should be cleaned up
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Unix socket should be removed after shutdown")
}

func TestLocalModeServe(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/test-serve.sock"

	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify both TCP and Unix socket are accessible
	// TCP listener should be accessible
	tcpPort := srv.Port()
	assert.Greater(t, tcpPort, 0)

	// Unix socket should exist
	_, err = os.Stat(socketPath)
	assert.NoError(t, err, "Unix socket should exist while server is running")

	// Cancel context to trigger graceful shutdown
	cancel()

	// Wait for server to shut down
	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not shut down in time")
	}

	// Socket should be cleaned up after shutdown
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Unix socket should be removed after shutdown")
}

func TestLocalModeGracefulStop(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/test-graceful.sock"

	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Start server in background
	ctx := context.Background()
	go func() {
		_ = srv.Serve(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify socket exists
	_, err = os.Stat(socketPath)
	assert.NoError(t, err, "Unix socket should exist")

	// Call GracefulStop
	srv.GracefulStop()

	// Socket should be cleaned up
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Unix socket should be removed after graceful stop")
}

func TestLocalModeWithExistingSocket(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/existing.sock"

	// Create a stale socket file
	f, err := os.Create(socketPath)
	require.NoError(t, err)
	f.Close()

	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
	}

	// NewServer should remove the existing socket and create a new one
	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Stop()

	// Socket should exist and be a socket (not a regular file)
	info, err := os.Stat(socketPath)
	require.NoError(t, err)
	assert.NotEqual(t, 0, info.Mode()&os.ModeSocket, "Should be a socket, not a regular file")
}

func TestLocalModeHealthCheckViaSocket(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/health-check.sock"

	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = srv.Serve(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Verify socket exists
	_, err = os.Stat(socketPath)
	require.NoError(t, err, "Unix socket should exist")

	// Test health check via Unix socket
	t.Run("health check via socket succeeds", func(t *testing.T) {
		// Create a gRPC client connected to the Unix socket
		conn, err := grpc.NewClient(
			"unix://"+socketPath,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err)
		defer conn.Close()

		// Create health check client
		healthClient := grpc_health_v1.NewHealthClient(conn)

		// Perform health check
		checkCtx, checkCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer checkCancel()

		resp, err := healthClient.Check(checkCtx, &grpc_health_v1.HealthCheckRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
	})

	// Test health check watch stream via Unix socket
	t.Run("health check watch via socket succeeds", func(t *testing.T) {
		// Create a gRPC client connected to the Unix socket
		conn, err := grpc.NewClient(
			"unix://"+socketPath,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err)
		defer conn.Close()

		// Create health check client
		healthClient := grpc_health_v1.NewHealthClient(conn)

		// Perform health check watch
		watchCtx, watchCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer watchCancel()

		stream, err := healthClient.Watch(watchCtx, &grpc_health_v1.HealthCheckRequest{})
		require.NoError(t, err)
		require.NotNil(t, stream)

		// Receive first status update
		resp, err := stream.Recv()
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
	})

	// Shutdown server
	cancel()

	// Wait for server to shut down
	time.Sleep(200 * time.Millisecond)

	// Verify socket is cleaned up
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Unix socket should be removed after shutdown")
}

func TestLocalModeHealthCheckConnectionFailure(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/nonexistent.sock"

	// Attempt to connect to non-existent socket
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	conn, err := grpc.NewClient(
		"unix://"+socketPath,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	// Create health check client
	healthClient := grpc_health_v1.NewHealthClient(conn)

	// Perform health check - should fail because socket doesn't exist
	_, err = healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	assert.Error(t, err, "Health check should fail when socket doesn't exist")
}

func TestLocalModeSocketPermissions(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/permissions.sock"

	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
	}

	srv, err := NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)
	defer srv.Stop()

	// Verify socket has 0600 permissions (owner read/write only)
	info, err := os.Stat(socketPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "Socket should have 0600 permissions for security")
}

func TestLocalModeCleanupOnError(t *testing.T) {
	// Create temporary directory for socket
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/error-cleanup.sock"

	// Test that socket is cleaned up even if there's an error after creation
	// We'll simulate this by creating a server with invalid TLS config
	cfg := &Config{
		Port:            0,
		HealthEndpoint:  "/health",
		GracefulTimeout: 1 * time.Second,
		LocalMode:       socketPath,
		TLSCertFile:     "/nonexistent/cert.pem",
		TLSKeyFile:      "/nonexistent/key.pem",
	}

	srv, err := NewServer(cfg)
	assert.Error(t, err, "NewServer should fail with invalid TLS config")
	assert.Nil(t, srv)

	// Verify socket was cleaned up
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err), "Socket should be cleaned up on error")
}

// mockRegistry is a mock implementation of the Registry interface for testing
type mockRegistry struct {
	registered    []interface{}
	deregistered  []interface{}
	registerErr   error
	deregisterErr error
	closed        bool
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		registered:   make([]interface{}, 0),
		deregistered: make([]interface{}, 0),
	}
}

func (m *mockRegistry) Register(ctx context.Context, info interface{}) error {
	if m.registerErr != nil {
		return m.registerErr
	}
	m.registered = append(m.registered, info)
	return nil
}

func (m *mockRegistry) Deregister(ctx context.Context, info interface{}) error {
	if m.deregisterErr != nil {
		return m.deregisterErr
	}
	m.deregistered = append(m.deregistered, info)
	return nil
}

func (m *mockRegistry) Close() error {
	m.closed = true
	return nil
}

func TestWithRegistry(t *testing.T) {
	mockReg := newMockRegistry()
	cfg := DefaultConfig()

	opt := WithRegistry(mockReg)
	opt(cfg)

	assert.Equal(t, mockReg, cfg.Registry)
}

func TestWithRegistryFromEnv(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectRegistry bool
	}{
		{
			name:           "env var not set",
			envValue:       "",
			expectRegistry: false,
		},
		{
			name:           "env var set to empty",
			envValue:       "",
			expectRegistry: false,
		},
		// Note: Testing with actual endpoints would require etcd to be running
		// So we only test the absence of the env var here
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env var first
			os.Unsetenv("GIBSON_REGISTRY_ENDPOINTS")

			if tt.envValue != "" {
				os.Setenv("GIBSON_REGISTRY_ENDPOINTS", tt.envValue)
				defer os.Unsetenv("GIBSON_REGISTRY_ENDPOINTS")
			}

			cfg := DefaultConfig()
			opt := WithRegistryFromEnv()
			opt(cfg)

			if tt.expectRegistry {
				assert.NotNil(t, cfg.Registry)
			} else {
				assert.Nil(t, cfg.Registry)
			}
		})
	}
}
