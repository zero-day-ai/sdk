# Plugin Package

The `plugin` package provides a comprehensive framework for creating extensible SDK plugins with type-safe method invocation, schema validation, and lifecycle management.

## Features

- **Plugin Interface**: Well-defined interface for plugin implementations
- **Builder Pattern**: Easy plugin creation using the Config builder
- **Schema Validation**: Automatic input/output validation using JSON schemas
- **Lifecycle Management**: Initialize, operate, and shutdown phases
- **Health Monitoring**: Built-in health check support
- **Thread-Safe**: Safe for concurrent use
- **100% Test Coverage**: Comprehensive test suite with examples

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/zero-day-ai/sdk/plugin"
    "github.com/zero-day-ai/sdk/schema"
)

func main() {
    // Create plugin configuration
    cfg := plugin.NewConfig()
    cfg.SetName("calculator")
    cfg.SetVersion("1.0.0")
    cfg.SetDescription("A simple calculator")

    // Add a method
    cfg.AddMethodWithDesc(
        "add",
        "Adds two numbers",
        func(ctx context.Context, params map[string]any) (any, error) {
            a := params["a"].(float64)
            b := params["b"].(float64)
            return map[string]any{"result": a + b}, nil
        },
        schema.Object(map[string]schema.JSON{
            "a": schema.Number(),
            "b": schema.Number(),
        }, "a", "b"),
        schema.Object(map[string]schema.JSON{
            "result": schema.Number(),
        }, "result"),
    )

    // Build the plugin
    p, err := plugin.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Initialize
    if err := p.Initialize(ctx, nil); err != nil {
        log.Fatal(err)
    }

    // Use the plugin
    result, err := p.Query(ctx, "add", map[string]any{"a": 5.0, "b": 3.0})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Result: %.0f\n", result.(map[string]any)["result"])

    // Shutdown
    if err := p.Shutdown(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Plugin Lifecycle

1. **Create**: Build plugin with `plugin.New(cfg)`
2. **Initialize**: Call `Initialize(ctx, config)` to prepare the plugin
3. **Operate**: Invoke methods using `Query(ctx, method, params)`
4. **Monitor**: Check health with `Health(ctx)`
5. **Shutdown**: Release resources with `Shutdown(ctx)`

## Architecture

### Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Methods() []MethodDescriptor
    Query(ctx context.Context, method string, params map[string]any) (any, error)
    Initialize(ctx context.Context, config map[string]any) error
    Shutdown(ctx context.Context) error
    Health(ctx context.Context) types.HealthStatus
}
```

### Configuration

The `Config` type provides a fluent API for building plugins:

- `SetName(name string)` - Set plugin name
- `SetVersion(version string)` - Set plugin version
- `SetDescription(desc string)` - Set plugin description
- `AddMethod(...)` - Register a method
- `AddMethodWithDesc(...)` - Register a method with description
- `SetInitFunc(fn InitFunc)` - Set initialization handler
- `SetShutdownFunc(fn ShutdownFunc)` - Set shutdown handler

### Method Handlers

Methods are implemented as handlers:

```go
type MethodHandler func(ctx context.Context, params map[string]any) (any, error)
```

Each method has:
- **Input Schema**: Validates parameters before invocation
- **Output Schema**: Validates results after invocation
- **Handler Function**: Implements the method logic

## Schema Validation

All method inputs and outputs are validated against JSON schemas:

```go
// Define schemas
inputSchema := schema.Object(map[string]schema.JSON{
    "name": schema.String(),
    "age": schema.Int(),
}, "name", "age") // Required fields

outputSchema := schema.Object(map[string]schema.JSON{
    "greeting": schema.String(),
}, "greeting")
```

Invalid inputs or outputs result in descriptive errors.

## Health Monitoring

Plugins report their health status:

```go
status := p.Health(ctx)
if status.IsHealthy() {
    fmt.Println("Plugin is healthy")
} else {
    fmt.Printf("Plugin unhealthy: %s\n", status.Message)
}
```

Health states:
- **Healthy**: Plugin is operational (after initialization)
- **Unhealthy**: Plugin not initialized or shut down

## Thread Safety

The plugin implementation is thread-safe and uses read-write locks to protect internal state. Multiple goroutines can safely invoke methods concurrently.

## Testing

Run tests with:

```bash
go test ./plugin/...
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./plugin
go tool cover -html=coverage.out
```

Run race detector:

```bash
go test -race ./plugin/...
```

## Examples

See `example_test.go` for complete working examples:
- Basic plugin creation and usage
- Plugin initialization with configuration
- Health check monitoring

## File Structure

```
plugin/
├── plugin.go          # Plugin interface definition
├── types.go           # MethodDescriptor and Descriptor types
├── builder.go         # Config and builder implementation
├── doc.go             # Package documentation
├── plugin_test.go     # Plugin interface tests
├── types_test.go      # Types tests
├── builder_test.go    # Builder and implementation tests
├── example_test.go    # Runnable examples
└── README.md          # This file
```

## Dependencies

- `github.com/zero-day-ai/sdk/schema` - JSON schema validation
- `github.com/zero-day-ai/sdk/types` - Health status types

## License

Part of the Gibson Framework SDK.
