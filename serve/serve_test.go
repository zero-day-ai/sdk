package serve

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
