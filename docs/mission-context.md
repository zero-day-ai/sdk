# Mission Execution Context

This guide explains how agents can access rich mission context beyond the basic `Mission()` method, including run history, memory continuity, and GraphRAG query scoping for resumable and multi-run missions.

## Table of Contents

- [Overview](#overview)
- [Getting Full Mission Context](#getting-full-mission-context)
- [Checking Run History](#checking-run-history)
- [GraphRAG Query Scoping](#graphrag-query-scoping)
- [Memory Continuity](#memory-continuity)
- [CLI Commands](#cli-commands)
- [Best Practices](#best-practices)

---

## Overview

Gibson missions support multiple execution runs with sophisticated context tracking, memory continuity, and query scoping capabilities. This enables:

- **Resumable missions**: Resume from specific workflow nodes after failures
- **Run continuity**: Access previous run results and build on prior work
- **Scoped queries**: Control which runs are included in GraphRAG searches
- **Memory modes**: Configure how memory state persists across runs

### Execution Context Hierarchy

```
Mission
  └─ Run 1 (initial execution)
       ├─ Findings: 3
       ├─ Status: completed
       └─ Memory: isolated
  └─ Run 2 (resumed from node X)
       ├─ Findings: 5
       ├─ Status: in_progress
       ├─ Previous Run: Run 1
       └─ Memory: inherit (read Run 1)
  └─ Run 3 (retry with refinements)
       ├─ Findings: 8
       ├─ Status: completed
       ├─ Previous Run: Run 2
       └─ Memory: shared (all runs)
```

---

## Getting Full Mission Context

### Basic Mission Context

The traditional `Mission()` method returns basic mission information:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    // Basic mission context
    mission := harness.Mission()

    logger := harness.Logger()
    logger.Info("mission info",
        "id", mission.ID,
        "name", mission.Name,
        "phase", mission.Phase,
    )

    // Check constraints
    if mission.Constraints.MaxFindings > 0 {
        logger.Info("findings limit",
            "max", mission.Constraints.MaxFindings,
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Extended Execution Context

Access full execution context including run tracking and resume information:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    // Get full execution context
    execCtx := harness.MissionExecutionContext()

    logger := harness.Logger()
    logger.Info("execution context",
        "mission_id", execCtx.MissionID,
        "mission_name", execCtx.MissionName,
        "run_number", execCtx.RunNumber,
        "is_resumed", execCtx.IsResumed,
    )

    // Check if this is a resumed run
    if execCtx.IsResumed {
        logger.Info("resumed execution",
            "from_node", execCtx.ResumedFromNode,
            "previous_run_id", execCtx.PreviousRunID,
            "previous_status", execCtx.PreviousRunStatus,
        )

        // Adjust behavior for resumed runs
        // For example, skip already-completed phases
    }

    // Check run position
    if execCtx.IsFirstRun() {
        logger.Info("first run of mission")
        // Initialize fresh state
    } else {
        logger.Info("continuation run",
            "previous_run", execCtx.PreviousRunID,
            "total_findings_so_far", execCtx.TotalFindingsAllRuns,
        )
        // Build on previous runs
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Context Properties

**MissionExecutionContext fields:**

| Field | Type | Description |
|-------|------|-------------|
| `MissionID` | `string` | Unique mission identifier |
| `MissionName` | `string` | Human-readable mission name |
| `RunNumber` | `int` | Sequential run number (1-based) |
| `IsResumed` | `bool` | True if resumed from prior run |
| `ResumedFromNode` | `string` | Workflow node where execution resumed |
| `PreviousRunID` | `string` | ID of the prior run (if any) |
| `PreviousRunStatus` | `string` | Status of prior run (completed, failed, etc.) |
| `TotalFindingsAllRuns` | `int` | Cumulative findings across all runs |
| `MemoryContinuity` | `string` | Memory mode (isolated/inherit/shared) |
| `Constraints` | `MissionConstraints` | Mission execution constraints |

**Helper methods:**

```go
// Check if there's a previous run
if execCtx.HasPreviousRun() {
    // Access previous run information
    prevRunID := execCtx.PreviousRunID
}

// Check if this is the first run
if execCtx.IsFirstRun() {
    // Perform first-run initialization
}
```

---

## Checking Run History

Access historical information about previous mission runs to understand patterns, success rates, and accumulated findings.

### Querying Run History

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    execCtx := harness.MissionExecutionContext()
    logger := harness.Logger()

    // Get run history for this mission
    history, err := harness.GetMissionRunHistory(ctx)
    if err != nil {
        logger.Error("failed to get run history", "error", err)
        return agent.NewFailedResult(err), err
    }

    logger.Info("mission run history",
        "total_runs", len(history),
        "current_run", execCtx.RunNumber,
    )

    // Analyze previous runs
    for _, run := range history {
        logger.Info("previous run",
            "run_number", run.RunNumber,
            "status", run.Status,
            "findings", run.FindingsCount,
            "created_at", run.CreatedAt,
        )

        // Check completion status
        if run.CompletedAt != nil {
            duration := run.CompletedAt.Sub(run.CreatedAt)
            logger.Info("run completed",
                "duration", duration,
            )
        }
    }

    // Calculate success metrics
    successCount := 0
    totalFindings := 0
    for _, run := range history {
        if run.Status == "completed" {
            successCount++
        }
        totalFindings += run.FindingsCount
    }

    logger.Info("mission metrics",
        "success_rate", float64(successCount)/float64(len(history)),
        "total_findings", totalFindings,
        "avg_findings_per_run", float64(totalFindings)/float64(len(history)),
    )

    return agent.NewSuccessResult("completed"), nil
}
```

### Run Summary Structure

**MissionRunSummary fields:**

| Field | Type | Description |
|-------|------|-------------|
| `MissionID` | `string` | Mission identifier |
| `RunNumber` | `int` | Sequential run number |
| `Status` | `string` | Run status (completed, failed, in_progress, etc.) |
| `FindingsCount` | `int` | Number of findings discovered in this run |
| `CreatedAt` | `time.Time` | When the run was created |
| `CompletedAt` | `*time.Time` | When the run completed (nil if still running) |

### Adaptive Behavior Based on History

```go
func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    history, err := harness.GetMissionRunHistory(ctx)
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    logger := harness.Logger()

    // Adapt strategy based on previous runs
    if len(history) > 0 {
        lastRun := history[len(history)-1]

        if lastRun.Status == "failed" {
            logger.Warn("previous run failed, using conservative approach")
            // Increase timeouts, reduce parallelism, etc.
        }

        if lastRun.FindingsCount == 0 {
            logger.Info("previous run found nothing, trying alternative techniques")
            // Switch to different attack patterns
        }

        if lastRun.FindingsCount > 10 {
            logger.Info("previous run very successful, focusing on similar areas")
            // Continue with proven approaches
        }
    }

    return agent.NewSuccessResult("completed"), nil
}
```

---

## GraphRAG Query Scoping

Control which mission runs are included in GraphRAG queries using scope options. This enables precise control over knowledge retrieval in multi-run scenarios.

### Scope Options

Gibson provides three scope levels for GraphRAG queries:

| Scope | Constant | Description | Use Case |
|-------|----------|-------------|----------|
| **Current Run** | `graphrag.ScopeCurrentRun` | Only current run's data | Isolated analysis, fresh perspective |
| **Same Mission** | `graphrag.ScopeSameMission` | All runs of this mission | Build on prior mission work |
| **All Missions** | `graphrag.ScopeAll` | All missions (default) | Cross-mission pattern discovery |

### Querying Current Run Only

Isolate queries to only the current execution run:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/graphrag"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()

    // Query only current run's findings and artifacts
    query := "SQL injection vulnerabilities discovered"
    results, err := harness.QueryGraphRAGScoped(ctx, query, graphrag.ScopeCurrentRun)
    if err != nil {
        logger.Error("graphrag query failed", "error", err)
        return agent.NewFailedResult(err), err
    }

    logger.Info("current run findings",
        "query", query,
        "results", len(results),
    )

    for _, result := range results {
        logger.Info("finding",
            "content", result.Node.Content,
            "score", result.Score,
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Querying Same Mission (All Runs)

Access all runs of the current mission for continuity:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/graphrag"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()

    // Query all runs of this mission
    query := "authentication bypass techniques attempted"
    results, err := harness.QueryGraphRAGScoped(ctx, query, graphrag.ScopeSameMission)
    if err != nil {
        logger.Error("graphrag query failed", "error", err)
        return agent.NewFailedResult(err), err
    }

    logger.Info("mission-wide results",
        "query", query,
        "results", len(results),
    )

    // Analyze patterns across runs
    runNumbers := make(map[int]int)
    for _, result := range results {
        // Extract run number from node metadata if available
        if runNum, ok := result.Node.Properties["run_number"].(int); ok {
            runNumbers[runNum]++
        }
    }

    logger.Info("findings distribution",
        "by_run", runNumbers,
    )

    return agent.NewSuccessResult("completed"), nil
}
```

### Querying All Missions (Default)

Search across all missions for broader pattern discovery:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/graphrag"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()

    // Query across all missions (default scope)
    query := "prompt injection techniques that succeeded"
    results, err := harness.QueryGraphRAGScoped(ctx, query, graphrag.ScopeAll)
    if err != nil {
        logger.Error("graphrag query failed", "error", err)
        return agent.NewFailedResult(err), err
    }

    logger.Info("cross-mission results",
        "query", query,
        "results", len(results),
    )

    // Learn from historical successes across all missions
    for _, result := range results {
        logger.Info("historical success",
            "content", result.Node.Content,
            "score", result.Score,
            "mission", result.Node.MissionID,
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Advanced Query Scoping

Combine scope with other query parameters:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/graphrag"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    execCtx := harness.MissionExecutionContext()
    logger := harness.Logger()

    // Build scoped query with advanced options
    query := graphrag.NewQuery("jailbreak attempts").
        WithMissionScope(graphrag.ScopeSameMission).
        WithMissionName(execCtx.MissionName).
        WithRunNumber(execCtx.RunNumber - 1).  // Previous run only
        WithTopK(20).
        WithMinScore(0.8).
        WithNodeTypes("finding", "artifact").
        WithIncludeRunMetadata(true)

    results, err := harness.GraphRAG().Query(ctx, query)
    if err != nil {
        logger.Error("advanced query failed", "error", err)
        return agent.NewFailedResult(err), err
    }

    logger.Info("scoped query results",
        "scope", "previous_run_only",
        "results", len(results),
    )

    // Access run metadata in results
    for _, result := range results {
        if runNum, ok := result.Node.Properties["run_number"].(int); ok {
            logger.Info("finding from previous run",
                "content", result.Node.Content,
                "run_number", runNum,
                "score", result.Score,
            )
        }
    }

    return agent.NewSuccessResult("completed"), nil
}
```

---

## Memory Continuity

Configure how mission memory behaves across multiple runs. Memory continuity enables agents to build on prior work while maintaining appropriate isolation or sharing boundaries.

### Memory Continuity Modes

| Mode | Behavior | Use Case |
|------|----------|----------|
| **isolated** (default) | Each run has separate memory | Independent parallel runs, clean state testing |
| **inherit** | Read prior run's memory, write to current | Sequential runs building on previous results |
| **shared** | All runs share same memory | Collaborative multi-agent scenarios |

### Isolated Mode (Default)

Each run starts with empty memory state:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()
    mission := harness.Memory().Mission()

    // Check continuity mode
    mode := mission.ContinuityMode()
    logger.Info("memory mode", "continuity", mode)

    if mode == memory.MemoryIsolated {
        logger.Info("isolated mode - starting fresh")

        // All memory operations are scoped to this run only
        err := mission.Set(ctx, "discovered_hosts", []string{"10.0.1.5"}, nil)
        if err != nil {
            return agent.NewFailedResult(err), err
        }
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Inherit Mode

Read previous run's memory in copy-on-write fashion:

```go
import (
    "context"
    "errors"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/memory"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()
    missionMem := harness.Memory().Mission()

    // Only works with inherit or shared mode
    if missionMem.ContinuityMode() == memory.MemoryInherit {
        logger.Info("inherit mode - accessing previous run")

        // Read value from previous run
        prevHosts, err := missionMem.GetPreviousRunValue(ctx, "discovered_hosts")
        if errors.Is(err, memory.ErrNoPreviousRun) {
            logger.Info("first run, no previous data")
            prevHosts = []string{}
        } else if err != nil {
            return agent.NewFailedResult(err), err
        }

        logger.Info("inherited hosts",
            "from_previous_run", prevHosts,
        )

        // Add new discoveries (writes to current run's memory)
        newHosts := append(prevHosts.([]string), "10.0.1.6", "10.0.1.7")
        err = missionMem.Set(ctx, "discovered_hosts", newHosts, nil)
        if err != nil {
            return agent.NewFailedResult(err), err
        }

        logger.Info("updated hosts",
            "current_run_total", len(newHosts),
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Shared Mode

All runs read and write to the same memory namespace:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/memory"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()
    missionMem := harness.Memory().Mission()

    if missionMem.ContinuityMode() == memory.MemoryShared {
        logger.Info("shared mode - coordinating with other runs")

        // Read shared state (visible across all runs)
        item, err := missionMem.Get(ctx, "scan_progress")
        if errors.Is(err, memory.ErrNotFound) {
            // Initialize shared state
            err = missionMem.Set(ctx, "scan_progress", map[string]any{
                "completed_targets": []string{},
                "in_progress": []string{},
            }, nil)
        }

        // Update shared state (immediately visible to other runs)
        progress := item.Value.(map[string]any)
        completed := progress["completed_targets"].([]string)
        completed = append(completed, "target-123")
        progress["completed_targets"] = completed

        err = missionMem.Set(ctx, "scan_progress", progress, nil)
        if err != nil {
            return agent.NewFailedResult(err), err
        }

        logger.Info("updated shared progress",
            "completed_count", len(completed),
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Accessing Value History

Query how a value has changed across runs:

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()
    missionMem := harness.Memory().Mission()

    // Get value history across all runs
    history, err := missionMem.GetValueHistory(ctx, "discovered_hosts")
    if err != nil {
        logger.Error("failed to get history", "error", err)
        return agent.NewFailedResult(err), err
    }

    logger.Info("value history",
        "key", "discovered_hosts",
        "versions", len(history),
    )

    for _, h := range history {
        hosts := h.Value.([]string)
        logger.Info("historical value",
            "run_number", h.RunNumber,
            "stored_at", h.StoredAt,
            "host_count", len(hosts),
            "hosts", hosts,
        )
    }

    // Analyze growth patterns
    if len(history) > 1 {
        firstRun := history[0].Value.([]string)
        lastRun := history[len(history)-1].Value.([]string)

        logger.Info("discovery progress",
            "initial_hosts", len(firstRun),
            "current_hosts", len(lastRun),
            "growth", len(lastRun)-len(firstRun),
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### Memory Continuity Error Handling

```go
import (
    "context"
    "errors"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/memory"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    logger := harness.Logger()
    missionMem := harness.Memory().Mission()

    // Attempt to access previous run value
    prevValue, err := missionMem.GetPreviousRunValue(ctx, "attack_state")

    if errors.Is(err, memory.ErrNoPreviousRun) {
        logger.Info("first run - initializing state")
        // Handle first run case
        prevValue = map[string]any{"phase": "reconnaissance"}

    } else if errors.Is(err, memory.ErrContinuityNotSupported) {
        logger.Warn("continuity not supported in isolated mode")
        // Fall back to isolated behavior
        prevValue = nil

    } else if err != nil {
        logger.Error("failed to get previous value", "error", err)
        return agent.NewFailedResult(err), err
    } else {
        logger.Info("retrieved previous run value",
            "value", prevValue,
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

---

## CLI Commands

### Mission Context

View full mission execution context:

```bash
# Get mission context
gibson mission context <mission-id>

# Example output:
# Mission: penetration-test-2024
# Run Number: 3
# Is Resumed: true
# Resumed From: workflow-node-5
# Previous Run: run-uuid-abc123
# Previous Status: failed
# Total Findings (All Runs): 12
# Memory Continuity: inherit
```

### Scoped Findings

Query findings with mission scope filters:

```bash
# Get findings from current run only
gibson mission findings <mission-name> --scope current_run

# Get findings from all runs of this mission
gibson mission findings <mission-name> --scope same_mission

# Get findings from all missions (default)
gibson mission findings <mission-name> --scope all

# Get findings from specific run number
gibson mission findings <mission-name> --scope same_mission --run 2
```

### Run History

View mission run history:

```bash
# List all runs for a mission
gibson mission runs <mission-name>

# Example output:
# Run 1: completed (5 findings) - 2024-01-05 10:00:00
# Run 2: failed (3 findings) - 2024-01-05 14:30:00
# Run 3: completed (8 findings) - 2024-01-05 16:45:00

# Get details for specific run
gibson mission run <mission-name> --run 2
```

### Memory Continuity

Run missions with different memory modes:

```bash
# Run with isolated memory (default)
gibson mission run workflow.yaml

# Run with inherit mode (read previous run)
gibson mission run workflow.yaml --memory-continuity inherit

# Run with shared mode (all runs share memory)
gibson mission run workflow.yaml --memory-continuity shared

# Resume mission with inherit mode
gibson mission resume <mission-id> --from-node <node-name> --memory-continuity inherit
```

### GraphRAG Queries

Query GraphRAG with scope filters:

```bash
# Query current run only
gibson graphrag query "SQL injection findings" --scope current_run

# Query same mission (all runs)
gibson graphrag query "authentication bypasses" --scope same_mission --mission <mission-name>

# Query specific run
gibson graphrag query "jailbreak attempts" --scope same_mission --mission <mission-name> --run 2

# Include run metadata in results
gibson graphrag query "vulnerabilities" --scope same_mission --include-metadata
```

---

## Best Practices

### 1. Choose Appropriate Memory Continuity

**Use `isolated` mode when:**
- Running independent parallel tests
- Requiring strict run isolation
- Testing requires clean state
- Security-sensitive operations

**Use `inherit` mode when:**
- Building on previous run results
- Progressive refinement workflows
- Learning from historical data
- Sequential execution is expected

**Use `shared` mode when:**
- Multiple agents collaborate
- Real-time coordination needed
- Shared state accumulation required
- Concurrent runs must see each other's changes

### 2. Scope GraphRAG Queries Appropriately

```go
// Good: Scope to current run for fresh analysis
results, err := harness.QueryGraphRAGScoped(ctx,
    "new vulnerabilities",
    graphrag.ScopeCurrentRun,
)

// Good: Scope to same mission for continuity
results, err := harness.QueryGraphRAGScoped(ctx,
    "what did we try before?",
    graphrag.ScopeSameMission,
)

// Good: Scope to all missions for pattern discovery
results, err := harness.QueryGraphRAGScoped(ctx,
    "successful jailbreak techniques",
    graphrag.ScopeAll,
)
```

### 3. Check Run Context Before Acting

```go
func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    execCtx := harness.MissionExecutionContext()
    logger := harness.Logger()

    // Adapt behavior based on run context
    if execCtx.IsResumed {
        logger.Info("resuming from previous run",
            "node", execCtx.ResumedFromNode,
        )

        // Skip already-completed work
        // Jump to appropriate phase
    }

    if execCtx.RunNumber > 1 {
        // Check what previous runs discovered
        history, _ := harness.GetMissionRunHistory(ctx)
        if len(history) > 0 {
            lastRun := history[len(history)-1]

            if lastRun.Status == "failed" {
                logger.Info("previous run failed, adjusting strategy")
                // Use more conservative approach
            }
        }
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### 4. Handle Memory Continuity Gracefully

```go
func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    missionMem := harness.Memory().Mission()
    logger := harness.Logger()

    // Always check continuity mode before accessing previous runs
    mode := missionMem.ContinuityMode()
    logger.Info("memory mode", "continuity", mode)

    var state map[string]any

    if mode != memory.MemoryIsolated {
        // Attempt to inherit previous state
        prevState, err := missionMem.GetPreviousRunValue(ctx, "execution_state")
        if err == nil {
            state = prevState.(map[string]any)
            logger.Info("inherited state from previous run")
        } else if errors.Is(err, memory.ErrNoPreviousRun) {
            // First run - initialize
            state = map[string]any{"phase": "init"}
            logger.Info("first run - initializing state")
        }
    } else {
        // Isolated mode - start fresh
        state = map[string]any{"phase": "init"}
        logger.Info("isolated mode - fresh state")
    }

    // Continue with execution...

    return agent.NewSuccessResult("completed"), nil
}
```

### 5. Track Metrics Across Runs

```go
func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    execCtx := harness.MissionExecutionContext()
    logger := harness.Logger()

    // Get historical performance
    history, err := harness.GetMissionRunHistory(ctx)
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    // Calculate metrics
    var totalDuration time.Duration
    var completedRuns int

    for _, run := range history {
        if run.CompletedAt != nil {
            duration := run.CompletedAt.Sub(run.CreatedAt)
            totalDuration += duration
            completedRuns++
        }
    }

    if completedRuns > 0 {
        avgDuration := totalDuration / time.Duration(completedRuns)
        logger.Info("historical performance",
            "avg_duration", avgDuration,
            "completed_runs", completedRuns,
            "total_findings", execCtx.TotalFindingsAllRuns,
        )
    }

    return agent.NewSuccessResult("completed"), nil
}
```

### 6. Include Run Metadata in GraphRAG Nodes

```go
import (
    "context"
    "github.com/zero-day-ai/sdk/agent"
    "github.com/zero-day-ai/sdk/graphrag"
)

func executeFunc(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
    execCtx := harness.MissionExecutionContext()

    // Create GraphRAG node with run metadata
    node := &graphrag.Node{
        Type:      "finding",
        MissionID: execCtx.MissionID,
        Content:   "SQL injection discovered in login form",
        Properties: map[string]any{
            "severity":     "high",
            "technique":    "AML.T0051",
            "run_number":   execCtx.RunNumber,
            "mission_name": execCtx.MissionName,
            "is_resumed":   execCtx.IsResumed,
        },
    }

    err := harness.GraphRAG().StoreNode(ctx, node)
    if err != nil {
        return agent.NewFailedResult(err), err
    }

    return agent.NewSuccessResult("completed"), nil
}
```

---

## Summary

Mission execution context in Gibson provides powerful capabilities for:

- **Run Tracking**: Access current run number, resume state, and previous run information
- **History Queries**: Analyze patterns across multiple mission runs
- **Scoped Queries**: Control which runs are included in GraphRAG searches
- **Memory Continuity**: Configure how memory persists across runs (isolated, inherit, shared)
- **Adaptive Behavior**: Adjust agent strategies based on historical performance

These features enable resumable missions, progressive refinement, collaborative multi-agent scenarios, and sophisticated knowledge retrieval across mission boundaries.

For more information, see:
- [Agent Development Guide](AGENT.md)
- [Memory System Documentation](../memory/doc.go)
- [GraphRAG Package](../graphrag/)
