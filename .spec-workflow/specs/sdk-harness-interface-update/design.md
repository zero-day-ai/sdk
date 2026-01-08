# Design: SDK Harness Interface Update

## Overview

This design addresses the implementation of missing `agent.Harness` interface methods across all harness types in the SDK, updating test mocks, pushing the SDK, rebuilding all agents, and running E2E tests.

## Architecture

### Current State

The SDK's `agent.Harness` interface (in `agent/harness.go`) has been extended with 5 new methods:

```go
// Mission Execution Context Methods
MissionExecutionContext() types.MissionExecutionContext
GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error)
GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)
GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)
QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error)
```

The `memory.MissionMemory` interface has 3 new methods:

```go
GetPreviousRunValue(ctx context.Context, key string) (any, error)
GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error)
ContinuityMode() memory.MemoryContinuityMode
```

### Implementation Status

| Component | Harness Methods | Memory Methods | Status |
|-----------|-----------------|----------------|--------|
| `serve/callback_harness.go` | ✅ Done | N/A | Complete |
| `serve/callback_memory.go` | N/A | ✅ Done | Complete |
| `serve/local_harness.go` | ✅ Done | ✅ Done (stub) | Complete |
| `eval/recording_harness.go` | ✅ Done | N/A | Complete |
| `agent/harness_test.go` mocks | ❌ Missing | ✅ Done | Partial |
| `agent/streaming_test.go` mocks | ❌ Missing | N/A | Missing |
| `eval/feedback_harness_test.go` mocks | ❌ Missing | N/A | Missing |
| `integration/agent_test.go` mocks | ❌ Missing | N/A | Missing |
| `serve/streaming_harness_test.go` mocks | ❌ Missing | N/A | Missing |

## Design Decisions

### D1: Stub Implementations Return Empty/Default Values

For methods that require orchestrator callback support not yet implemented:
- `GetMissionRunHistory` → returns empty `[]types.MissionRunSummary{}`
- `GetPreviousRunFindings` → returns empty `[]*finding.Finding{}`
- `GetAllRunFindings` → returns empty `[]*finding.Finding{}`

This allows agents to compile and run without breaking, while full functionality waits for proto updates.

### D2: Mock Harnesses Embed Base Mock

Test mocks like `mockStreamingHarness` embed `mockHarness`, so adding methods to `mockHarness` propagates to all embedders.

### D3: Planning Methods Already Implemented

`PlanContext()` and `ReportStepHints()` are already implemented on all harness types.

## Component Design

### Test Mock Updates

Each test file has mock harness implementations that need the 5 new methods:

**Pattern for all mocks:**
```go
func (m *mockHarness) MissionExecutionContext() types.MissionExecutionContext {
    return types.MissionExecutionContext{}
}

func (m *mockHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
    return []types.MissionRunSummary{}, nil
}

func (m *mockHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
    return []*finding.Finding{}, nil
}

func (m *mockHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
    return []*finding.Finding{}, nil
}

func (m *mockHarness) QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error) {
    return nil, nil
}
```

### Files to Modify

1. **`agent/harness_test.go`** - Add 5 methods to `mockHarness`
2. **`agent/streaming_test.go`** - Methods inherited from embedded `mockHarness`
3. **`eval/feedback_harness_test.go`** - Add 5 methods to `mockHarness`
4. **`eval/recording_harness_test.go`** - Add 5 methods to `mockHarness` (if exists)
5. **`integration/agent_test.go`** - Add 5 methods to `mockHarness`
6. **`serve/streaming_harness_test.go`** - Add 5 methods to `mockStreamHarness`

### Agent Rebuild Process

1. Push SDK changes to GitHub
2. For each agent repo:
   - `go get github.com/zero-day-ai/sdk@latest`
   - `go mod tidy`
   - `go build`
   - Install to `~/.gibson/agents/bin`

### Repos to Update

| Repo | Agents |
|------|--------|
| `gibson-enterprise-agents` | whistler, crease, carl, bishop, k8skiller |
| `gibson-oss-tools` | Various OSS tools |

## Data Flow

```
SDK Push → Agent go.mod update → Agent rebuild → Agent install → E2E test
```

## Error Handling

- Mock implementations return `nil` errors
- Stub implementations log debug messages when called
- Build failures abort the process

## Testing Strategy

1. `go build ./...` - Verify SDK compiles
2. `go test ./...` - Verify all tests pass
3. Rebuild agents - Verify they compile
4. E2E mission - Verify runtime works

## Dependencies

- SDK: `github.com/zero-day-ai/sdk`
- Enterprise agents: Local at `/home/anthony/Code/zero-day.ai/closed/agents`
- OSS tools: Local at `/home/anthony/Code/zero-day.ai/opensource/gibson-oss-tools`

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Proto not updated | Stub implementations return empty values |
| Agent build fails | Fix issues incrementally |
| E2E test fails | Debug agent execution logs |

## Success Criteria

1. `go build ./...` succeeds in SDK
2. `go test ./...` passes all tests in SDK
3. All 5 enterprise agents rebuild successfully
4. OSS tools rebuild successfully
5. E2E demo mission runs with LLM calls visible in Langfuse
