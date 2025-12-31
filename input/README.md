# sdk/input

Type-safe helpers for extracting values from `map[string]any`.

## Purpose

This package simplifies working with JSON unmarshaled data or configuration maps where types may vary. Instead of manual type assertions with error handling, these helpers provide:

- Type-safe extraction with automatic coercion
- Sensible defaults on type mismatch (no errors/panics)
- Graceful handling of nil maps and values
- Zero allocations in hot paths

## Installation

```bash
go get github.com/zero-day-ai/sdk/input
```

## Quick Start

```go
import "github.com/zero-day-ai/sdk/input"

// JSON unmarshaled into map[string]any
config := map[string]any{
    "host":     "example.com",
    "port":     8080,
    "timeout":  "30s",
    "enabled":  true,
    "tags":     []string{"web", "api"},
}

// Extract with type safety and defaults
host := input.GetString(config, "host", "localhost")
port := input.GetInt(config, "port", 80)
timeout := input.GetTimeout(config, "timeout", 10*time.Second)
enabled := input.GetBool(config, "enabled", false)
tags := input.GetStringSlice(config, "tags")
```

## API Reference

### Basic Types

- `GetString(m, key, default) string` - Extract string value
- `GetInt(m, key, default) int` - Extract int with coercion from int64, float64, string
- `GetBool(m, key, default) bool` - Extract boolean value
- `GetFloat64(m, key, default) float64` - Extract float with coercion

### Complex Types

- `GetStringSlice(m, key) []string` - Extract string slice (handles []string, []interface{}, single string)
- `GetMap(m, key) map[string]any` - Extract nested map
- `GetTimeout(m, key, default) time.Duration` - Extract duration (handles int seconds, "5m" strings, time.Duration)

## Type Coercion

The package automatically handles JSON unmarshaling quirks:

### Numbers
JSON numbers unmarshal to `float64`, but may be `int`, `int64`, or numeric strings:

```go
config := map[string]any{
    "a": 42,           // int
    "b": int64(100),   // int64
    "c": 123.5,        // float64
    "d": "456",        // string
}

input.GetInt(config, "a", 0)  // 42
input.GetInt(config, "b", 0)  // 100
input.GetInt(config, "c", 0)  // 123 (truncated)
input.GetInt(config, "d", 0)  // 456 (parsed)
```

### String Slices
JSON arrays may be `[]string` or `[]interface{}`:

```go
config := map[string]any{
    "a": []string{"x", "y"},
    "b": []interface{}{"a", 123, true},  // mixed types
    "c": "single",                        // single string
}

input.GetStringSlice(config, "a")  // ["x", "y"]
input.GetStringSlice(config, "b")  // ["a", "123", "true"]
input.GetStringSlice(config, "c")  // ["single"]
```

### Timeouts
Timeouts can be specified multiple ways:

```go
config := map[string]any{
    "a": 30,                    // int seconds
    "b": "5m",                  // duration string
    "c": 45 * time.Second,      // time.Duration
    "d": "1h30m",               // complex duration
}

input.GetTimeout(config, "a", 0)  // 30s
input.GetTimeout(config, "b", 0)  // 5m0s
input.GetTimeout(config, "c", 0)  // 45s
input.GetTimeout(config, "d", 0)  // 1h30m0s
```

## Design Philosophy

This package follows the Robustness Principle: "be liberal in what you accept." It handles real-world data variations without requiring extensive error handling, making tool development simpler and more robust.

**Key principles:**
- Never panic or return errors
- Return sensible defaults on type mismatch
- Handle nil values gracefully
- Optimize for common JSON unmarshaling patterns

## Performance

All functions achieve zero allocations in hot paths:

```
BenchmarkGetString-8           100000000    10.81 ns/op    0 B/op    0 allocs/op
BenchmarkGetInt-8               92764384    13.96 ns/op    0 B/op    0 allocs/op
BenchmarkGetIntCoercion-8       90910854    14.25 ns/op    0 B/op    0 allocs/op
BenchmarkGetStringSlice-8       86249184    13.25 ns/op    0 B/op    0 allocs/op
BenchmarkGetTimeout-8           23897042    52.12 ns/op    0 B/op    0 allocs/op
```

## Testing

100% test coverage with comprehensive edge case testing:

```bash
go test ./input/...
go test -cover ./input/...
go test -bench=. ./input/...
```

## Use Cases

### Tool Input Parsing

```go
func (t *Tool) Execute(ctx context.Context, input map[string]any) (*types.ToolResponse, error) {
    target := input.GetString(input, "target", "localhost")
    timeout := input.GetTimeout(input, "timeout", 30*time.Second)
    ports := input.GetStringSlice(input, "ports")

    // Use values safely without manual type checking
}
```

### Configuration Loading

```go
func LoadConfig(data map[string]any) Config {
    return Config{
        Host:    input.GetString(data, "host", "localhost"),
        Port:    input.GetInt(data, "port", 8080),
        Timeout: input.GetTimeout(data, "timeout", 10*time.Second),
        Debug:   input.GetBool(data, "debug", false),
    }
}
```

### Nested Configuration

```go
serverConfig := input.GetMap(config, "server")
if serverConfig != nil {
    host := input.GetString(serverConfig, "host", "0.0.0.0")
    port := input.GetInt(serverConfig, "port", 8080)
}
```

## Dependencies

- Go standard library only (no external dependencies)
- Requires Go 1.21+

## Related Packages

- `sdk/types` - Type definitions for tools and responses
- `sdk/exec` - Command execution utilities
- `sdk/toolerr` - Structured error types
- `sdk/health` - Health check utilities
