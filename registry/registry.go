// Package registry provides service discovery and registration interfaces for Gibson components.
//
// The registry enables dynamic service discovery by allowing agents, tools, and plugins
// to register themselves at runtime. Gibson supports two registry modes:
//
//   - Embedded: Zero-ops local development using in-process etcd
//   - External: Production-grade etcd cluster for distributed deployments
//
// Components use the Registry interface to register on startup, maintain presence via
// lease keepalives, and deregister on graceful shutdown. The registry serves as the
// single source of truth for component discovery, replacing legacy PID files and
// socket-based discovery.
package registry

import (
	"context"
	"time"
)

// ServiceInfo describes a registered service instance.
//
// Each running component (agent, tool, or plugin) registers a ServiceInfo entry
// that includes identifying information, network endpoint, and custom metadata.
// Multiple instances of the same component can run simultaneously, each with
// a unique InstanceID.
type ServiceInfo struct {
	// Kind identifies the component type: "agent", "tool", or "plugin"
	Kind string `json:"kind"`

	// Name is the component name (e.g., "k8skiller", "nmap", "davinci")
	Name string `json:"name"`

	// Version is the semantic version of the component (e.g., "1.2.3")
	Version string `json:"version"`

	// InstanceID is a unique identifier for this specific instance (typically UUID)
	// This allows multiple instances of the same component to run concurrently
	InstanceID string `json:"instance_id"`

	// Endpoint is the network address where this component can be reached
	// Format: "host:port" for TCP (e.g., "localhost:50051")
	//         "unix:///path/to/socket" for Unix domain sockets
	Endpoint string `json:"endpoint"`

	// Metadata contains component-specific attributes such as:
	//   - capabilities: comma-separated list (e.g., "prompt_injection,jailbreak")
	//   - target_types: supported target types (e.g., "llm_chat,llm_api")
	//   - mitre_techniques: ATT&CK technique IDs
	//   - any other custom key-value pairs
	Metadata map[string]string `json:"metadata"`

	// StartedAt is the timestamp when this instance started
	StartedAt time.Time `json:"started_at"`
}

// Registry defines the service registration and discovery interface.
//
// Implementations must provide thread-safe access to service registration,
// discovery, and watch capabilities. The registry uses etcd leases with TTL
// to automatically remove stale entries when components crash or disconnect.
//
// Example usage:
//
//	reg, _ := registry.NewClient(config)
//	defer reg.Close()
//
//	info := ServiceInfo{
//	    Kind:       "agent",
//	    Name:       "davinci",
//	    Version:    "1.0.0",
//	    InstanceID: uuid.New().String(),
//	    Endpoint:   "localhost:50051",
//	    Metadata:   map[string]string{"capabilities": "jailbreak"},
//	    StartedAt:  time.Now(),
//	}
//
//	reg.Register(ctx, info)
//	defer reg.Deregister(ctx, info)
type Registry interface {
	// Register adds this service instance to the registry.
	//
	// The service will be discoverable by other components immediately.
	// The implementation must create an etcd lease with the configured TTL
	// and associate the service entry with that lease. A background goroutine
	// should renew the lease periodically (typically every TTL/3).
	//
	// If the service instance is already registered (same InstanceID), this
	// should update the existing entry rather than creating a duplicate.
	//
	// Returns an error if the registry is unavailable or if the lease cannot
	// be created.
	Register(ctx context.Context, info ServiceInfo) error

	// Deregister removes this service instance from the registry.
	//
	// This should be called during graceful shutdown to immediately remove
	// the service from discovery. The implementation should revoke the
	// associated etcd lease, which will delete the service entry.
	//
	// If the service is not registered, this is a no-op (not an error).
	//
	// Returns an error if the registry is unavailable.
	Deregister(ctx context.Context, info ServiceInfo) error

	// Discover finds all instances of a service by kind and name.
	//
	// For example, to find all instances of the "k8skiller" agent:
	//   instances, _ := reg.Discover(ctx, "agent", "k8skiller")
	//
	// The returned slice may be empty if no instances are currently registered.
	// Instances are returned in arbitrary order (use DiscoverWithLoadBalancing
	// if ordering/selection strategy is important).
	//
	// Returns an error if the registry is unavailable or if the query fails.
	Discover(ctx context.Context, kind, name string) ([]ServiceInfo, error)

	// DiscoverAll finds all instances of a given kind.
	//
	// For example, to find all registered agents:
	//   agents, _ := reg.DiscoverAll(ctx, "agent")
	//
	// This is useful for status displays and dashboards that want to show
	// all available components of a particular type.
	//
	// The returned slice may be empty if no instances are registered.
	//
	// Returns an error if the registry is unavailable or if the query fails.
	DiscoverAll(ctx context.Context, kind string) ([]ServiceInfo, error)

	// Watch returns a channel that receives updates when services change.
	//
	// The channel will emit the current list of instances whenever:
	//   - A new instance registers
	//   - An existing instance deregisters
	//   - An instance's lease expires (crashed or disconnected)
	//
	// The initial state is sent immediately upon calling Watch. Subsequent
	// updates are sent as changes occur.
	//
	// The watch is scoped to a specific kind and name. To watch all agents,
	// call Watch multiple times or use DiscoverAll periodically.
	//
	// The channel is closed when:
	//   - The provided context is canceled
	//   - Close() is called on the registry
	//   - An unrecoverable error occurs
	//
	// Example:
	//   ch, _ := reg.Watch(ctx, "agent", "davinci")
	//   for instances := range ch {
	//       log.Printf("Davinci agents: %d instances", len(instances))
	//   }
	//
	// Returns an error if the registry is unavailable or if the watch cannot
	// be established.
	Watch(ctx context.Context, kind, name string) (<-chan []ServiceInfo, error)

	// Close releases registry resources and stops all background goroutines.
	//
	// This should be called during application shutdown. After Close() is called,
	// all other methods will return errors.
	//
	// For embedded registries, this will gracefully stop the in-process etcd server.
	// For external registries, this will close the etcd client connection.
	//
	// All active watches will be terminated and their channels closed.
	//
	// Returns an error if cleanup fails, though this is typically ignored during
	// shutdown.
	Close() error
}

// Config holds registry connection configuration.
//
// The registry can operate in two modes determined by the Type field:
//
//  1. Embedded mode (Type="embedded"):
//     - Starts an in-process etcd server using go.etcd.io/etcd/server/v3/embed
//     - Zero external dependencies, perfect for local development
//     - Data persists to DataDir (default: ~/.gibson/etcd-data)
//     - Listens on ListenAddress (default: localhost:2379)
//
//  2. External mode (Type="etcd"):
//     - Connects to an external etcd cluster
//     - Production-grade deployment with HA support
//     - Requires Endpoints to be configured
//     - Optionally uses TLS for secure communication
type Config struct {
	// Type specifies the registry mode: "embedded" or "etcd"
	// Default: "embedded"
	Type string `json:"type"`

	// Endpoints is the list of etcd endpoints for external mode
	// Format: ["host1:2379", "host2:2379", "host3:2379"]
	// Required if Type="etcd", ignored if Type="embedded"
	Endpoints []string `json:"endpoints"`

	// Namespace is the etcd key prefix for all Gibson service entries
	// All services are stored under /{namespace}/{kind}/{name}/{instance-id}
	// Default: "gibson"
	Namespace string `json:"namespace"`

	// TTL is the lease time-to-live in seconds
	// Services must renew their lease within this interval or be removed
	// Default: 30 seconds
	// Recommended: 15-60 seconds depending on failure detection requirements
	TTL int `json:"ttl"`

	// DataDir is the directory where embedded etcd persists data
	// Only used if Type="embedded"
	// Default: "~/.gibson/etcd-data"
	DataDir string `json:"data_dir"`

	// ListenAddress is the address where embedded etcd listens for clients
	// Only used if Type="embedded"
	// Default: "localhost:2379"
	// Format: "host:port"
	ListenAddress string `json:"listen_address"`

	// TLS holds TLS configuration for secure etcd communication
	// Optional for both embedded and external modes
	// If nil, TLS is disabled
	TLS *TLSConfig `json:"tls"`
}

// TLSConfig holds TLS certificate configuration for secure registry communication.
//
// When TLS is enabled, all communication with etcd is encrypted and authenticated
// using mutual TLS (mTLS). This is strongly recommended for production deployments.
type TLSConfig struct {
	// Enabled determines whether TLS is active
	// If false, all other fields are ignored
	Enabled bool `json:"enabled"`

	// CertFile is the path to the client certificate file (PEM format)
	// Required if Enabled=true
	CertFile string `json:"cert_file"`

	// KeyFile is the path to the client private key file (PEM format)
	// Required if Enabled=true
	KeyFile string `json:"key_file"`

	// CAFile is the path to the certificate authority file (PEM format)
	// Used to verify the etcd server's certificate
	// Required if Enabled=true
	CAFile string `json:"ca_file"`
}
