# Harness Callback Implementation Status

## Overview

This document tracks the implementation status of the SDK Harness Full Implementation spec for the Gibson security testing framework. The goal is to replace the `nil` harness passed to agents in standalone mode with a fully functional CallbackHarness that forwards all operations to the orchestrator via gRPC.

## Completed Tasks (SDK Side)

### Phase 1: Proto Definitions ✅
- **Task 1**: Created `<sdk-root>/api/proto/harness_callback.proto`
  - Defined HarnessCallbackService with 34 RPC methods
  - Includes all LLM, Tool, Plugin, Agent, Finding, Memory, and GraphRAG operations
  - Uses unique message names to avoid conflicts with existing protos
  - Generated Go code successfully

- **Task 2**: Generated Go code from proto
  - Files created: `harness_callback.pb.go` and `harness_callback_grpc.pb.go`
  - Location: `<sdk-root>/api/gen/proto/`

### Phase 2: SDK Client ✅
- **Task 3**: Created `callback_client.go`
  - Full gRPC client with connection management
  - Thread-safe context tracking (task_id, agent_name, trace_id, span_id)
  - All 34 RPC wrapper methods implemented
  - TLS and authentication token support
  - Proper error handling and connection lifecycle

- **Task 4**: Created `callback_memory.go`
  - CallbackMemoryStore implementing agent.MemoryStore interface
  - Forwards Get/Set/Delete/List operations to orchestrator
  - JSON serialization for any-typed values

- **Task 5**: Created `callback_token_tracker.go`
  - Thread-safe token tracker implementing llm.TokenTracker interface
  - Tracks usage by slot and total across all slots
  - Thread-safe Add/Total/BySlot/Reset/Slots/HasSlot methods

### Phase 3: Harness Implementation ✅
- **Tasks 6-12**: Created `callback_harness.go` (comprehensive ~850 lines)
  - Core harness methods: Logger(), Tracer(), TokenUsage(), Mission(), Target(), Memory()
  - LLM operations: Complete(), CompleteWithTools(), Stream()
  - Tool operations: CallTool(), ListTools() with caching
  - Plugin operations: QueryPlugin(), ListPlugins() with caching
  - Agent delegation: DelegateToAgent(), ListAgents() with caching
  - Finding operations: SubmitFinding(), GetFindings()
  - GraphRAG queries: QueryGraphRAG(), FindSimilarAttacks(), FindSimilarFindings(), GetAttackChains(), GetRelatedFindings()
  - GraphRAG storage: StoreGraphNode(), CreateGraphRelationship(), StoreGraphBatch(), TraverseGraph(), GraphRAGHealth()
  - Helper methods for proto <-> SDK type conversions

### Phase 4: Integration ✅
- **Task 13**: Modified `options.go`
  - Added WithOrchestratorEndpoint(endpoint string)
  - Added WithOrchestratorTLS(conf *tls.Config)
  - Added WithOrchestratorToken(token string)

- **Task 14**: Modified `agent.go` and `serve.go`
  - Added orchestrator fields to Config struct
  - Modified agentServiceServer to store Config
  - Updated Execute() to create CallbackClient and CallbackHarness when orchestrator is configured
  - Falls back to nil harness if orchestrator not configured (backward compatible)

## Remaining Tasks (Gibson Side & Testing)

### Phase 5: Orchestrator Service ⏳
- **Task 15**: Create `<gibson-root>/internal/harness/callback_service.go`
  - Implement HarnessCallbackService server
  - Bridge callback requests to Gibson's real harness implementations
  - Handle context extraction and validation
  - Implement all 34 RPC methods

- **Task 16**: Register HarnessCallbackService in Gibson core
  - Find appropriate location in Gibson's server setup
  - Register the service with the gRPC server
  - Ensure service is available before agent execution

### Phase 6: Testing ⏳
- **Task 17**: Create `callback_client_test.go`
  - Test connection lifecycle
  - Test all RPC wrapper methods
  - Test error handling and reconnection

- **Task 18**: Create `callback_harness_test.go`
  - Test all harness interface methods
  - Test proto conversions
  - Test caching behavior

- **Task 19**: Create Gibson `callback_service_test.go`
  - Test service implementation
  - Test context propagation
  - Test error scenarios

- **Task 20**: Create `callback_integration_test.go`
  - End-to-end test with real agent + orchestrator
  - Test all harness operations through the callback chain
  - Test streaming operations
  - Test concurrent access

### Phase 7: Cleanup ⏳
- **Task 21**: Update `<project-root>/enterprise/gibson-ent-agents/whistler/main.go`
  - Add nil checks for harness before use
  - Provide graceful fallback behavior

- **Task 22**: Create documentation at `opensource/sdk/docs/harness_callbacks.md`
  - Architecture overview with diagrams
  - Usage examples for agent developers
  - Configuration guide for orchestrator setup
  - Troubleshooting guide

## Key Implementation Details

### Callback Architecture
```
Agent Process                        Orchestrator Process
┌─────────────────┐                 ┌──────────────────┐
│  Agent Execute  │                 │ HarnessCallback  │
│       ↓         │                 │    Service       │
│ CallbackHarness │ ←─ gRPC ───────→ │                  │
│       ↓         │   RPC calls     │        ↓         │
│  Complete()     │                 │  Real LLM        │
│  CallTool()     │                 │  Real Tools      │
│  QueryPlugin()  │                 │  Real Plugins    │
│  SubmitFinding()│                 │  Real Finding DB │
│  Memory()       │                 │  Real Memory     │
│  GraphRAG()     │                 │  Real GraphRAG   │
└─────────────────┘                 └──────────────────┘
```

### Context Propagation
Every RPC request includes:
- `task_id`: Unique identifier for the task execution
- `agent_name`: Name of the calling agent
- `trace_id`: OpenTelemetry trace ID for distributed tracing
- `span_id`: OpenTelemetry span ID for the operation

### Proto Message Design
- All descriptor types renamed to avoid conflicts (e.g., `HarnessAgentDescriptor`)
- JSON serialization used for generic types (map[string]any, any)
- Error handling via `gibson.common.Error` message
- Token usage tracked in all LLM operations

### Caching Strategy
List operations are cached per task execution:
- `ListTools()` - cached after first call
- `ListPlugins()` - cached after first call
- `ListAgents()` - cached after first call

This reduces round-trips to the orchestrator for repeated discovery operations.

### Thread Safety
- CallbackClient: Thread-safe context tracking with sync.RWMutex
- CallbackTokenTracker: Thread-safe token accumulation with sync.RWMutex
- CallbackHarness: Thread-safe cache access with sync.RWMutex

## Files Created

### SDK Package (`opensource/sdk`)
```
api/proto/harness_callback.proto          (proto definitions - 449 lines)
api/gen/proto/harness_callback.pb.go      (generated)
api/gen/proto/harness_callback_grpc.pb.go (generated)
serve/callback_client.go                  (gRPC client - 476 lines)
serve/callback_memory.go                  (memory store - 95 lines)
serve/callback_token_tracker.go           (token tracker - 96 lines)
serve/callback_harness.go                 (main harness - 858 lines)
```

### Modified Files
```
serve/options.go     (added 3 orchestrator option functions)
serve/serve.go       (added 4 orchestrator fields to Config)
serve/agent.go       (modified Execute to create callback harness)
```

## Usage Example

```go
// In agent main.go
func main() {
    agent := &MyAgent{}

    // Configure with orchestrator endpoint
    err := serve.Agent(agent,
        serve.WithPort(50051),
        serve.WithOrchestratorEndpoint("localhost:50052"),
        serve.WithOrchestratorToken("secret-token"),
    )
    if err != nil {
        log.Fatal(err)
    }
}
```

## Next Steps

1. **Implement Gibson callback_service.go** - This is the most critical remaining piece
2. **Register service in Gibson** - Ensure service is available when agents connect
3. **Write tests** - Comprehensive testing of the callback chain
4. **Update whistler** - Add nil checks as reference implementation
5. **Write documentation** - Complete docs for agent developers

## Performance Considerations

- Each harness operation incurs a gRPC round-trip
- Caching reduces repeated list operations
- Streaming operations use server-side streaming for efficiency
- Connection pooling handled by gRPC client
- Consider connection reuse across multiple task executions

## Security Considerations

- TLS configuration supported via WithOrchestratorTLS
- Bearer token authentication via WithOrchestratorToken
- Context includes agent_name to verify caller identity
- All operations go through orchestrator for audit trail

## Backward Compatibility

The implementation is fully backward compatible:
- If no orchestrator endpoint is configured, agents receive nil harness (existing behavior)
- Agents can check for nil and handle gracefully
- No breaking changes to existing agent implementations
