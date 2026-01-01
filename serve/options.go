package serve

import (
	"context"
	"fmt"
	"time"

	"github.com/zero-day-ai/sdk/registry"
)

// Option is a functional option for configuring a Server.
// Options provide a flexible way to customize server behavior
// without requiring a large number of constructor parameters.
type Option func(*Config)

// WithPort sets the TCP port for the gRPC server.
// The port must be between 1 and 65535.
// Use port 0 to automatically select an available port.
//
// Example:
//
//	serve.Agent(myAgent, serve.WithPort(8080))
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithHealthEndpoint sets the path for the health check endpoint.
// This is reserved for future HTTP health endpoint support.
// Currently, health checks are exposed via gRPC health checking protocol.
//
// Example:
//
//	serve.Agent(myAgent, serve.WithHealthEndpoint("/healthz"))
func WithHealthEndpoint(path string) Option {
	return func(c *Config) {
		c.HealthEndpoint = path
	}
}

// WithGracefulShutdown sets the maximum duration to wait for active
// requests to complete during graceful shutdown.
// After this timeout, the server will force shutdown.
//
// A longer timeout gives more time for long-running requests to complete,
// but delays shutdown. A shorter timeout causes faster shutdown but may
// interrupt active requests.
//
// Example:
//
//	serve.Agent(myAgent, serve.WithGracefulShutdown(60*time.Second))
func WithGracefulShutdown(timeout time.Duration) Option {
	return func(c *Config) {
		c.GracefulTimeout = timeout
	}
}

// WithTLS enables TLS encryption for the gRPC server.
// Both certFile and keyFile must be valid paths to PEM-encoded files.
// If either path is empty, TLS will be disabled.
//
// The certificate file should contain the server's certificate chain.
// The key file should contain the server's private key.
//
// Example:
//
//	serve.Agent(myAgent, serve.WithTLS("/etc/certs/server.crt", "/etc/certs/server.key"))
func WithTLS(certFile, keyFile string) Option {
	return func(c *Config) {
		c.TLSCertFile = certFile
		c.TLSKeyFile = keyFile
	}
}

// WithLocalMode enables Unix domain socket listening alongside TCP.
// The server will create a Unix socket at the specified path with 0600 permissions
// (owner read/write only) for secure local IPC communication.
// The socket is automatically cleaned up on server shutdown.
//
// This is useful for local component deployment where the Gibson framework
// needs to communicate with agents/tools via Unix sockets for better performance
// and security compared to TCP localhost connections.
//
// Example:
//
//	serve.Agent(myAgent, serve.WithLocalMode("/var/run/gibson/agents/my-agent.sock"))
func WithLocalMode(socketPath string) Option {
	return func(c *Config) {
		c.LocalMode = socketPath
	}
}

// WithRegistry enables automatic service registration with the provided registry.
// When configured, components will automatically:
//   - Register themselves with the registry after the gRPC server starts
//   - Maintain their registration via lease keepalives
//   - Deregister during graceful shutdown
//
// The registry parameter must implement the Register, Deregister, and Close methods.
// Typically, this is a registry.Registry from the SDK's registry package.
//
// Example:
//
//	reg, _ := registry.NewClient(config)
//	serve.Agent(myAgent, serve.WithRegistry(reg))
func WithRegistry(reg interface {
	Register(ctx context.Context, info interface{}) error
	Deregister(ctx context.Context, info interface{}) error
	Close() error
}) Option {
	return func(c *Config) {
		c.Registry = reg
	}
}

// WithRegistryFromEnv creates a registry client from the GIBSON_REGISTRY_ENDPOINTS
// environment variable. If the env var is not set, registration is skipped silently
// (the component works but isn't registered).
//
// The env var should be comma-separated endpoints: "localhost:2379,etcd2:2379"
//
// This is a convenience wrapper around WithRegistry that automatically creates
// a registry client from environment configuration. It's the easiest way for
// components to support optional service discovery without requiring explicit
// configuration.
//
// Example:
//
//	serve.Agent(myAgent, serve.WithRegistryFromEnv())
//
// If GIBSON_REGISTRY_ENDPOINTS is set, the component will register itself.
// If not set, the component will work normally but won't be discoverable.
func WithRegistryFromEnv() Option {
	return func(c *Config) {
		client, err := registry.NewClientFromEnv()
		if err != nil {
			// Log warning but don't fail - component should still work
			// In a production system, you might want to use a proper logger here
			return
		}
		if client != nil {
			// Wrap the client in an adapter to match the expected interface
			c.Registry = &registryAdapter{client: client}
		}
		// If client is nil (env not set), leave c.Registry unset
	}
}

// registryAdapter adapts registry.Client to the generic interface expected by Config.
// This is needed because the agent/tool/plugin serve functions pass map[string]interface{}
// as service info, but the registry.Client expects registry.ServiceInfo.
type registryAdapter struct {
	client *registry.Client
}

func (r *registryAdapter) Register(ctx context.Context, info interface{}) error {
	// Convert map to ServiceInfo
	serviceInfo, err := mapToServiceInfo(info)
	if err != nil {
		return err
	}
	return r.client.Register(ctx, serviceInfo)
}

func (r *registryAdapter) Deregister(ctx context.Context, info interface{}) error {
	// Convert map to ServiceInfo
	serviceInfo, err := mapToServiceInfo(info)
	if err != nil {
		return err
	}
	return r.client.Deregister(ctx, serviceInfo)
}

func (r *registryAdapter) Close() error {
	return r.client.Close()
}

// mapToServiceInfo converts a map[string]interface{} to registry.ServiceInfo.
// This is used to bridge between the generic interface{} used in serve functions
// and the typed ServiceInfo expected by the registry client.
func mapToServiceInfo(info interface{}) (registry.ServiceInfo, error) {
	m, ok := info.(map[string]interface{})
	if !ok {
		return registry.ServiceInfo{}, fmt.Errorf("expected map[string]interface{}, got %T", info)
	}

	// Extract fields from map
	serviceInfo := registry.ServiceInfo{}

	if v, ok := m["kind"].(string); ok {
		serviceInfo.Kind = v
	}
	if v, ok := m["name"].(string); ok {
		serviceInfo.Name = v
	}
	if v, ok := m["version"].(string); ok {
		serviceInfo.Version = v
	}
	if v, ok := m["instance_id"].(string); ok {
		serviceInfo.InstanceID = v
	}
	if v, ok := m["endpoint"].(string); ok {
		serviceInfo.Endpoint = v
	}
	if v, ok := m["metadata"].(map[string]string); ok {
		serviceInfo.Metadata = v
	}
	if v, ok := m["started_at"].(time.Time); ok {
		serviceInfo.StartedAt = v
	}

	return serviceInfo, nil
}
