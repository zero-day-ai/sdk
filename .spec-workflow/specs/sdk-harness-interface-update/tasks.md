# Tasks: SDK Harness Interface Update

## Task Overview

| Task | Description | Status |
|------|-------------|--------|
| 1 | Fix mock harnesses in agent/harness_test.go | [ ] |
| 2 | Fix mock harnesses in eval/feedback_harness_test.go | [ ] |
| 3 | Fix mock harnesses in integration/agent_test.go | [ ] |
| 4 | Fix mock harnesses in serve/streaming_harness_test.go | [ ] |
| 5 | Verify SDK builds and tests pass | [ ] |
| 6 | Commit and push SDK to GitHub | [ ] |
| 7 | Update and rebuild enterprise agents | [ ] |
| 8 | Update and rebuild OSS tools | [ ] |
| 9 | Run E2E demo mission | [ ] |

---

## Task 1: Fix mock harnesses in agent/harness_test.go

- [ ] Add 5 new Harness interface methods to `mockHarness` struct

**Files:** `agent/harness_test.go`

**Requirements:** FR-1

**_Prompt:**
```
Role: Go SDK Developer
Task: Add missing Harness interface methods to mockHarness in agent/harness_test.go

Add these methods after existing mockHarness methods:
- MissionExecutionContext() types.MissionExecutionContext - return empty struct
- GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) - return empty slice, nil
- GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) - return empty slice, nil
- GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) - return empty slice, nil
- QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error) - return nil, nil

Also add PlanContext() and ReportStepHints() if missing.

Add required imports: "github.com/zero-day-ai/sdk/planning"

_Leverage: Look at serve/local_harness.go for implementation pattern
_Requirements: FR-1, US-1, US-2
Success: mockHarness implements full agent.Harness interface

After completion, mark task as [-] in progress, then use log-implementation tool, then mark [x] complete.
```

---

## Task 2: Fix mock harnesses in eval/feedback_harness_test.go

- [ ] Add 5 new Harness interface methods to `mockHarness` struct

**Files:** `eval/feedback_harness_test.go`

**Requirements:** FR-1

**_Prompt:**
```
Role: Go SDK Developer
Task: Add missing Harness interface methods to mockHarness in eval/feedback_harness_test.go

Add these methods:
- MissionExecutionContext() types.MissionExecutionContext
- GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error)
- GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)
- GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)
- QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error)
- PlanContext() planning.PlanningContext
- ReportStepHints(ctx context.Context, hints *planning.StepHints) error

All return empty/nil values.

_Leverage: Pattern from task 1
_Requirements: FR-1, US-1, US-2
Success: mockHarness implements full agent.Harness interface

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Task 3: Fix mock harnesses in integration/agent_test.go

- [ ] Add all missing Harness interface methods to `mockHarness` struct

**Files:** `integration/agent_test.go`

**Requirements:** FR-1

**_Prompt:**
```
Role: Go SDK Developer
Task: Add missing Harness interface methods to mockHarness in integration/agent_test.go

This mock may be missing many methods. Add all required by agent.Harness interface.
Check what's already there and add what's missing.

_Leverage: Compare with agent/harness.go interface definition
_Requirements: FR-1, US-1, US-2
Success: mockHarness implements full agent.Harness interface

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Task 4: Fix mock harnesses in serve/streaming_harness_test.go

- [ ] Add 5 new Harness interface methods to `mockStreamHarness` struct

**Files:** `serve/streaming_harness_test.go`

**Requirements:** FR-1

**_Prompt:**
```
Role: Go SDK Developer
Task: Add missing Harness interface methods to mockStreamHarness in serve/streaming_harness_test.go

Add these methods:
- MissionExecutionContext() types.MissionExecutionContext
- GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error)
- GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)
- GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)
- QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error)
- PlanContext() planning.PlanningContext (if missing)
- ReportStepHints(ctx context.Context, hints *planning.StepHints) error (if missing)

_Leverage: Pattern from previous tasks
_Requirements: FR-1, US-1, US-2
Success: mockStreamHarness implements full agent.Harness interface

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Task 5: Verify SDK builds and tests pass

- [ ] Run `go build ./...` and `go test ./...`

**Files:** All SDK files

**Requirements:** US-1, US-2

**_Prompt:**
```
Role: Go SDK Developer
Task: Verify SDK compiles and all tests pass

Commands:
1. cd /home/anthony/Code/zero-day.ai/opensource/sdk
2. go build ./...
3. go test ./...

Fix any remaining issues found.

_Requirements: US-1, US-2
Success: Zero build errors, zero test failures

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Task 6: Commit and push SDK to GitHub

- [ ] Commit all changes and push to origin

**Files:** All modified SDK files

**Requirements:** US-3

**_Prompt:**
```
Role: DevOps Engineer
Task: Commit and push SDK changes to GitHub

Commands:
1. cd /home/anthony/Code/zero-day.ai/opensource/sdk
2. git status
3. git add -A
4. git commit -m "feat: implement mission execution context and scoped GraphRAG methods

- Add MissionExecutionContext, GetMissionRunHistory, GetPreviousRunFindings,
  GetAllRunFindings, QueryGraphRAGScoped to all harness implementations
- Add ContinuityMode, GetPreviousRunValue, GetValueHistory to memory implementations
- Update all test mocks to implement full interfaces
- Stub implementations return empty values (proto support pending)"
5. git push origin main

_Requirements: US-3
Success: Changes pushed to GitHub successfully

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Task 7: Update and rebuild enterprise agents

- [ ] Update SDK dependency and rebuild all enterprise agents

**Files:** Enterprise agent repos

**Requirements:** US-4

**_Prompt:**
```
Role: DevOps Engineer
Task: Update enterprise agents to latest SDK and rebuild

For each agent in /home/anthony/Code/zero-day.ai/closed/agents/:
1. cd to agent directory
2. go get github.com/zero-day-ai/sdk@latest
3. go mod tidy
4. go build -o <agent-name> .
5. cp <agent-name> ~/.gibson/agents/bin/

Agents: whistler, crease, carl, bishop, k8skiller

_Requirements: US-4
Success: All 5 agents rebuilt and installed to ~/.gibson/agents/bin/

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Task 8: Update and rebuild OSS tools

- [ ] Update SDK dependency and rebuild OSS tools

**Files:** gibson-oss-tools repo

**Requirements:** US-4

**_Prompt:**
```
Role: DevOps Engineer
Task: Update OSS tools to latest SDK and rebuild

1. cd /home/anthony/Code/zero-day.ai/opensource/gibson-oss-tools
2. go get github.com/zero-day-ai/sdk@latest
3. go mod tidy
4. ./build.sh (or go build for each tool)
5. Install tools to ~/.gibson/tools/bin/

_Requirements: US-4
Success: All OSS tools rebuilt and installed

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Task 9: Run E2E demo mission

- [ ] Execute full E2E demo mission and verify LLM calls in Langfuse

**Files:** Demo mission YAML

**Requirements:** US-5

**_Prompt:**
```
Role: QA Engineer
Task: Run E2E demo mission and verify it works

Commands:
1. Ensure Gibson daemon is running
2. gibson mission run /home/anthony/Code/zero-day.ai/dev/gibson-demo-mission.yaml
3. Monitor agent execution in logs
4. Verify LLM calls appear in Langfuse at http://localhost:3000
5. Confirm mission completes or produces meaningful output

_Requirements: US-5
Success:
- Mission executes without immediate agent failures
- LLM calls visible in Langfuse
- Agents produce output (findings or analysis)

After completion, use log-implementation tool, then mark [x] complete.
```

---

## Completion Checklist

- [ ] All test mocks implement full Harness interface
- [ ] SDK builds without errors
- [ ] SDK tests pass
- [ ] SDK pushed to GitHub
- [ ] All enterprise agents rebuilt
- [ ] OSS tools rebuilt
- [ ] E2E mission runs successfully
- [ ] LLM calls visible in Langfuse
