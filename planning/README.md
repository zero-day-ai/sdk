# Planning Package

The `planning` package provides planning-aware interfaces for Gibson agent developers. It enables agents to be aware of their position in mission execution, access budget constraints, and provide feedback to the planning system.

## Overview

This package is part of the Gibson SDK and provides two main types:

1. **`PlanningContext`** - Read-only interface for accessing mission plan state
2. **`StepHints`** - Builder for providing feedback to the planning system

## Features

- **Mission awareness**: Agents can see their position in the overall plan
- **Budget tracking**: Access to step-level and mission-level token budgets
- **Feedback mechanism**: Report confidence, findings, and planning suggestions
- **Tactical replanning**: Agents can recommend when replanning is needed
- **Standalone design**: No dependencies on Gibson internals, uses only standard library

## Usage

### Reading Planning Context

Agents receive the planning context through the harness:

```go
func (a *MyAgent) Execute(ctx context.Context, task agent.Task, harness harness.AgentHarness) (agent.Result, error) {
    planCtx := harness.PlanContext()
    if planCtx == nil {
        // Planning not enabled, proceed normally
        return a.executeNormal(ctx, task, harness)
    }

    // Adapt based on plan position
    if planCtx.CurrentStepIndex() == 0 {
        // First step - initialize mission state
        log.Info("Starting mission", "goal", planCtx.OriginalGoal())
    }

    // Adapt to budget constraints
    if planCtx.StepBudget() < 1000 {
        log.Info("Low step budget, using efficient strategy")
        return a.executeEfficient(ctx, task, harness)
    }

    // Final step handling
    if planCtx.CurrentStepIndex() == planCtx.TotalSteps()-1 {
        log.Info("Last step - summarizing findings")
        return a.summarize(ctx, harness)
    }

    // Normal execution
    return a.executeNormal(ctx, task, harness)
}
```

### Providing Step Hints

Use the fluent builder to construct hints:

```go
// Build hints
hints := planning.NewStepHints().
    WithConfidence(0.85).
    WithKeyFinding("Admin panel discovered at /admin").
    WithKeyFinding("Default credentials may be in use").
    WithSuggestion("auth_bypass_agent").
    WithSuggestion("credential_stuffing_agent")

// Report to framework
harness.ReportStepHints(ctx, hints)
```

### Recommending Replanning

When an agent discovers that the current plan is ineffective:

```go
hints := planning.NewStepHints().
    WithConfidence(0.2).
    RecommendReplan("Target uses custom auth - standard attacks ineffective")

harness.ReportStepHints(ctx, hints)
```

## API Reference

### PlanningContext

| Method | Returns | Description |
|--------|---------|-------------|
| `OriginalGoal()` | `string` | Immutable mission goal statement |
| `CurrentStepIndex()` | `int` | 0-based index of current step |
| `TotalSteps()` | `int` | Total number of planned steps |
| `RemainingSteps()` | `[]string` | Node IDs executing after this step |
| `StepBudget()` | `int` | Token budget for this step (0 = unlimited) |
| `MissionBudgetRemaining()` | `int` | Total remaining mission budget (0 = no tracking) |

### StepHints Builder

| Method | Returns | Description |
|--------|---------|-------------|
| `NewStepHints()` | `*StepHints` | Create new hints with defaults (confidence=0.5) |
| `WithConfidence(float64)` | `*StepHints` | Set confidence (0.0-1.0, clamped) |
| `WithSuggestion(string)` | `*StepHints` | Add suggested next step |
| `WithKeyFinding(string)` | `*StepHints` | Add key finding |
| `RecommendReplan(string)` | `*StepHints` | Recommend replanning with reason |

### StepHints Getters

| Method | Returns | Description |
|--------|---------|-------------|
| `Confidence()` | `float64` | Agent's self-assessed confidence |
| `SuggestedNext()` | `[]string` | List of suggested next steps (copy) |
| `KeyFindings()` | `[]string` | List of key findings (copy) |
| `ReplanReason()` | `string` | Reason for replanning or empty |
| `HasReplanRecommendation()` | `bool` | True if replanning recommended |

## Design Principles

1. **Standalone**: No dependencies on `internal/` packages - external agents can use it
2. **Immutable context**: Agents cannot modify mission state, only read it
3. **Defensive copies**: Getters return copies to prevent external modification
4. **Fluent builder**: Chain method calls for clean, readable code
5. **Safe defaults**: NewStepHints() creates sensible defaults (confidence=0.5)
6. **Input validation**: Confidence values are clamped to [0.0, 1.0]
7. **Empty string filtering**: Empty suggestions and findings are ignored

## Testing

Run the test suite:

```bash
go test ./planning/...
go test -race -cover ./planning/...
```

View examples:

```bash
go test -v -run Example ./planning/...
```

## Related Packages

- `github.com/zero-day-ai/sdk/agent` - Agent interface definition
- `github.com/zero-day-ai/sdk` - Core SDK types and harness interface

## License

See the main Gibson repository for license information.
