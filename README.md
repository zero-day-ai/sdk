# Gibson SDK

The official Go SDK for building Gibson agents, tools, and plugins.

## Installation

```bash
go get github.com/zero-day-ai/sdk@latest
```

## Overview

The SDK provides everything needed to build components for the Gibson framework:

- **Agent Development** - Build autonomous, LLM-powered security agents
- **Tool Development** - Create proto-based tool wrappers
- **Plugin Development** - Build stateful service integrations
- **Memory APIs** - Access the three-tier memory system
- **gRPC Serving** - Distribute components over the network

## Package Structure

```
sdk/
├── agent/       # Agent interfaces and types
├── tool/        # Tool interfaces and utilities
├── plugin/      # Plugin interfaces
├── llm/         # LLM abstractions and message types
├── memory/      # Memory access APIs
├── finding/     # Finding submission types
├── mission/     # Mission context types
├── result/      # Execution result types
├── health/      # Health check utilities
├── exec/        # Tool execution helpers
├── input/       # Input parsing utilities
├── toolerr/     # Tool error handling
├── schema/      # JSON Schema support
├── serve/       # gRPC serving utilities
├── eval/        # Evaluation framework
├── registry/    # Component registry access
├── graphrag/    # GraphRAG integration
│   ├── domain/      # Generated domain types (Host, Port, Finding, etc.)
│   ├── validation/  # CEL-based validators
│   └── id/          # Node ID generation
├── taxonomy/    # YAML-driven taxonomy (single source of truth)
└── examples/    # Reference implementations
    ├── minimal-agent/
    ├── custom-tool/
    └── custom-plugin/
```

## Core Concepts

### Agent Harness

The `AgentHarness` interface is how agents interact with Gibson:

```go
type AgentHarness interface {
    // LLM Access
    Complete(ctx context.Context, slot string, messages []llm.Message, opts ...CompletionOption) (*llm.CompletionResponse, error)
    StreamComplete(ctx context.Context, slot string, messages []llm.Message, opts ...StreamOption) (chan *llm.StreamChunk, error)
    CompleteStructured(ctx context.Context, slot string, schema *types.ResponseFormat, messages []llm.Message, opts ...CompletionOption) (*StructuredCompletionResult, error)

    // Tool Execution
    ExecuteTool(ctx context.Context, toolName string, input proto.Message) (proto.Message, error)
    CallToolProto(ctx context.Context, toolName string, input, output proto.Message) error

    // Plugin Queries
    QueryPlugin(ctx context.Context, plugin, method string, params map[string]any) (any, error)

    // Agent Delegation
    DelegateToAgent(ctx context.Context, agentName string, task agent.Task) (agent.Result, error)
    ListAgents(ctx context.Context) ([]agent.AgentDescriptor, error)

    // Memory Access
    Memory() memory.MemoryManager

    // Finding Submission
    SubmitFinding(ctx context.Context, finding agent.Finding) error

    // Observability
    Logger() *slog.Logger
    Tracer() trace.Tracer

    // Context
    MissionContext() MissionContext
    TargetInfo() TargetInfo
}
```

### LLM Slots

Agents declare abstract LLM requirements that are resolved at runtime:

```go
type SlotDefinition struct {
    Name        string
    Description string
    Required    bool
    Default     SlotConfig
    Constraints SlotConstraints
}

type SlotConstraints struct {
    MinContextWindow int
    RequiredFeatures []string  // "tool_use", "vision", "streaming", "json_mode"
}

// Helper constructor
slot := agent.NewSlotDefinition("primary", "Main LLM", true).
    WithConstraints(agent.SlotConstraints{
        MinContextWindow: 8000,
        RequiredFeatures: []string{agent.FeatureToolUse},
    })
```

### Three-Tier Memory

```go
type MemoryManager interface {
    Working() WorkingMemory    // Ephemeral, task-scoped
    Mission() MissionMemory     // Persistent, mission-scoped
    LongTerm() LongTermMemory  // Vector storage, cross-mission
}

// Working Memory - in-memory key-value
harness.Memory().Working().Set(ctx, "key", value)
harness.Memory().Working().Get(ctx, "key")

// Mission Memory - SQLite with FTS5 search
harness.Memory().Mission().Set(ctx, "key", value)
harness.Memory().Mission().Search(ctx, "query", opts)

// Long-Term Memory - vector embeddings
harness.Memory().LongTerm().Store(ctx, "text data", metadata)
harness.Memory().LongTerm().Search(ctx, "semantic query", threshold, limit)
```

## Building an Agent

### 1. Implement the Agent Interface

```go
package myagent

import (
    "context"

    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/component"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/types"
)

type MyAgent struct {
    config Config
}

// Identity
func (a *MyAgent) Name() string        { return "myagent" }
func (a *MyAgent) Version() string     { return "1.0.0" }
func (a *MyAgent) Description() string { return "My custom agent" }

// Capabilities
func (a *MyAgent) Capabilities() []string {
    return []string{"scanning", "enumeration"}
}

func (a *MyAgent) TargetTypes() []component.TargetType {
    return []component.TargetType{component.TargetTypeWeb}
}

func (a *MyAgent) TechniqueTypes() []component.TechniqueType {
    return []component.TechniqueType{component.TechniqueReconnaissance}
}

// LLM Requirements
func (a *MyAgent) LLMSlots() []agent.SlotDefinition {
    return []agent.SlotDefinition{
        agent.NewSlotDefinition("primary", "Main reasoning LLM", true).
            WithConstraints(agent.SlotConstraints{
                MinContextWindow: 8000,
                RequiredFeatures: []string{agent.FeatureJSONMode},
            }),
    }
}

// Lifecycle
func (a *MyAgent) Initialize(ctx context.Context, cfg agent.AgentConfig) error {
    // Parse configuration
    return nil
}

func (a *MyAgent) Shutdown(ctx context.Context) error {
    return nil
}

func (a *MyAgent) Health(ctx context.Context) types.HealthStatus {
    return types.HealthStatus{Status: types.HealthStatusHealthy}
}

// Core Execution
func (a *MyAgent) Execute(ctx context.Context, task agent.Task, h agent.Harness) (agent.Result, error) {
    result := agent.NewResult(task.ID)
    result.Start()

    // Use LLM
    messages := []llm.Message{
        llm.NewSystemMessage("You are a security analyst"),
        llm.NewUserMessage(task.Goal),
    }
    resp, err := h.Complete(ctx, "primary", messages)
    if err != nil {
        result.Fail(err)
        return result, err
    }

    // Execute tools
    toolOutput, err := h.ExecuteTool(ctx, "nmap", &pb.NmapRequest{...})

    // Submit findings
    finding := agent.Finding{
        Title:      "Vulnerability Found",
        Severity:   agent.SeverityHigh,
        Confidence: 0.9,
    }
    h.SubmitFinding(ctx, finding)

    result.Complete(map[string]any{"analysis": resp.Content})
    return result, nil
}
```

### 2. Create Main Entry Point

```go
package main

import (
    "github.com/myorg/myagent"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    agent := &myagent.MyAgent{}
    serve.Agent(agent, serve.WithPort(50051))
}
```

## Building a Tool

Tools are stateless operations that wrap security utilities. They use Protocol Buffers for type-safe I/O and can automatically populate the GraphRAG knowledge graph.

For complete tool development documentation, see **[docs/TOOLS.md](docs/TOOLS.md)**.

### 1. Define Proto Messages

```protobuf
syntax = "proto3";

package gibson.tools;

import "graphrag.proto";  // For automatic graph storage

message MyToolRequest {
    string target = 1;
    repeated string options = 2;
}

message MyToolResponse {
    bool success = 1;
    string output = 2;

    // Automatic graph storage (field 100 reserved)
    gibson.graphrag.DiscoveryResult discovery = 100;
}
```

### 2. Implement Tool Interface

```go
package mytool

import (
    "context"

    pb "github.com/myorg/mytool/proto"
    "github.com/zero-day-ai/sdk/types"
    "google.golang.org/protobuf/proto"
)

type MyTool struct{}

func (t *MyTool) Name() string        { return "mytool" }
func (t *MyTool) Version() string     { return "1.0.0" }
func (t *MyTool) Description() string { return "My security tool" }
func (t *MyTool) Tags() []string      { return []string{"scanning"} }

func (t *MyTool) InputMessageType() string  { return "gibson.tools.MyToolRequest" }
func (t *MyTool) OutputMessageType() string { return "gibson.tools.MyToolResponse" }

func (t *MyTool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
    req := input.(*pb.MyToolRequest)

    // Execute tool logic
    output, err := runTool(req.Target, req.Options)
    if err != nil {
        return &pb.MyToolResponse{Success: false}, err
    }

    // Populate discovery for automatic graph storage
    discovery := &graphragpb.DiscoveryResult{
        Hosts: []*graphragpb.Host{
            {Ip: "192.168.1.100", Hostname: "server01", State: "up"},
        },
    }

    return &pb.MyToolResponse{
        Success:   true,
        Output:    output,
        Discovery: discovery,  // Gibson automatically persists to Neo4j
    }, nil
}

func (t *MyTool) Health(ctx context.Context) types.HealthStatus {
    return types.HealthStatus{Status: types.HealthStatusHealthy}
}
```

### 3. Serve the Tool

```go
func main() {
    tool := &mytool.MyTool{}
    serve.Tool(tool, serve.WithPort(50052))
}
```

## Building a Plugin

### 1. Implement Plugin Interface

```go
package myplugin

import (
    "context"

    "github.com/zero-day-ai/sdk/plugin"
    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/types"
)

type MyPlugin struct {
    client *APIClient
}

func (p *MyPlugin) Name() string    { return "myplugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }

func (p *MyPlugin) Initialize(ctx context.Context, cfg plugin.PluginConfig) error {
    apiKey := cfg.Settings["api_key"].(string)
    p.client = NewAPIClient(apiKey)
    return nil
}

func (p *MyPlugin) Shutdown(ctx context.Context) error {
    return p.client.Close()
}

func (p *MyPlugin) Methods() []plugin.MethodDescriptor {
    return []plugin.MethodDescriptor{
        {
            Name:        "search",
            Description: "Search for data",
            InputSchema: schema.JSON(`{"type":"object","properties":{"query":{"type":"string"}}}`),
        },
    }
}

func (p *MyPlugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
    switch method {
    case "search":
        return p.client.Search(ctx, params["query"].(string))
    default:
        return nil, fmt.Errorf("unknown method: %s", method)
    }
}

func (p *MyPlugin) Health(ctx context.Context) types.HealthStatus {
    return types.HealthStatus{Status: types.HealthStatusHealthy}
}
```

### 2. Serve the Plugin

```go
func main() {
    plugin := &myplugin.MyPlugin{}
    serve.Plugin(plugin, serve.WithPort(50053))
}
```

## LLM Message Types

```go
// Create messages
system := llm.NewSystemMessage("You are a security analyst")
user := llm.NewUserMessage("Analyze this target")
assistant := llm.NewAssistantMessage("I'll analyze the target...")

// Tool calls
toolCall := llm.ToolCall{
    ID:       "call_123",
    Name:     "nmap",
    Arguments: `{"target": "192.168.1.1"}`,
}
assistantWithTool := llm.NewAssistantMessageWithToolCalls([]llm.ToolCall{toolCall})

// Tool results
toolResult := llm.NewToolMessage("call_123", `{"hosts": [...]}`)

// Completion request
resp, err := harness.Complete(ctx, "primary", messages,
    agent.WithTemperature(0.7),
    agent.WithMaxTokens(4000),
)

// Structured output
schema := &types.ResponseFormat{
    Type: "json_schema",
    JSONSchema: &types.JSONSchema{
        Name: "analysis",
        Schema: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "findings": map[string]any{"type": "array"},
            },
        },
    },
}
result, err := harness.CompleteStructured(ctx, "primary", schema, messages)
parsed := result.StructuredData // Automatically parsed JSON
```

## Task and Result Types

### Task

```go
type Task struct {
    ID           types.ID
    Name         string
    Description  string
    Goal         string            // Primary objective
    Context      map[string]any    // Additional context
    Timeout      time.Duration
    MissionID    *types.ID
    TargetID     *types.ID
    Priority     int
    Tags         []string
}
```

### Result

```go
type Result struct {
    TaskID      types.ID
    Status      ResultStatus  // pending, running, completed, failed, cancelled
    Output      map[string]any
    Findings    []Finding
    Error       *ResultError
    Metrics     TaskMetrics
    StartedAt   time.Time
    CompletedAt time.Time
}

// Constructors
result := agent.NewResult(task.ID)
result.Start()
result.Complete(outputMap)
// or
result.Fail(err)
```

### Finding

```go
type Finding struct {
    ID          types.ID
    Title       string
    Description string
    Severity    FindingSeverity  // critical, high, medium, low, info
    Confidence  float64          // 0.0 - 1.0
    Category    string
    TargetID    *types.ID
    Evidence    []Evidence
    CVSS        *CVSSScore
    CWE         []string
    Metadata    map[string]any
}

// Severity constants
agent.SeverityCritical
agent.SeverityHigh
agent.SeverityMedium
agent.SeverityLow
agent.SeverityInfo
```

## GraphRAG Domain Types

The SDK provides type-safe domain types for storing security data in the knowledge graph. All types are generated from a YAML taxonomy and support fluent builders, validation, and automatic parent-child relationships.

### Core Domain Types

```go
import "github.com/zero-day-ai/sdk/graphrag/domain"

// Create a host
host := domain.NewHost().
    SetIp("192.168.1.1").
    SetHostname("server.local").
    SetOs("Linux").
    SetState("up")

// Create a port belonging to the host
port := domain.NewPort(443, "tcp").
    BelongsTo(host).
    SetState("open")

// Create a service running on the port
service := domain.NewService("https").
    BelongsTo(port).
    SetProduct("nginx").
    SetVersion("1.18.0")

// Create a finding
finding := domain.NewFinding("SQL Injection", "critical").
    SetDescription("SQL injection in login form").
    SetConfidence(0.95).
    SetCategory("injection")

// Create a domain
domain := domain.NewDomain("example.com")

// Create a subdomain
subdomain := domain.NewSubdomain("api").
    BelongsTo(domain)
```

### The GraphNode Interface

All domain types implement the `GraphNode` interface:

```go
type GraphNode interface {
    NodeType() string                     // e.g., "host", "port", "finding"
    Properties() map[string]any           // All properties
    IdentifyingProperties() map[string]any // For deduplication
    ParentRef() *NodeRef                  // Parent relationship (nil for root nodes)
    Validate() error                      // CEL-based validation
    ToProto() *taxonomypb.GraphNode       // Convert to proto
    ID() string                           // Get node ID
    SetID(string)                         // Set node ID
}
```

### Validation

Core types are validated automatically using CEL (Common Expression Language):

```go
// Host requires either IP or hostname
host := domain.NewHost() // No IP or hostname
err := host.Validate()   // Returns error: "host requires either ip or hostname"

// Port number must be 1-65535
port := domain.NewPort(0, "tcp")
err := port.Validate() // Returns error: "port number must be between 1 and 65535"

// Child nodes require parents
port := domain.NewPort(443, "tcp") // No parent
err := port.Validate()             // Returns error: "port requires a parent of type host"

// Valid node
host := domain.NewHost().SetIp("192.168.1.1")
port := domain.NewPort(443, "tcp").BelongsTo(host)
err := port.Validate() // Returns nil
```

### Parent-Child Relationships

Some node types require parents:

| Child Type | Parent Type | Relationship |
|------------|-------------|--------------|
| `port` | `host` | `HAS_PORT` |
| `service` | `port` | `RUNS_SERVICE` |
| `endpoint` | `service` | `HAS_ENDPOINT` |
| `subdomain` | `domain` | `HAS_SUBDOMAIN` |
| `evidence` | `finding` | `HAS_EVIDENCE` |

Use `BelongsTo()` to set the parent:

```go
host := domain.NewHost().SetIp("192.168.1.1")
port := domain.NewPort(443, "tcp").BelongsTo(host)
service := domain.NewService("https").BelongsTo(port)
```

### Type Constants

Use generated constants for type safety:

```go
import "github.com/zero-day-ai/sdk/graphrag"

// Node type constants
nodeType := graphrag.NodeTypeHost     // "host"
nodeType := graphrag.NodeTypePort     // "port"
nodeType := graphrag.NodeTypeFinding  // "finding"

// Relationship type constants
relType := graphrag.RelTypeHasPort      // "HAS_PORT"
relType := graphrag.RelTypeRunsService  // "RUNS_SERVICE"
relType := graphrag.RelTypeAffects      // "AFFECTS"
```

### Custom Types

Custom (non-core) types pass validation without rules:

```go
import "github.com/zero-day-ai/sdk/graphrag/validation"

// Custom types are always valid
isCore := validation.IsCoreType("my_custom_type") // false
err := validation.ValidateNode("my_custom_type", props, false) // nil
```

### Creating Extension Taxonomies

Agents can define custom types via extension YAML:

```yaml
# my-agent/taxonomy/extension.yaml
version: "1.0.0"
kind: extension
extends: "3.0.0"  # Core taxonomy version

node_types:
  - name: my_custom_node
    category: custom
    description: My agent's custom node type
    properties:
      - name: custom_field
        type: string
        required: true

relationship_types:
  - name: MY_CUSTOM_REL
    description: Custom relationship
    from_types: [my_custom_node]
    to_types: [finding]
```

Then generate types:

```bash
taxonomy-gen \
  --base sdk/taxonomy/core.yaml \
  --extension my-agent/taxonomy/extension.yaml \
  --output-domain my-agent/domain/domain_generated.go
```

## gRPC Serving Options

```go
serve.Agent(agent,
    serve.WithPort(50051),
    serve.WithTLS(certFile, keyFile),
    serve.WithLogger(logger),
    serve.WithHealthCheck(true),
)

serve.Tool(tool,
    serve.WithPort(50052),
)

serve.Plugin(plugin,
    serve.WithPort(50053),
)
```

## Health Checks

```go
type HealthStatus struct {
    Status    HealthStatusType  // healthy, degraded, unhealthy
    Message   string
    Details   map[string]any
    CheckedAt time.Time
}

const (
    HealthStatusHealthy   = "healthy"
    HealthStatusDegraded  = "degraded"
    HealthStatusUnhealthy = "unhealthy"
)

// Check binary dependency
func (t *MyTool) Health(ctx context.Context) types.HealthStatus {
    if _, err := exec.LookPath("mytool"); err != nil {
        return types.HealthStatus{
            Status:  types.HealthStatusUnhealthy,
            Message: "mytool binary not found",
        }
    }
    return types.HealthStatus{Status: types.HealthStatusHealthy}
}
```

## Examples

### examples/minimal-agent/

A minimal agent implementation demonstrating core concepts.

### examples/custom-tool/

A custom tool wrapper with proto definitions.

### examples/custom-plugin/

A plugin integrating with an external API.

## Dependencies

```go
require (
    google.golang.org/grpc v1.78.0
    google.golang.org/protobuf v1.36.5
    go.etcd.io/etcd/client/v3 v3.5.18
    go.opentelemetry.io/otel v1.34.0
)
```

## Documentation

### Core Guides

| Guide | Description |
|-------|-------------|
| [Tool Development](docs/TOOLS.md) | Complete guide to building security tools with automatic graph storage |
| [Agent Development](docs/AGENTS.md) | Building autonomous LLM-powered security agents |
| [Plugin Development](docs/PLUGINS.md) | Creating stateful service integrations |

### API Reference

| Reference | Description |
|-----------|-------------|
| [DiscoveryResult Proto](api/proto/DISCOVERY_RESULT.md) | Automatic graph storage proto message reference |
| [Proto Definitions](api/proto/) | Complete protocol buffer schemas |

### Examples

| Example | Description |
|---------|-------------|
| [examples/minimal-agent/](examples/minimal-agent/) | Minimal agent implementation |
| [examples/custom-tool/](examples/custom-tool/) | Custom tool with discovery support |
| [examples/custom-plugin/](examples/custom-plugin/) | Plugin integrating external API |

## Related Repositories

| Repository | Description |
|------------|-------------|
| [gibson](https://github.com/zero-day-ai/gibson) | Core framework |
| [tools](https://github.com/zero-day-ai/tools) | Security tool wrappers |
| [network-recon](https://github.com/zero-day-ai/network-recon) | Network recon agent |
| [tech-stack-fingerprinting](https://github.com/zero-day-ai/tech-stack-fingerprinting) | Tech detection agent |

## License

Proprietary - Zero-Day.AI
