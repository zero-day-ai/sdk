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
- [GraphRAG Taxonomy Integration](#graphrag-taxonomy-integration)
- [Future Tool Development Roadmap](#future-tool-development-roadmap)

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

---

## GraphRAG Taxonomy Integration

Tools can embed GraphRAG taxonomy directly in their schema, enabling automatic knowledge graph population. When a tool outputs data, Gibson extracts entities and relationships based on the embedded taxonomy.

### How Taxonomy Works

When you define a tool's `OutputSchema()`, you can include a `TaxonomyMapping` that tells Gibson how to extract knowledge graph nodes and relationships from the tool's output:

```go
func (t *NmapTool) OutputSchema() schema.JSON {
    return schema.Object(map[string]schema.JSON{
        "hosts": schema.Array(schema.Object(map[string]schema.JSON{
            "ip":       schema.StringWithDesc("IP address"),
            "hostname": schema.StringWithDesc("Hostname"),
            "ports":    schema.Array(portSchema, "Open ports"),
        }), "Discovered hosts"),
    }).WithTaxonomy(schema.TaxonomyMapping{
        Extracts: []schema.ExtractRule{
            {
                JSONPath:   "$.hosts[*]",
                NodeType:   "host",
                IDTemplate: "host:{.ip}",
                Properties: []schema.PropertyMapping{
                    {Source: "ip", Target: "ip"},
                    {Source: "hostname", Target: "hostname"},
                },
                Relationships: []schema.RelationshipMapping{
                    {
                        Type:         "DISCOVERED_BY",
                        FromTemplate: "host:{.ip}",
                        ToTemplate:   "agent_run:{_context.agent_run_id}",
                    },
                },
            },
            {
                JSONPath:   "$.hosts[*].ports[*]",
                NodeType:   "port",
                IDTemplate: "port:{_parent.ip}:{.port}",
                Properties: []schema.PropertyMapping{
                    {Source: "port", Target: "port_number"},
                    {Source: "protocol", Target: "protocol"},
                    {Source: "service", Target: "service"},
                },
                Relationships: []schema.RelationshipMapping{
                    {
                        Type:         "HAS_PORT",
                        FromTemplate: "host:{_parent.ip}",
                        ToTemplate:   "port:{_parent.ip}:{.port}",
                    },
                },
            },
        },
    })
}
```

### Key Concepts

| Concept | Description |
|---------|-------------|
| **Node Types** | Entity categories in the knowledge graph (host, port, domain, finding, etc.) |
| **ID Template** | How to construct unique IDs for nodes using JSONPath expressions |
| **Properties** | Field mappings from tool output to node properties |
| **Relationships** | Edges between nodes with directional type labels |
| **JSONPath** | Path to extract arrays/objects from tool output (`$.hosts[*]`, `$.results[*].ports[*]`) |

### Current Node Types

Tools should use these standard node types for consistency:

| Node Type | Description | Example ID Template |
|-----------|-------------|---------------------|
| `domain` | Root domain | `domain:{.name}` |
| `subdomain` | Subdomain of a domain | `subdomain:{.name}` |
| `host` | IP address or hostname | `host:{.ip}` |
| `port` | Network port | `port:{.host}:{.port}` |
| `endpoint` | HTTP/API endpoint | `endpoint:{.url}` |
| `technology` | Software/framework detected | `technology:{.name}:{.version}` |
| `finding` | Vulnerability or security finding | `finding:{.template_id}:{.host}` |
| `asn` | Autonomous System Number | `asn:{.number}` |
| `dns_record` | DNS record | `dns:{.type}:{.name}` |

---

## Future Tool Development Roadmap

This section catalogs **all security tools** that bug bounty hunters and security researchers commonly use, organized by MITRE ATT&CK taxonomy phases. Each tool includes its primary use case, expected GraphRAG node types, and implementation priority.

### Currently Implemented (6 tools)

These tools have embedded GraphRAG taxonomy and are production-ready:

| Tool | Phase | Node Types | Description |
|------|-------|------------|-------------|
| **nmap** | Discovery | host, port | Network exploration and port scanning |
| **masscan** | Discovery | host, port | High-speed TCP port scanner |
| **subfinder** | Reconnaissance | domain, subdomain | Passive subdomain enumeration |
| **httpx** | Reconnaissance | endpoint, technology | HTTP probing and technology detection |
| **nuclei** | Reconnaissance | finding | Template-based vulnerability scanning |
| **amass** | Reconnaissance | domain, subdomain, host, asn, dns_record | Attack surface mapping |

---

### Reconnaissance (TA0043) - Attack Surface Discovery

Tools for discovering targets, subdomains, and gathering OSINT.

#### Subdomain Enumeration

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **subfinder** | ✅ Done | domain, subdomain | Passive subdomain discovery using APIs |
| **amass** | ✅ Done | domain, subdomain, host, asn, dns_record | Comprehensive attack surface mapping |
| **assetfinder** | High | domain, subdomain | Find domains and subdomains from public sources |
| **findomain** | High | domain, subdomain | Fast subdomain enumeration |
| **knockpy** | Medium | domain, subdomain | DNS enumeration with wordlist |
| **sublist3r** | Medium | domain, subdomain | Subdomain enumeration using search engines |
| **crt.sh** | Medium | domain, subdomain, certificate | Certificate transparency log search |
| **massdns** | High | domain, subdomain, dns_record | High-performance DNS resolver |
| **shuffledns** | Medium | domain, subdomain | Wrapper around massdns for active bruteforce |
| **puredns** | Medium | domain, subdomain | Fast domain resolver and bruteforcer |
| **dnsrecon** | Medium | domain, dns_record | DNS enumeration and zone transfer |
| **dnsx** | High | domain, subdomain, dns_record | Fast DNS toolkit with wildcard filtering |
| **alterx** | Low | domain, subdomain | Subdomain permutation generator |

#### DNS & Network Analysis

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **dnsx** | High | dns_record, subdomain | DNS query toolkit |
| **asnmap** | High | asn, host, cidr | ASN to CIDR mapping |
| **whois** | Medium | domain, registrant, organization | Domain registration lookup |
| **dig** | Low | dns_record | DNS query tool |
| **nslookup** | Low | dns_record | DNS query utility |
| **host** | Low | dns_record, host | DNS lookup utility |

#### Web Enumeration & Crawling

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **httpx** | ✅ Done | endpoint, technology | HTTP probing with technology detection |
| **katana** | High | endpoint, url, parameter | Next-gen web crawler |
| **gospider** | High | endpoint, url, parameter | Fast web spider |
| **hakrawler** | Medium | endpoint, url | Simple web crawler |
| **gau** | High | endpoint, url | Fetch known URLs from AlienVault, Wayback |
| **waybackurls** | High | endpoint, url | Fetch URLs from Wayback Machine |
| **getallurls** | Medium | endpoint, url | URL extraction from various sources |
| **paramspider** | Medium | endpoint, parameter | Mining parameters from web archives |
| **arjun** | Medium | endpoint, parameter | HTTP parameter discovery |
| **x8** | Medium | endpoint, parameter | Hidden parameter discovery |
| **feroxbuster** | High | endpoint, directory | Fast content discovery |
| **gobuster** | High | endpoint, directory | Directory/DNS/VHost brute-forcer |
| **ffuf** | High | endpoint, directory, parameter | Fast web fuzzer |
| **dirsearch** | Medium | endpoint, directory | Web path discovery |
| **dirb** | Low | endpoint, directory | Web content scanner |
| **wfuzz** | Medium | endpoint, parameter | Web application fuzzer |

#### Technology & Service Detection

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **whatweb** | High | technology, endpoint | Web technology identification |
| **wappalyzer** | High | technology, endpoint | Technology profiler |
| **webanalyze** | Medium | technology, endpoint | Port of Wappalyzer |
| **builtwith** | Medium | technology | Technology lookup API |
| **retire.js** | Medium | technology, vulnerability | JavaScript library vulnerability scanner |
| **wafw00f** | Medium | technology, waf | WAF detection tool |
| **identywaf** | Low | technology, waf | WAF identification |

#### OSINT & Intelligence Gathering

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **shodan** | High | host, port, technology, banner | Internet-wide scanner database |
| **censys** | High | host, port, certificate, technology | Internet scan data search |
| **zoomeye** | Medium | host, port, technology | Chinese internet scanner |
| **fofa** | Medium | host, port, technology | Network asset search |
| **binaryedge** | Medium | host, port, vulnerability | Threat intelligence platform |
| **securitytrails** | High | domain, subdomain, dns_record, history | Historical DNS data |
| **passivetotal** | Medium | domain, host, threat | Threat intelligence |
| **virustotal** | Medium | domain, host, file, malware | Multi-AV scanner and intelligence |
| **urlscan.io** | Medium | endpoint, screenshot, technology | Website scanner |
| **spiderfoot** | High | domain, host, email, social | OSINT automation |
| **theHarvester** | High | domain, email, host, employee | Email/subdomain/IP harvester |
| **maltego** | Low | entity, relationship | OSINT visualization |
| **recon-ng** | High | domain, host, email, credential | Web reconnaissance framework |
| **osintframework** | Low | various | OSINT resource collection |

#### Cloud & Infrastructure Recon

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **cloud_enum** | High | bucket, blob, host | Cloud asset enumeration |
| **S3Scanner** | High | bucket, file | AWS S3 bucket enumeration |
| **GCPBucketBrute** | Medium | bucket | GCP bucket enumeration |
| **AzureHound** | Medium | azure_resource, user, group | Azure AD enumeration |
| **ScoutSuite** | High | cloud_resource, misconfiguration | Multi-cloud security audit |
| **Prowler** | High | aws_resource, finding | AWS security assessment |
| **CloudMapper** | Medium | aws_resource, network | AWS environment analysis |
| **pacu** | Medium | aws_resource, exploit | AWS exploitation framework |

#### Email & Social OSINT

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **hunter.io** | High | email, domain, employee | Email finder API |
| **phonebook.cz** | Medium | email, domain, subdomain | Email/subdomain search |
| **emailrep.io** | Medium | email, reputation | Email reputation lookup |
| **haveibeenpwned** | High | email, breach | Breach database search |
| **dehashed** | Medium | credential, breach | Leaked credential search |
| **intelx.io** | Medium | credential, breach, document | Intelligence X search |
| **linkedin2username** | Medium | employee, email | LinkedIn username generator |
| **socialscan** | Low | social_account | Social media username check |

#### GitHub & Code Recon

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **gitrob** | High | repository, secret, credential | GitHub secret scanner |
| **trufflehog** | High | secret, credential, repository | Credential scanner |
| **gitleaks** | High | secret, credential | Git secret scanner |
| **shhgit** | Medium | secret, credential | Real-time GitHub secret finder |
| **gitdorker** | Medium | repository, finding | GitHub dork automation |
| **github-search** | Medium | repository, code | GitHub search automation |

---

### Discovery (TA0007) - Network & Service Discovery

Tools for active scanning and service enumeration.

#### Port Scanning

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **nmap** | ✅ Done | host, port, service, os | Network mapper and port scanner |
| **masscan** | ✅ Done | host, port | Fastest port scanner |
| **rustscan** | High | host, port | Fast Rust-based port scanner |
| **zmap** | Medium | host, port | Internet-wide scanner |
| **unicornscan** | Low | host, port | Asynchronous port scanner |
| **naabu** | High | host, port | Fast port scanner in Go |

#### Service Enumeration

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **nikto** | High | endpoint, vulnerability, technology | Web server scanner |
| **whatweb** | High | technology, endpoint | Web fingerprinting |
| **wpscan** | High | technology, vulnerability, user | WordPress scanner |
| **joomscan** | Medium | technology, vulnerability | Joomla scanner |
| **droopescan** | Medium | technology, vulnerability | CMS scanner |
| **drupwn** | Low | technology, vulnerability | Drupal scanner |
| **CMSmap** | Medium | technology, vulnerability | CMS detection and exploitation |
| **sslscan** | High | certificate, cipher, vulnerability | SSL/TLS scanner |
| **sslyze** | High | certificate, cipher, vulnerability | SSL configuration analyzer |
| **testssl.sh** | Medium | certificate, vulnerability | SSL/TLS testing |

#### Active Directory & Windows

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **bloodhound** | High | ad_user, ad_group, ad_computer, path | AD attack path mapping |
| **sharphound** | High | ad_object, relationship | BloodHound data collector |
| **ldapdomaindump** | High | ad_user, ad_group, ad_computer | LDAP enumeration |
| **adidnsdump** | Medium | dns_record, ad_object | AD integrated DNS dump |
| **kerbrute** | High | ad_user, credential | Kerberos bruteforce |
| **enum4linux** | High | smb_share, ad_user, ad_group | SMB/NetBIOS enumeration |
| **enum4linux-ng** | High | smb_share, ad_user, ad_group | Next-gen enum4linux |
| **crackmapexec** | High | host, credential, smb_share | Swiss army knife for Windows |
| **netexec** | High | host, credential, smb_share | CrackMapExec successor |
| **rpcclient** | Medium | ad_user, ad_group | RPC client for Windows |
| **smbmap** | High | smb_share, file | SMB share enumeration |
| **smbclient** | Medium | smb_share | SMB client |

#### Network Service Scanning

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **snmpwalk** | Medium | snmp_data, host | SNMP enumeration |
| **onesixtyone** | Medium | snmp_community, host | SNMP scanner |
| **nbtscan** | Low | netbios_name, host | NetBIOS scanner |
| **fierce** | Medium | domain, host | DNS reconnaissance |
| **dnsmap** | Low | subdomain | DNS bruteforcing |

---

### Vulnerability Assessment (TA0043/TA0007)

Tools for identifying vulnerabilities.

#### Vulnerability Scanners

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **nuclei** | ✅ Done | finding, vulnerability | Template-based scanner |
| **nikto** | High | vulnerability, endpoint | Web vulnerability scanner |
| **openvas** | Medium | vulnerability, host | Open-source vulnerability scanner |
| **nessus** | Medium | vulnerability, host | Commercial vulnerability scanner |
| **trivy** | High | vulnerability, container, package | Container vulnerability scanner |
| **grype** | High | vulnerability, package | Container/SBOM vulnerability scanner |
| **snyk** | Medium | vulnerability, package | Developer-first security |
| **retire.js** | Medium | vulnerability, javascript | JavaScript vulnerability scanner |

#### Web Application Scanners

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **burpsuite** | High | vulnerability, endpoint, parameter | Web security testing platform |
| **zaproxy** | High | vulnerability, endpoint | OWASP ZAP proxy |
| **arachni** | Medium | vulnerability, endpoint | Web application scanner |
| **w3af** | Medium | vulnerability, endpoint | Web application attack framework |
| **skipfish** | Low | vulnerability, endpoint | Web application security scanner |
| **vega** | Low | vulnerability, endpoint | Web vulnerability scanner |

#### API Security Testing

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **postman** | Medium | endpoint, parameter | API platform |
| **insomnia** | Medium | endpoint | API client |
| **kiterunner** | High | endpoint, api | API endpoint discovery |
| **mitmproxy** | High | endpoint, request, response | Interactive HTTPS proxy |
| **fuzzapi** | Medium | endpoint, parameter | REST API fuzzer |

---

### Initial Access (TA0001) - Exploitation

Tools for gaining initial access through vulnerabilities.

#### Web Exploitation

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **sqlmap** | High | vulnerability, database, credential | SQL injection automation |
| **nosqlmap** | Medium | vulnerability, database | NoSQL injection |
| **commix** | Medium | vulnerability, command | Command injection exploitation |
| **tplmap** | Medium | vulnerability | Template injection exploitation |
| **sstimap** | Medium | vulnerability | SSTI exploitation |
| **xsstrike** | High | vulnerability, parameter | XSS scanner and exploiter |
| **dalfox** | High | vulnerability, parameter | XSS scanner |
| **kxss** | Medium | vulnerability | XSS detection |
| **xxeinjector** | Medium | vulnerability | XXE exploitation |
| **crlfuzz** | Medium | vulnerability | CRLF injection scanner |
| **corsscanner** | Medium | vulnerability | CORS misconfiguration scanner |

#### Authentication Attacks

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **hydra** | High | credential, service | Network login cracker |
| **medusa** | Medium | credential, service | Parallel login brute-forcer |
| **ncrack** | Medium | credential, service | Network authentication cracker |
| **patator** | Medium | credential | Multi-purpose brute-forcer |
| **crowbar** | Medium | credential | Brute-forcing tool for services |
| **sprayhound** | Medium | credential, ad_user | Password spraying for AD |
| **trevorspray** | Medium | credential | Microsoft Online password sprayer |
| **o365spray** | Medium | credential | Microsoft O365 password sprayer |

#### Exploit Frameworks

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **metasploit** | High | exploit, payload, session | Penetration testing framework |
| **exploitdb** | High | exploit, vulnerability | Exploit database (searchsploit) |
| **routersploit** | Medium | exploit, vulnerability | Router exploitation |
| **autosploit** | Low | exploit | Automated mass exploitation |
| **pwntools** | Medium | exploit | CTF/exploit development |
| **ropper** | Medium | gadget | ROP gadget finder |
| **one_gadget** | Medium | gadget | One-shot RCE gadget finder |

---

### Credential Access (TA0006) - Credential Theft

Tools for obtaining credentials.

#### Password Cracking

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **hashcat** | High | credential, hash | GPU-accelerated password cracker |
| **john** | High | credential, hash | John the Ripper |
| **ophcrack** | Low | credential | Windows password cracker |
| **rainbowcrack** | Low | credential | Rainbow table cracker |
| **hash-identifier** | Low | hash | Hash type identification |
| **haiti** | Medium | hash | Hash type identifier |
| **name-that-hash** | Medium | hash | Modern hash identifier |

#### Credential Extraction

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **mimikatz** | High | credential, token | Windows credential dumper |
| **secretsdump** | High | credential, hash | Impacket secret dumping |
| **pypykatz** | High | credential | Mimikatz in Python |
| **LaZagne** | Medium | credential | Credentials recovery |
| **credentialfileview** | Low | credential | Windows credential viewer |
| **dploot** | Medium | credential | DPAPI looting |
| **dpat** | Medium | credential | Domain Password Audit Tool |

#### Network Credential Capture

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **responder** | High | credential, hash | LLMNR/NBT-NS poisoner |
| **ntlmrelayx** | High | credential, session | NTLM relay attacks |
| **mitm6** | Medium | credential | IPv6 MITM attack |
| **bettercap** | Medium | credential, packet | Network attack framework |
| **ettercap** | Low | credential, packet | MITM attacks |
| **wireshark** | Medium | packet, credential | Network protocol analyzer |
| **tcpdump** | Low | packet | Command-line packet analyzer |

---

### Privilege Escalation (TA0004) - Elevating Access

Tools for escalating privileges on compromised systems.

#### Linux Privilege Escalation

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **linpeas** | High | misconfiguration, credential, suid | Linux privilege escalation scanner |
| **linuxprivchecker** | Medium | misconfiguration | Linux privesc checker |
| **linux-exploit-suggester** | Medium | vulnerability, kernel | Kernel exploit suggester |
| **linux-smart-enumeration** | Medium | misconfiguration | Linux enumeration |
| **pspy** | High | process, cron | Process snooping |
| **gtfobins-cli** | Medium | binary, privesc | GTFOBins lookup |

#### Windows Privilege Escalation

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **winpeas** | High | misconfiguration, credential, service | Windows privilege escalation scanner |
| **powerup** | Medium | misconfiguration | PowerShell privesc checker |
| **seatbelt** | High | misconfiguration, credential | Windows security audit |
| **sharpup** | Medium | misconfiguration | SharpUp for privesc |
| **watson** | Medium | vulnerability | Windows exploit suggester |
| **beroot** | Medium | misconfiguration | Common misconfiguration check |
| **privesccheck** | Medium | misconfiguration | Windows privilege escalation check |
| **accesschk** | Medium | permission | Windows access checker |

#### Container & Cloud Escalation

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **deepce** | High | container, misconfiguration | Docker enumeration |
| **botb** | Medium | container, escape | Container breakout |
| **CDK** | Medium | container, misconfiguration | Container penetration toolkit |
| **kubeletctl** | Medium | kubernetes, misconfiguration | Kubelet exploitation |
| **peirates** | Medium | kubernetes, credential | Kubernetes pentesting |

---

### Lateral Movement (TA0008) - Moving Through Networks

Tools for moving between systems.

#### Remote Execution

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **impacket** | High | session, credential | Python network protocols |
| **psexec** | High | session | Remote command execution |
| **wmiexec** | High | session | WMI execution |
| **smbexec** | High | session | SMB execution |
| **atexec** | Medium | session | AT scheduled task execution |
| **dcomexec** | Medium | session | DCOM execution |
| **evil-winrm** | High | session | WinRM shell |
| **crackmapexec** | High | session, credential | Network penetration |
| **netexec** | High | session, credential | CrackMapExec successor |

#### Tunneling & Pivoting

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **chisel** | High | tunnel | Fast TCP/UDP tunnel |
| **ligolo-ng** | High | tunnel | Tunneling with TUN interface |
| **proxychains** | High | tunnel, proxy | Proxy chaining |
| **socat** | Medium | tunnel | Multipurpose relay |
| **sshuttle** | Medium | tunnel | SSH-based VPN |
| **plink** | Medium | tunnel | PuTTY command-line |
| **rpivot** | Medium | tunnel | Reverse SOCKS proxy |
| **revsocks** | Medium | tunnel | Reverse SOCKS5 tunnel |

#### Pass-the-Hash/Ticket

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **impacket-ticketer** | High | ticket, credential | Ticket forging |
| **getTGT** | High | ticket | TGT retrieval |
| **getST** | High | ticket | Service ticket retrieval |
| **rubeus** | High | ticket, credential | Kerberos abuse |
| **kekeo** | Medium | ticket, credential | Kerberos toolkit |
| **ticketconverter** | Medium | ticket | Ticket format conversion |

---

### Post-Exploitation & C2 (TA0011)

Command and control frameworks.

#### C2 Frameworks

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **sliver** | High | implant, session | Modern C2 framework |
| **covenant** | Medium | implant, session | .NET C2 framework |
| **cobalt-strike** | Medium | beacon, session | Commercial C2 |
| **havoc** | Medium | implant, session | Modern C2 |
| **mythic** | Medium | implant, session | Collaborative C2 |
| **villain** | Low | session | C2 framework |
| **poshc2** | Medium | implant, session | PowerShell C2 |

#### Post-Exploitation

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **powershell-empire** | Medium | session, module | PowerShell post-exploitation |
| **silenttrinity** | Medium | session | Post-exploitation agent |
| **pupy** | Medium | session | Cross-platform RAT |
| **merlin** | Medium | session | Cross-platform C2 |

---

### Specialized Tools

#### Mobile Security

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **mobsf** | High | apk, finding, vulnerability | Mobile security framework |
| **frida** | High | hook, function | Dynamic instrumentation |
| **objection** | High | hook, bypass | Runtime mobile exploration |
| **jadx** | Medium | apk, source | Android decompiler |
| **apktool** | Medium | apk | Android reverse engineering |
| **drozer** | Medium | android, vulnerability | Android security audit |
| **needle** | Medium | ios, vulnerability | iOS security testing |

#### Wireless Security

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **aircrack-ng** | Medium | wireless, credential | WiFi security suite |
| **wifite** | Medium | wireless, credential | Automated WiFi attacks |
| **bettercap** | Medium | wireless, packet | Network attacks |
| **kismet** | Medium | wireless, device | Wireless detector |
| **fluxion** | Low | wireless, credential | WiFi phishing |

#### Binary Analysis

| Tool | Priority | Node Types | Description |
|------|----------|------------|-------------|
| **ghidra** | Medium | binary, function | NSA reverse engineering |
| **ida** | Medium | binary, function | Interactive disassembler |
| **radare2** | Medium | binary, function | Reverse engineering framework |
| **binary-ninja** | Low | binary, function | Binary analysis platform |
| **angr** | Medium | binary, symbolic | Binary analysis framework |
| **cutter** | Medium | binary, function | Radare2 GUI |

---

### Tool Development Guidelines

When implementing new tools, follow these guidelines:

1. **Use Standard Node Types**: Reuse existing node types (host, port, domain, etc.) for consistency
2. **Meaningful ID Templates**: IDs should be deterministic and unique across runs
3. **Capture Relationships**: Connect findings to hosts, ports, domains appropriately
4. **Include Metadata**: Add timestamps, confidence scores, sources where applicable
5. **Test Taxonomy Extraction**: Verify the knowledge graph populates correctly

### Priority Definitions

| Priority | Criteria |
|----------|----------|
| **High** | Core bug bounty workflow, frequently used, unique capability |
| **Medium** | Specialized use case, good to have, complements high-priority tools |
| **Low** | Niche use case, redundant with other tools, legacy |
| **✅ Done** | Already implemented with embedded taxonomy |

### Contributing

To add a new tool with GraphRAG taxonomy:

1. Create tool directory: `mkdir -p {phase}/{tool-name}`
2. Initialize Go module: `go mod init github.com/zero-day-ai/gibson-oss-tools/{phase}/{tool-name}`
3. Implement `tool.go` with the Tool interface
4. Implement `schema.go` with TaxonomyMapping in OutputSchema()
5. Implement `main.go` with --schema support
6. Add to `go.work` and `Makefile`
7. Test with `./bin/{tool} --schema` to verify taxonomy embedding
8. Submit PR with example output showing knowledge graph extraction
