# toolerr - Structured Tool Errors

Package `toolerr` provides structured error types for Gibson tools with standard error codes, context, and full compatibility with Go's standard error handling.

## Features

- **Structured Errors**: Rich error context including tool name, operation, error code, and details
- **Standard Error Codes**: Predefined constants for common error scenarios
- **Method Chaining**: Fluent API for building errors with cause and context
- **Standard Library Integration**: Full compatibility with `errors.Is()`, `errors.As()`, and `errors.Unwrap()`
- **Zero Allocations**: Creating errors has zero memory allocations (compiler optimized)
- **High Performance**: Error formatting is fast with minimal allocations

## Installation

```bash
go get github.com/zero-day-ai/sdk/toolerr
```

## Quick Start

```go
import "github.com/zero-day-ai/sdk/toolerr"

// Create a simple error
err := toolerr.New("nmap", "scan", toolerr.ErrCodeBinaryNotFound,
    "nmap binary not found in PATH")

// Add context with method chaining
err := toolerr.New("kubectl", "apply", toolerr.ErrCodeExecutionFailed,
    "command failed").
    WithCause(execErr).
    WithDetails(map[string]any{
        "namespace": "default",
        "resource": "deployment",
    })
```

## Error Codes

The package provides standard error codes for consistent error handling:

| Code | Description |
|------|-------------|
| `ErrCodeBinaryNotFound` | Required binary not in PATH |
| `ErrCodeExecutionFailed` | Command execution failed |
| `ErrCodeTimeout` | Operation timed out |
| `ErrCodeParseError` | Failed to parse output or data |
| `ErrCodeInvalidInput` | Invalid input parameters |
| `ErrCodeDependencyMissing` | Required dependency missing |
| `ErrCodePermissionDenied` | Insufficient permissions |
| `ErrCodeNetworkError` | Network-related error |

## Usage Examples

### Basic Error

```go
err := toolerr.New("nmap", "scan", toolerr.ErrCodeBinaryNotFound,
    "nmap binary not found in PATH")
fmt.Println(err)
// Output: nmap [scan/BINARY_NOT_FOUND]: nmap binary not found in PATH
```

### Error with Cause

```go
execErr := errors.New("exit status 1")
err := toolerr.New("kubectl", "apply", toolerr.ErrCodeExecutionFailed,
    "command failed").
    WithCause(execErr)
fmt.Println(err)
// Output: kubectl [apply/EXECUTION_FAILED]: command failed: exit status 1
```

### Error with Details

```go
err := toolerr.New("terraform", "plan", toolerr.ErrCodeTimeout,
    "operation timed out").
    WithDetails(map[string]any{
        "timeout": "30s",
        "target": "vpc-12345",
    })
```

### Error Chain Checking

```go
baseErr := errors.New("connection refused")
err := toolerr.New("terraform", "plan", toolerr.ErrCodeNetworkError,
    "failed to connect to AWS").
    WithCause(baseErr)

// Check if error chain contains specific error
if errors.Is(err, baseErr) {
    // Handle specific error
}
```

### Type Assertion

```go
var toolErr *toolerr.Error
if errors.As(err, &toolErr) {
    fmt.Printf("Tool: %s\n", toolErr.Tool)
    fmt.Printf("Operation: %s\n", toolErr.Operation)
    fmt.Printf("Code: %s\n", toolErr.Code)
    fmt.Printf("Details: %v\n", toolErr.Details)
}
```

### Comparing Errors

```go
err1 := toolerr.New("nmap", "scan", toolerr.ErrCodeBinaryNotFound, "msg1")
err2 := toolerr.New("nmap", "scan", toolerr.ErrCodeBinaryNotFound, "msg2")

// Errors are considered equal if Tool, Operation, and Code match
if errors.Is(err1, err2) {
    // Same error type
}
```

## API Reference

### Error Type

```go
type Error struct {
    Tool      string         // Tool name (e.g., "nmap", "kubectl")
    Operation string         // Operation that failed (e.g., "scan", "apply")
    Code      string         // Error code constant
    Message   string         // Human-readable message
    Details   map[string]any // Additional context
    Cause     error          // Underlying error
}
```

### Functions

#### New

```go
func New(tool, operation, code, message string) *Error
```

Creates a new structured tool error.

#### WithCause

```go
func (e *Error) WithCause(err error) *Error
```

Adds an underlying error. Returns the same error instance for chaining.

#### WithDetails

```go
func (e *Error) WithDetails(details map[string]any) *Error
```

Adds additional context. Returns the same error instance for chaining.

#### Error

```go
func (e *Error) Error() string
```

Implements the error interface. Format: `tool [operation/code]: message: cause`

#### Unwrap

```go
func (e *Error) Unwrap() error
```

Returns the underlying cause error for `errors.Unwrap()`.

#### Is

```go
func (e *Error) Is(target error) bool
```

Implements error equality checking for `errors.Is()`.

#### As

```go
func (e *Error) As(target any) bool
```

Implements error type assertion for `errors.As()`.

## Sentinel Errors

The package also provides sentinel errors for common scenarios:

```go
var (
    ErrBinaryNotFound = errors.New("binary not found")
    ErrTimeout        = errors.New("operation timed out")
    ErrInvalidInput   = errors.New("invalid input")
)
```

## Performance

Benchmarks on Intel Core i7-4770K @ 3.50GHz:

```
BenchmarkNew-8               	1000000000	   0.30 ns/op	   0 B/op	   0 allocs/op
BenchmarkWithCause-8         	1000000000	   0.29 ns/op	   0 B/op	   0 allocs/op
BenchmarkWithDetails-8       	1000000000	   0.30 ns/op	   0 B/op	   0 allocs/op
BenchmarkErrorFormatting-8   	 2439633	 482.80 ns/op	 232 B/op	   8 allocs/op
```

- Creating errors has **zero allocations** (compiler optimized)
- Error formatting uses only **8 allocations** and **232 bytes**

## Testing

Run tests:

```bash
go test ./toolerr/...
```

Run with coverage:

```bash
go test ./toolerr/... -cover
```

Run with race detector:

```bash
go test ./toolerr/... -race
```

Run benchmarks:

```bash
go test ./toolerr/... -bench=. -benchmem
```

## Design Philosophy

1. **Simplicity**: Use standard library patterns and idioms
2. **Compatibility**: Full integration with Go's error handling
3. **Performance**: Zero allocations for error creation
4. **Consistency**: Standard error codes across all tools
5. **Context**: Rich error information for debugging

## License

Part of the Gibson SDK - see main SDK license.
