package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client implements Registry for SDK components to self-register with etcd.
//
// The client connects to an etcd cluster (either embedded or external) and
// provides service registration, discovery, and watch capabilities. It handles
// lease management automatically, renewing leases every TTL/3 to maintain
// service presence.
//
// Example usage:
//
//	cfg := registry.Config{
//	    Endpoints: []string{"localhost:2379"},
//	    Namespace: "gibson",
//	    TTL:       30,
//	}
//	client, err := registry.NewClient(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
// Thread-safety: All methods are safe for concurrent use.
type Client struct {
	client    *clientv3.Client
	namespace string
	ttl       int

	// Lease tracking for keepalive
	mu         sync.RWMutex
	leases     map[string]clientv3.LeaseID // key: instance ID, value: lease ID
	cancelFns  map[string]context.CancelFunc
	wg         sync.WaitGroup // tracks background goroutines
	closed     bool
	closedChan chan struct{}
}

// NewClient creates a registry client from the provided configuration.
//
// This establishes a connection to the etcd cluster and verifies connectivity
// by performing a health check. If the connection fails, an error is returned.
//
// The client must be closed using Close() when no longer needed to release
// resources and stop background keepalive goroutines.
//
// Parameters:
//   - cfg: Configuration containing endpoints, namespace, TTL, and TLS settings
//
// Returns:
//   - *Client: A connected registry client
//   - error: Connection error if etcd is unreachable or authentication fails
func NewClient(cfg Config) (*Client, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("registry endpoints cannot be empty")
	}

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = "gibson"
	}

	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 30
	}

	// Build etcd client config
	clientCfg := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: 5 * time.Second,
	}

	// Configure TLS if enabled
	if cfg.TLS != nil && cfg.TLS.Enabled {
		tlsInfo, err := newTLSInfo(cfg.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		clientCfg.TLS = tlsConfig
	}

	// Create etcd client
	cli, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	// Verify connectivity with a quick health check
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err = cli.Get(ctx, "health-check")
	if err != nil && err != context.DeadlineExceeded {
		cli.Close()
		return nil, fmt.Errorf("etcd health check failed: %w", err)
	}

	return &Client{
		client:     cli,
		namespace:  namespace,
		ttl:        ttl,
		leases:     make(map[string]clientv3.LeaseID),
		cancelFns:  make(map[string]context.CancelFunc),
		closedChan: make(chan struct{}),
	}, nil
}

// NewClientFromEnv creates a registry client using the GIBSON_REGISTRY_ENDPOINTS
// environment variable.
//
// The environment variable should contain a comma-separated list of etcd endpoints:
//   GIBSON_REGISTRY_ENDPOINTS=localhost:2379,localhost:2380,localhost:2381
//
// If the environment variable is not set, this function returns (nil, nil) to
// allow components to work gracefully without registry integration. This is NOT
// considered an error - the component will function but won't be discoverable.
//
// Returns:
//   - *Client: A connected registry client, or nil if env var not set
//   - error: Connection error if env var is set but connection fails
func NewClientFromEnv() (*Client, error) {
	endpoints := os.Getenv("GIBSON_REGISTRY_ENDPOINTS")
	if endpoints == "" {
		// Not an error - component works but isn't registered
		return nil, nil
	}

	// Parse comma-separated endpoints
	endpointList := strings.Split(endpoints, ",")
	for i, ep := range endpointList {
		endpointList[i] = strings.TrimSpace(ep)
	}

	// Create config with defaults
	cfg := Config{
		Endpoints: endpointList,
		Namespace: "gibson",
		TTL:       30,
	}

	return NewClient(cfg)
}

// Register adds this service instance to the registry.
//
// The service will be discoverable immediately and will remain registered as
// long as the lease is kept alive. A background goroutine is started to renew
// the lease every TTL/3 seconds.
//
// If the service instance is already registered (same InstanceID), this updates
// the existing entry and restarts the keepalive goroutine.
//
// Parameters:
//   - ctx: Context for the registration operation (not for keepalive)
//   - info: Service information to register
//
// Returns:
//   - error: Registration error if etcd is unavailable or lease creation fails
func (c *Client) Register(ctx context.Context, info ServiceInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("registry client is closed")
	}

	// Cancel existing keepalive if re-registering
	if cancelFn, exists := c.cancelFns[info.InstanceID]; exists {
		cancelFn()
		delete(c.cancelFns, info.InstanceID)
	}

	// Create lease with TTL
	leaseResp, err := c.client.Grant(ctx, int64(c.ttl))
	if err != nil {
		return fmt.Errorf("failed to create lease: %w", err)
	}

	// Serialize service info to JSON
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("failed to marshal service info: %w", err)
	}

	// Build registry key: /namespace/kind/name/instance-id
	key := c.buildKey(info.Kind, info.Name, info.InstanceID)

	// Put service info with lease
	_, err = c.client.Put(ctx, key, string(data), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// Store lease ID
	c.leases[info.InstanceID] = leaseResp.ID

	// Start keepalive goroutine
	keepaliveCtx, cancel := context.WithCancel(context.Background())
	c.cancelFns[info.InstanceID] = cancel

	c.wg.Add(1)
	go c.keepalive(keepaliveCtx, leaseResp.ID, info.InstanceID)

	return nil
}

// Deregister removes this service instance from the registry.
//
// This revokes the etcd lease, which immediately deletes the service entry.
// The background keepalive goroutine is stopped.
//
// If the service is not registered, this is a no-op (not an error).
//
// Parameters:
//   - ctx: Context for the deregistration operation
//   - info: Service information to deregister (InstanceID is used)
//
// Returns:
//   - error: Deregistration error if etcd is unavailable
func (c *Client) Deregister(ctx context.Context, info ServiceInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("registry client is closed")
	}

	// Stop keepalive goroutine
	if cancelFn, exists := c.cancelFns[info.InstanceID]; exists {
		cancelFn()
		delete(c.cancelFns, info.InstanceID)
	}

	// Revoke lease (deletes the service entry)
	leaseID, exists := c.leases[info.InstanceID]
	if !exists {
		// Not registered, this is a no-op
		return nil
	}

	_, err := c.client.Revoke(ctx, leaseID)
	if err != nil {
		return fmt.Errorf("failed to revoke lease: %w", err)
	}

	delete(c.leases, info.InstanceID)

	return nil
}

// Discover finds all instances of a service by kind and name.
//
// Returns all currently registered instances. The slice may be empty if no
// instances are running. Instances are returned in arbitrary order.
//
// Parameters:
//   - ctx: Context for the discovery operation
//   - kind: Component type ("agent", "tool", or "plugin")
//   - name: Component name (e.g., "k8skiller", "nmap")
//
// Returns:
//   - []ServiceInfo: List of registered instances
//   - error: Discovery error if etcd is unavailable
func (c *Client) Discover(ctx context.Context, kind, name string) ([]ServiceInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("registry client is closed")
	}

	// Query all instances: /namespace/kind/name/
	prefix := fmt.Sprintf("/%s/%s/%s/", c.namespace, kind, name)

	resp, err := c.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	// Parse service info from each entry
	instances := make([]ServiceInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info ServiceInfo
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			// Skip invalid entries
			continue
		}
		instances = append(instances, info)
	}

	return instances, nil
}

// DiscoverAll finds all instances of a given kind.
//
// This is useful for status displays that want to show all agents, all tools,
// or all plugins currently registered.
//
// Parameters:
//   - ctx: Context for the discovery operation
//   - kind: Component type ("agent", "tool", or "plugin")
//
// Returns:
//   - []ServiceInfo: List of all registered instances of this kind
//   - error: Discovery error if etcd is unavailable
func (c *Client) DiscoverAll(ctx context.Context, kind string) ([]ServiceInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("registry client is closed")
	}

	// Query all instances of this kind: /namespace/kind/
	prefix := fmt.Sprintf("/%s/%s/", c.namespace, kind)

	resp, err := c.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to discover all services: %w", err)
	}

	// Parse service info from each entry
	instances := make([]ServiceInfo, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info ServiceInfo
		if err := json.Unmarshal(kv.Value, &info); err != nil {
			// Skip invalid entries
			continue
		}
		instances = append(instances, info)
	}

	return instances, nil
}

// Watch returns a channel that receives updates when services change.
//
// The channel emits the current list of instances whenever a service registers,
// deregisters, or its lease expires. The initial state is sent immediately.
//
// The channel is closed when the context is canceled or Close() is called.
//
// Parameters:
//   - ctx: Context that controls the watch lifetime
//   - kind: Component type to watch
//   - name: Component name to watch
//
// Returns:
//   - <-chan []ServiceInfo: Channel that receives instance updates
//   - error: Watch error if etcd is unavailable
func (c *Client) Watch(ctx context.Context, kind, name string) (<-chan []ServiceInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.closed {
		return nil, fmt.Errorf("registry client is closed")
	}

	ch := make(chan []ServiceInfo, 1)

	// Send initial state
	instances, err := c.Discover(ctx, kind, name)
	if err != nil {
		return nil, err
	}
	ch <- instances

	// Watch for changes
	prefix := fmt.Sprintf("/%s/%s/%s/", c.namespace, kind, name)
	watchChan := c.client.Watch(ctx, prefix, clientv3.WithPrefix())

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-c.closedChan:
				return
			case watchResp, ok := <-watchChan:
				if !ok {
					return
				}
				if watchResp.Err() != nil {
					return
				}

				// Fetch current state after any change
				instances, err := c.Discover(context.Background(), kind, name)
				if err != nil {
					// Skip this update if we can't query
					continue
				}

				select {
				case ch <- instances:
				case <-ctx.Done():
					return
				case <-c.closedChan:
					return
				}
			}
		}
	}()

	return ch, nil
}

// Close releases all resources and stops background goroutines.
//
// After Close() is called, all other methods will return errors. All active
// watches are terminated and their channels closed. All keepalive goroutines
// are stopped.
//
// Returns:
//   - error: Cleanup error, typically ignored during shutdown
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true

	// Cancel all keepalive goroutines
	for _, cancel := range c.cancelFns {
		cancel()
	}
	c.cancelFns = make(map[string]context.CancelFunc)

	close(c.closedChan)
	c.mu.Unlock()

	// Wait for all goroutines to finish
	c.wg.Wait()

	// Close etcd client
	return c.client.Close()
}

// keepalive renews the lease every TTL/3 seconds to maintain service presence.
//
// This runs in a background goroutine started by Register(). It stops when:
//   - The context is canceled (via Deregister or Close)
//   - The lease becomes invalid
//   - An unrecoverable error occurs
func (c *Client) keepalive(ctx context.Context, leaseID clientv3.LeaseID, instanceID string) {
	defer c.wg.Done()

	// Renew every TTL/3 seconds
	interval := time.Duration(c.ttl) * time.Second / 3
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.closedChan:
			return
		case <-ticker.C:
			_, err := c.client.KeepAliveOnce(context.Background(), leaseID)
			if err != nil {
				// Lease is invalid, stop keepalive
				c.mu.Lock()
				delete(c.leases, instanceID)
				delete(c.cancelFns, instanceID)
				c.mu.Unlock()
				return
			}
		}
	}
}

// buildKey constructs the etcd key for a service instance.
//
// Format: /namespace/kind/name/instance-id
func (c *Client) buildKey(kind, name, instanceID string) string {
	return fmt.Sprintf("/%s/%s/%s/%s", c.namespace, kind, name, instanceID)
}
