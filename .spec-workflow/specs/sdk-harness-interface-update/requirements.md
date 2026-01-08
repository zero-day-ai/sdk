# Requirements: SDK Harness Interface Update

## Overview

Complete the SDK Harness interface implementation to enable full E2E testing of Gibson agents. The SDK has new interface methods that need to be implemented across all harness types and mock implementations to allow agents to compile and execute properly.

## Problem Statement

The Gibson SDK's `agent.Harness` interface was extended with new methods for mission execution context and scoped GraphRAG queries. These methods are not fully implemented across:
1. `CallbackHarness` (serve package) - PARTIALLY DONE
2. `LocalHarness` (serve package) - PARTIALLY DONE
3. `RecordingHarness` (eval package) - PARTIALLY DONE
4. Mock harnesses in test files - NOT DONE
5. All enterprise agents need to be rebuilt against the updated SDK

## User Stories

### US-1: SDK Compiles Successfully
**As a** developer
**I want** the SDK to compile without errors
**So that** I can build agents against it

**Acceptance Criteria:**
- `go build ./...` completes with no errors
- `go test ./...` completes with no test build failures

### US-2: All Tests Pass
**As a** developer
**I want** all SDK tests to pass
**So that** I know the implementation is correct

**Acceptance Criteria:**
- All mock harnesses implement the full `agent.Harness` interface
- All mock memory stores implement the full `memory.MissionMemory` interface
- Test suite runs green

### US-3: SDK Published to GitHub
**As a** Gibson user
**I want** the latest SDK on GitHub
**So that** agents can depend on the updated version

**Acceptance Criteria:**
- All changes committed with proper message
- Changes pushed to GitHub
- SDK version updated if needed

### US-4: All Agents Rebuilt
**As a** Gibson operator
**I want** all agents rebuilt against latest SDK
**So that** they work with the orchestrator

**Acceptance Criteria:**
- Enterprise agents (whistler, crease, carl, bishop, k8skiller) rebuilt
- OSS tools rebuilt
- All agents installed to `~/.gibson/agents/bin`

### US-5: E2E Test Succeeds
**As a** Gibson user
**I want** to run a full E2E demo mission
**So that** I can start hacking stuff

**Acceptance Criteria:**
- Demo mission executes without agent failures
- Agents make LLM calls (visible in Langfuse)
- Mission completes or produces meaningful output

## Functional Requirements

### FR-1: Interface Implementation
The following methods must be implemented on all harness types:
- `MissionExecutionContext() types.MissionExecutionContext`
- `GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error)`
- `GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)`
- `GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)`
- `QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error)`

### FR-2: Memory Interface Implementation
The `MissionMemory` interface requires:
- `GetPreviousRunValue(ctx context.Context, key string) (any, error)`
- `GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error)`
- `ContinuityMode() memory.MemoryContinuityMode`

### FR-3: Planning Interface Implementation
The harness must implement:
- `PlanContext() planning.PlanningContext`
- `ReportStepHints(ctx context.Context, hints *planning.StepHints) error`

## Non-Functional Requirements

### NFR-1: Backwards Compatibility
- Stub implementations return empty slices/default values (not errors) where appropriate
- Existing agent code continues to work without modification

### NFR-2: Build Time
- Full rebuild of all agents should complete in under 5 minutes

## Out of Scope
- Proto/gRPC updates for callback implementations (stubbed for now)
- Full implementation of mission run history (returns empty)
- Full implementation of cross-run findings (returns empty)

## Dependencies
- Go 1.21+
- Access to enterprise agents repo
- Access to gibson-oss-tools repo
- GitHub push access

## Success Metrics
- Zero build errors
- Zero test failures
- Successful E2E mission execution with LLM calls visible in Langfuse
