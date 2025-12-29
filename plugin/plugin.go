package plugin

import (
	"context"

	"github.com/zero-day-ai/sdk/types"
)

// Plugin is the interface for SDK plugins.
// A plugin extends the SDK functionality by providing named methods that can be invoked
// with parameters and return results. Plugins support initialization, shutdown, and health checks.
type Plugin interface {
	// Name returns the unique identifier for the plugin.
	Name() string

	// Version returns the semantic version of the plugin.
	Version() string

	// Description returns a human-readable description of the plugin's purpose.
	Description() string

	// Methods returns a list of method descriptors that this plugin provides.
	// Each descriptor includes the method name, description, and input/output schemas.
	Methods() []MethodDescriptor

	// Query invokes a named method with the given parameters.
	// Returns the method result or an error if the method doesn't exist or fails.
	Query(ctx context.Context, method string, params map[string]any) (any, error)

	// Initialize prepares the plugin for use with the given configuration.
	// This is called once before any Query calls.
	Initialize(ctx context.Context, config map[string]any) error

	// Shutdown gracefully shuts down the plugin and releases any resources.
	// This is called once when the plugin is no longer needed.
	Shutdown(ctx context.Context) error

	// Health returns the current health status of the plugin.
	// This can be used for monitoring and diagnostics.
	Health(ctx context.Context) types.HealthStatus
}
