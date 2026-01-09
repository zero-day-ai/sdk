# Gibson Tools Guide

This guide explains how to build, deploy, and use tools in the Gibson framework. Tools are reusable executable components that agents can call to perform specific operations.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Tool Execution Modes](#tool-execution-modes)
- [When to Use Which Mode](#when-to-use-which-mode)
- [Quick Start: Subprocess Tool](#quick-start-subprocess-tool)
- [Quick Start: gRPC Tool](#quick-start-grpc-tool)
- [The Harness Abstraction](#the-harness-abstraction)
- [Scaling with the Gibson Daemon](#scaling-with-the-gibson-daemon)
- [Architectural Considerations](#architectural-considerations)
- [Best Practices](#best-practices)

---

## Architecture Overview

Tools in Gibson follow a **unified interface** pattern. From an agent's perspective, calling a tool is always the same:

```go
result, err := harness.CallTool(ctx, "my-tool", input)
```

The agent has **no knowledge** of whether the tool is:
- A subprocess binary running locally
- A gRPC service running on a remote machine
- An in-process Go function

This abstraction is handled by the **harness**, which routes tool calls to the appropriate execution backend.

### Component Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                         Agent                                │
│                           │                                  │
│                    harness.CallTool()                        │
│                           │                                  │
└───────────────────────────┼──────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Tool Registry (Harness)                   │
│                           │                                  │
│         ┌─────────────────┼─────────────────┐               │
│         ▼                 ▼                 ▼               │
│  ┌────────────┐    ┌────────────┐    ┌────────────┐        │
│  │ In-Process │    │   gRPC     │    │ Subprocess │        │
│  │   Tool     │    │   Client   │    │  Executor  │        │
│  └────────────┘    └────────────┘    └────────────┘        │
│                           │                 │               │
└───────────────────────────┼─────────────────┼───────────────┘
                            │                 │
                            ▼                 ▼
                    ┌────────────┐    ┌────────────┐
                    │   Remote   │    │   Local    │
                    │   gRPC     │    │  Binary    │
                    │   Server   │    │  Process   │
                    └────────────┘    └────────────┘
```

---

## Tool Execution Modes

### 1. Subprocess Tools

**Location**: `~/.gibson/tools/bin/`

Subprocess tools are standalone binaries that communicate via stdin/stdout using JSON.

**How they work:**
1. The `ToolExecutorService` scans `~/.gibson/tools/bin/` on daemon startup
2. For each binary, it calls `./tool --schema` to fetch metadata
3. When an agent calls the tool, the daemon:
   - Spawns the binary as a subprocess
   - Writes JSON input to stdin
   - Reads JSON output from stdout
   - Applies timeout and handles errors

**Protocol:**
```bash
# Get tool schema
./my-tool --schema
# Output: {"name": "...", "version": "...", "input_schema": {...}, "output_schema": {...}}

# Execute tool
echo '{"param": "value"}' | ./my-tool
# Output: {"result": "..."}
```

**Characteristics:**
- Stateless - fresh process per invocation
- Complete process isolation
- Any language that can read/write JSON
- Higher latency (process spawn overhead)
- Simple to develop and test

### 2. gRPC Tools

**Registration**: Via etcd service registry

gRPC tools are long-running services that register themselves in etcd and communicate via gRPC.

**How they work:**
1. Tool starts and registers itself in etcd with endpoint and metadata
2. Daemon discovers the tool via registry lookup
3. When an agent calls the tool, the daemon:
   - Looks up the tool in etcd
   - Creates a gRPC client connection (or reuses existing)
   - Makes an RPC call with JSON-encoded input
   - Returns JSON-decoded output

**Characteristics:**
- Can be stateful (connection pools, caches, sessions)
- Lower latency (persistent connections)
- Network-distributed deployment
- More infrastructure complexity
- Supports health checks and graceful shutdown

### 3. In-Process Tools (Internal)

In-process tools are Go functions registered directly in the daemon's tool registry. These are typically used for built-in framework tools and are not the primary focus of this guide.

---

## When to Use Which Mode

### Use Subprocess Tools When:

| Scenario | Why Subprocess |
|----------|----------------|
| **Simple utilities** | Minimal overhead, easy to write |
| **Prototyping** | Quick iteration without service infrastructure |
| **Language flexibility** | Python, Rust, Node.js, shell scripts |
| **Stateless operations** | No benefit from persistent connections |
| **Isolated execution** | Security through process boundaries |
| **Infrequent calls** | Process spawn overhead is acceptable |

**Examples:**
- DNS lookup tool
- File hash calculator
- Simple HTTP request tool
- Text parsing utility
- Image analysis tool (wrapping CLI)

### Use gRPC Tools When:

| Scenario | Why gRPC |
|----------|----------|
| **High-frequency calls** | Avoid process spawn overhead |
| **Stateful operations** | Database connections, API sessions |
| **Distributed deployment** | Run on different machines |
| **Warm caches** | Keep data in memory between calls |
| **Complex initialization** | Load models, establish connections once |
| **Real-time monitoring** | Health checks, metrics exposure |

**Examples:**
- Database query tool (connection pooling)
- ML inference tool (model loaded in memory)
- Browser automation tool (persistent browser instance)
- Rate-limited API client (token bucket in memory)
- Streaming data processor

### Decision Flowchart

```
                    ┌───────────────────────┐
                    │ Does your tool need   │
                    │ persistent state?     │
                    └───────────┬───────────┘
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
                   YES                      NO
                    │                       │
                    ▼                       ▼
            ┌───────────────┐       ┌───────────────────┐
            │   Use gRPC    │       │ Called frequently │
            │               │       │ (>10 calls/sec)?  │
            └───────────────┘       └─────────┬─────────┘
                                              │
                                  ┌───────────┴───────────┐
                                  ▼                       ▼
                                 YES                      NO
                                  │                       │
                                  ▼                       ▼
                          ┌───────────────┐       ┌───────────────┐
                          │   Use gRPC    │       │ Use Subprocess│
                          └───────────────┘       └───────────────┘
```

---

## Quick Start: Subprocess Tool

### Step 1: Implement the Tool Interface

Create a new Go file for your tool:

```go
// main.go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/serve"
    "github.com/zero-day-ai/sdk/types"
)

// DNSLookupTool performs DNS lookups
type DNSLookupTool struct{}

func (t *DNSLookupTool) Name() string {
    return "dns-lookup"
}

func (t *DNSLookupTool) Version() string {
    return "1.0.0"
}

func (t *DNSLookupTool) Description() string {
    return "Perform DNS lookups for domain names"
}

func (t *DNSLookupTool) Tags() []string {
    return []string{"network", "reconnaissance", "dns"}
}

func (t *DNSLookupTool) InputSchema() schema.JSON {
    return schema.Object(map[string]schema.JSON{
        "domain":      schema.StringWithDesc("Domain name to lookup"),
        "record_type": schema.StringWithDesc("DNS record type (A, AAAA, MX, TXT, etc.)"),
    }, "domain") // "domain" is required
}

func (t *DNSLookupTool) OutputSchema() schema.JSON {
    return schema.Object(map[string]schema.JSON{
        "domain":  schema.StringWithDesc("Queried domain"),
        "records": schema.Array(schema.String(), "DNS records found"),
        "ttl":     schema.IntegerWithDesc("Time to live in seconds"),
    }, "domain", "records")
}

func (t *DNSLookupTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    domain := input["domain"].(string)
    recordType := "A"
    if rt, ok := input["record_type"].(string); ok && rt != "" {
        recordType = rt
    }

    // Your DNS lookup logic here
    records, ttl, err := performDNSLookup(domain, recordType)
    if err != nil {
        return nil, fmt.Errorf("DNS lookup failed: %w", err)
    }

    return map[string]any{
        "domain":  domain,
        "records": records,
        "ttl":     ttl,
    }, nil
}

func (t *DNSLookupTool) Health(ctx context.Context) types.HealthStatus {
    return types.NewHealthyStatus("DNS resolver available")
}

func main() {
    tool := &DNSLookupTool{}

    // Handle --schema flag (required for subprocess discovery)
    if len(os.Args) > 1 && os.Args[1] == "--schema" {
        if err := serve.OutputSchema(tool); err != nil {
            fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
            os.Exit(1)
        }
        os.Exit(0)
    }

    // Run in subprocess mode (reads JSON from stdin, writes to stdout)
    if err := serve.RunSubprocess(tool); err != nil {
        os.Exit(1)
    }
}

func performDNSLookup(domain, recordType string) ([]string, int, error) {
    // Implementation details...
    return []string{"192.168.1.1"}, 300, nil
}
```

### Step 2: Build and Deploy

```bash
# Build the tool
go build -o dns-lookup ./main.go

# Deploy to Gibson tools directory
cp dns-lookup ~/.gibson/tools/bin/

# Verify the tool is discoverable
./dns-lookup --schema
```

### Step 3: Test Manually

```bash
# Test execution
echo '{"domain": "example.com", "record_type": "A"}' | ./dns-lookup
```

### Step 4: Refresh the Daemon

The daemon will automatically discover new tools on startup. To hot-reload:

```bash
# Via gibson CLI (if available)
gibson tool refresh

# Or restart the daemon
gibson daemon restart
```

---

## Quick Start: gRPC Tool

### Step 1: Implement the Tool with gRPC Server

```go
// main.go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/serve"
    "github.com/zero-day-ai/sdk/types"
)

// RateLimitedHTTPTool makes HTTP requests with rate limiting
type RateLimitedHTTPTool struct {
    mu          sync.Mutex
    lastRequest time.Time
    rateLimit   time.Duration
}

func NewRateLimitedHTTPTool(requestsPerSecond int) *RateLimitedHTTPTool {
    return &RateLimitedHTTPTool{
        rateLimit: time.Second / time.Duration(requestsPerSecond),
    }
}

func (t *RateLimitedHTTPTool) Name() string {
    return "rate-limited-http"
}

func (t *RateLimitedHTTPTool) Version() string {
    return "1.0.0"
}

func (t *RateLimitedHTTPTool) Description() string {
    return "Make HTTP requests with built-in rate limiting"
}

func (t *RateLimitedHTTPTool) Tags() []string {
    return []string{"http", "network", "rate-limited"}
}

func (t *RateLimitedHTTPTool) InputSchema() schema.JSON {
    return schema.Object(map[string]schema.JSON{
        "url":    schema.StringWithDesc("URL to request"),
        "method": schema.StringWithDesc("HTTP method (GET, POST, etc.)"),
    }, "url")
}

func (t *RateLimitedHTTPTool) OutputSchema() schema.JSON {
    return schema.Object(map[string]schema.JSON{
        "status": schema.IntegerWithDesc("HTTP status code"),
        "body":   schema.StringWithDesc("Response body"),
    }, "status")
}

func (t *RateLimitedHTTPTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    // Enforce rate limiting (stateful - benefits from gRPC)
    t.mu.Lock()
    elapsed := time.Since(t.lastRequest)
    if elapsed < t.rateLimit {
        time.Sleep(t.rateLimit - elapsed)
    }
    t.lastRequest = time.Now()
    t.mu.Unlock()

    url := input["url"].(string)
    method := "GET"
    if m, ok := input["method"].(string); ok {
        method = m
    }

    // Make HTTP request...
    status, body, err := makeHTTPRequest(ctx, method, url)
    if err != nil {
        return nil, err
    }

    return map[string]any{
        "status": status,
        "body":   body,
    }, nil
}

func (t *RateLimitedHTTPTool) Health(ctx context.Context) types.HealthStatus {
    return types.NewHealthyStatus("Rate limiter active")
}

func main() {
    tool := NewRateLimitedHTTPTool(10) // 10 requests per second

    // serve.Tool automatically handles:
    // - --schema flag for subprocess mode
    // - GIBSON_TOOL_MODE=subprocess for stdin/stdout mode
    // - Default: starts gRPC server with registry registration
    err := serve.Tool(tool,
        serve.WithPort(50052),                        // gRPC server port
        serve.WithGracefulShutdown(30*time.Second),   // Graceful shutdown timeout
        // serve.WithRegistry(registry),              // Optional: auto-register with etcd
    )
    if err != nil {
        log.Fatalf("Failed to serve tool: %v", err)
    }
}

func makeHTTPRequest(ctx context.Context, method, url string) (int, string, error) {
    // Implementation...
    return 200, "OK", nil
}
```

### Step 2: Build and Run

```bash
# Build the tool
go build -o rate-limited-http ./main.go

# Run as gRPC server
./rate-limited-http
# Output: tool server started component=tool name=rate-limited-http version=1.0.0 port=50052
```

### Step 3: Register with etcd (Production)

For production deployments, the tool should register itself with etcd:

```go
// In main()
import "github.com/zero-day-ai/sdk/registry"

// Create registry client
reg, err := registry.NewManager(registry.Config{
    Endpoints: []string{"localhost:2379"},
})
if err != nil {
    log.Fatal(err)
}

// Serve with auto-registration
err = serve.Tool(tool,
    serve.WithPort(50052),
    serve.WithRegistry(reg),
    serve.WithAdvertiseAddr("tool-server.internal:50052"), // Address other services use to reach this tool
)
```

### Step 4: Verify Registration

```bash
# Check etcd for registration
etcdctl get --prefix /gibson/tools/
```

---

## The Harness Abstraction

The harness provides a **unified interface** for agents to call tools, regardless of their execution mode.

### How Routing Works

When an agent calls `harness.CallTool(ctx, "my-tool", input)`:

```go
// Simplified routing logic in harness implementation
func (h *DefaultAgentHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
    // 1. Check local tool registry (in-process and gRPC clients)
    tool, err := h.toolRegistry.Get(name)
    if err == nil {
        return tool.Execute(ctx, input)
    }

    // 2. Try remote discovery via etcd registry
    if h.registryAdapter != nil {
        remoteTool, err := h.registryAdapter.DiscoverTool(ctx, name)
        if err == nil {
            return remoteTool.Execute(ctx, input)
        }
    }

    // 3. Check subprocess tool executor service
    if h.toolExecutorService != nil {
        return h.toolExecutorService.Execute(ctx, name, input, timeout)
    }

    return nil, ErrToolNotFound
}
```

### Priority Order

1. **Local in-process tools** - Registered directly in the daemon
2. **gRPC tools (via registry)** - Discovered through etcd
3. **Subprocess tools** - Binaries in `~/.gibson/tools/bin/`

This means if you have both a gRPC tool and a subprocess tool with the same name, the gRPC tool takes precedence.

---

## Scaling with the Gibson Daemon

### Single Daemon Deployment

```
┌─────────────────────────────────────────────────────────────┐
│                      Gibson Daemon                           │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Tool Executor Service                   │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐               │    │
│  │  │ dns-    │ │ http-   │ │ nmap-   │  Subprocess   │    │
│  │  │ lookup  │ │ client  │ │ scanner │  Tools        │    │
│  │  └─────────┘ └─────────┘ └─────────┘               │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              etcd Registry (Embedded)                │    │
│  └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Distributed Deployment

```
┌─────────────────────────────────────────────────────────────┐
│                       Machine A                              │
│  ┌───────────────────────┐                                  │
│  │    Gibson Daemon      │                                  │
│  │  ┌─────────────────┐  │                                  │
│  │  │ Subprocess Tools│  │                                  │
│  │  └─────────────────┘  │                                  │
│  └───────────┬───────────┘                                  │
│              │                                               │
└──────────────┼───────────────────────────────────────────────┘
               │
               │ gRPC
               │
┌──────────────┼───────────────────────────────────────────────┐
│              │            Machine B                          │
│  ┌───────────▼───────────┐  ┌────────────────────────────┐  │
│  │   etcd Cluster        │  │   gRPC Tool: ML Inference  │  │
│  │   (External)          │  │   (GPU-enabled)            │  │
│  └───────────────────────┘  └────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
               │
               │ gRPC
               │
┌──────────────┼───────────────────────────────────────────────┐
│              │            Machine C                          │
│  ┌───────────▼───────────┐  ┌────────────────────────────┐  │
│  │   gRPC Tool:          │  │   gRPC Tool:               │  │
│  │   Browser Automation  │  │   Database Client          │  │
│  │   (Chromium)          │  │   (Connection Pool)        │  │
│  └───────────────────────┘  └────────────────────────────┘  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### Kubernetes Deployment

```yaml
# Example: gRPC tool as a Kubernetes deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ml-inference-tool
spec:
  replicas: 3
  selector:
    matchLabels:
      app: ml-inference-tool
  template:
    spec:
      containers:
      - name: tool
        image: gibson/ml-inference-tool:latest
        ports:
        - containerPort: 50051
        env:
        - name: ETCD_ENDPOINTS
          value: "etcd-cluster:2379"
        - name: ADVERTISE_ADDR
          value: "ml-inference-tool.gibson.svc.cluster.local:50051"
        resources:
          limits:
            nvidia.com/gpu: 1
---
apiVersion: v1
kind: Service
metadata:
  name: ml-inference-tool
spec:
  ports:
  - port: 50051
  selector:
    app: ml-inference-tool
```

---

## Architectural Considerations

### Strengths of This Architecture

1. **Clean Abstraction**
   - Agents are decoupled from tool implementation details
   - Same agent code works with any tool deployment mode
   - Easy to migrate tools between modes without agent changes

2. **Flexible Deployment**
   - Start simple with subprocess tools
   - Scale to gRPC when needed
   - Mix and match based on requirements

3. **Language Agnostic (Subprocess)**
   - Tools can be written in any language
   - Simple JSON protocol is universally supported
   - Great for wrapping existing CLI tools

4. **Hot Reloading**
   - `ToolExecutorService.RefreshTools()` rescans without daemon restart
   - gRPC tools can register/deregister dynamically via etcd TTL

5. **Observability Built-in**
   - Execution metrics tracked per tool
   - Health checks for gRPC tools
   - Distributed tracing through harness

### Known Limitations

1. **Subprocess Overhead**
   - Process spawn adds ~10-50ms latency per call
   - Not suitable for high-frequency operations (>100 calls/sec)
   - Memory overhead for each spawned process

2. **No Streaming Support (Subprocess)**
   - Subprocess tools must return complete output at once
   - gRPC tools could support streaming but interface doesn't expose it to agents

3. **Schema Discovery Timing**
   - Subprocess schemas fetched at daemon startup or refresh
   - If schema changes, daemon must be refreshed
   - gRPC tools fetch schema on connection, slightly more dynamic

4. **Error Propagation**
   - Subprocess errors limited to exit codes and stderr
   - Rich error types only available with gRPC tools
   - Error context can be lost across process boundaries

5. **Resource Management**
   - Subprocess tools can't manage shared resources across calls
   - No connection pooling, caching, or warm state
   - Each invocation starts fresh

6. **Security Boundaries**
   - Subprocess tools run with daemon's permissions
   - gRPC tools need separate authentication/authorization
   - etcd registry is trusted - compromised etcd = compromised tools

### When NOT to Use Subprocess Tools

- **Database operations**: Connection overhead per call is prohibitive
- **ML inference**: Model loading time dominates execution
- **Browser automation**: Browser startup is expensive
- **Stateful operations**: Session management, transactions
- **High-throughput**: >100 calls per second per tool

---

## Best Practices

### General

1. **Start with Subprocess, Graduate to gRPC**
   - Prototype with subprocess for simplicity
   - Migrate to gRPC when you hit performance limits or need state

2. **Schema Design**
   - Keep input schemas simple and flat when possible
   - Use descriptive field names and descriptions
   - Mark truly required fields as required

3. **Error Handling**
   - Return structured errors with context
   - Use appropriate error codes
   - Log errors with relevant details

### Subprocess Tools

1. **Handle `--schema` First**
   ```go
   if len(os.Args) > 1 && os.Args[1] == "--schema" {
       serve.OutputSchema(tool)
       os.Exit(0)
   }
   ```

2. **Exit Codes Matter**
   - `0`: Success
   - `1`: Error (check stderr)
   - Write errors to stderr, not stdout

3. **Timeout Handling**
   - Respect context cancellation
   - Clean up resources on timeout
   - Don't hang indefinitely

### gRPC Tools

1. **Register on Startup, Deregister on Shutdown**
   ```go
   // Register
   registry.Register(ctx, serviceInfo)

   // Graceful shutdown
   defer registry.Deregister(ctx, serviceInfo)
   ```

2. **Implement Health Checks**
   - Return accurate health status
   - Check dependencies (DB connections, etc.)
   - Use gRPC health protocol

3. **Connection Management**
   - Use connection pooling for downstream services
   - Implement circuit breakers for external dependencies
   - Handle connection failures gracefully

### Testing

```go
// Unit test your tool execution
func TestMyTool_Execute(t *testing.T) {
    tool := &MyTool{}

    input := map[string]any{
        "param": "value",
    }

    output, err := tool.Execute(context.Background(), input)

    require.NoError(t, err)
    assert.Equal(t, "expected", output["result"])
}

// Integration test subprocess mode
func TestMyTool_Subprocess(t *testing.T) {
    // Build the tool
    cmd := exec.Command("go", "build", "-o", "test-tool", ".")
    require.NoError(t, cmd.Run())
    defer os.Remove("test-tool")

    // Test schema output
    schemaCmd := exec.Command("./test-tool", "--schema")
    schemaOut, err := schemaCmd.Output()
    require.NoError(t, err)

    var schema map[string]any
    require.NoError(t, json.Unmarshal(schemaOut, &schema))
    assert.Equal(t, "my-tool", schema["name"])

    // Test execution
    execCmd := exec.Command("./test-tool")
    execCmd.Stdin = strings.NewReader(`{"param": "value"}`)
    execOut, err := execCmd.Output()
    require.NoError(t, err)

    var output map[string]any
    require.NoError(t, json.Unmarshal(execOut, &output))
    assert.Equal(t, "expected", output["result"])
}
```

---

## Reference

### Tool Interface

```go
type Tool interface {
    Name() string                                           // Unique identifier
    Version() string                                        // Semantic version
    Description() string                                    // Human-readable description
    Tags() []string                                         // Discovery tags
    InputSchema() schema.JSON                               // JSON Schema for input
    OutputSchema() schema.JSON                              // JSON Schema for output
    Execute(ctx context.Context, input map[string]any) (map[string]any, error)
    Health(ctx context.Context) types.HealthStatus
}
```

### Key Files

| Component | Location |
|-----------|----------|
| Tool Interface | `opensource/sdk/tool/tool.go` |
| Tool Builder | `opensource/sdk/tool/builder.go` |
| Serve Functions | `opensource/sdk/serve/tool.go` |
| Subprocess Protocol | `opensource/sdk/serve/subprocess.go` |
| Tool Executor Service | `opensource/gibson/internal/daemon/toolexec/` |
| gRPC Client | `opensource/gibson/internal/tool/grpc_client.go` |
| Tool Registry | `opensource/gibson/internal/tool/registry.go` |
| Harness Implementation | `opensource/gibson/internal/harness/implementation.go` |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GIBSON_TOOL_MODE=subprocess` | Force subprocess mode (stdin/stdout JSON) |

### CLI Commands

```bash
# List registered tools
gibson tool list

# Get tool schema
gibson tool schema <name>

# Execute tool manually
gibson tool exec <name> --input '{"param": "value"}'

# Refresh tool registry
gibson tool refresh

# Check tool health
gibson tool health <name>
```
