package serve

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithPort(t *testing.T) {
	cfg := DefaultConfig()
	opt := WithPort(8080)
	opt(cfg)

	assert.Equal(t, 8080, cfg.Port)
}

func TestWithHealthEndpoint(t *testing.T) {
	cfg := DefaultConfig()
	opt := WithHealthEndpoint("/healthz")
	opt(cfg)

	assert.Equal(t, "/healthz", cfg.HealthEndpoint)
}

func TestWithGracefulShutdown(t *testing.T) {
	cfg := DefaultConfig()
	opt := WithGracefulShutdown(60 * time.Second)
	opt(cfg)

	assert.Equal(t, 60*time.Second, cfg.GracefulTimeout)
}

func TestWithTLS(t *testing.T) {
	cfg := DefaultConfig()
	opt := WithTLS("/etc/certs/server.crt", "/etc/certs/server.key")
	opt(cfg)

	assert.Equal(t, "/etc/certs/server.crt", cfg.TLSCertFile)
	assert.Equal(t, "/etc/certs/server.key", cfg.TLSKeyFile)
}

func TestMultipleOptions(t *testing.T) {
	cfg := DefaultConfig()

	opts := []Option{
		WithPort(9090),
		WithHealthEndpoint("/ready"),
		WithGracefulShutdown(45 * time.Second),
		WithTLS("cert.pem", "key.pem"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "/ready", cfg.HealthEndpoint)
	assert.Equal(t, 45*time.Second, cfg.GracefulTimeout)
	assert.Equal(t, "cert.pem", cfg.TLSCertFile)
	assert.Equal(t, "key.pem", cfg.TLSKeyFile)
}
