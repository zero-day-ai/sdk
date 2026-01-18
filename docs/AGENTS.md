# Agents Quickstart Guide

Build autonomous AI security testing agents with the Gibson SDK.

## What is an Agent?

An agent is an autonomous security testing component that:
- Executes security tests against target systems
- Uses LLMs for reasoning and decision-making
- Invokes tools to perform actions
- Reports findings when vulnerabilities are discovered
- Manages state across execution phases

## Minimal Agent (5 Minutes)

```go
package main

import (
    "context"
    "log"

    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/types"
)

func main() {
    cfg := agent.NewConfig().
        SetName("my-agent").
        SetVersion("1.0.0").
        SetDescription("My first security testing agent").
        AddCapability(agent.CapabilityPromptInjection).
        AddTargetType(types.TargetTypeLLMChat).
        AddLLMSlot("primary", llm.SlotRequirements{
            MinContextWindow: 8000,
        }).
        SetExecuteFunc(execute)

    myAgent, err := agent.New(cfg)
    if err != nil {
        log.Fatal(err)
    }

    // Agent is ready - register with Gibson or run standalone
    log.Printf("Agent created: %s v%s", myAgent.Name(), myAgent.Version())
}

func execute(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    // Use the harness to interact with LLMs, tools, memory, etc.
    resp, err := h.Complete(ctx, "primary", []llm.Message{
        {Role: llm.RoleSystem, Content: "You are a security tester."},
        {Role: llm.RoleUser, Content: task.Goal},
    })
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    return agent.NewSuccessResult(resp.Content), nil
}
```

## Agent Configuration

### Required Fields

| Field | Description |
|-------|-------------|
| `Name` | Unique identifier (kebab-case, e.g., "prompt-injector") |
| `Version` | Semantic version (e.g., "1.0.0") |
| `Description` | Human-readable purpose |
| `ExecuteFunc` | The main execution logic |

### Optional Fields

| Field | Description |
|-------|-------------|
| `Capabilities` | What the agent can test for |
| `TargetTypes` | What systems it can test |
| `TechniqueTypes` | Attack techniques used |
| `LLMSlots` | LLM requirements |
| `InitFunc` | Initialization logic |
| `ShutdownFunc` | Cleanup logic |
| `HealthFunc` | Health check logic |

## LLM Slots

Slots define your agent's LLM requirements. Gibson routes requests to appropriate providers.

```go
// Single slot for simple agents
cfg.AddLLMSlot("primary", llm.SlotRequirements{
    MinContextWindow: 8000,
})

// Multiple slots for complex agents
cfg.AddLLMSlot("planner", llm.SlotRequirements{
    MinContextWindow: 32000,
    RequiredFeatures: []string{"function_calling"},
    PreferredModels:  []string{"claude-3-opus", "gpt-4-turbo"},
})

cfg.AddLLMSlot("executor", llm.SlotRequirements{
    MinContextWindow: 8000,
    PreferredModels:  []string{"claude-3-haiku", "gpt-4o-mini"},
})
```

## The Harness

The `agent.Harness` is your runtime environment. It provides access to everything:

### LLM Access

```go
// Simple completion
resp, err := h.Complete(ctx, "primary", messages)

// With options
resp, err := h.Complete(ctx, "primary", messages,
    llm.WithTemperature(0.7),
    llm.WithMaxTokens(1000),
)

// Streaming
stream, err := h.Stream(ctx, "primary", messages)
for chunk := range stream {
    fmt.Print(chunk.Delta)
}

// With tool calling
resp, err := h.CompleteWithTools(ctx, "primary", messages, toolDefs)

// Structured output (parses JSON into struct)
type Analysis struct {
    Vulnerabilities []string `json:"vulnerabilities"`
}
result, err := h.CompleteStructured(ctx, "primary", messages, Analysis{})
```

### Tool Execution

```go
// Call a tool
output, err := h.CallTool(ctx, "http-request", map[string]any{
    "url":    "https://target.com/api",
    "method": "GET",
})

// Parallel tool calls
results, err := h.CallToolsParallel(ctx, []agent.ToolCall{
    {Name: "http-request", Input: map[string]any{"url": url1}},
    {Name: "http-request", Input: map[string]any{"url": url2}},
}, 10) // max concurrency
```

### Memory

```go
store := h.Memory()

// Working memory - ephemeral scratch space
store.Working().Set(ctx, "step", 1)
val, _ := store.Working().Get(ctx, "step")

// Mission memory - persists within mission
store.Mission().Set(ctx, "discovered", endpoints, metadata)
results, _ := store.Mission().Search(ctx, "endpoints", 10)

// Long-term memory - vector-based, cross-mission
id, _ := store.LongTerm().Store(ctx, "attack pattern worked", meta)
similar, _ := store.LongTerm().Search(ctx, "bypass techniques", 5, nil)
```

### Findings

```go
import "github.com/zero-day-ai/sdk/finding"

f := finding.NewFinding(
    h.Mission().ID,
    "my-agent",
    "SQL Injection Found",
    "The login endpoint is vulnerable to SQL injection",
    finding.CategoryInjection,
    finding.SeverityHigh,
)
f.Confidence = 0.95
f.AddEvidence(finding.Evidence{
    Type:    finding.EvidenceTypeRequest,
    Content: "POST /login username=' OR '1'='1",
})

err := h.SubmitFinding(ctx, f)
```

### Context

```go
// Mission info
mission := h.Mission()
log.Printf("Mission: %s", mission.ID)

// Target info
target := h.Target()
log.Printf("Target: %s (%s)", target.URL, target.Type)

// Logging
logger := h.Logger()
logger.Info("testing endpoint", "url", endpoint)

// Tracing
tracer := h.Tracer()
ctx, span := tracer.Start(ctx, "analyze-response")
defer span.End()
```

## Agentic Loop Pattern

Most agents follow an iterative loop pattern:

```go
func execute(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
    logger := h.Logger()

    messages := []llm.Message{
        {Role: llm.RoleSystem, Content: systemPrompt},
        {Role: llm.RoleUser, Content: task.Goal},
    }

    tools := getToolDefinitions()
    maxIterations := 20

    for i := 0; i < maxIterations; i++ {
        logger.Info("iteration", "count", i+1)

        resp, err := h.CompleteWithTools(ctx, "primary", messages, tools)
        if err != nil {
            return agent.NewFailedResult(err), err
        }

        // No tool calls = agent is done
        if len(resp.ToolCalls) == 0 {
            return agent.NewSuccessResult(resp.Content), nil
        }

        // Add assistant message
        messages = append(messages, llm.Message{
            Role:      llm.RoleAssistant,
            Content:   resp.Content,
            ToolCalls: resp.ToolCalls,
        })

        // Execute each tool call
        for _, tc := range resp.ToolCalls {
            output, err := h.CallTool(ctx, tc.Name, tc.Arguments)

            messages = append(messages, llm.Message{
                Role:       llm.RoleTool,
                ToolCallID: tc.ID,
                Content:    formatToolResult(output, err),
            })
        }
    }

    return agent.NewPartialResult("max iterations reached", nil), nil
}
```

## Lifecycle Hooks

Add initialization and cleanup:

```go
cfg := agent.NewConfig().
    SetName("stateful-agent").
    SetVersion("1.0.0").
    SetExecuteFunc(execute).
    SetInitFunc(func(ctx context.Context, config map[string]any) error {
        // Load models, open connections, etc.
        return nil
    }).
    SetShutdownFunc(func(ctx context.Context) error {
        // Close connections, save state, etc.
        return nil
    }).
    SetHealthFunc(func(ctx context.Context) types.HealthStatus {
        // Report health status
        return types.NewHealthyStatus("all systems operational")
    })
```

## Serving Your Agent

### Local Mode (Development)

```go
import "github.com/zero-day-ai/sdk/serve"

err := serve.Agent(myAgent,
    serve.WithPort(50051),
    serve.WithLocalMode("~/.gibson/run/agents/my-agent.sock"),
)
```

### Remote Mode (Production)

```go
err := serve.Agent(myAgent,
    serve.WithPort(50051),
    serve.WithTLS("cert.pem", "key.pem"),
    serve.WithGracefulShutdown(30*time.Second),
)
```

## Running with Gibson

```bash
# Build your agent
make build

# Register with Gibson
gibson agent install ./my-agent

# Run a mission
gibson agent run my-agent --target https://example.com
```

## Best Practices

1. **Use meaningful slot names** - "planner", "executor", "analyzer" not "slot1", "slot2"
2. **Log extensively** - Use `h.Logger()` for debugging and audit trails
3. **Handle errors gracefully** - Return `agent.NewFailedResult(err)` with context
4. **Respect task constraints** - Check `task.Constraints.MaxTurns`, `MaxTokens`, etc.
5. **Submit findings early** - Don't wait until the end; submit as you discover
6. **Use memory tiers appropriately**:
   - Working: Temporary calculations, loop counters
   - Mission: Discovered data, conversation history
   - Long-term: Patterns that help future missions

## Next Steps

- See `examples/minimal-agent/` for a complete working example
- Read the [Tools Guide](TOOLS.md) to understand tool integration
- Check the [main README](../README.md) for advanced features
