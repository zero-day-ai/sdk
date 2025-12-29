// Package plugin provides a framework for creating extensible SDK plugins.
//
// The plugin package enables developers to create self-contained, reusable components
// that extend the SDK's functionality. Each plugin exposes named methods that can be
// invoked with validated parameters and return typed results.
//
// # Core Concepts
//
// A Plugin is a component that:
//   - Has a unique name and version
//   - Provides one or more named methods
//   - Validates inputs and outputs using JSON schemas
//   - Supports initialization and graceful shutdown
//   - Reports health status for monitoring
//
// # Creating a Plugin
//
// Plugins are created using the builder pattern with the Config type:
//
//	cfg := plugin.NewConfig()
//	cfg.SetName("example")
//	cfg.SetVersion("1.0.0")
//	cfg.SetDescription("An example plugin")
//
//	// Add a method
//	cfg.AddMethodWithDesc(
//	    "greet",
//	    "Returns a greeting message",
//	    func(ctx context.Context, params map[string]any) (any, error) {
//	        name := params["name"].(string)
//	        return map[string]any{"message": "Hello, " + name}, nil
//	    },
//	    schema.Object(map[string]schema.JSON{
//	        "name": schema.String(),
//	    }, "name"),
//	    schema.Object(map[string]schema.JSON{
//	        "message": schema.String(),
//	    }, "message"),
//	)
//
//	// Build the plugin
//	p, err := plugin.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Using a Plugin
//
// Once created, a plugin must be initialized before use:
//
//	// Initialize with configuration
//	err := p.Initialize(ctx, map[string]any{
//	    "timeout": 30,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Invoke a method
//	result, err := p.Query(ctx, "greet", map[string]any{
//	    "name": "World",
//	})
//
//	// Check health
//	status := p.Health(ctx)
//	if !status.IsHealthy() {
//	    log.Printf("Plugin unhealthy: %s", status.Message)
//	}
//
//	// Shutdown when done
//	err = p.Shutdown(ctx)
//
// # Schema Validation
//
// All method inputs and outputs are validated against their JSON schemas.
// This ensures type safety and provides clear error messages when invalid
// data is passed or returned.
//
// # Lifecycle Management
//
// Plugins have a well-defined lifecycle:
//
//  1. Creation - Build the plugin with New()
//  2. Initialization - Call Initialize() with configuration
//  3. Operation - Invoke methods with Query()
//  4. Shutdown - Call Shutdown() to release resources
//
// # Thread Safety
//
// The plugin implementation is thread-safe and can handle concurrent
// method invocations. Internal state is protected with read-write locks.
//
// # Error Handling
//
// Errors are returned in the following cases:
//   - Invalid configuration during plugin creation
//   - Method not found during Query
//   - Schema validation failures for inputs or outputs
//   - Method handler errors
//   - Initialization or shutdown failures
//
// All errors include descriptive messages to aid in debugging.
package plugin
