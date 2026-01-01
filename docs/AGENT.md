# Gibson SDK Agent Development Guide

This guide explains how to build agents, tools, and plugins for the Gibson Framework, with a focus on the LLM slot system that enables agents to use multiple LLMs from different vendors.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [LLM Slot System](#llm-slot-system)
3. [Configuration](#configuration)
4. [Building Agents](#building-agents)
5. [Component Manifest (component.yaml)](#component-manifest-componentyaml)
6. [Agent Installation](#agent-installation)
7. [Health Checks](#health-checks)
8. [Checking Agent Status](#checking-agent-status)
9. [Building Tools](#building-tools)
10. [Building Plugins](#building-plugins)
11. [Multi-Vendor LLM Usage](#multi-vendor-llm-usage)
12. [Complete Examples](#complete-examples)
13. [Agent Streaming and TUI Integration](#agent-streaming-and-tui-integration)
14. [Streaming Agents](#streaming-agents)

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Gibson Framework                             │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────────────┐   │
│  │   Agent     │───▶│   Harness    │───▶│    Slot Manager      │   │
│  │             │    │              │    │                      │   │
│  │ LLMSlots()  │    │ Complete()   │    │  ResolveSlot()       │   │
│  │ Execute()   │    │ CallTool()   │    │  ValidateConstraints │   │
│  └─────────────┘    └──────────────┘    └──────────┬───────────┘   │
│                                                     │               │
│                                    ┌────────────────┴───────────┐   │
│                                    │       LLM Registry         │   │
│                                    ├─────────────────────────────┤   │
│                                    │ ┌─────────┐ ┌─────────┐    │   │
│                                    │ │Anthropic│ │ OpenAI  │    │   │
│                                    │ └─────────┘ └─────────┘    │   │
│                                    │ ┌─────────┐ ┌─────────┐    │   │
│                                    │ │ Google  │ │ Ollama  │    │   │
│                                    │ └─────────┘ └─────────┘    │   │
│                                    └────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

**Key Concepts:**

- **Slots**: Named LLM requirements declared by agents (e.g., "primary", "fast", "vision")
- **Providers**: LLM vendors (Anthropic, OpenAI, Google, Ollama)
- **Harness**: Runtime interface providing controlled access to LLMs, tools, and memory
- **Slot Manager**: Resolves slot requirements to specific provider/model combinations

---

## LLM Slot System

### What Are Slots?

Slots are **named LLM requirements** that agents declare. Instead of hardcoding "use GPT-4" or "use Claude", agents declare what capabilities they need, and the framework resolves these to actual models at runtime.

```
┌──────────────────────────────────────────────────────────────────┐
│                    Agent Slot Declaration                         │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Slot: "primary"                 Slot: "fast"                     │
│  ├─ MinContextWindow: 100000     ├─ MinContextWindow: 8192        │
│  ├─ RequiredFeatures:            ├─ RequiredFeatures: []          │
│  │   - tool_use                  └─ PreferredModels:              │
│  │   - vision                        - gpt-3.5-turbo              │
│  └─ PreferredModels:                                              │
│      - claude-3-opus                                              │
│                                                                   │
├───────────────────────┬──────────────────────────────────────────┤
│                       │                                           │
│                       ▼                                           │
│              ┌─────────────────┐                                  │
│              │  Slot Manager   │                                  │
│              │                 │                                  │
│              │  ResolveSlot()  │                                  │
│              └────────┬────────┘                                  │
│                       │                                           │
│         ┌─────────────┴─────────────┐                            │
│         ▼                           ▼                             │
│  ┌─────────────────┐       ┌─────────────────┐                   │
│  │   Anthropic     │       │    OpenAI       │                   │
│  │ claude-3-opus   │       │ gpt-3.5-turbo   │                   │
│  └─────────────────┘       └─────────────────┘                   │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

### Slot Definition

```go
import "github.com/zero-day-ai/sdk/llm"

type SlotDefinition struct {
    Name             string   // Unique identifier: "primary", "fast", "vision"
    Description      string   // What this slot is used for
    Required         bool     // Must this slot be available?
    MinContextWindow int      // Minimum tokens the model must support
    RequiredFeatures []string // Capabilities: "tool_use", "vision", "streaming"
    PreferredModels  []string // Hints for preferred models (not strict)
}
```

### Available Features

| Feature | Description | Providers |
|---------|-------------|-----------|
| `tool_use` | Function/tool calling | Anthropic, OpenAI, Google |
| `vision` | Image understanding | Anthropic, OpenAI, Google |
| `streaming` | Token-by-token response | All providers |
| `json_mode` | Structured JSON output | OpenAI, Anthropic |

### Slot Resolution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     Slot Resolution Process                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Agent calls: harness.Complete(ctx, "primary", messages)      │
│                         │                                        │
│                         ▼                                        │
│  2. SlotManager.ResolveSlot("primary")                          │
│         │                                                        │
│         ├─► Check agent's default config for "primary"           │
│         │   Provider: "anthropic"                                │
│         │   Model: "claude-3-opus-20240229"                      │
│         │                                                        │
│         ├─► Check mission-level override (if any)                │
│         │                                                        │
│         ├─► Merge configs (override wins)                        │
│         │                                                        │
│         ├─► Lookup provider in LLMRegistry                       │
│         │                                                        │
│         ├─► Verify model exists in provider                      │
│         │                                                        │
│         ├─► Validate constraints:                                │
│         │   ✓ Context window >= MinContextWindow                 │
│         │   ✓ Model has RequiredFeatures                         │
│         │                                                        │
│         └─► Return (provider, modelInfo)                         │
│                         │                                        │
│                         ▼                                        │
│  3. provider.Complete(ctx, request)                              │
│                         │                                        │
│                         ▼                                        │
│  4. Return response + track token usage by slot                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Configuration

### Config File Location

Gibson reads configuration from `~/.gibson/config.yaml`. This file defines:
- LLM provider credentials
- Model configurations
- Default providers
- Agent slot overrides

### Configuration Structure

```yaml
# ~/.gibson/config.yaml

# Core settings
core:
  log_level: info
  data_dir: ~/.gibson

# Database settings
database:
  path: ~/.gibson/gibson.db
  wal_mode: true

# LLM Provider Configuration
llm:
  # Default provider when not specified
  default_provider: anthropic

  # Provider definitions
  providers:
    # Anthropic (Claude models)
    anthropic:
      type: anthropic
      api_key: ${ANTHROPIC_API_KEY}    # Environment variable interpolation
      base_url: https://api.anthropic.com
      default_model: claude-3-5-sonnet-20241022
      models:
        claude-3-5-sonnet-20241022:
          context_window: 200000
          max_output: 8192
          features: [chat, streaming, tools, vision]
          pricing_input: 0.003
          pricing_output: 0.015
        claude-3-opus-20240229:
          context_window: 200000
          max_output: 4096
          features: [chat, streaming, tools, vision]
          pricing_input: 0.015
          pricing_output: 0.075
        claude-3-haiku-20240307:
          context_window: 200000
          max_output: 4096
          features: [chat, streaming, tools, vision]
          pricing_input: 0.00025
          pricing_output: 0.00125

    # OpenAI (GPT models)
    openai:
      type: openai
      api_key: ${OPENAI_API_KEY}
      base_url: https://api.openai.com/v1
      default_model: gpt-4-turbo
      models:
        gpt-4-turbo:
          context_window: 128000
          max_output: 4096
          features: [chat, tools, vision, json_mode]
          pricing_input: 0.01
          pricing_output: 0.03
        gpt-4o:
          context_window: 128000
          max_output: 16384
          features: [chat, tools, vision, json_mode, streaming]
          pricing_input: 0.005
          pricing_output: 0.015
        gpt-3.5-turbo:
          context_window: 16385
          max_output: 4096
          features: [chat, tools, json_mode]
          pricing_input: 0.0005
          pricing_output: 0.0015

    # Google (Gemini models)
    google:
      type: google
      api_key: ${GOOGLE_API_KEY}
      default_model: gemini-1.5-pro
      models:
        gemini-1.5-pro:
          context_window: 1000000
          max_output: 8192
          features: [chat, tools, vision, streaming]
          pricing_input: 0.00125
          pricing_output: 0.005
        gemini-1.5-flash:
          context_window: 1000000
          max_output: 8192
          features: [chat, tools, vision, streaming]
          pricing_input: 0.000075
          pricing_output: 0.0003

    # Ollama (Local models)
    ollama:
      type: ollama
      base_url: http://localhost:11434
      default_model: llama3.1:70b
      models:
        llama3.1:70b:
          context_window: 128000
          max_output: 4096
          features: [chat, streaming]
        codellama:34b:
          context_window: 16384
          max_output: 4096
          features: [chat, streaming]

# Agent-specific slot overrides
agents:
  prompt-injection:
    slots:
      primary:
        provider: anthropic
        model: claude-3-5-sonnet-20241022
        temperature: 0.7
      fast:
        provider: openai
        model: gpt-3.5-turbo
        temperature: 0.3

  recon:
    slots:
      primary:
        provider: openai
        model: gpt-4-turbo
      vision:
        provider: google
        model: gemini-1.5-pro
```

### Environment Variable Interpolation

Config values can reference environment variables:

```yaml
api_key: ${ANTHROPIC_API_KEY}           # Required - fails if not set
api_key: ${ANTHROPIC_API_KEY:-default}  # With default value
base_url: ${CUSTOM_URL:-https://api.anthropic.com}
```

### How Config Gets Read

```
┌────────────────────────────────────────────────────────────────┐
│                   Configuration Loading Flow                    │
├────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. Framework Start                                             │
│         │                                                       │
│         ▼                                                       │
│  2. Load ~/.gibson/config.yaml                                  │
│         │                                                       │
│         ▼                                                       │
│  3. Interpolate environment variables                           │
│     ${ANTHROPIC_API_KEY} → "sk-ant-..."                        │
│         │                                                       │
│         ▼                                                       │
│  4. Validate schema                                             │
│     - Required fields present                                   │
│     - Types correct                                             │
│         │                                                       │
│         ▼                                                       │
│  5. Create ProviderConfig for each provider                     │
│         │                                                       │
│         ▼                                                       │
│  6. Initialize providers via factory                            │
│     NewAnthropicProvider(cfg)                                   │
│     NewOpenAIProvider(cfg)                                      │
│         │                                                       │
│         ▼                                                       │
│  7. Register providers in LLMRegistry                           │
│         │                                                       │
│         ▼                                                       │
│  8. Create SlotManager with registry                            │
│         │                                                       │
│         ▼                                                       │
│  9. Framework ready - agents can use slots                      │
│                                                                 │
└────────────────────────────────────────────────────────────────┘
```

---

## Building Agents

### Agent Interface

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/types"
)

type Agent interface {
    // Metadata
    Name() string
    Version() string
    Description() string

    // Capabilities
    Capabilities() []Capability
    TargetTypes() []types.TargetType
    TechniqueTypes() []types.TechniqueType

    // LLM Requirements - THE KEY METHOD
    LLMSlots() []llm.SlotDefinition

    // Execution
    Execute(ctx context.Context, harness Harness, task Task) (Result, error)

    // Lifecycle
    Initialize(ctx context.Context, config map[string]any) error
    Shutdown(ctx context.Context) error
    Health(ctx context.Context) types.HealthStatus
}
```

### Creating an Agent with the SDK

```go
package main

import (
    "context"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/types"
)

func main() {
    myAgent, err := sdk.NewAgent(
        // Basic metadata
        sdk.WithName("my-recon-agent"),
        sdk.WithVersion("1.0.0"),
        sdk.WithDescription("Performs reconnaissance on AI targets"),

        // What targets this agent can attack
        sdk.WithTargetTypes(
            types.TargetTypeLLMChat,
            types.TargetTypeLLMAPI,
            types.TargetTypeRAG,
        ),

        // What techniques this agent uses
        sdk.WithTechniqueTypes(
            types.TechniqueTypeReconnaissance,
            types.TechniqueTypeInformationGathering,
        ),

        // LLM Slot Requirements
        sdk.WithLLMSlot("primary", llm.SlotRequirements{
            MinContextWindow: 100000,
            RequiredFeatures: []string{"tool_use"},
        }),
        sdk.WithLLMSlot("fast", llm.SlotRequirements{
            MinContextWindow: 8192,
            RequiredFeatures: []string{},
        }),

        // Tools this agent needs
        sdk.WithTools("http-request", "dns-lookup", "screenshot"),

        // The execution function
        sdk.WithExecuteFunc(executeRecon),
    )
    if err != nil {
        panic(err)
    }

    // Serve as gRPC service
    sdk.ServeAgent(myAgent, sdk.WithPort(50051))
}

func executeRecon(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Use the primary slot for main reasoning
    resp, err := h.Complete(ctx, "primary", []llm.Message{
        {Role: "system", Content: "You are a security reconnaissance agent..."},
        {Role: "user", Content: task.Goal},
    })
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Use the fast slot for quick classifications
    classification, err := h.Complete(ctx, "fast", []llm.Message{
        {Role: "user", Content: "Classify: " + resp.Content},
    })

    return agent.Result{
        Status: agent.StatusSuccess,
        Output: resp.Content,
    }, nil
}
```

### Agent Harness Interface

The harness is the runtime interface providing controlled access to framework capabilities:

```go
type Harness interface {
    // LLM Access - Uses Slot System
    Complete(ctx context.Context, slot string, messages []llm.Message,
             opts ...CompletionOption) (*llm.CompletionResponse, error)
    CompleteWithTools(ctx context.Context, slot string, messages []llm.Message,
                      tools []llm.ToolDef) (*llm.CompletionResponse, error)
    Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error)

    // Tool Access
    CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error)
    ListTools(ctx context.Context) ([]tool.Descriptor, error)

    // Plugin Access
    QueryPlugin(ctx context.Context, name string, method string,
                params map[string]any) (any, error)
    ListPlugins(ctx context.Context) ([]plugin.Descriptor, error)

    // Agent Delegation
    DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error)
    ListAgents(ctx context.Context) ([]agent.Descriptor, error)

    // Findings
    SubmitFinding(ctx context.Context, finding finding.Finding) error
    GetFindings(ctx context.Context, filter finding.Filter) ([]finding.Finding, error)

    // Memory
    Memory() memory.Store

    // Context
    Mission() types.MissionContext
    Target() types.TargetInfo

    // Observability
    Tracer() trace.Tracer
    Logger() *slog.Logger
    Metrics() MetricsRecorder
    TokenUsage() llm.TokenTracker
}
```

---

## Component Manifest (component.yaml)

Every agent must have a `component.yaml` manifest file in its root directory. This file tells Gibson how to build, run, and validate the agent.

### Manifest File Requirements

- **File name**: Must be exactly `component.yaml` (or `component.json`)
- **Location**: Root directory of the agent
- **For mono-repos**: Each agent subdirectory must have its own `component.yaml`

### Agent Manifest Structure

```yaml
kind: agent                   # Component type (must be "agent")
name: security-scanner        # Unique agent name (lowercase, alphanumeric, hyphens)
version: 1.0.0               # Semantic version
description: Comprehensive security scanning agent with LLM-powered analysis
author: Gibson Security Team
license: MIT
repository: https://github.com/zero-day-ai/gibson-agents

# Agent-specific metadata
agent:
  capabilities:               # What this agent can do
    - prompt_injection
    - jailbreak
    - data_extraction
    - reconnaissance

  target_types:               # What targets this agent can attack
    - llm_chat
    - api
    - rag_system

  technique_types:            # Attack technique categories
    - reconnaissance
    - initial_access
    - credential_access

  mitre_attack:               # MITRE ATT&CK technique IDs
    - T1190                   # Exploit Public-Facing Application
    - T1552                   # Unsecured Credentials

# LLM slot requirements
llm_slots:
  primary:
    description: Main reasoning LLM for analysis and planning
    required: true
    constraints:
      min_context_window: 100000
      required_features:
        - tool_use
        - vision
    default:
      provider: anthropic
      model: claude-sonnet-4-20250514

  fast:
    description: Fast LLM for quick classification tasks
    required: false
    constraints:
      min_context_window: 8000
    default:
      provider: anthropic
      model: claude-3-5-haiku-20241022

# Tool dependencies (from gibson-tools-official)
tools:
  - name: http-scanner
    path: reconnaissance/http-scanner
    required: true
  - name: nuclei
    path: reconnaissance/nuclei
    required: false

# Build configuration
build:
  command: go build -o security-scanner ./cmd/scanner
  artifacts:
    - security-scanner
  workdir: .
  env:
    CGO_ENABLED: "0"

# Runtime configuration
runtime:
  type: go                    # Runtime type: go, python, docker, binary
  entrypoint: ./security-scanner
  port: 0                     # gRPC port (0 = dynamic assignment)
  args: []                    # Command-line arguments
  health_check:
    protocol: grpc            # Protocol: "http", "grpc", or "auto" (default: auto)
    interval: 30s             # Health check interval for monitoring
    timeout: 5s               # Timeout per health check request
    endpoint: /health         # HTTP endpoint (only used with http protocol)
    service_name: ""          # gRPC service name (empty = overall server health)

# Environment variables
env:
  SCANNER_LOG_LEVEL: info
  SCANNER_SAFE_MODE: "true"

# Dependencies
dependencies:
  gibson: ">=1.0.0"           # Minimum Gibson version
  components:                  # Other Gibson components required
    - http-scanner@1.0.0
  system: []                   # System binaries required
  env:                         # Required environment variables
    API_KEY: ""               # Empty = optional, non-empty = required
```

### Field Reference

#### Top-Level Fields

| Field | Required | Description |
|-------|----------|-------------|
| `kind` | Yes | Must be `agent` |
| `name` | Yes | Unique identifier (alphanumeric, dash, underscore) |
| `version` | Yes | Semantic version (e.g., `1.0.0`) |
| `description` | No | Brief description of the agent |
| `author` | No | Author name or organization |
| `license` | No | License identifier (MIT, Apache-2.0, etc.) |
| `repository` | No | Source repository URL |

#### Agent Metadata (`agent` section)

| Field | Required | Description |
|-------|----------|-------------|
| `agent.capabilities` | Yes | List of agent capabilities (prompt_injection, jailbreak, etc.) |
| `agent.target_types` | Yes | Target types the agent can attack (llm_chat, api, rag_system) |
| `agent.technique_types` | No | Attack technique categories |
| `agent.mitre_attack` | No | MITRE ATT&CK technique IDs |

#### LLM Slots (`llm_slots` section)

| Field | Required | Description |
|-------|----------|-------------|
| `llm_slots.<name>.description` | No | Description of what this slot is used for |
| `llm_slots.<name>.required` | No | Whether this slot must be available (default: true) |
| `llm_slots.<name>.constraints.min_context_window` | No | Minimum context window size |
| `llm_slots.<name>.constraints.required_features` | No | Required model features (tool_use, vision, etc.) |
| `llm_slots.<name>.default.provider` | No | Default LLM provider |
| `llm_slots.<name>.default.model` | No | Default model name |

#### Tool Dependencies (`tools` section)

| Field | Required | Description |
|-------|----------|-------------|
| `tools[].name` | Yes | Tool name |
| `tools[].path` | Yes | Path in gibson-tools-official repo |
| `tools[].required` | No | Whether this tool is required (default: true) |

#### Build Configuration (`build` section)

| Field | Required | Description |
|-------|----------|-------------|
| `build.command` | No | Build command (default: `make build`) |
| `build.artifacts` | No | Expected output files |
| `build.workdir` | No | Build working directory |
| `build.env` | No | Build environment variables |

#### Runtime Configuration (`runtime` section)

| Field | Required | Description |
|-------|----------|-------------|
| `runtime.type` | Yes | Runtime type: `go`, `python`, `docker`, `binary` |
| `runtime.entrypoint` | Yes | Executable path or command |
| `runtime.port` | No | gRPC port (0 = dynamic) |
| `runtime.args` | No | Command-line arguments |
| `runtime.health_check.protocol` | No | Health check protocol: `http`, `grpc`, or `auto` (default: `auto`) |
| `runtime.health_check.interval` | No | Health check interval for ongoing monitoring (default: `10s`) |
| `runtime.health_check.timeout` | No | Timeout per health check request (default: `5s`) |
| `runtime.health_check.endpoint` | No | HTTP health endpoint path (default: `/health`) |
| `runtime.health_check.service_name` | No | gRPC service name to check (default: empty = server health) |

---

## Agent Installation

### Installing from Repository

Use the Gibson CLI to install agents from a git repository:

```bash
# Install from dedicated repository
gibson agent install https://github.com/user/my-agent

# Install from mono-repo subdirectory (use # fragment)
gibson agent install https://github.com/user/agents#security/scanner

# Install using SSH URL with subdirectory
gibson agent install git@github.com:user/agents.git#path/to/agent

# Install with specific branch
gibson agent install https://github.com/user/my-agent --branch main

# Install with specific tag
gibson agent install https://github.com/user/my-agent --tag v1.0.0

# Force reinstall
gibson agent install https://github.com/user/my-agent --force

# Bulk install all agents from mono-repo
gibson agent install-all https://github.com/user/gibson-agents
```

### Installation Process

When you run `gibson agent install <repo-url>`:

1. **Clone Repository**: Clone to temporary directory
2. **Locate Manifest**: Look for `component.yaml` in root (or subdirectory if specified with `#`)
3. **Validate Manifest**: Parse and validate manifest structure
4. **Check Dependencies**: Verify Gibson version, tools, and system dependencies
5. **Build Component**: Execute build command
6. **Install**: Move to `~/.gibson/agents/<name>/`
7. **Register**: Add to component registry

### Installation Directory Structure

```
~/.gibson/
├── agents/
│   ├── security-scanner/
│   │   ├── component.yaml
│   │   ├── go.mod
│   │   ├── main.go
│   │   └── security-scanner    # Built binary
│   ├── k8skiller/
│   └── ...
├── tools/
├── plugins/
└── config.yaml
```

### Managing Installed Agents

```bash
# List installed agents
gibson agent list

# Get agent info
gibson agent info security-scanner

# Update an agent
gibson agent update security-scanner

# Update all agents
gibson agent update --all

# Uninstall an agent
gibson agent uninstall security-scanner

# Check agent health
gibson agent health security-scanner
```

### Common Installation Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `MANIFEST_NOT_FOUND` | No `component.yaml` in expected location | Ensure file is named exactly `component.yaml` |
| `INVALID_KIND` | Kind field missing or not `agent` | Add `kind: agent` to manifest |
| `BUILD_FAILED` | Build command failed | Check build command and dependencies |
| `DEPENDENCY_FAILED` | Missing tool or system dependency | Install required dependencies first |
| `SLOT_VALIDATION_FAILED` | LLM slot constraints not met | Check LLM provider configuration |

---

## Health Checks

The Gibson Framework uses health checks to determine when a component has successfully started and to monitor ongoing health during operation. Components can expose health checks via HTTP or gRPC protocols.

### How SDK Components Report Health

When you use `sdk.ServeAgent()`, `sdk.ServeTool()`, or `sdk.ServePlugin()`, the SDK automatically registers a gRPC health service using the standard `grpc_health_v1` protocol. This is the recommended approach as it provides:

- Automatic health reporting without additional code
- Standard protocol supported by Kubernetes, load balancers, and monitoring tools
- Service-specific health status (per-service granularity)

```go
// The SDK automatically implements grpc_health_v1.HealthServer
// No additional code needed - health checks work out of the box
sdk.ServeAgent(myAgent, sdk.WithPort(50051))
```

The gRPC health service responds to `grpc.health.v1.Health/Check` requests with status:
- `SERVING` - Component is healthy and ready
- `NOT_SERVING` - Component is unhealthy
- `SERVICE_UNKNOWN` - Requested service name not found

### How Gibson CLI Performs Health Checks

Gibson CLI supports three health check protocols configured via the component manifest:

1. **`grpc`** - Uses the standard gRPC health protocol (`grpc_health_v1`)
2. **`http`** - Makes HTTP GET requests to a health endpoint
3. **`auto`** (default) - Tries gRPC first, falls back to HTTP if gRPC connection fails

```
                        Protocol Detection Flow

    ┌─────────────────────────────────────────────────────────────┐
    │                   Component Startup                          │
    └────────────────────────────┬────────────────────────────────┘
                                 │
                                 ▼
    ┌─────────────────────────────────────────────────────────────┐
    │              Check health_check.protocol                     │
    └────────────────────────────┬────────────────────────────────┘
                                 │
           ┌─────────────────────┼─────────────────────┐
           ▼                     ▼                     ▼
    ┌────────────┐        ┌────────────┐        ┌────────────┐
    │  "grpc"    │        │  "http"    │        │  "auto"    │
    └─────┬──────┘        └─────┬──────┘        └─────┬──────┘
          │                     │                     │
          ▼                     ▼                     ▼
    ┌────────────┐        ┌────────────┐        ┌────────────┐
    │ gRPC Check │        │ HTTP Check │        │ Try gRPC   │
    │ grpc.health│        │ GET /health│        │            │
    └────────────┘        └────────────┘        └─────┬──────┘
                                                      │
                                               ┌──────┴──────┐
                                               │ Connection  │
                                               │   Error?    │
                                               └──────┬──────┘
                                                 No   │   Yes
                                            ┌─────────┴─────────┐
                                            ▼                   ▼
                                      ┌──────────┐       ┌──────────┐
                                      │  Done    │       │ Try HTTP │
                                      │  (gRPC)  │       │ Fallback │
                                      └──────────┘       └──────────┘
```

### Configuring Health Checks in component.yaml

Add the `health_check` section under `runtime` to configure health check behavior:

```yaml
# component.yaml
kind: agent
name: my-agent
version: 1.0.0

runtime:
  type: go
  entrypoint: ./my-agent
  port: 0

  # Health check configuration
  health_check:
    protocol: auto        # "http", "grpc", or "auto" (default: auto)
    interval: 10s         # Check interval for monitoring (default: 10s)
    timeout: 5s           # Timeout per check (default: 5s)
    endpoint: /health     # HTTP endpoint path (default: /health)
    service_name: ""      # gRPC service name (default: "" = overall health)
```

#### Health Check Configuration Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `protocol` | string | `auto` | Protocol: `http`, `grpc`, or `auto` |
| `interval` | duration | `10s` | Interval between health checks during monitoring |
| `timeout` | duration | `5s` | Timeout for each health check request |
| `endpoint` | string | `/health` | HTTP endpoint path (only for HTTP protocol) |
| `service_name` | string | `""` | gRPC service name to check (empty = overall server health) |

### Protocol Selection Guidelines

| Scenario | Recommended Protocol | Reason |
|----------|---------------------|--------|
| SDK-built components | `grpc` or `auto` | SDK uses gRPC health by default |
| HTTP-only services | `http` | Component only exposes HTTP |
| Mixed environments | `auto` | Auto-detects the correct protocol |
| Kubernetes deployments | `grpc` | K8s natively supports gRPC health probes |

### Example: Agent with gRPC Health Check

```yaml
# component.yaml for a Go agent using SDK
kind: agent
name: security-scanner
version: 1.0.0

runtime:
  type: go
  entrypoint: ./security-scanner
  port: 0
  health_check:
    protocol: grpc        # Explicitly use gRPC (recommended for SDK agents)
    timeout: 10s          # Allow more time for complex agents
```

### Example: Tool with HTTP Health Check

```yaml
# component.yaml for a tool that exposes HTTP health
kind: tool
name: legacy-scanner
version: 1.0.0

runtime:
  type: binary
  entrypoint: ./legacy-scanner
  port: 8080
  health_check:
    protocol: http        # Use HTTP for non-SDK components
    endpoint: /healthz    # Custom health endpoint
    timeout: 3s
```

### Troubleshooting Health Check Issues

#### Error: `[TIMEOUT] component=<name> timeout during health check`

This error occurs when the component doesn't respond to health checks within the startup timeout. Common causes:

| Cause | Solution |
|-------|----------|
| Protocol mismatch | If using SDK, set `protocol: grpc` in manifest |
| Slow startup | Increase startup timeout or optimize initialization |
| Wrong port | Ensure `port` in manifest matches actual listening port |
| Component crashed | Check component logs for errors |

#### Error: `[HEALTH_CHECK_FAILED] grpc health check failed`

| Cause | Solution |
|-------|----------|
| gRPC service not registered | Ensure using `sdk.ServeAgent()` or equivalent |
| Service name mismatch | Set correct `service_name` in manifest or use empty string |
| Network issues | Check firewall rules and port availability |

#### Debugging Health Checks

```bash
# Check if component is responding on expected port
nc -zv localhost 50051

# Test gRPC health check manually (requires grpcurl)
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check

# Test HTTP health check manually
curl -v http://localhost:8080/health

# View component logs for health check errors
gibson agent logs my-agent
```

---

## Checking Agent Status

The `gibson agent status` command displays detailed runtime status for an agent including health check results, uptime, and recent errors.

### Basic Usage

```bash
gibson agent status <agent-name>
```

### Sample Output

```
Agent: k8skiller
Status: running
PID: 12345
Port: 50000
Uptime: 2h 15m 30s
Started: 2025-12-31T16:10:24-06:00

Health Check:
  Status: SERVING
  Protocol: gRPC
  Response Time: 2.3ms

Health Configuration:
  Protocol: grpc
  Interval: 30s
  Timeout: 5s

Recent Errors: (none)
```

### Command Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--watch` | `-w` | false | Enable continuous monitoring mode |
| `--interval` | | 2s | Refresh interval for watch mode (minimum 1s) |
| `--errors` | | 5 | Number of recent errors to display |
| `--json` | | false | Output status as JSON |

### JSON Output

Use `--json` for machine-readable output suitable for monitoring integrations:

```bash
gibson agent status k8skiller --json
```

```json
{
  "name": "k8skiller",
  "status": "running",
  "pid": 12345,
  "port": 50000,
  "process_state": "running",
  "uptime": "2h15m30s",
  "uptime_seconds": 8130,
  "started_at": "2025-12-31T16:10:24-06:00",
  "health_check": {
    "status": "SERVING",
    "protocol": "grpc",
    "response_time_ms": 2.3
  },
  "health_config": {
    "protocol": "grpc",
    "interval": "30s",
    "timeout": "5s"
  },
  "recent_errors": []
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Agent is running and healthy |
| 1 | Agent is stopped, unhealthy, or not found |

### Watch Mode

Monitor agent status continuously:

```bash
gibson agent status k8skiller --watch
gibson agent status k8skiller --watch --interval 5s
```

Press Ctrl+C to exit watch mode.

---

## Building Tools

Tools are stateless functions that agents can call:

```go
package main

import (
    "context"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/schema"
)

func main() {
    httpTool, err := sdk.NewTool(
        sdk.WithToolName("http-request"),
        sdk.WithToolVersion("1.0.0"),
        sdk.WithToolDescription("Make HTTP requests to target endpoints"),
        sdk.WithToolTags("http", "network", "web"),

        // Input schema (what the tool accepts)
        sdk.WithInputSchema(schema.JSON{
            Type: "object",
            Properties: map[string]schema.JSON{
                "url": {
                    Type:        "string",
                    Description: "The URL to request",
                },
                "method": {
                    Type:        "string",
                    Enum:        []string{"GET", "POST", "PUT", "DELETE"},
                    Description: "HTTP method",
                },
                "headers": {
                    Type:        "object",
                    Description: "HTTP headers to include",
                },
                "body": {
                    Type:        "string",
                    Description: "Request body (for POST/PUT)",
                },
            },
            Required: []string{"url"},
        }),

        // Output schema (what the tool returns)
        sdk.WithOutputSchema(schema.JSON{
            Type: "object",
            Properties: map[string]schema.JSON{
                "status_code": {Type: "integer"},
                "headers":     {Type: "object"},
                "body":        {Type: "string"},
                "error":       {Type: "string"},
            },
        }),

        // The execution handler
        sdk.WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
            url := input["url"].(string)
            method := "GET"
            if m, ok := input["method"].(string); ok {
                method = m
            }

            // Make the HTTP request...
            resp, err := makeHTTPRequest(method, url, input)
            if err != nil {
                return map[string]any{"error": err.Error()}, nil
            }

            return map[string]any{
                "status_code": resp.StatusCode,
                "headers":     resp.Header,
                "body":        resp.Body,
            }, nil
        }),
    )
    if err != nil {
        panic(err)
    }

    sdk.ServeTool(httpTool, sdk.WithPort(50052))
}
```

---

## Building Plugins

Plugins are stateful services with multiple methods:

```go
package main

import (
    "context"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/schema"
)

func main() {
    dbPlugin, err := sdk.NewPlugin(
        sdk.WithPluginName("payload-database"),
        sdk.WithPluginVersion("1.0.0"),
        sdk.WithPluginDescription("Manages attack payloads"),

        // Define available methods
        sdk.WithMethod("search", searchHandler,
            schema.JSON{
                Type: "object",
                Properties: map[string]schema.JSON{
                    "query":    {Type: "string"},
                    "category": {Type: "string"},
                    "limit":    {Type: "integer"},
                },
            },
            schema.JSON{
                Type: "array",
                Items: &schema.JSON{Type: "object"},
            },
        ),

        sdk.WithMethod("get", getHandler,
            schema.JSON{
                Type: "object",
                Properties: map[string]schema.JSON{
                    "id": {Type: "string"},
                },
                Required: []string{"id"},
            },
            schema.JSON{Type: "object"},
        ),

        sdk.WithMethod("add", addHandler,
            schema.JSON{
                Type: "object",
                Properties: map[string]schema.JSON{
                    "content":  {Type: "string"},
                    "category": {Type: "string"},
                    "tags":     {Type: "array", Items: &schema.JSON{Type: "string"}},
                },
                Required: []string{"content", "category"},
            },
            schema.JSON{
                Type: "object",
                Properties: map[string]schema.JSON{
                    "id": {Type: "string"},
                },
            },
        ),
    )
    if err != nil {
        panic(err)
    }

    sdk.ServePlugin(dbPlugin, sdk.WithPort(50053))
}

func searchHandler(ctx context.Context, params map[string]any) (any, error) {
    query := params["query"].(string)
    // Search payloads...
    return []map[string]any{
        {"id": "1", "content": "payload1"},
        {"id": "2", "content": "payload2"},
    }, nil
}

func getHandler(ctx context.Context, params map[string]any) (any, error) {
    id := params["id"].(string)
    // Get payload by ID...
    return map[string]any{"id": id, "content": "..."}, nil
}

func addHandler(ctx context.Context, params map[string]any) (any, error) {
    // Add new payload...
    return map[string]any{"id": "new-id"}, nil
}
```

---

## Multi-Vendor LLM Usage

### The Power of Slots

Slots enable agents to use the **best LLM for each task** without hardcoding vendors:

```
┌──────────────────────────────────────────────────────────────────┐
│              Multi-Vendor LLM Architecture                        │
├──────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                     My Agent                                 │ │
│  ├─────────────────────────────────────────────────────────────┤ │
│  │                                                              │ │
│  │  Slot: "primary"          Slot: "fast"        Slot: "code"   │ │
│  │  ├─ 100K context          ├─ 8K context       ├─ 32K context │ │
│  │  ├─ tool_use, vision      ├─ (none)           ├─ (none)      │ │
│  │  └─ Main reasoning        └─ Classifications  └─ Code gen    │ │
│  │                                                              │ │
│  └───────┬───────────────────────┬───────────────────┬─────────┘ │
│          │                       │                   │           │
│          ▼                       ▼                   ▼           │
│  ┌───────────────┐      ┌───────────────┐   ┌───────────────┐   │
│  │   Anthropic   │      │    OpenAI     │   │    Ollama     │   │
│  │ Claude Opus   │      │ GPT-3.5-turbo │   │  CodeLlama    │   │
│  │               │      │               │   │   (Local)     │   │
│  │ Best at:      │      │ Best at:      │   │ Best at:      │   │
│  │ - Reasoning   │      │ - Speed       │   │ - Code        │   │
│  │ - Tool use    │      │ - Cost        │   │ - Privacy     │   │
│  │ - Vision      │      │               │   │ - Offline     │   │
│  └───────────────┘      └───────────────┘   └───────────────┘   │
│                                                                   │
└──────────────────────────────────────────────────────────────────┘
```

### Example: Agent Using Three Vendors

```go
package main

import (
    "context"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/llm"
)

func main() {
    myAgent, err := sdk.NewAgent(
        sdk.WithName("multi-vendor-agent"),
        sdk.WithVersion("1.0.0"),
        sdk.WithDescription("Uses multiple LLM vendors for optimal performance"),

        // PRIMARY: Claude for complex reasoning and tool use
        sdk.WithLLMSlot("primary", llm.SlotRequirements{
            MinContextWindow: 100000,
            RequiredFeatures: []string{"tool_use", "vision"},
        }),

        // FAST: GPT-3.5 for quick, cheap classifications
        sdk.WithLLMSlot("fast", llm.SlotRequirements{
            MinContextWindow: 4096,
            RequiredFeatures: []string{},
        }),

        // CODE: Local CodeLlama for code generation (private, offline)
        sdk.WithLLMSlot("code", llm.SlotRequirements{
            MinContextWindow: 16384,
            RequiredFeatures: []string{},
        }),

        // VISION: Gemini for image analysis (1M context!)
        sdk.WithLLMSlot("vision", llm.SlotRequirements{
            MinContextWindow: 500000,
            RequiredFeatures: []string{"vision"},
        }),

        sdk.WithExecuteFunc(executeMultiVendor),
    )
    if err != nil {
        panic(err)
    }

    sdk.ServeAgent(myAgent, sdk.WithPort(50051))
}

func executeMultiVendor(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Step 1: Use PRIMARY (Claude) for main reasoning with tools
    tools := []llm.ToolDef{
        {
            Name:        "analyze_endpoint",
            Description: "Analyze an API endpoint",
            InputSchema: map[string]any{"type": "object", "properties": map[string]any{
                "url": map[string]any{"type": "string"},
            }},
        },
    }

    planResp, err := h.CompleteWithTools(ctx, "primary", []llm.Message{
        {Role: "system", Content: "You are a security analyst. Plan the attack."},
        {Role: "user", Content: task.Goal},
    }, tools)
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Step 2: Use FAST (GPT-3.5) for quick severity classification
    // This is cheap and fast - perfect for simple decisions
    classResp, err := h.Complete(ctx, "fast", []llm.Message{
        {Role: "user", Content: "Classify severity (low/medium/high/critical): " + planResp.Content},
    })
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Step 3: Use CODE (CodeLlama) for payload generation
    // Runs locally - no data leaves the machine
    codeResp, err := h.Complete(ctx, "code", []llm.Message{
        {Role: "user", Content: "Generate a test payload for: " + planResp.Content},
    })
    if err != nil {
        // CODE slot might not be required - handle gracefully
        h.Logger().Warn("code slot unavailable, skipping payload generation")
    }

    // Step 4: Use VISION (Gemini) for screenshot analysis
    // Gemini has 1M context - great for large documents
    if hasScreenshot(task) {
        visionResp, err := h.Complete(ctx, "vision", []llm.Message{
            {Role: "user", Content: []any{
                map[string]any{"type": "text", "text": "Analyze this screenshot for vulnerabilities"},
                map[string]any{"type": "image_url", "image_url": task.ScreenshotURL},
            }},
        })
        if err != nil {
            h.Logger().Warn("vision analysis failed", "error", err)
        }
    }

    return agent.Result{
        Status: agent.StatusSuccess,
        Output: planResp.Content,
        Metadata: map[string]any{
            "severity": classResp.Content,
            "payload":  codeResp.Content,
        },
    }, nil
}
```

### Config for Multi-Vendor Agent

```yaml
# ~/.gibson/config.yaml

llm:
  default_provider: anthropic

  providers:
    anthropic:
      type: anthropic
      api_key: ${ANTHROPIC_API_KEY}
      default_model: claude-3-opus-20240229
      models:
        claude-3-opus-20240229:
          context_window: 200000
          max_output: 4096
          features: [chat, streaming, tools, vision]

    openai:
      type: openai
      api_key: ${OPENAI_API_KEY}
      default_model: gpt-3.5-turbo
      models:
        gpt-3.5-turbo:
          context_window: 16385
          max_output: 4096
          features: [chat, tools]

    ollama:
      type: ollama
      base_url: http://localhost:11434
      default_model: codellama:34b
      models:
        codellama:34b:
          context_window: 16384
          max_output: 4096
          features: [chat]

    google:
      type: google
      api_key: ${GOOGLE_API_KEY}
      default_model: gemini-1.5-pro
      models:
        gemini-1.5-pro:
          context_window: 1000000
          max_output: 8192
          features: [chat, vision, streaming]

# Override slots for this specific agent
agents:
  multi-vendor-agent:
    slots:
      primary:
        provider: anthropic
        model: claude-3-opus-20240229
        temperature: 0.7
      fast:
        provider: openai
        model: gpt-3.5-turbo
        temperature: 0.3
        max_tokens: 100
      code:
        provider: ollama
        model: codellama:34b
        temperature: 0.2
      vision:
        provider: google
        model: gemini-1.5-pro
        temperature: 0.5
```

---

## Complete Examples

### Example 1: Prompt Injection Agent

A complete agent that tests for prompt injection vulnerabilities:

```go
package main

import (
    "context"
    "fmt"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/finding"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/types"
)

func main() {
    injectionAgent, err := sdk.NewAgent(
        sdk.WithName("prompt-injection"),
        sdk.WithVersion("1.0.0"),
        sdk.WithDescription("Tests LLM applications for prompt injection vulnerabilities"),

        sdk.WithTargetTypes(
            types.TargetTypeLLMChat,
            types.TargetTypeLLMAPI,
        ),

        sdk.WithTechniqueTypes(
            types.TechniqueTypePromptInjection,
        ),

        sdk.WithCapabilities(
            agent.CapabilityAttack,
            agent.CapabilityFindingGeneration,
        ),

        // Two slots: primary for attack generation, fast for response analysis
        sdk.WithLLMSlot("primary", llm.SlotRequirements{
            MinContextWindow: 100000,
            RequiredFeatures: []string{"tool_use"},
        }),
        sdk.WithLLMSlot("analyzer", llm.SlotRequirements{
            MinContextWindow: 32000,
            RequiredFeatures: []string{},
        }),

        sdk.WithTools("http-request"),
        sdk.WithPlugins("payload-database"),

        sdk.WithExecuteFunc(executeInjection),
    )
    if err != nil {
        panic(err)
    }

    sdk.ServeAgent(injectionAgent, sdk.WithPort(50051))
}

func executeInjection(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    target := h.Target()
    logger := h.Logger()

    logger.Info("starting prompt injection test", "target", target.URL)

    // Step 1: Get payloads from database plugin
    payloads, err := h.QueryPlugin(ctx, "payload-database", "search", map[string]any{
        "category": "prompt_injection",
        "limit":    50,
    })
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, fmt.Errorf("failed to load payloads: %w", err)
    }

    // Step 2: Use PRIMARY slot to generate attack strategy
    strategyResp, err := h.Complete(ctx, "primary", []llm.Message{
        {Role: "system", Content: `You are a security researcher testing for prompt injection.
Target: ` + target.URL + `
Available payloads: ` + fmt.Sprintf("%v", payloads) + `

Generate an attack plan using the most promising payloads.`},
        {Role: "user", Content: task.Goal},
    })
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Step 3: Execute attacks using http-request tool
    for _, payload := range selectTopPayloads(payloads, 10) {
        response, err := h.CallTool(ctx, "http-request", map[string]any{
            "url":    target.URL,
            "method": "POST",
            "body":   payload,
        })
        if err != nil {
            logger.Warn("request failed", "error", err)
            continue
        }

        // Step 4: Use ANALYZER slot (cheaper/faster) to check for success
        analysisResp, err := h.Complete(ctx, "analyzer", []llm.Message{
            {Role: "user", Content: fmt.Sprintf(`Analyze if this response indicates successful prompt injection:
Payload: %s
Response: %v

Reply with JSON: {"success": true/false, "confidence": 0-100, "evidence": "..."}`,
                payload, response)},
        })
        if err != nil {
            continue
        }

        // Step 5: Submit finding if successful
        if isSuccessful(analysisResp.Content) {
            h.SubmitFinding(ctx, finding.Finding{
                Title:       "Prompt Injection Vulnerability",
                Description: "Target is vulnerable to prompt injection",
                Category:    finding.CategoryPromptInjection,
                Severity:    finding.SeverityHigh,
                Confidence:  85,
                Evidence: []finding.Evidence{
                    {
                        Type:    finding.EvidenceTypePayload,
                        Title:   "Successful Payload",
                        Content: payload.(string),
                    },
                    {
                        Type:    finding.EvidenceTypeHTTPResponse,
                        Title:   "Vulnerable Response",
                        Content: fmt.Sprintf("%v", response),
                    },
                },
            })
        }
    }

    return agent.Result{
        Status: agent.StatusSuccess,
        Output: strategyResp.Content,
    }, nil
}
```

### Example 2: Multi-Agent Coordination

An orchestrator agent that delegates to specialized agents:

```go
func executeOrchestrator(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Phase 1: Recon agent gathers information
    reconResult, err := h.DelegateToAgent(ctx, "recon", agent.Task{
        Goal: "Gather information about: " + task.Goal,
    })
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Phase 2: Use PRIMARY slot to decide next steps
    planResp, err := h.Complete(ctx, "primary", []llm.Message{
        {Role: "system", Content: "Based on recon results, plan the attack."},
        {Role: "user", Content: reconResult.Output},
    })
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Phase 3: Delegate to specialized attack agents
    agents := parseAgentsFromPlan(planResp.Content)
    for _, agentName := range agents {
        result, err := h.DelegateToAgent(ctx, agentName, agent.Task{
            Goal:    task.Goal,
            Context: reconResult.Output,
        })
        if err != nil {
            h.Logger().Warn("agent failed", "agent", agentName, "error", err)
            continue
        }

        h.Logger().Info("agent completed", "agent", agentName, "status", result.Status)
    }

    return agent.Result{Status: agent.StatusSuccess}, nil
}
```

---

## Token Usage Tracking

The framework tracks token usage per slot:

```go
func executeWithTracking(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Make some completions
    h.Complete(ctx, "primary", messages1)
    h.Complete(ctx, "primary", messages2)
    h.Complete(ctx, "fast", messages3)

    // Get usage statistics
    tracker := h.TokenUsage()

    primaryUsage := tracker.BySlot("primary")
    fastUsage := tracker.BySlot("fast")
    totalUsage := tracker.Total()

    h.Logger().Info("token usage",
        "primary_input", primaryUsage.InputTokens,
        "primary_output", primaryUsage.OutputTokens,
        "fast_input", fastUsage.InputTokens,
        "fast_output", fastUsage.OutputTokens,
        "total", totalUsage.TotalTokens,
    )

    return agent.Result{Status: agent.StatusSuccess}, nil
}
```

---

## Best Practices

### 1. Slot Naming Conventions

| Slot Name | Purpose | Typical Provider |
|-----------|---------|------------------|
| `primary` | Main reasoning, complex tasks | Claude Opus, GPT-4 |
| `fast` | Quick classifications, simple tasks | Claude Haiku, GPT-3.5 |
| `code` | Code generation | CodeLlama, GPT-4 |
| `vision` | Image analysis | Gemini, GPT-4V |
| `embedding` | Vector embeddings | OpenAI Ada, local |

### 2. Required vs Optional Slots

```go
// Required slot - agent fails if unavailable
sdk.WithLLMSlot("primary", llm.SlotRequirements{
    Required: true,  // default
    MinContextWindow: 100000,
})

// Optional slot - agent continues without it
sdk.WithLLMSlot("vision", llm.SlotRequirements{
    Required: false,
    RequiredFeatures: []string{"vision"},
})
```

### 3. Graceful Degradation

```go
func execute(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Try vision slot, fall back to primary
    resp, err := h.Complete(ctx, "vision", visionMessages)
    if err != nil {
        h.Logger().Warn("vision unavailable, using primary")
        resp, err = h.Complete(ctx, "primary", textMessages)
    }
    // ...
}
```

### 4. Cost Optimization

```go
// Use cheaper models for simple tasks
classResp, _ := h.Complete(ctx, "fast", []llm.Message{
    {Role: "user", Content: "Yes or No: " + question},
}, sdk.WithMaxTokens(10))  // Limit output

// Use expensive models only when needed
if needsDeepAnalysis(classResp.Content) {
    h.Complete(ctx, "primary", complexMessages)
}
```

---

## Troubleshooting

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `ErrNoMatchingProvider` | Provider not configured | Add provider to config.yaml |
| `ErrModelNotFound` | Model not in provider's list | Check model name spelling |
| `ErrInvalidSlotConfig` | Model doesn't meet constraints | Use model with required features |
| `ErrProviderUnavailable` | API unreachable | Check network, API keys |

### Debugging Slot Resolution

```go
func execute(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Log available slots
    h.Logger().Debug("attempting slot resolution", "slot", "primary")

    resp, err := h.Complete(ctx, "primary", messages)
    if err != nil {
        h.Logger().Error("slot resolution failed",
            "slot", "primary",
            "error", err,
            "target", h.Target().URL,
        )
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Log token usage
    usage := h.TokenUsage().BySlot("primary")
    h.Logger().Info("completion successful",
        "slot", "primary",
        "input_tokens", usage.InputTokens,
        "output_tokens", usage.OutputTokens,
    )

    return agent.Result{Status: agent.StatusSuccess}, nil
}
```

---

## Agent Streaming and TUI Integration

### Overview

When agents are executed through the Gibson TUI, real-time streaming of agent events is displayed in the console view. This enables operators to monitor agent progress, see tool calls, view findings as they're discovered, and interact with agents during execution.

### Stream Event Types

Agents emit events that are displayed in the TUI console:

| Event Type | Description | TUI Display |
|------------|-------------|-------------|
| `output` | Agent LLM reasoning/output text | Agent output styled text |
| `tool_call` | Agent invoking a tool | ">>> Calling tool: toolname" with arguments |
| `tool_result` | Tool execution response | "<<< Tool result: success/failed" with output |
| `finding` | Security vulnerability discovered | Severity-colored finding block |
| `status` | Agent state change (running, paused, completed) | Status message with context |
| `steering_ack` | Acknowledgment of operator steering | Acceptance/rejection message |
| `error` | Agent error occurred | Error-styled message with details |

### Focusing on an Agent

The `/focus` command in the TUI console subscribes to an agent's event stream:

```
> /focus davinci
```

This:
1. Verifies the agent is registered and running
2. Subscribes to the StreamManager for that agent
3. Starts an EventProcessor that forwards events to the TUI
4. Displays real-time agent output in the console

### Sending Steering Commands

While focused on an agent, operators can send commands:

| Command | Description |
|---------|-------------|
| `/steer <message>` | Send guidance to the agent |
| `/interrupt` | Request the agent to pause |
| `/resume` | Resume a paused agent |
| `/mode auto\|interactive` | Switch agent operation mode |
| `/unfocus` | Stop streaming and unfocus |

### Event Content Structures

Agents should emit events with JSON content matching these structures:

**Output Event:**
```json
{
  "text": "Agent reasoning or output text",
  "complete": true
}
```

**Tool Call Event:**
```json
{
  "tool_name": "http-request",
  "tool_id": "call-123",
  "arguments": {"url": "https://target.com", "method": "GET"}
}
```

**Tool Result Event:**
```json
{
  "tool_id": "call-123",
  "success": true,
  "output": {"status_code": 200, "body": "..."},
  "error": ""
}
```

**Finding Event:**
```json
{
  "id": "finding-uuid",
  "title": "SQL Injection in Login Form",
  "severity": "critical",
  "category": "injection",
  "description": "Parameter 'username' is vulnerable..."
}
```

**Status Event:**
```json
{
  "status": "running|paused|completed|failed|waiting_for_input|interrupted",
  "message": "Optional status message",
  "reason": "Optional reason for status change"
}
```

**Steering Acknowledgment Event:**
```json
{
  "sequence": 5,
  "accepted": true,
  "message": "Optional response message"
}
```

**Error Event:**
```json
{
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

### Best Practices for Agent Streaming

1. **Emit Complete Output**: Set `complete: true` when a reasoning block is finished to improve TUI rendering.

2. **Provide Tool Context**: Include meaningful `tool_name` and structured `arguments` in tool calls.

3. **Acknowledge Steering**: Always emit `steering_ack` events when steering messages are received.

4. **Report Status Changes**: Emit `status` events when the agent's operational state changes.

5. **Structure Findings**: Include `id`, `title`, `severity`, and `category` for proper TUI display.

### Stream Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                       Agent Streaming Flow                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐   │
│  │   Agent     │────▶│  StreamManager   │────▶│ EventProcessor  │   │
│  │  (gRPC)     │     │                  │     │                 │   │
│  │             │     │  Subscribe()     │     │  Start()        │   │
│  │ Emit Events │     │  Unsubscribe()   │     │  processLoop()  │   │
│  └─────────────┘     └────────┬─────────┘     └───────┬─────────┘   │
│                               │                       │              │
│                               │                       ▼              │
│                               │              ┌─────────────────┐    │
│                               │              │   tea.Program   │    │
│                               │              │                 │    │
│                               │              │  Send(msg)      │    │
│                               │              └───────┬─────────┘    │
│                               │                      │               │
│                               ▼                      ▼               │
│                      ┌─────────────────┐    ┌─────────────────┐     │
│                      │   SessionDAO    │    │  ConsoleView    │     │
│                      │                 │    │                 │     │
│                      │ Persist Events  │    │ EventRenderer   │     │
│                      │                 │    │ Display Output  │     │
│                      └─────────────────┘    └─────────────────┘     │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

### Steering Message Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Operator Steering Flow                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────┐     ┌──────────────────┐     ┌─────────────────┐   │
│  │ TUI Console │────▶│  StreamManager   │────▶│     Agent       │   │
│  │             │     │                  │     │                 │   │
│  │ /steer msg  │     │ SendSteering()   │     │ Receive steer   │   │
│  │ /interrupt  │     │ SendInterrupt()  │     │ Process command │   │
│  │ /resume     │     │ Resume()         │     │ Emit steering   │   │
│  │ /mode xxx   │     │ SetMode()        │     │   _ack event    │   │
│  └─────────────┘     └──────────────────┘     └─────────────────┘   │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Streaming Agents

### Overview of Streaming Support

Streaming enables agents to provide **real-time feedback** during execution rather than just returning a final result. When an agent is executed through the Gibson TUI or connected via the `StreamExecute` RPC, streaming events are displayed live in the console view, allowing operators to:

- Monitor agent reasoning and decision-making in real-time
- See tool invocations as they happen
- Receive findings immediately upon discovery
- Interact with agents through steering messages
- Switch between autonomous and interactive modes
- Interrupt and resume agent execution with guidance

**When to Use Streaming vs Regular Execution:**

| Use Streaming When... | Use Regular Execution When... |
|----------------------|-------------------------------|
| Agent has long-running operations | Agent completes quickly (< 5 seconds) |
| Interactive oversight is needed | Fully autonomous operation is acceptable |
| Real-time progress feedback improves UX | Final result is sufficient |
| Agent needs to request user input | No user interaction is required |
| Debugging complex agent behavior | Simple, deterministic operations |

### StreamingAgent Interface

Agents can opt into streaming support by implementing the `StreamingAgent` interface:

```go
type StreamingAgent interface {
    agent.Agent

    // ExecuteStreaming runs the agent with streaming event emission support.
    // The StreamingHarness provides methods to emit events during execution.
    ExecuteStreaming(ctx context.Context, harness StreamingHarness, task agent.Task) (agent.Result, error)
}
```

**Key Points:**
- `StreamingAgent` embeds the base `Agent` interface
- The `ExecuteStreaming` method receives a `StreamingHarness` instead of the base `Harness`
- Agents that don't implement `StreamingAgent` still work with streaming - the framework automatically emits events by intercepting harness calls

### StreamingHarness Interface

The `StreamingHarness` extends the base `agent.Harness` with bidirectional streaming capabilities:

```go
type StreamingHarness interface {
    // Embed the base Harness interface
    agent.Harness

    // Event Emission Methods
    EmitOutput(content string, isReasoning bool) error
    EmitToolCall(toolName string, input map[string]any, callID string) error
    EmitToolResult(callID string, output map[string]any, success bool) error
    EmitFinding(finding agent.Finding) error
    EmitStatus(status proto.AgentStatus, message string) error
    EmitError(code string, message string, fatal bool) error

    // Steering and Mode Methods
    Steering() <-chan *proto.SteeringMessage
    Mode() proto.AgentMode
    SetMode(mode proto.AgentMode)
}
```

#### Event Emission Methods

**EmitOutput(content string, isReasoning bool)**
- Emits text output chunks to the client
- Set `isReasoning=true` for internal reasoning/thinking output
- Set `isReasoning=false` for final user-facing output

**EmitToolCall(toolName string, input map[string]any, callID string)**
- Emits an event indicating a tool invocation is starting
- `callID` should be a unique identifier for correlating with the result
- The framework displays ">>> Calling tool: toolname" in the TUI

**EmitToolResult(callID string, output map[string]any, success bool)**
- Emits the result of a tool invocation
- `callID` must match the ID from the corresponding `EmitToolCall`
- Set `success=true` if the tool executed successfully, `false` if it failed

**EmitFinding(finding agent.Finding)**
- Emits a security finding discovered during testing
- The finding will be both streamed to the client and recorded via `SubmitFinding`
- TUI displays findings with severity-colored formatting

**EmitStatus(status proto.AgentStatus, message string)**
- Emits an agent status change (running, paused, waiting, completed, failed)
- The message provides additional context about the status change
- Available statuses:
  - `AGENT_STATUS_RUNNING` - Agent is executing
  - `AGENT_STATUS_PAUSED` - Agent is paused (interrupted)
  - `AGENT_STATUS_WAITING_FOR_INPUT` - Agent needs user input
  - `AGENT_STATUS_COMPLETED` - Agent finished successfully
  - `AGENT_STATUS_FAILED` - Agent encountered an error
  - `AGENT_STATUS_INTERRUPTED` - Agent was interrupted by client

**EmitError(code string, message string, fatal bool)**
- Emits an error event to the client
- Set `fatal=true` if the error should terminate execution
- Set `fatal=false` for recoverable errors that the agent can handle

#### Steering and Mode Methods

**Steering() <-chan *proto.SteeringMessage**
- Returns a receive-only channel for steering messages from the client
- Agents can listen on this channel to receive user input, approvals, or guidance
- Steering messages contain:
  - `Id` - Unique message identifier
  - `Content` - The steering message text
  - `Timestamp` - When the message was sent

**Mode() proto.AgentMode**
- Returns the current execution mode
- Available modes:
  - `AGENT_MODE_AUTONOMOUS` - Agent operates independently
  - `AGENT_MODE_INTERACTIVE` - Agent waits for user approval before actions

**SetMode(mode proto.AgentMode)**
- Updates the current execution mode atomically
- Called by the framework when the client sends a `SetModeRequest`
- Agents generally should not call this directly

### Creating a Streaming Agent

Use the `WithStreamingExecuteFunc` option to create an agent with streaming support:

```go
package main

import (
    "context"
    "fmt"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/api/gen/proto"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    myAgent, err := sdk.NewAgent(
        sdk.WithName("streaming-recon"),
        sdk.WithVersion("1.0.0"),
        sdk.WithDescription("Reconnaissance agent with streaming support"),

        sdk.WithLLMSlot("primary", llm.SlotRequirements{
            MinContextWindow: 100000,
            RequiredFeatures: []string{"tool_use"},
        }),

        sdk.WithTools("http-request", "dns-lookup"),

        // Add streaming execution function
        sdk.WithStreamingExecuteFunc(executeStreaming),
    )
    if err != nil {
        panic(err)
    }

    serve.Agent(myAgent, serve.WithPort(50051))
}

func executeStreaming(ctx context.Context, h serve.StreamingHarness, task agent.Task) (agent.Result, error) {
    // Emit status to indicate we're starting
    h.EmitStatus(proto.AgentStatus_AGENT_STATUS_RUNNING, "Starting reconnaissance")

    // Emit reasoning output
    h.EmitOutput("Analyzing target: "+h.Target().URL, true)

    // Phase 1: DNS lookup
    h.EmitOutput("Phase 1: DNS Reconnaissance", false)

    dnsResult, err := h.CallTool(ctx, "dns-lookup", map[string]any{
        "domain": h.Target().URL,
    })
    if err != nil {
        h.EmitError("DNS_LOOKUP_FAILED", err.Error(), false)
        // Continue despite error
    }

    // Phase 2: HTTP analysis
    h.EmitOutput("Phase 2: HTTP Analysis", false)

    httpResult, err := h.CallTool(ctx, "http-request", map[string]any{
        "url":    h.Target().URL,
        "method": "GET",
    })
    if err != nil {
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Check if we found a vulnerability
    if isVulnerable(httpResult) {
        finding := agent.Finding{
            Title:       "Information Disclosure",
            Description: "Server reveals sensitive information in headers",
            Severity:    "high",
            Category:    "information_disclosure",
        }
        h.EmitFinding(finding)
    }

    h.EmitStatus(proto.AgentStatus_AGENT_STATUS_COMPLETED, "Reconnaissance complete")

    return agent.Result{
        Status: agent.StatusSuccess,
        Output: "Reconnaissance completed successfully",
    }, nil
}
```

### Emitting Events

#### When to Emit Each Event Type

| Event Type | When to Emit | Example Use Case |
|------------|--------------|------------------|
| `EmitOutput` (reasoning) | Internal agent thinking, planning, analysis | "Analyzing RBAC permissions for privilege escalation paths" |
| `EmitOutput` (result) | Final output, phase transitions, summaries | "Found 3 vulnerable service accounts" |
| `EmitToolCall` | Before invoking a tool (manual calls only) | Before calling kubectl to list pods |
| `EmitToolResult` | After tool completes (manual calls only) | After kubectl returns pod list |
| `EmitFinding` | When a vulnerability is discovered | SQL injection found in login form |
| `EmitStatus` | Agent state changes, phase transitions | Starting phase 2, waiting for approval |
| `EmitError` | Recoverable or fatal errors occur | API rate limit exceeded (recoverable) |

#### Example: Emitting Events

```go
func executeStreaming(ctx context.Context, h serve.StreamingHarness, task agent.Task) (agent.Result, error) {
    // Emit reasoning output
    h.EmitOutput("Planning attack strategy...", true)

    // Get attack plan from LLM
    plan, err := h.Complete(ctx, "primary", []llm.Message{
        {Role: "system", Content: "You are a security tester."},
        {Role: "user", Content: "Plan an attack on: " + h.Target().URL},
    })
    if err != nil {
        // Emit error event
        h.EmitError("LLM_ERROR", fmt.Sprintf("Failed to generate plan: %v", err), true)
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Emit the plan as output
    h.EmitOutput(plan.Content, false)

    // Execute each step of the plan
    for i, step := range parseSteps(plan.Content) {
        // Emit status for phase transition
        h.EmitStatus(proto.AgentStatus_AGENT_STATUS_RUNNING,
            fmt.Sprintf("Executing step %d of %d", i+1, len(steps)))

        // Execute the step (CallTool automatically emits ToolCall and ToolResult)
        result, err := h.CallTool(ctx, step.ToolName, step.Input)
        if err != nil {
            // Emit non-fatal error
            h.EmitError("TOOL_ERROR", fmt.Sprintf("Step %d failed: %v", i+1, err), false)
            continue
        }

        // Check for findings
        if finding := analyzeResult(result); finding != nil {
            h.EmitFinding(*finding)
        }
    }

    return agent.Result{Status: agent.StatusSuccess}, nil
}
```

### Automatic Event Emission

**The `StreamingHarness` automatically emits events when agents use harness methods**, even if the agent doesn't explicitly implement `StreamingAgent`. This means you get streaming support for free by using the harness methods:

#### Automatic Event Interception

| Harness Method | Automatic Events Emitted |
|----------------|-------------------------|
| `CallTool(ctx, name, input)` | `ToolCallEvent` before, `ToolResultEvent` after |
| `SubmitFinding(ctx, finding)` | `FindingEvent` before submitting |
| `Complete(ctx, slot, messages)` | `OutputChunk` event with LLM response content |
| `Stream(ctx, slot, messages)` | `OutputChunk` event for each streaming chunk |

#### Example: Non-Streaming Agent with Automatic Events

```go
// This agent doesn't implement StreamingAgent, but still gets streaming events!
func executeRegular(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // When this runs through StreamExecute, the framework wraps the harness
    // and automatically emits events for all harness method calls

    // This automatically emits ToolCallEvent and ToolResultEvent
    result, err := h.CallTool(ctx, "nmap", map[string]any{
        "target": h.Target().URL,
    })

    // This automatically emits OutputChunk event with the LLM response
    resp, err := h.Complete(ctx, "primary", []llm.Message{
        {Role: "user", Content: "Analyze: " + result["output"].(string)},
    })

    // This automatically emits FindingEvent
    h.SubmitFinding(ctx, myFinding)

    return agent.Result{Status: agent.StatusSuccess}, nil
}
```

### Steering and Interactive Mode

Agents can receive steering messages from operators during execution and adjust their behavior based on the current mode:

```go
func executeStreaming(ctx context.Context, h serve.StreamingHarness, task agent.Task) (agent.Result, error) {
    // Check the current mode
    if h.Mode() == proto.AgentMode_AGENT_MODE_INTERACTIVE {
        // Wait for user approval before proceeding
        h.EmitStatus(proto.AgentStatus_AGENT_STATUS_WAITING_FOR_INPUT,
            "Waiting for approval to execute exploit")

        select {
        case msg := <-h.Steering():
            // User sent steering message
            if msg.Content == "approve" {
                h.EmitOutput("Proceeding with approval", false)
            } else {
                return agent.Result{Status: agent.StatusCancelled}, nil
            }
        case <-ctx.Done():
            return agent.Result{Status: agent.StatusCancelled}, ctx.Err()
        }
    }

    // Listen for steering messages while executing
    go func() {
        for {
            select {
            case msg := <-h.Steering():
                // Handle steering message
                h.EmitOutput(fmt.Sprintf("Received guidance: %s", msg.Content), true)
                // Adjust agent behavior based on guidance
            case <-ctx.Done():
                return
            }
        }
    }()

    // Continue with normal execution
    // ...

    return agent.Result{Status: agent.StatusSuccess}, nil
}
```

### Steering Commands from TUI

When an operator focuses on an agent in the TUI, they can send commands:

| Command | Description | Agent Handling |
|---------|-------------|----------------|
| `/steer <message>` | Send guidance to the agent | Received via `Steering()` channel |
| `/interrupt` | Request the agent to pause | Framework emits `PAUSED` status |
| `/resume [guidance]` | Resume a paused agent | Framework emits `RUNNING` status, optional guidance via `Steering()` |
| `/mode auto` | Switch to autonomous mode | `Mode()` returns `AGENT_MODE_AUTONOMOUS` |
| `/mode interactive` | Switch to interactive mode | `Mode()` returns `AGENT_MODE_INTERACTIVE` |
| `/unfocus` | Stop streaming and unfocus | Stream is closed |

### Best Practices for Streaming Agents

1. **Emit Status Changes**: Always emit status events at major phase transitions so operators understand what the agent is doing.

2. **Use Reasoning Output**: Emit reasoning output (isReasoning=true) for internal thoughts and planning to provide transparency.

3. **Handle Interrupts Gracefully**: Check `ctx.Done()` regularly and respond to interrupts promptly.

4. **Acknowledge Steering**: When receiving steering messages, emit output to acknowledge receipt and show how the guidance is being used.

5. **Provide Context in Errors**: Include actionable context in error messages so operators can intervene if needed.

6. **Don't Over-Emit**: Balance real-time feedback with noise - emit meaningful events, not every minor operation.

7. **Use Automatic Emission When Possible**: Leverage the automatic event emission from harness methods rather than manually emitting ToolCall/ToolResult for every operation.

8. **Test Both Modes**: Test your agent in both autonomous and interactive modes to ensure mode switching works correctly.

### Complete Streaming Agent Example

Here's a complete example of a Kubernetes security testing agent with streaming support:

```go
package main

import (
    "context"
    "fmt"
    "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/api/gen/proto"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    k8sAgent, err := sdk.NewAgent(
        sdk.WithName("k8s-security-scanner"),
        sdk.WithVersion("1.0.0"),
        sdk.WithDescription("Kubernetes security scanner with streaming"),

        sdk.WithLLMSlot("primary", llm.SlotRequirements{
            MinContextWindow: 100000,
            RequiredFeatures: []string{"tool_use"},
        }),

        sdk.WithTools("kubectl", "rbac-analyzer"),

        sdk.WithStreamingExecuteFunc(executeK8sScanning),
    )
    if err != nil {
        panic(err)
    }

    serve.Agent(k8sAgent, serve.WithPort(50051))
}

func executeK8sScanning(ctx context.Context, h serve.StreamingHarness, task agent.Task) (agent.Result, error) {
    h.EmitStatus(proto.AgentStatus_AGENT_STATUS_RUNNING, "Starting Kubernetes security scan")

    // Phase 1: Reconnaissance
    h.EmitOutput("Phase 1: Reconnaissance", false)
    h.EmitOutput("Gathering cluster information...", true)

    clusterInfo, err := h.CallTool(ctx, "kubectl", map[string]any{
        "args": []string{"cluster-info"},
    })
    if err != nil {
        h.EmitError("RECON_FAILED", fmt.Sprintf("Failed to get cluster info: %v", err), true)
        return agent.Result{Status: agent.StatusFailed}, err
    }

    // Phase 2: RBAC Analysis
    h.EmitOutput("Phase 2: RBAC Analysis", false)

    // Check if we're in interactive mode - if so, wait for approval
    if h.Mode() == proto.AgentMode_AGENT_MODE_INTERACTIVE {
        h.EmitStatus(proto.AgentStatus_AGENT_STATUS_WAITING_FOR_INPUT,
            "Waiting for approval to analyze RBAC permissions")

        select {
        case msg := <-h.Steering():
            if msg.Content != "approve" {
                h.EmitOutput("RBAC analysis cancelled by operator", false)
                return agent.Result{Status: agent.StatusCancelled}, nil
            }
        case <-ctx.Done():
            return agent.Result{Status: agent.StatusCancelled}, ctx.Err()
        }
    }

    rbacResult, err := h.CallTool(ctx, "rbac-analyzer", map[string]any{
        "namespace": "default",
    })
    if err != nil {
        h.EmitError("RBAC_ANALYSIS_FAILED", err.Error(), false)
        // Continue despite error
    } else {
        // Analyze results with LLM
        h.EmitOutput("Analyzing RBAC configuration for privilege escalation paths...", true)

        analysis, err := h.Complete(ctx, "primary", []llm.Message{
            {Role: "system", Content: "You are a Kubernetes security expert."},
            {Role: "user", Content: fmt.Sprintf("Analyze this RBAC configuration for vulnerabilities: %v", rbacResult)},
        })
        if err != nil {
            h.EmitError("LLM_ERROR", err.Error(), false)
        } else if containsVulnerability(analysis.Content) {
            finding := agent.Finding{
                Title:       "Overprivileged Service Account",
                Description: analysis.Content,
                Severity:    "high",
                Category:    "privilege_escalation",
            }
            h.EmitFinding(finding)
        }
    }

    // Phase 3: Pod Security Analysis
    h.EmitOutput("Phase 3: Pod Security Analysis", false)

    pods, err := h.CallTool(ctx, "kubectl", map[string]any{
        "args": []string{"get", "pods", "-o", "json"},
    })
    if err != nil {
        h.EmitError("POD_ENUM_FAILED", err.Error(), false)
    }

    h.EmitStatus(proto.AgentStatus_AGENT_STATUS_COMPLETED, "Security scan complete")

    return agent.Result{
        Status: agent.StatusSuccess,
        Output: "Kubernetes security scan completed successfully",
    }, nil
}

func containsVulnerability(analysis string) bool {
    // Parse LLM analysis for vulnerability indicators
    // Implementation details omitted for brevity
    return false
}
```

---

## Summary

The Gibson SDK's slot system provides:

1. **Vendor Abstraction**: Agents declare needs, not specific models
2. **Runtime Flexibility**: Config changes don't require code changes
3. **Multi-Vendor Support**: Use best model for each task
4. **Cost Control**: Route cheap tasks to cheap models
5. **Offline Capability**: Local models (Ollama) for air-gapped environments
6. **Token Tracking**: Monitor usage per slot for cost analysis
7. **Real-Time Streaming**: Live event streaming for interactive agent oversight
8. **Bidirectional Control**: Steering messages enable operator guidance during execution

For more examples, see the `examples/` directory in the SDK.
