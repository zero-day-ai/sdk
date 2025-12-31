# SDK Exec Package

The `exec` package provides a simple, context-aware API for executing shell commands with timeout support. It wraps Go's standard `os/exec` package with additional features for command execution in tool development.

## Features

- Context-aware command execution with timeout support
- Automatic process cleanup on timeout or cancellation
- Structured result with stdout, stderr, exit code, and duration
- No shell interpretation (prevents shell injection)
- Binary existence and path resolution utilities
- 100% test coverage
- Stdlib-only dependencies

## Installation

The package is part of the Gibson SDK:

```bash
go get github.com/zero-day-ai/sdk
```

## Usage

### Basic Command Execution

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/zero-day-ai/sdk/exec"
)

func main() {
    ctx := context.Background()

    cfg := exec.Config{
        Command: "echo",
        Args:    []string{"hello", "world"},
        Timeout: 5 * time.Second,
    }

    result, err := exec.Run(ctx, cfg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Exit code: %d\n", result.ExitCode)
    fmt.Printf("Output: %s", result.Stdout)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

### Command with Timeout

```go
cfg := exec.Config{
    Command: "sleep",
    Args:    []string{"10"},
    Timeout: 1 * time.Second,  // Will timeout after 1 second
}

result, err := exec.Run(ctx, cfg)
if err != nil {
    fmt.Printf("Command timed out: %v\n", err)
}
```

### Command with Working Directory and Environment

```go
cfg := exec.Config{
    Command: "git",
    Args:    []string{"status"},
    WorkDir: "/path/to/repo",
    Env:     []string{"GIT_TERMINAL_PROMPT=0"},
    Timeout: 10 * time.Second,
}

result, err := exec.Run(ctx, cfg)
```

### Command with Stdin

```go
cfg := exec.Config{
    Command:   "cat",
    StdinData: []byte("hello from stdin"),
}

result, err := exec.Run(ctx, cfg)
fmt.Println(string(result.Stdout))  // Output: hello from stdin
```

### Handling Non-Zero Exit Codes

The `Run` function does not treat non-zero exit codes as errors. This allows you to handle command failures appropriately:

```go
cfg := exec.Config{
    Command: "grep",
    Args:    []string{"pattern", "file.txt"},
}

result, err := exec.Run(ctx, cfg)
if err != nil {
    // Actual execution error (binary not found, permission denied, etc.)
    return err
}

if result.ExitCode != 0 {
    // Command ran but failed (pattern not found, file doesn't exist, etc.)
    fmt.Printf("Command failed with exit code %d\n", result.ExitCode)
    fmt.Printf("Error: %s\n", result.Stderr)
}
```

### Checking Binary Existence

```go
if !exec.BinaryExists("docker") {
    return fmt.Errorf("docker is not installed")
}

// Get full path to binary
path, err := exec.BinaryPath("docker")
if err != nil {
    return err
}
fmt.Printf("Docker is at: %s\n", path)
```

## API Reference

### Types

#### Config

```go
type Config struct {
    Command   string        // Required: command to execute
    Args      []string      // Optional: command arguments
    WorkDir   string        // Optional: working directory
    Env       []string      // Optional: environment variables (KEY=value)
    Timeout   time.Duration // Optional: execution timeout
    StdinData []byte        // Optional: stdin input
}
```

#### Result

```go
type Result struct {
    Stdout   []byte        // Captured stdout
    Stderr   []byte        // Captured stderr
    ExitCode int           // Process exit code (0 = success)
    Duration time.Duration // Actual execution time
}
```

### Functions

#### Run

```go
func Run(ctx context.Context, cfg Config) (*Result, error)
```

Executes a command with the given configuration. Returns a Result containing stdout, stderr, exit code, and duration.

- Respects context cancellation and configured timeout
- Non-zero exit codes are NOT treated as errors (check `Result.ExitCode`)
- Only execution failures (binary not found, etc.) return an error
- Automatically kills the process on timeout/cancellation

#### BinaryExists

```go
func BinaryExists(name string) bool
```

Checks if a binary exists in the system PATH. Returns true if found and executable.

#### BinaryPath

```go
func BinaryPath(name string) (string, error)
```

Returns the full path to a binary in the system PATH. Returns an error if not found.

## Design Principles

1. **No Shell Interpretation**: Commands are executed directly without shell interpretation, preventing shell injection vulnerabilities.

2. **Context-Aware**: All execution respects context cancellation and timeouts for proper resource management.

3. **Explicit Error Handling**: Non-zero exit codes are distinguished from execution errors, allowing tools to handle each appropriately.

4. **Minimal Dependencies**: Uses only standard library packages (os/exec, context, bytes, time).

5. **Process Cleanup**: Ensures processes are properly killed on timeout or cancellation to prevent resource leaks.

## Performance

Benchmark results on Intel i7-4770K @ 3.50GHz:

```
BenchmarkRun_SimpleEcho-8     1982    593359 ns/op    18272 B/op    119 allocs/op
BenchmarkBinaryExists-8      61480     19882 ns/op     4800 B/op     53 allocs/op
BenchmarkBinaryPath-8        61082     20778 ns/op     4800 B/op     53 allocs/op
```

Command execution adds minimal overhead (<1ms) for simple commands.

## Testing

The package has 100% test coverage with comprehensive tests for:

- Successful command execution
- Non-zero exit codes
- Timeout handling
- Context cancellation
- Working directory and environment variables
- Stdin input
- Binary existence checks
- Error conditions

Run tests:

```bash
go test ./exec/...
go test ./exec/... -cover
go test ./exec/... -bench=.
```

## Thread Safety

All functions in this package are safe for concurrent use. Multiple goroutines can call `Run`, `BinaryExists`, and `BinaryPath` simultaneously.

## Error Handling

The package distinguishes between three types of failures:

1. **Configuration Errors**: Empty command, invalid parameters
   - Returns error immediately

2. **Execution Errors**: Binary not found, permission denied, timeout, cancellation
   - Returns error with descriptive message
   - Result struct is still returned with partial data

3. **Command Failures**: Command runs but exits with non-zero code
   - Does NOT return error
   - Returns Result with ExitCode set
   - Caller decides how to handle

This design allows tools to handle different failure modes appropriately without losing access to partial output.
