# Harness Callbacks Architecture

## Overview

The Harness Callbacks system enables agents running in **standalone mode** (as separate gRPC services) to access the full capabilities of the Gibson orchestrator's harness. This allows remote agents to perform LLM completions, execute tools, submit findings, and access all other harness operations via gRPC callbacks.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Gibson Orchestrator                       │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         HarnessCallbackService (gRPC)                 │   │
│  │  - Exposes harness operations via RPC                │   │
│  │  - Manages harness registry by task ID               │   │
│  └──────────────────────────────────────────────────────┘   │
│                          ▲                                   │
│                          │ gRPC                              │
│                          │                                   │
└──────────────────────────┼───────────────────────────────────┘
                           │
                           │ Network Boundary
                           │
┌──────────────────────────┼───────────────────────────────────┐
│                          │                                   │
│                    Standalone Agent                          │
│                                                              │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │         CallbackClient (gRPC Client)                  │   │
│  │  - Connects to orchestrator's callback service       │   │
│  │  - Sends task context with each RPC                  │   │
│  └──────────────────────────────────────────────────────┘   │
│                          │                                   │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │         CallbackHarness (Harness Implementation)      │   │
│  │  - Implements agent.Harness interface                │   │
│  │  - Forwards all operations to CallbackClient         │   │
│  └──────────────────────────────────────────────────────┘   │
│                          │                                   │
│                          ▼                                   │
│                   Agent Execute Function                     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. HarnessCallbackService (Gibson)

**Location**: `opensource/gibson/internal/harness/callback_service.go`

The gRPC service that runs within the Gibson orchestrator. It:

- Exposes all harness operations as gRPC RPCs
- Maintains a registry of active harnesses indexed by task ID
- Delegates operations to the appropriate harness instance

**Registration Pattern**:

```go
service := harness.NewHarnessCallbackService(logger)

// Before executing agent task
service.RegisterHarness(taskID, agentHarness)
defer service.UnregisterHarness(taskID)

// Execute agent...
```

### 2. CallbackClient (SDK)

**Location**: `opensource/sdk/serve/callback_client.go`

The gRPC client that standalone agents use to connect to the orchestrator. It:

- Manages connection lifecycle to the orchestrator
- Tracks context information (task ID, agent name, trace ID)
- Provides typed methods for all harness operations
- Handles proto message serialization/deserialization

**Usage Pattern**:

```go
client, err := serve.NewCallbackClient("localhost:50052")
if err != nil {
    return err
}
defer client.Close()

// Connect to orchestrator
if err := client.Connect(ctx); err != nil {
    return err
}

// Set task context
client.SetContext(taskID, agentName, traceID, spanID)

// Make RPC calls
resp, err := client.LLMComplete(ctx, "primary", messages)
```

### 3. CallbackHarness (SDK)

**Location**: `opensource/sdk/serve/callback_harness.go`

Implements the `agent.Harness` interface by forwarding all operations to the CallbackClient. It:

- Provides the standard harness interface to agents
- Maintains local state (logger, tracer, mission context)
- Implements caching for list operations
- Provides callback-based memory store

**Usage Pattern**:

```go
harness := serve.NewCallbackHarness(
    client,
    logger,
    tracer,
    missionContext,
    targetInfo,
)

// Use like any harness
resp, err := harness.Complete(ctx, "primary", messages)
tools := harness.ListTools(ctx)
```

### 4. CallbackServer (Gibson Helper)

**Location**: `opensource/gibson/internal/harness/callback_server.go`

A convenience wrapper that simplifies starting the callback service. It:

- Creates and manages a gRPC server
- Registers the HarnessCallbackService
- Handles graceful shutdown
- Provides health checking and reflection

**Usage Pattern**:

```go
server := harness.NewCallbackServer(logger, 50052)

go func() {
    if err := server.Start(ctx); err != nil {
        log.Fatalf("Callback server failed: %v", err)
    }
}()

// Register harnesses as needed
server.RegisterHarness(taskID, harness)
defer server.UnregisterHarness(taskID)
```

## Configuration

### Agent Configuration

Agents configure callback support via serve options:

```go
err := serve.Agent(myAgent,
    serve.WithPort(50051),
    serve.WithOrchestratorEndpoint("localhost:50052"),  // Enable callbacks
    serve.WithOrchestratorTLS(tlsConfig),               // Optional TLS
    serve.WithOrchestratorToken("secret-token"),        // Optional auth
)
```

**Environment Variables**:

- `GIBSON_ORCHESTRATOR_ENDPOINT`: Orchestrator callback service address
- `GIBSON_ORCHESTRATOR_TOKEN`: Authentication token for callbacks

### Orchestrator Configuration

The orchestrator needs to:

1. **Start the callback server**:
   ```go
   callbackServer := harness.NewCallbackServer(logger, 50052)
   go callbackServer.Start(ctx)
   ```

2. **Register harnesses before agent execution**:
   ```go
   // Generate unique task ID
   taskID := uuid.New().String()

   // Register harness for this execution
   callbackServer.RegisterHarness(taskID, agentHarness)
   defer callbackServer.UnregisterHarness(taskID)

   // Execute agent with task ID
   result, err := agentClient.Execute(ctx, &proto.AgentExecuteRequest{
       TaskJson: taskJSON,
       // Task JSON should include the task ID
   })
   ```

## Supported Operations

### LLM Operations

- **Complete**: Synchronous completion
- **CompleteWithTools**: Tool-enabled completion
- **Stream**: Streaming completion

### Tool Operations

- **CallTool**: Execute a tool
- **ListTools**: List available tools

### Plugin Operations

- **QueryPlugin**: Query a plugin method
- **ListPlugins**: List available plugins

### Agent Operations

- **DelegateToAgent**: Delegate to another agent
- **ListAgents**: List available agents

### Finding Operations

- **SubmitFinding**: Submit a security finding
- **GetFindings**: Retrieve findings with filters

### Memory Operations

- **Get**: Get value from working memory
- **Set**: Set value in working memory
- **Delete**: Delete from working memory
- **List**: List keys in working memory

### GraphRAG Operations (Optional)

- **QueryGraphRAG**: Semantic graph queries
- **FindSimilarAttacks**: Find attack patterns
- **FindSimilarFindings**: Find related findings
- **GetAttackChains**: Get attack chain sequences
- **StoreGraphNode**: Store knowledge graph nodes
- **CreateGraphRelationship**: Create relationships
- **TraverseGraph**: Graph traversal

## Complete Example

### Gibson Orchestrator Side

```go
package main

import (
    "context"
    "log/slog"

    "github.com/google/uuid"
    "github.com/zero-day-ai/gibson/internal/harness"
)

func main() {
    logger := slog.Default()

    // Create and start callback server
    callbackServer := harness.NewCallbackServer(logger, 50052)

    ctx := context.Background()
    go func() {
        if err := callbackServer.Start(ctx); err != nil {
            logger.Error("callback server failed", "error", err)
        }
    }()

    // When executing an agent task...
    taskID := uuid.New().String()

    // Create harness for this task
    agentHarness := createAgentHarness(ctx, missionID, taskID)

    // Register harness
    callbackServer.RegisterHarness(taskID, agentHarness)
    defer callbackServer.UnregisterHarness(taskID)

    // Execute agent (agent will callback via gRPC)
    result, err := executeAgent(ctx, agentName, task)
    if err != nil {
        logger.Error("agent execution failed", "error", err)
    }
}
```

### Agent Side

```go
package main

import (
    "context"
    "log"

    sdk "github.com/zero-day-ai/sdk"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/llm"
    "github.com/zero-day-ai/sdk/serve"
)

func main() {
    myAgent, _ := sdk.NewAgent(
        sdk.WithName("my-agent"),
        sdk.WithVersion("1.0.0"),
        sdk.WithExecuteFunc(executeAgent),
    )

    // Serve agent with callback support
    err := serve.Agent(myAgent,
        serve.WithPort(50051),
        serve.WithOrchestratorEndpoint("localhost:50052"),  // Enable callbacks
    )
    if err != nil {
        log.Fatal(err)
    }
}

func executeAgent(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    // Harness nil check is critical!
    if harness == nil {
        return agent.Result{
            TaskID:       task.ID,
            Status:       agent.StatusFailed,
            ErrorMessage: "harness required - ensure orchestrator is configured",
        }, fmt.Errorf("harness required")
    }

    // Use harness normally
    logger := harness.Logger()
    logger.Info("executing task", "task_id", task.ID)

    // LLM completion via callback
    messages := []llm.Message{
        {Role: llm.RoleUser, Content: "Analyze this target"},
    }
    resp, err := harness.Complete(ctx, "primary", messages)
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    // Tool execution via callback
    output, err := harness.CallTool(ctx, "nmap", map[string]any{
        "target": harness.Target().URL,
    })

    // Submit findings via callback
    harness.SubmitFinding(ctx, agent.Finding{
        Title:      "Open Port Found",
        Severity:   agent.SeverityMedium,
        Confidence: agent.ConfidenceHigh,
    })

    return agent.NewSuccessResult(map[string]any{
        "completion": resp.Message.Content,
        "scan":       output,
    }), nil
}
```

## Security Considerations

### 1. Authentication

The callback service should require authentication:

```go
// Orchestrator side
callbackServer := harness.NewCallbackServer(logger, 50052)
// TODO: Add TLS and token verification

// Agent side
serve.Agent(myAgent,
    serve.WithOrchestratorEndpoint("localhost:50052"),
    serve.WithOrchestratorToken(os.Getenv("ORCHESTRATOR_TOKEN")),
)
```

### 2. TLS Encryption

Production deployments should use TLS:

```go
// Load TLS config
tlsConfig, err := loadTLSConfig()

// Agent side
serve.Agent(myAgent,
    serve.WithOrchestratorEndpoint("orchestrator.example.com:50052"),
    serve.WithOrchestratorTLS(tlsConfig),
)
```

### 3. Task Isolation

- Each task gets its own harness instance
- Harnesses are unregistered after task completion
- Task IDs must be unique and unpredictable (use UUIDs)

### 4. Network Security

- Callback service should run on a private network
- Use firewalls to restrict access
- Consider mutual TLS for agent authentication

## Error Handling

### Connection Failures

```go
client, err := serve.NewCallbackClient(endpoint)
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}

if err := client.Connect(ctx); err != nil {
    return fmt.Errorf("failed to connect to orchestrator: %w", err)
}
```

### Operation Failures

All harness operations can fail. Agents must handle errors gracefully:

```go
resp, err := harness.Complete(ctx, "primary", messages)
if err != nil {
    logger.Error("LLM completion failed", "error", err)
    return agent.NewFailedResult(err), err
}
```

### Timeout Handling

Operations respect context deadlines:

```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

resp, err := harness.Complete(ctx, "primary", messages)
if err == context.DeadlineExceeded {
    logger.Warn("operation timed out")
}
```

## Performance Considerations

### 1. Network Latency

Each harness operation becomes a network RPC. Consider:

- Batch operations when possible
- Cache list results (CallbackHarness does this automatically)
- Use streaming for large responses

### 2. Connection Pooling

The CallbackClient maintains a single gRPC connection. gRPC handles multiplexing automatically.

### 3. Memory Efficiency

- Working memory operations are synchronous RPCs
- Large data should be stored in mission or long-term memory
- Consider compression for large payloads

## Testing

### Unit Tests

Test individual components with mocks:

```go
// Test CallbackClient
func TestCallbackClient(t *testing.T) {
    client, err := NewCallbackClient("localhost:50051")
    require.NoError(t, err)

    client.SetContext("task-123", "agent-1", "", "")
    // Test methods...
}
```

### Integration Tests

Test the full callback flow:

```go
// Start mock server
server, mockSvc, addr := startMockServer(t)
defer server.Stop()

// Create client and harness
client, _ := NewCallbackClient(addr)
client.Connect(ctx)
harness := NewCallbackHarness(client, logger, tracer, mission, target)

// Test operations
resp, err := harness.Complete(ctx, "primary", messages)
```

See:
- `opensource/sdk/serve/callback_client_test.go`
- `opensource/sdk/serve/callback_harness_test.go`
- `opensource/sdk/serve/callback_integration_test.go`
- `opensource/gibson/internal/harness/callback_service_test.go`

## Troubleshooting

### Agent receives nil harness

**Symptom**: Agent crashes with nil pointer dereference

**Cause**: No orchestrator endpoint configured

**Solution**:
1. Add nil check in agent code (see Whistler example)
2. Configure orchestrator endpoint via `WithOrchestratorEndpoint()`

### Connection refused errors

**Symptom**: `failed to connect to orchestrator: connection refused`

**Cause**: Callback service not running or wrong address

**Solution**:
1. Verify callback server is running: `netstat -an | grep 50052`
2. Check endpoint configuration matches server address
3. Verify firewall rules

### Task not found errors

**Symptom**: `no active harness for task: task-xyz`

**Cause**: Harness not registered or wrong task ID

**Solution**:
1. Ensure harness is registered before agent execution
2. Verify task ID matches between registration and agent call
3. Check harness wasn't unregistered prematurely

### Authentication failures

**Symptom**: `Unauthorized` or `permission denied` errors

**Cause**: Missing or invalid authentication token

**Solution**:
1. Set `GIBSON_ORCHESTRATOR_TOKEN` environment variable
2. Use `WithOrchestratorToken()` option
3. Verify token matches on both sides

## Migration Guide

### From Embedded Mode

If your agent currently expects a harness:

```go
// Before (embedded mode)
func execute(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    resp, _ := harness.Complete(ctx, "primary", messages)
    // ...
}
```

No changes needed! Just:

1. Configure orchestrator endpoint when serving:
   ```go
   serve.Agent(myAgent, serve.WithOrchestratorEndpoint("localhost:50052"))
   ```

2. Add nil check for backward compatibility:
   ```go
   if harness == nil {
       return agent.NewFailedResult(fmt.Errorf("harness required")), nil
   }
   ```

### To Standalone Mode

To enable your agent for standalone deployment:

1. **Add serve configuration**:
   ```go
   serve.Agent(myAgent,
       serve.WithPort(50051),
       serve.WithOrchestratorEndpoint(os.Getenv("ORCHESTRATOR_ENDPOINT")),
   )
   ```

2. **Add harness nil check** (critical!):
   ```go
   if harness == nil {
       return agent.Result{
           TaskID:       task.ID,
           Status:       agent.StatusFailed,
           ErrorMessage: "harness required - ensure orchestrator is configured",
       }, fmt.Errorf("harness required")
   }
   ```

3. **Configure orchestrator** to start callback service and register harnesses

## References

- **Specification**: `.spec-workflow/specs/sdk-harness-full-implementation/`
- **Implementation Status**: `opensource/sdk/HARNESS_CALLBACK_IMPLEMENTATION_STATUS.md`
- **Example Agent**: `enterprise/agents/whistler/main.go`
- **Proto Definitions**: `opensource/sdk/api/proto/harness_callback.proto`
