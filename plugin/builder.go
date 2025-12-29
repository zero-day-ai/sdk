package plugin

import (
	"context"
	"fmt"
	"sync"

	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

// MethodHandler is a function that handles a plugin method invocation.
// It receives the context and input parameters, and returns the result or an error.
type MethodHandler func(ctx context.Context, params map[string]any) (any, error)

// InitFunc is called to initialize the plugin with configuration.
type InitFunc func(ctx context.Context, config map[string]any) error

// ShutdownFunc is called to gracefully shutdown the plugin.
type ShutdownFunc func(ctx context.Context) error

// methodEntry represents a registered method with its descriptor and handler.
type methodEntry struct {
	descriptor MethodDescriptor
	handler    MethodHandler
}

// Config holds the configuration for building a plugin.
// Use NewConfig to create a new configuration, then use the setter methods
// to configure the plugin before calling New to build it.
type Config struct {
	name         string
	version      string
	description  string
	methods      []methodEntry
	initFunc     InitFunc
	shutdownFunc ShutdownFunc
}

// NewConfig creates a new plugin configuration with default values.
func NewConfig() *Config {
	return &Config{
		methods: make([]methodEntry, 0),
		initFunc: func(ctx context.Context, config map[string]any) error {
			return nil
		},
		shutdownFunc: func(ctx context.Context) error {
			return nil
		},
	}
}

// SetName sets the plugin name.
func (c *Config) SetName(name string) {
	c.name = name
}

// SetVersion sets the plugin version.
func (c *Config) SetVersion(version string) {
	c.version = version
}

// SetDescription sets the plugin description.
func (c *Config) SetDescription(desc string) {
	c.description = desc
}

// AddMethod registers a new method with the plugin.
// The method will be available for invocation via Query.
func (c *Config) AddMethod(name string, handler MethodHandler, inputSchema, outputSchema schema.JSON) {
	entry := methodEntry{
		descriptor: MethodDescriptor{
			Name:         name,
			Description:  "", // Can be enhanced to accept description parameter
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
		},
		handler: handler,
	}
	c.methods = append(c.methods, entry)
}

// AddMethodWithDesc registers a new method with the plugin including a description.
// The method will be available for invocation via Query.
func (c *Config) AddMethodWithDesc(name, description string, handler MethodHandler, inputSchema, outputSchema schema.JSON) {
	entry := methodEntry{
		descriptor: MethodDescriptor{
			Name:         name,
			Description:  description,
			InputSchema:  inputSchema,
			OutputSchema: outputSchema,
		},
		handler: handler,
	}
	c.methods = append(c.methods, entry)
}

// SetInitFunc sets the initialization function.
func (c *Config) SetInitFunc(fn InitFunc) {
	c.initFunc = fn
}

// SetShutdownFunc sets the shutdown function.
func (c *Config) SetShutdownFunc(fn ShutdownFunc) {
	c.shutdownFunc = fn
}

// New creates a new Plugin from the configuration.
// Returns an error if the configuration is invalid.
func New(cfg *Config) (Plugin, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if cfg.name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}

	if cfg.version == "" {
		return nil, fmt.Errorf("plugin version is required")
	}

	// Build method map for fast lookup
	methodMap := make(map[string]methodEntry)
	for _, entry := range cfg.methods {
		if entry.descriptor.Name == "" {
			return nil, fmt.Errorf("method name cannot be empty")
		}
		if _, exists := methodMap[entry.descriptor.Name]; exists {
			return nil, fmt.Errorf("duplicate method name: %s", entry.descriptor.Name)
		}
		methodMap[entry.descriptor.Name] = entry
	}

	return &sdkPlugin{
		name:         cfg.name,
		version:      cfg.version,
		description:  cfg.description,
		methods:      cfg.methods,
		methodMap:    methodMap,
		initFunc:     cfg.initFunc,
		shutdownFunc: cfg.shutdownFunc,
		initialized:  false,
	}, nil
}

// sdkPlugin is the private implementation of the Plugin interface.
type sdkPlugin struct {
	name         string
	version      string
	description  string
	methods      []methodEntry
	methodMap    map[string]methodEntry
	initFunc     InitFunc
	shutdownFunc ShutdownFunc
	initialized  bool
	mu           sync.RWMutex
}

// Name returns the plugin's unique identifier.
func (p *sdkPlugin) Name() string {
	return p.name
}

// Version returns the plugin's semantic version.
func (p *sdkPlugin) Version() string {
	return p.version
}

// Description returns the plugin's description.
func (p *sdkPlugin) Description() string {
	return p.description
}

// Methods returns a list of all method descriptors.
func (p *sdkPlugin) Methods() []MethodDescriptor {
	descriptors := make([]MethodDescriptor, 0, len(p.methods))
	for _, entry := range p.methods {
		descriptors = append(descriptors, entry.descriptor)
	}
	return descriptors
}

// Query invokes a named method with the given parameters.
func (p *sdkPlugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, exists := p.methodMap[method]
	if !exists {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	// Validate input parameters against schema
	if err := entry.descriptor.InputSchema.Validate(params); err != nil {
		return nil, fmt.Errorf("invalid input parameters: %w", err)
	}

	// Invoke the method handler
	result, err := entry.handler(ctx, params)
	if err != nil {
		return nil, err
	}

	// Validate output against schema
	if err := entry.descriptor.OutputSchema.Validate(result); err != nil {
		return nil, fmt.Errorf("invalid output: %w", err)
	}

	return result, nil
}

// Initialize prepares the plugin for use.
func (p *sdkPlugin) Initialize(ctx context.Context, config map[string]any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.initialized {
		return fmt.Errorf("plugin already initialized")
	}

	if err := p.initFunc(ctx, config); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	p.initialized = true
	return nil
}

// Shutdown gracefully shuts down the plugin.
func (p *sdkPlugin) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.initialized {
		return fmt.Errorf("plugin not initialized")
	}

	if err := p.shutdownFunc(ctx); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}

	p.initialized = false
	return nil
}

// Health returns the current health status of the plugin.
func (p *sdkPlugin) Health(ctx context.Context) types.HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.initialized {
		return types.NewUnhealthyStatus("plugin not initialized", nil)
	}

	return types.NewHealthyStatus("plugin operational")
}
