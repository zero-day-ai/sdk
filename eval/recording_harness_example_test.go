package eval_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/eval"
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/memory"
	"github.com/zero-day-ai/sdk/mission"
	"github.com/zero-day-ai/sdk/planning"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/tool"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/trace"
)

// This example demonstrates how to use RecordingHarness to capture
// agent execution trajectories for evaluation.
func ExampleRecordingHarness() {
	// Create a mock harness for demonstration
	// In real usage, this would be the actual agent harness
	mockHarness := &minimalMockHarness{}

	// Wrap it with a recording harness
	recorder := eval.NewRecordingHarness(mockHarness)

	// Execute agent operations through the recording harness
	ctx := context.Background()

	// LLM completion
	_, _ = recorder.Complete(ctx, "primary", []llm.Message{
		{Role: "user", Content: "What is 2+2?"},
	})

	// Tool invocation
	_, _ = recorder.CallTool(ctx, "calculator", map[string]any{
		"operation": "add",
		"a":         2,
		"b":         2,
	})

	// Memory operation
	_ = recorder.Memory().Working().Set(ctx, "result", 4)

	// Get the recorded trajectory
	trajectory := recorder.Trajectory()

	// Print trajectory summary
	fmt.Printf("Recorded %d operations\n", len(trajectory.Steps))
	for i, step := range trajectory.Steps {
		// Round duration to milliseconds for consistent output
		durationMs := step.Duration.Round(time.Millisecond)
		fmt.Printf("%d. %s: %s (took %v)\n", i+1, step.Type, step.Name, durationMs)
	}

	// Output:
	// Recorded 3 operations
	// 1. llm: primary (took 0s)
	// 2. tool: calculator (took 0s)
	// 3. memory.working: set (took 0s)
}

// minimalMockHarness is a minimal harness implementation for the example.
type minimalMockHarness struct{}

func (m *minimalMockHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{Content: "4"}, nil
}

func (m *minimalMockHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	return map[string]any{"result": 4}, nil
}

func (m *minimalMockHarness) Memory() memory.Store {
	return &minimalMemoryStore{}
}

// Stub implementations for other required methods (not shown for brevity)
func (m *minimalMockHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	return nil, nil
}
func (m *minimalMockHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	return nil, nil
}
func (m *minimalMockHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	return nil, nil
}
func (m *minimalMockHarness) CallToolsParallel(ctx context.Context, calls []agent.ToolCall, maxConcurrency int) ([]agent.ToolResult, error) {
	results := make([]agent.ToolResult, len(calls))
	for i, call := range calls {
		output, err := m.CallTool(ctx, call.Name, call.Input)
		results[i] = agent.ToolResult{Name: call.Name, Output: output, Error: err}
	}
	return results, nil
}
func (m *minimalMockHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	return nil, nil
}
func (m *minimalMockHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	return nil, nil
}
func (m *minimalMockHarness) DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error) {
	return agent.Result{}, nil
}
func (m *minimalMockHarness) ListAgents(ctx context.Context) ([]agent.Descriptor, error) {
	return nil, nil
}
func (m *minimalMockHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error { return nil }
func (m *minimalMockHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return nil, nil
}
func (m *minimalMockHarness) PlanContext() planning.PlanningContext { return nil }
func (m *minimalMockHarness) ReportStepHints(ctx context.Context, hints *planning.StepHints) error {
	return nil
}
func (m *minimalMockHarness) Mission() types.MissionContext { return types.MissionContext{} }
func (m *minimalMockHarness) Target() types.TargetInfo      { return types.TargetInfo{} }
func (m *minimalMockHarness) Tracer() trace.Tracer          { return nil }
func (m *minimalMockHarness) Logger() *slog.Logger          { return nil }
func (m *minimalMockHarness) TokenUsage() llm.TokenTracker  { return nil }
func (m *minimalMockHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}
func (m *minimalMockHarness) QuerySemantic(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}
func (m *minimalMockHarness) QueryStructured(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}
func (m *minimalMockHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	return nil, nil
}
func (m *minimalMockHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	return nil, nil
}
func (m *minimalMockHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	return nil, nil
}
func (m *minimalMockHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	return nil, nil
}
func (m *minimalMockHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "", nil
}
func (m *minimalMockHarness) StoreSemantic(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "", nil
}
func (m *minimalMockHarness) StoreStructured(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "", nil
}
func (m *minimalMockHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	return nil
}
func (m *minimalMockHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	return nil, nil
}
func (m *minimalMockHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	return nil, nil
}
func (m *minimalMockHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	return types.HealthStatus{}
}

func (m *minimalMockHarness) MissionExecutionContext() types.MissionExecutionContext {
	return types.MissionExecutionContext{}
}

func (m *minimalMockHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
	return []types.MissionRunSummary{}, nil
}

func (m *minimalMockHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (m *minimalMockHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (m *minimalMockHarness) QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error) {
	return nil, nil
}

// MissionManager methods - stubs for testing
func (m *minimalMockHarness) CreateMission(ctx context.Context, workflow any, targetID string, opts *mission.CreateMissionOpts) (*mission.MissionInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *minimalMockHarness) RunMission(ctx context.Context, missionID string, opts *mission.RunMissionOpts) error {
	return errors.New("not implemented")
}

func (m *minimalMockHarness) GetMissionStatus(ctx context.Context, missionID string) (*mission.MissionStatusInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *minimalMockHarness) WaitForMission(ctx context.Context, missionID string, timeout time.Duration) (*mission.MissionResult, error) {
	return nil, errors.New("not implemented")
}

func (m *minimalMockHarness) ListMissions(ctx context.Context, filter *mission.MissionFilter) ([]*mission.MissionInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *minimalMockHarness) CancelMission(ctx context.Context, missionID string) error {
	return errors.New("not implemented")
}

func (m *minimalMockHarness) GetMissionResults(ctx context.Context, missionID string) (*mission.MissionResult, error) {
	return nil, errors.New("not implemented")
}

func (m *minimalMockHarness) GetCredential(ctx context.Context, name string) (*types.Credential, error) {
	return &types.Credential{
		Name:   name,
		Type:   "api-key",
		Secret: "mock-secret-value",
	}, nil
}

// CompleteStructured methods
func (m *minimalMockHarness) CompleteStructured(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return nil, errors.New("not implemented")
}

func (m *minimalMockHarness) CompleteStructuredAny(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return m.CompleteStructured(ctx, slot, messages, schema)
}

type minimalMemoryStore struct{}

func (m *minimalMemoryStore) Working() memory.WorkingMemory   { return &minimalWorkingMemory{} }
func (m *minimalMemoryStore) Mission() memory.MissionMemory   { return &minimalMissionMemory{} }
func (m *minimalMemoryStore) LongTerm() memory.LongTermMemory { return &minimalLongTermMemory{} }

type minimalMissionMemory struct{}

func (m *minimalMissionMemory) Get(ctx context.Context, key string) (*memory.Item, error) {
	return nil, nil
}
func (m *minimalMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	return nil
}
func (m *minimalMissionMemory) Delete(ctx context.Context, key string) error { return nil }
func (m *minimalMissionMemory) Search(ctx context.Context, query string, limit int) ([]memory.Result, error) {
	return nil, nil
}
func (m *minimalMissionMemory) History(ctx context.Context, limit int) ([]memory.Item, error) {
	return nil, nil
}

func (m *minimalMissionMemory) GetPreviousRunValue(ctx context.Context, key string) (any, error) {
	return nil, memory.ErrNotFound
}

func (m *minimalMissionMemory) GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error) {
	return nil, nil
}

func (m *minimalMissionMemory) ContinuityMode() memory.MemoryContinuityMode {
	return memory.MemoryIsolated
}

type minimalLongTermMemory struct{}

func (m *minimalLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	return "", nil
}
func (m *minimalLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]memory.Result, error) {
	return nil, nil
}
func (m *minimalLongTermMemory) Delete(ctx context.Context, id string) error { return nil }

type minimalWorkingMemory struct{}

func (m *minimalWorkingMemory) Get(ctx context.Context, key string) (any, error)     { return nil, nil }
func (m *minimalWorkingMemory) Set(ctx context.Context, key string, value any) error { return nil }
func (m *minimalWorkingMemory) Delete(ctx context.Context, key string) error         { return nil }
func (m *minimalWorkingMemory) Clear(ctx context.Context) error                      { return nil }
func (m *minimalWorkingMemory) Keys(ctx context.Context) ([]string, error)           { return nil, nil }
