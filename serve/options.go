package serve

import "time"

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
