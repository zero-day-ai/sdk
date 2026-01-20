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
	"go.opentelemetry.io/otel/trace/noop"
)

// ExampleNewFeedbackHarness demonstrates basic usage of FeedbackHarness
// for real-time agent evaluation with feedback.
func ExampleNewFeedbackHarness() {
	// Create a mock harness (in real usage, this would be your actual harness)
	innerHarness := createMockHarness()

	// Create a streaming scorer (in real usage, you'd use actual scorers)
	scorer := createMockScorer()

	// Configure feedback options
	opts := eval.FeedbackOptions{
		Scorers:           []eval.StreamingScorer{scorer},
		WarningThreshold:  0.5,
		CriticalThreshold: 0.2,
		Frequency: eval.FeedbackFrequency{
			EveryNSteps: 2, // Evaluate every 2 steps
		},
		AutoInject: false, // Manually retrieve feedback
	}

	// Create feedback harness
	fh := eval.NewFeedbackHarness(innerHarness, opts)
	defer fh.Close()

	ctx := context.Background()

	// Perform some agent operations
	messages := []llm.Message{{Role: "user", Content: "Analyze target"}}
	_, _ = fh.Complete(ctx, "primary", messages)
	_, _ = fh.CallTool(ctx, "nmap", map[string]any{"target": "example.com"})

	// Wait for async evaluation
	time.Sleep(500 * time.Millisecond)

	// Check for feedback
	if feedback := fh.GetFeedback(); feedback != nil {
		fmt.Printf("Score: %.2f\n", feedback.Overall.Score)
		fmt.Printf("Action: %s\n", feedback.Overall.Action)
		if len(feedback.Alerts) > 0 {
			fmt.Printf("Alerts: %d\n", len(feedback.Alerts))
		}
	}

	// Access full trajectory
	trajectory := fh.RecordingHarness().Trajectory()
	fmt.Printf("Steps recorded: %d\n", len(trajectory.Steps))

	// Output:
	// Score: 0.80
	// Action: continue
	// Steps recorded: 2
}

// ExampleFeedbackHarness_autoInject demonstrates automatic feedback injection
// into LLM calls for agent self-correction.
func ExampleFeedbackHarness_autoInject() {
	innerHarness := createMockHarness()
	scorer := createMockScorer()

	opts := eval.FeedbackOptions{
		Scorers: []eval.StreamingScorer{scorer},
		Frequency: eval.FeedbackFrequency{
			EveryNSteps: 1,
		},
		AutoInject: true, // Automatically inject feedback into LLM calls
	}

	fh := eval.NewFeedbackHarness(innerHarness, opts)
	defer fh.Close()

	ctx := context.Background()

	// First LLM call
	_, _ = fh.Complete(ctx, "primary", []llm.Message{{Role: "user", Content: "Step 1"}})
	time.Sleep(100 * time.Millisecond)

	// Second LLM call will automatically receive feedback as a system message
	_, _ = fh.Complete(ctx, "primary", []llm.Message{{Role: "user", Content: "Step 2"}})

	fmt.Println("Feedback automatically injected into LLM calls")

	// Output:
	// Feedback automatically injected into LLM calls
}

// ExampleFeedbackOptions_frequency demonstrates different frequency configurations.
func ExampleFeedbackOptions_frequency() {
	// Evaluate after every step
	freq1 := eval.FeedbackFrequency{
		EveryNSteps: 1,
	}

	// Evaluate every 5 steps with debouncing
	freq2 := eval.FeedbackFrequency{
		EveryNSteps: 5,
		Debounce:    500 * time.Millisecond,
	}

	// Always evaluate on threshold breach
	freq3 := eval.FeedbackFrequency{
		EveryNSteps: 10,
		OnThreshold: true, // Override step frequency if threshold crossed
	}

	fmt.Printf("Frequency 1: Every %d step(s)\n", freq1.EveryNSteps)
	fmt.Printf("Frequency 2: Every %d step(s) with %v debounce\n", freq2.EveryNSteps, freq2.Debounce)
	fmt.Printf("Frequency 3: Every %d step(s), on threshold: %v\n", freq3.EveryNSteps, freq3.OnThreshold)

	// Output:
	// Frequency 1: Every 1 step(s)
	// Frequency 2: Every 5 step(s) with 500ms debounce
	// Frequency 3: Every 10 step(s), on threshold: true
}

// Helper functions for examples

type exampleHarness struct{}

func (e *exampleHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{Content: "response"}, nil
}

func (e *exampleHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{Content: "response"}, nil
}

func (e *exampleHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

func (e *exampleHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	return map[string]any{"result": "success"}, nil
}

func (e *exampleHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	return nil, nil
}

func (e *exampleHarness) CallToolsParallel(ctx context.Context, calls []agent.ToolCall, maxConcurrency int) ([]agent.ToolResult, error) {
	results := make([]agent.ToolResult, len(calls))
	for i, call := range calls {
		output, err := e.CallTool(ctx, call.Name, call.Input)
		results[i] = agent.ToolResult{Name: call.Name, Output: output, Error: err}
	}
	return results, nil
}

func (e *exampleHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	return nil, nil
}

func (e *exampleHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	return nil, nil
}

func (e *exampleHarness) DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error) {
	return agent.NewSuccessResult("done"), nil
}

func (e *exampleHarness) ListAgents(ctx context.Context) ([]agent.Descriptor, error) {
	return nil, nil
}

func (e *exampleHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	return nil
}

func (e *exampleHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return nil, nil
}

func (e *exampleHarness) Memory() memory.Store {
	return nil
}

func (e *exampleHarness) PlanContext() planning.PlanningContext {
	return nil
}

func (e *exampleHarness) ReportStepHints(ctx context.Context, hints *planning.StepHints) error {
	return nil
}

func (e *exampleHarness) Mission() types.MissionContext {
	return types.MissionContext{}
}

func (e *exampleHarness) Target() types.TargetInfo {
	return types.TargetInfo{}
}

func (e *exampleHarness) Tracer() trace.Tracer {
	return noop.NewTracerProvider().Tracer("test")
}

func (e *exampleHarness) Logger() *slog.Logger {
	return slog.Default()
}

func (e *exampleHarness) TokenUsage() llm.TokenTracker {
	return nil
}

func (e *exampleHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}

func (e *exampleHarness) QuerySemantic(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}

func (e *exampleHarness) QueryStructured(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}

func (e *exampleHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	return nil, nil
}

func (e *exampleHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (e *exampleHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	return nil, nil
}

func (e *exampleHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (e *exampleHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "node-id", nil
}

func (e *exampleHarness) StoreSemantic(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "node-id", nil
}

func (e *exampleHarness) StoreStructured(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "node-id", nil
}

func (e *exampleHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	return nil
}

func (e *exampleHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	return nil, nil
}

func (e *exampleHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	return nil, nil
}

func (e *exampleHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	return types.HealthStatus{Status: "healthy"}
}

func (e *exampleHarness) MissionExecutionContext() types.MissionExecutionContext {
	return types.MissionExecutionContext{}
}

func (e *exampleHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
	return []types.MissionRunSummary{}, nil
}

func (e *exampleHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (e *exampleHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (e *exampleHarness) QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error) {
	return nil, nil
}

// MissionManager methods - stubs for testing
func (e *exampleHarness) CreateMission(ctx context.Context, workflow any, targetID string, opts *mission.CreateMissionOpts) (*mission.MissionInfo, error) {
	return nil, errors.New("not implemented")
}

func (e *exampleHarness) RunMission(ctx context.Context, missionID string, opts *mission.RunMissionOpts) error {
	return errors.New("not implemented")
}

func (e *exampleHarness) GetMissionStatus(ctx context.Context, missionID string) (*mission.MissionStatusInfo, error) {
	return nil, errors.New("not implemented")
}

func (e *exampleHarness) WaitForMission(ctx context.Context, missionID string, timeout time.Duration) (*mission.MissionResult, error) {
	return nil, errors.New("not implemented")
}

func (e *exampleHarness) ListMissions(ctx context.Context, filter *mission.MissionFilter) ([]*mission.MissionInfo, error) {
	return nil, errors.New("not implemented")
}

func (e *exampleHarness) CancelMission(ctx context.Context, missionID string) error {
	return errors.New("not implemented")
}

func (e *exampleHarness) GetMissionResults(ctx context.Context, missionID string) (*mission.MissionResult, error) {
	return nil, errors.New("not implemented")
}

func (e *exampleHarness) GetCredential(ctx context.Context, name string) (*types.Credential, error) {
	return &types.Credential{
		Name:   name,
		Type:   "api-key",
		Secret: "mock-secret-value",
	}, nil
}

// CompleteStructured methods
func (e *exampleHarness) CompleteStructured(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return nil, errors.New("not implemented")
}

func (e *exampleHarness) CompleteStructuredAny(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return e.CompleteStructured(ctx, slot, messages, schema)
}

type exampleScorer struct{}

func (s *exampleScorer) Name() string {
	return "example-scorer"
}

func (s *exampleScorer) Score(ctx context.Context, sample eval.Sample) (eval.ScoreResult, error) {
	return eval.ScoreResult{Score: 0.8}, nil
}

func (s *exampleScorer) ScorePartial(ctx context.Context, trajectory eval.Trajectory) (eval.PartialScore, error) {
	return eval.PartialScore{
		Score:      0.8,
		Confidence: 0.9,
		Status:     eval.ScoreStatusPartial,
		Action:     eval.ActionContinue,
	}, nil
}

func (s *exampleScorer) SupportsStreaming() bool {
	return true
}

func createMockHarness() agent.Harness {
	return &exampleHarness{}
}

func createMockScorer() eval.StreamingScorer {
	return &exampleScorer{}
}
