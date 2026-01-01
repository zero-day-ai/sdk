# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

Specs should ALWAYS go in ~/Code/zero-day.ai/.spec-workflow


## Overview

The Gibson SDK is the official Go SDK for building AI security testing agents, tools, and plugins for the Gibson Framework. It provides APIs for creating autonomous security testing agents that discover vulnerabilities in LLMs and AI systems.

**Repository:** `github.com/zero-day-ai/sdk`
**Go Version:** 1.24.4+

## Build & Test Commands

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run example tests only
go test -v -run Example ./...

# Run benchmarks
go test -bench=. ./...

# Download/tidy dependencies
go mod download
go mod tidy
```

## Architecture

### Three-Tier Component Model

```
┌─────────────────────────────────────────────────────────────────┐
│                      GIBSON FRAMEWORK                            │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   MISSION ORCHESTRATOR                   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐              │
│         ▼                    ▼                    ▼              │
│   ┌───────────┐        ┌───────────┐        ┌───────────┐       │
│   │   AGENT   │        │   TOOL    │        │  PLUGIN   │       │
│   │           │ ──────▶│           │        │           │       │
│   │ Execute() │ ──────▶│ Execute() │◀───────│  Query()  │       │
│   └───────────┘        └───────────┘        └───────────┘       │
│         │                                                        │
│         ▼                                                        │
│   ┌─────────────────────────────────────────────────────────┐   │
│   │                    AGENT HARNESS                         │   │
│   │  LLM | Tools | Plugins | Memory | Findings | Tracing    │   │
│   └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Core Packages

| Package | Purpose |
|---------|---------|
| `agent/` | Agent interface, builder, harness, execution types |
| `tool/` | Tool interface, builder, execution types |
| `plugin/` | Plugin interface, builder, method handlers |
| `llm/` | LLM types, slots, completions, streaming, token tracking |
| `memory/` | Three-tier memory system (working, mission, long-term) |
| `finding/` | Security finding types, severity, categories, MITRE mappings |
| `schema/` | JSON Schema validation |
| `serve/` | gRPC server infrastructure |
| `exec/` | Context-aware shell command execution |
| `input/` | Type-safe map value extraction |
| `health/` | Health status types |
| `toolerr/` | Structured error types |
| `graphrag/` | Graph-based retrieval-augmented generation |
| `api/proto/` | Protocol buffer definitions |
| `api/gen/` | Generated proto code |

### Key Interfaces

**Agent**: Autonomous security testing component with capabilities, LLM slots, and execute logic
**Tool**: Stateless executable with schema-validated input/output
**Plugin**: Stateful service with multiple queryable methods
**Harness**: Runtime interface providing agents access to LLMs, tools, plugins, memory, and findings

### LLM Slot System

Agents declare named LLM requirements (slots) instead of hardcoding models. The framework resolves slots to actual providers/models at runtime:

```go
sdk.WithLLMSlot("primary", llm.SlotRequirements{
    MinContextWindow: 100000,
    RequiredFeatures: []string{"tool_use", "vision"},
})
sdk.WithLLMSlot("fast", llm.SlotRequirements{
    MinContextWindow: 8192,
})
```

### Three-Tier Memory

- **Working Memory**: Ephemeral in-memory key-value store (cleared after task)
- **Mission Memory**: SQLite-backed persistent storage with FTS5 search (mission-scoped)
- **Long-Term Memory**: Vector database for semantic storage across missions

## Component Patterns

### Creating an Agent

```go
cfg := agent.NewConfig().
    SetName("my-agent").
    SetVersion("1.0.0").
    SetDescription("Agent description").
    AddCapability(agent.CapabilityPromptInjection).
    AddTargetType(types.TargetTypeLLMChat).
    AddLLMSlot("primary", llm.SlotRequirements{MinContextWindow: 8000}).
    SetExecuteFunc(func(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
        // Agent logic
        return agent.NewSuccessResult(output), nil
    })

myAgent, err := agent.New(cfg)
```

### Creating a Tool

```go
cfg := tool.NewConfig().
    SetName("my-tool").
    SetVersion("1.0.0").
    SetInputSchema(schema.Object(map[string]schema.JSON{...}, required...)).
    SetOutputSchema(schema.Object(map[string]schema.JSON{...}, required...)).
    SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
        // Tool logic
        return output, nil
    })

myTool, err := tool.New(cfg)
```

### Creating a Plugin

```go
cfg := plugin.NewConfig().
    SetName("my-plugin").
    SetVersion("1.0.0")

cfg.AddMethodWithDesc("methodName", "description",
    func(ctx context.Context, params map[string]any) (any, error) {
        // Method logic
        return result, nil
    },
    inputSchema, outputSchema,
)

myPlugin, err := plugin.New(cfg)
```

### Serving Components via gRPC

```go
serve.Agent(myAgent, serve.WithPort(50051))
serve.Tool(myTool, serve.WithPort(50052))
serve.Plugin(myPlugin, serve.WithPort(50053))
```

## Component Manifests

Tools and agents require a `component.yaml` manifest for installation:

```yaml
kind: tool  # or: agent, plugin
name: my-component
version: 1.0.0
description: Component description

build:
  command: go build -o my-component .
  artifacts:
    - my-component

runtime:
  type: go
  entrypoint: ./my-component
  port: 0

dependencies:
  gibson: ">=1.0.0"
  system:
    - external-binary
```

## Dependencies

- `google.golang.org/grpc` - gRPC service communication
- `google.golang.org/protobuf` - Protocol Buffers serialization
- `go.opentelemetry.io/otel` - Distributed tracing
- `github.com/google/uuid` - UUID generation
- `github.com/stretchr/testify` - Testing assertions

## Security Policy

**No binaries in repository**: Only source code is permitted. All binaries must be built locally from source for supply chain security.
