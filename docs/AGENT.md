# Gibson SDK Agent Development Guide

This guide explains how to build agents, tools, and plugins for the Gibson Framework, with a focus on the LLM slot system that enables agents to use multiple LLMs from different vendors.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [LLM Slot System](#llm-slot-system)
3. [Configuration](#configuration)
4. [Building Agents](#building-agents)
5. [Building Tools](#building-tools)
6. [Building Plugins](#building-plugins)
7. [Multi-Vendor LLM Usage](#multi-vendor-llm-usage)
8. [Complete Examples](#complete-examples)

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

## Summary

The Gibson SDK's slot system provides:

1. **Vendor Abstraction**: Agents declare needs, not specific models
2. **Runtime Flexibility**: Config changes don't require code changes
3. **Multi-Vendor Support**: Use best model for each task
4. **Cost Control**: Route cheap tasks to cheap models
5. **Offline Capability**: Local models (Ollama) for air-gapped environments
6. **Token Tracking**: Monitor usage per slot for cost analysis

For more examples, see the `examples/` directory in the SDK.
