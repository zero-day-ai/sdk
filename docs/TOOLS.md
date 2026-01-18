# Tools Quickstart Guide

Build reusable capabilities that agents can invoke with the Gibson SDK.

## What is a Tool?

A tool is a reusable, schema-validated capability that:
- Performs a specific action (HTTP requests, scanning, parsing, etc.)
- Has well-defined input and output schemas
- Can be invoked by any agent via the harness
- Is stateless and thread-safe

## Minimal Tool (5 Minutes)

```go
package main

import (
    "context"
    "log"

    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/tool"
    "github.com/zero-day-ai/sdk/types"
)

func main() {
    cfg := tool.NewConfig().
        SetName("echo").
        SetVersion("1.0.0").
        SetDescription("Echoes the input message").
        SetTags([]string{"utility", "debug"}).
        SetInputSchema(schema.Object(map[string]schema.JSON{
            "message": schema.String(),
        }, "message")).  // "message" is required
        SetOutputSchema(schema.Object(map[string]schema.JSON{
            "echo": schema.String(),
        }, "echo")).
        SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
            msg := input["message"].(string)
            return map[string]any{"echo": msg}, nil
        })

    myTool, err := tool.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Test it
    result, err := myTool.Execute(context.Background(), map[string]any{
        "message": "Hello, World!",
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Result: %v", result["echo"]) // Output: Hello, World!
}
```

## Tool Configuration

### Required Fields

| Field | Description |
|-------|-------------|
| `Name` | Unique identifier (kebab-case, e.g., "http-client") |
| `Version` | Semantic version (e.g., "1.0.0") |
| `Description` | What the tool does |
| `InputSchema` | JSON Schema for input validation |
| `OutputSchema` | JSON Schema for output validation |
| `ExecuteFunc` | The execution logic |

### Optional Fields

| Field | Description |
|-------|-------------|
| `Tags` | Categories for discovery (e.g., "network", "http") |
| `HealthFunc` | Health check logic |

## Schema Building

The `schema` package provides builders for JSON Schema:

### Basic Types

```go
import "github.com/zero-day-ai/sdk/schema"

// String
schema.String()
schema.StringWithDesc("A description")

// Number (float64)
schema.Number()

// Integer
schema.Int()

// Boolean
schema.Bool()

// Enum (string with allowed values)
schema.Enum("GET", "POST", "PUT", "DELETE")
```

### Complex Types

```go
// Object with required fields
schema.Object(map[string]schema.JSON{
    "url":     schema.String(),
    "method":  schema.Enum("GET", "POST"),
    "headers": schema.Object(map[string]schema.JSON{}),  // optional nested object
    "body":    schema.String(),
}, "url", "method")  // url and method are required

// Array
schema.Array(schema.String())                    // []string
schema.Array(schema.Object(...))                 // []object
schema.Array(schema.String(), "List of URLs")   // with description
```

### Real-World Example

```go
inputSchema := schema.Object(map[string]schema.JSON{
    "url": schema.StringWithDesc("Target URL to request"),
    "method": schema.Enum("GET", "POST", "PUT", "DELETE", "PATCH"),
    "headers": schema.Object(map[string]schema.JSON{
        // Dynamic keys allowed
    }),
    "body": schema.StringWithDesc("Request body (for POST/PUT)"),
    "timeout_seconds": schema.Int(),
    "follow_redirects": schema.Bool(),
}, "url", "method")  // Only url and method are required

outputSchema := schema.Object(map[string]schema.JSON{
    "status_code": schema.Int(),
    "headers": schema.Object(map[string]schema.JSON{}),
    "body": schema.String(),
    "elapsed_ms": schema.Int(),
}, "status_code", "body")
```

## HTTP Client Tool Example

A complete, production-ready HTTP client tool:

```go
package main

import (
    "context"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/tool"
    "github.com/zero-day-ai/sdk/toolerr"
    "github.com/zero-day-ai/sdk/types"
)

func main() {
    cfg := tool.NewConfig().
        SetName("http-client").
        SetVersion("1.0.0").
        SetDescription("Makes HTTP requests to target URLs").
        SetTags([]string{"http", "network", "web"}).
        SetInputSchema(schema.Object(map[string]schema.JSON{
            "url":              schema.StringWithDesc("Target URL"),
            "method":           schema.Enum("GET", "POST", "PUT", "DELETE", "PATCH"),
            "headers":          schema.Object(map[string]schema.JSON{}),
            "body":             schema.String(),
            "timeout_seconds":  schema.Int(),
            "follow_redirects": schema.Bool(),
        }, "url", "method")).
        SetOutputSchema(schema.Object(map[string]schema.JSON{
            "status_code": schema.Int(),
            "headers":     schema.Object(map[string]schema.JSON{}),
            "body":        schema.String(),
            "elapsed_ms":  schema.Int(),
        }, "status_code", "body")).
        SetExecuteFunc(executeHTTP).
        SetHealthFunc(func(ctx context.Context) types.HealthStatus {
            return types.NewHealthyStatus("http client operational")
        })

    httpTool, _ := tool.New(cfg)
    // Register with Gibson or serve standalone
}

func executeHTTP(ctx context.Context, input map[string]any) (map[string]any, error) {
    url := input["url"].(string)
    method := input["method"].(string)

    // Build request
    var body io.Reader
    if b, ok := input["body"].(string); ok && b != "" {
        body = strings.NewReader(b)
    }

    req, err := http.NewRequestWithContext(ctx, method, url, body)
    if err != nil {
        return nil, toolerr.New("http-client", "request", toolerr.ErrCodeInvalidInput,
            "failed to create request").WithCause(err)
    }

    // Add headers
    if headers, ok := input["headers"].(map[string]any); ok {
        for k, v := range headers {
            req.Header.Set(k, v.(string))
        }
    }

    // Configure client
    timeout := 30 * time.Second
    if t, ok := input["timeout_seconds"].(float64); ok && t > 0 {
        timeout = time.Duration(t) * time.Second
    }

    client := &http.Client{Timeout: timeout}
    if follow, ok := input["follow_redirects"].(bool); ok && !follow {
        client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        }
    }

    // Execute
    start := time.Now()
    resp, err := client.Do(req)
    elapsed := time.Since(start).Milliseconds()

    if err != nil {
        return nil, toolerr.New("http-client", "execute", toolerr.ErrCodeNetworkError,
            "request failed").WithCause(err)
    }
    defer resp.Body.Close()

    // Read response
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, toolerr.New("http-client", "read", toolerr.ErrCodeExecutionFailed,
            "failed to read response body").WithCause(err)
    }

    // Convert headers
    respHeaders := make(map[string]any)
    for k, v := range resp.Header {
        if len(v) == 1 {
            respHeaders[k] = v[0]
        } else {
            respHeaders[k] = v
        }
    }

    return map[string]any{
        "status_code": resp.StatusCode,
        "headers":     respHeaders,
        "body":        string(bodyBytes),
        "elapsed_ms":  int(elapsed),
    }, nil
}
```

## Error Handling with toolerr

Use structured errors for better debugging:

```go
import "github.com/zero-day-ai/sdk/toolerr"

// Basic error
err := toolerr.New("my-tool", "operation", toolerr.ErrCodeExecutionFailed,
    "something went wrong")

// With cause
err := toolerr.New("my-tool", "scan", toolerr.ErrCodeTimeout,
    "scan timed out").WithCause(originalErr)

// With details
err := toolerr.New("kubectl", "apply", toolerr.ErrCodeExecutionFailed,
    "failed to apply manifest").
    WithCause(execErr).
    WithDetails(map[string]any{
        "namespace": "default",
        "manifest":  "deployment.yaml",
    })
```

### Error Codes

| Code | Constant | Description |
|------|----------|-------------|
| `BINARY_NOT_FOUND` | `ErrCodeBinaryNotFound` | Required binary not in PATH |
| `EXECUTION_FAILED` | `ErrCodeExecutionFailed` | Tool execution failed |
| `TIMEOUT` | `ErrCodeTimeout` | Operation timed out |
| `PARSE_ERROR` | `ErrCodeParseError` | Failed to parse input/output |
| `INVALID_INPUT` | `ErrCodeInvalidInput` | Input validation failed |
| `PERMISSION_DENIED` | `ErrCodePermissionDenied` | Insufficient permissions |
| `NETWORK_ERROR` | `ErrCodeNetworkError` | Network operation failed |

## Serving Your Tool

### gRPC Mode (High Performance)

```go
import "github.com/zero-day-ai/sdk/serve"

// Local mode (Unix socket)
err := serve.Tool(myTool,
    serve.WithPort(50052),
    serve.WithLocalMode("~/.gibson/run/tools/my-tool.sock"),
)

// Remote mode (TCP with TLS)
err := serve.Tool(myTool,
    serve.WithPort(50052),
    serve.WithTLS("cert.pem", "key.pem"),
)
```

### Subprocess Mode (Simple)

For simple tools, use stdin/stdout JSON protocol:

```go
package main

import (
    "os"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    myTool := createMyTool()

    // Handle --schema flag for discovery
    if len(os.Args) > 1 && os.Args[1] == "--schema" {
        serve.OutputSchema(myTool)
        os.Exit(0)
    }

    // Run in subprocess mode
    if err := serve.RunSubprocess(myTool); err != nil {
        os.Exit(1)
    }
}
```

Usage:
```bash
# Get schema
./my-tool --schema

# Execute
echo '{"message": "hello"}' | ./my-tool
```

## Wrapping External Binaries

Many tools wrap existing CLI tools:

```go
package main

import (
    "context"
    "encoding/json"
    "os/exec"

    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/tool"
    "github.com/zero-day-ai/sdk/toolerr"
)

func main() {
    cfg := tool.NewConfig().
        SetName("nmap-scanner").
        SetVersion("1.0.0").
        SetDescription("Port scanner using nmap").
        SetTags([]string{"network", "scanner", "ports"}).
        SetInputSchema(schema.Object(map[string]schema.JSON{
            "target": schema.StringWithDesc("Target host or IP"),
            "ports":  schema.StringWithDesc("Port range (e.g., '1-1000' or '22,80,443')"),
            "flags":  schema.Array(schema.String()),
        }, "target")).
        SetOutputSchema(schema.Object(map[string]schema.JSON{
            "hosts": schema.Array(schema.Object(map[string]schema.JSON{
                "ip":    schema.String(),
                "ports": schema.Array(schema.Object(map[string]schema.JSON{})),
            })),
            "raw_output": schema.String(),
        })).
        SetExecuteFunc(executeNmap)

    nmapTool, _ := tool.New(cfg)
    // ...
}

func executeNmap(ctx context.Context, input map[string]any) (map[string]any, error) {
    // Check if nmap is installed
    if _, err := exec.LookPath("nmap"); err != nil {
        return nil, toolerr.New("nmap-scanner", "init", toolerr.ErrCodeBinaryNotFound,
            "nmap not found in PATH")
    }

    target := input["target"].(string)

    // Build args
    args := []string{"-oX", "-", target}  // XML output to stdout

    if ports, ok := input["ports"].(string); ok && ports != "" {
        args = append([]string{"-p", ports}, args...)
    }

    if flags, ok := input["flags"].([]any); ok {
        for _, f := range flags {
            args = append([]string{f.(string)}, args...)
        }
    }

    // Execute
    cmd := exec.CommandContext(ctx, "nmap", args...)
    output, err := cmd.Output()
    if err != nil {
        return nil, toolerr.New("nmap-scanner", "execute", toolerr.ErrCodeExecutionFailed,
            "nmap execution failed").WithCause(err)
    }

    // Parse XML output (simplified)
    hosts := parseNmapXML(output)

    return map[string]any{
        "hosts":      hosts,
        "raw_output": string(output),
    }, nil
}
```

## Health Checks

Implement health checks for monitoring:

```go
cfg.SetHealthFunc(func(ctx context.Context) types.HealthStatus {
    // Check dependencies
    if _, err := exec.LookPath("nmap"); err != nil {
        return types.NewUnhealthyStatus(
            "nmap binary not found",
            map[string]any{"error": err.Error()},
        )
    }

    // Check connectivity (if applicable)
    // ...

    return types.NewHealthyStatus("all dependencies available")
})
```

## Best Practices

1. **Validate inputs thoroughly** - Don't trust input even with schema validation
2. **Use timeouts** - Always set context timeouts for external operations
3. **Return structured errors** - Use `toolerr` for debugging
4. **Keep tools focused** - One tool, one purpose
5. **Make tools stateless** - No persistent state between invocations
6. **Document schemas well** - Use `WithDesc()` for all fields
7. **Handle cancellation** - Respect `ctx.Done()` for long operations
8. **Log appropriately** - Tools shouldn't log to stdout (use stderr or structured logging)

## Testing Tools

```go
func TestMyTool(t *testing.T) {
    myTool, err := createMyTool()
    require.NoError(t, err)

    ctx := context.Background()

    // Test valid input
    result, err := myTool.Execute(ctx, map[string]any{
        "url":    "https://httpbin.org/get",
        "method": "GET",
    })
    require.NoError(t, err)
    assert.Equal(t, 200, result["status_code"])

    // Test invalid input (schema validation)
    _, err = myTool.Execute(ctx, map[string]any{
        "url": "https://example.com",
        // missing required "method"
    })
    assert.Error(t, err)
}
```

## Next Steps

- See `examples/custom-tool/` for more examples
- Read the [Agents Guide](AGENTS.md) to understand how agents use tools
- Check the [main README](../README.md) for deployment options
