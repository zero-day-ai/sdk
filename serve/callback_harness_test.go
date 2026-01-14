package serve

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/planning"
	"github.com/zero-day-ai/sdk/types"
)

// TestNewCallbackHarness tests the harness constructor.
func TestNewCallbackHarness(t *testing.T) {
	// Create a real callback client (not connected)
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mission := types.MissionContext{
		ID:   "mission-123",
		Name: "Test Mission",
	}

	target := types.TargetInfo{
		Connection: map[string]any{"url": "http://target.example.com"},
		Type:       "web",
	}

	harness := NewCallbackHarness(client, logger, nil, mission, target)

	assert.NotNil(t, harness)
	assert.Equal(t, mission.ID, harness.Mission().ID)
	assert.Equal(t, mission.Name, harness.Mission().Name)
	resultTarget := harness.Target()
	assert.Equal(t, "http://target.example.com", resultTarget.URL())
	assert.Equal(t, target.Type, resultTarget.Type)
}

// TestCallbackHarnessLogger tests the Logger method.
func TestCallbackHarnessLogger(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	harness := NewCallbackHarness(client, logger, nil, types.MissionContext{}, types.TargetInfo{})

	assert.NotNil(t, harness.Logger())
}

// TestCallbackHarnessMemory tests the Memory method.
func TestCallbackHarnessMemory(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	harness := NewCallbackHarness(client, logger, nil, types.MissionContext{}, types.TargetInfo{})

	memory := harness.Memory()
	assert.NotNil(t, memory)
}

// TestCallbackHarnessTokenUsage tests the TokenUsage method.
func TestCallbackHarnessTokenUsage(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	harness := NewCallbackHarness(client, logger, nil, types.MissionContext{}, types.TargetInfo{})

	tokens := harness.TokenUsage()
	assert.NotNil(t, tokens)
}

// TestCallbackHarnessMission tests the Mission method.
func TestCallbackHarnessMission(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mission := types.MissionContext{
		ID:   "mission-456",
		Name: "Another Mission",
	}

	harness := NewCallbackHarness(client, logger, nil, mission, types.TargetInfo{})

	result := harness.Mission()
	assert.Equal(t, mission.ID, result.ID)
	assert.Equal(t, mission.Name, result.Name)
}

// TestCallbackHarnessTarget tests the Target method.
func TestCallbackHarnessTarget(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	target := types.TargetInfo{
		Connection: map[string]any{"url": "http://example.com"},
		Type:       "api",
	}

	harness := NewCallbackHarness(client, logger, nil, types.MissionContext{}, target)

	result := harness.Target()
	assert.Equal(t, "http://example.com", result.URL())
	assert.Equal(t, target.Type, result.Type)
}

// TestCallbackHarnessPlanContext tests the PlanContext method.
func TestCallbackHarnessPlanContext(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	harness := NewCallbackHarness(client, logger, nil, types.MissionContext{}, types.TargetInfo{})

	// Initially nil since no planning context was set
	assert.Nil(t, harness.PlanContext())

	// Create a mock planning context
	mockCtx := &mockPlanningContext{
		currentStepIndex:       2,
		totalSteps:             5,
		remainingSteps:         []string{"step3", "step4", "step5"},
		stepBudget:             1000,
		missionBudgetRemaining: 5000,
	}

	// Set the planning context
	harness.SetPlanContext(mockCtx)

	// Verify it returns the context
	ctx := harness.PlanContext()
	assert.NotNil(t, ctx)
	assert.Equal(t, 2, ctx.CurrentStepIndex())
	assert.Equal(t, 5, ctx.TotalSteps())
	assert.Equal(t, []string{"step3", "step4", "step5"}, ctx.RemainingSteps())
	assert.Equal(t, 1000, ctx.StepBudget())
	assert.Equal(t, 5000, ctx.MissionBudgetRemaining())
}

// mockPlanningContext is a simple mock for testing PlanContext.
type mockPlanningContext struct {
	currentStepIndex       int
	totalSteps             int
	remainingSteps         []string
	stepBudget             int
	missionBudgetRemaining int
}

func (m *mockPlanningContext) CurrentStepIndex() int {
	return m.currentStepIndex
}

func (m *mockPlanningContext) TotalSteps() int {
	return m.totalSteps
}

func (m *mockPlanningContext) RemainingSteps() []string {
	return m.remainingSteps
}

func (m *mockPlanningContext) StepBudget() int {
	return m.stepBudget
}

func (m *mockPlanningContext) MissionBudgetRemaining() int {
	return m.missionBudgetRemaining
}

// TestStepHintsBuilder tests the StepHints builder pattern.
func TestStepHintsBuilder(t *testing.T) {
	hints := planning.NewStepHints().
		WithConfidence(0.85).
		WithSuggestion("next_agent").
		WithSuggestion("another_agent").
		WithKeyFinding("Found vulnerability XYZ").
		WithKeyFinding("Default credentials detected").
		RecommendReplan("Target uses custom authentication")

	assert.Equal(t, 0.85, hints.Confidence())
	assert.Equal(t, []string{"next_agent", "another_agent"}, hints.SuggestedNext())
	assert.Equal(t, []string{"Found vulnerability XYZ", "Default credentials detected"}, hints.KeyFindings())
	assert.Equal(t, "Target uses custom authentication", hints.ReplanReason())
	assert.True(t, hints.HasReplanRecommendation())
}

// TestStepHintsDefaultValues tests StepHints default values.
func TestStepHintsDefaultValues(t *testing.T) {
	hints := planning.NewStepHints()

	assert.Equal(t, 0.5, hints.Confidence()) // Default is neutral
	assert.Empty(t, hints.SuggestedNext())
	assert.Empty(t, hints.KeyFindings())
	assert.Equal(t, "", hints.ReplanReason())
	assert.False(t, hints.HasReplanRecommendation())
}

// TestStepHintsConfidenceClamping tests that confidence is clamped to [0, 1].
func TestStepHintsConfidenceClamping(t *testing.T) {
	// Test upper bound clamping
	hintsHigh := planning.NewStepHints().WithConfidence(1.5)
	assert.Equal(t, 1.0, hintsHigh.Confidence())

	// Test lower bound clamping
	hintsLow := planning.NewStepHints().WithConfidence(-0.5)
	assert.Equal(t, 0.0, hintsLow.Confidence())

	// Test valid range
	hintsValid := planning.NewStepHints().WithConfidence(0.7)
	assert.Equal(t, 0.7, hintsValid.Confidence())
}

// ============================================================================
// CallToolsParallel Tests
// ============================================================================

// mockToolCaller is a wrapper that implements CallToolsParallel using a mock CallTool.
// This tests the CallToolsParallel implementation by providing a controlled CallTool.
type mockToolCaller struct {
	callToolFunc func(ctx context.Context, name string, input map[string]any) (map[string]any, error)
}

// CallTool delegates to the mock function.
func (m *mockToolCaller) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, name, input)
	}
	return nil, assert.AnError
}

// CallToolsParallel uses the actual implementation but calls the mock's CallTool.
// This is a copy of the CallbackHarness.CallToolsParallel implementation for testing.
func (m *mockToolCaller) CallToolsParallel(ctx context.Context, calls []agent.ToolCall, maxConcurrency int) ([]agent.ToolResult, error) {
	if len(calls) == 0 {
		return []agent.ToolResult{}, nil
	}

	// Default concurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	// Create results slice (same length as calls, preserves order)
	results := make([]agent.ToolResult, len(calls))

	// Semaphore for concurrency limiting
	sem := make(chan struct{}, maxConcurrency)

	// WaitGroup for completion tracking
	var wg sync.WaitGroup

	// Execute calls in parallel
	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c agent.ToolCall) {
			defer wg.Done()

			// Acquire semaphore slot
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = agent.ToolResult{
					Name:  c.Name,
					Error: ctx.Err(),
				}
				return
			}

			// Execute tool call using mock's CallTool
			output, err := m.CallTool(ctx, c.Name, c.Input)
			results[idx] = agent.ToolResult{
				Name:   c.Name,
				Output: output,
				Error:  err,
			}
		}(i, call)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return results, nil
}

// newMockHarness creates a mock tool caller with customizable CallTool behavior.
func newMockHarness(callToolFunc func(ctx context.Context, name string, input map[string]any) (map[string]any, error)) *mockToolCaller {
	return &mockToolCaller{
		callToolFunc: callToolFunc,
	}
}

// TestCallToolsParallel_EmptyInput tests that empty input returns empty slice.
func TestCallToolsParallel_EmptyInput(t *testing.T) {
	harness := newMockHarness(nil)
	ctx := context.Background()

	results, err := harness.CallToolsParallel(ctx, []agent.ToolCall{}, 5)

	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestCallToolsParallel_SingleCall tests execution with a single tool call.
func TestCallToolsParallel_SingleCall(t *testing.T) {
	callCount := 0
	harness := newMockHarness(func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
		callCount++
		assert.Equal(t, "test_tool", name)
		assert.Equal(t, "value1", input["key1"])
		return map[string]any{"result": "success"}, nil
	})

	ctx := context.Background()
	calls := []agent.ToolCall{
		{Name: "test_tool", Input: map[string]any{"key1": "value1"}},
	}

	results, err := harness.CallToolsParallel(ctx, calls, 5)

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "test_tool", results[0].Name)
	assert.Equal(t, "success", results[0].Output["result"])
	assert.NoError(t, results[0].Error)
	assert.Equal(t, 1, callCount)
}

// TestCallToolsParallel_MultipleCallsSuccess tests multiple successful calls.
func TestCallToolsParallel_MultipleCallsSuccess(t *testing.T) {
	harness := newMockHarness(func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
		// Simulate different tool responses
		return map[string]any{
			"tool":  name,
			"input": input["param"],
		}, nil
	})

	ctx := context.Background()
	calls := []agent.ToolCall{
		{Name: "tool1", Input: map[string]any{"param": "a"}},
		{Name: "tool2", Input: map[string]any{"param": "b"}},
		{Name: "tool3", Input: map[string]any{"param": "c"}},
	}

	results, err := harness.CallToolsParallel(ctx, calls, 5)

	require.NoError(t, err)
	require.Len(t, results, 3)

	// Verify each result
	assert.Equal(t, "tool1", results[0].Name)
	assert.Equal(t, "tool1", results[0].Output["tool"])
	assert.Equal(t, "a", results[0].Output["input"])
	assert.NoError(t, results[0].Error)

	assert.Equal(t, "tool2", results[1].Name)
	assert.Equal(t, "tool2", results[1].Output["tool"])
	assert.Equal(t, "b", results[1].Output["input"])
	assert.NoError(t, results[1].Error)

	assert.Equal(t, "tool3", results[2].Name)
	assert.Equal(t, "tool3", results[2].Output["tool"])
	assert.Equal(t, "c", results[2].Output["input"])
	assert.NoError(t, results[2].Error)
}

// TestCallToolsParallel_PartialFailure tests that some calls can fail while others succeed.
func TestCallToolsParallel_PartialFailure(t *testing.T) {
	harness := newMockHarness(func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
		// Fail on tool2
		if name == "tool2" {
			return nil, assert.AnError
		}
		return map[string]any{"status": "ok"}, nil
	})

	ctx := context.Background()
	calls := []agent.ToolCall{
		{Name: "tool1", Input: map[string]any{}},
		{Name: "tool2", Input: map[string]any{}},
		{Name: "tool3", Input: map[string]any{}},
	}

	results, err := harness.CallToolsParallel(ctx, calls, 5)

	require.NoError(t, err)
	require.Len(t, results, 3)

	// tool1 should succeed
	assert.Equal(t, "tool1", results[0].Name)
	assert.Equal(t, "ok", results[0].Output["status"])
	assert.NoError(t, results[0].Error)

	// tool2 should fail
	assert.Equal(t, "tool2", results[1].Name)
	assert.Nil(t, results[1].Output)
	assert.Error(t, results[1].Error)

	// tool3 should succeed
	assert.Equal(t, "tool3", results[2].Name)
	assert.Equal(t, "ok", results[2].Output["status"])
	assert.NoError(t, results[2].Error)
}

// TestCallToolsParallel_OrderPreserved tests that results match input order.
func TestCallToolsParallel_OrderPreserved(t *testing.T) {
	harness := newMockHarness(func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
		// Simulate variable execution times to test ordering
		// Even if execution order varies, results should match input order
		return map[string]any{"name": name}, nil
	})

	ctx := context.Background()
	calls := []agent.ToolCall{
		{Name: "first", Input: map[string]any{}},
		{Name: "second", Input: map[string]any{}},
		{Name: "third", Input: map[string]any{}},
		{Name: "fourth", Input: map[string]any{}},
		{Name: "fifth", Input: map[string]any{}},
	}

	results, err := harness.CallToolsParallel(ctx, calls, 10)

	require.NoError(t, err)
	require.Len(t, results, 5)

	// Verify order is preserved
	assert.Equal(t, "first", results[0].Name)
	assert.Equal(t, "second", results[1].Name)
	assert.Equal(t, "third", results[2].Name)
	assert.Equal(t, "fourth", results[3].Name)
	assert.Equal(t, "fifth", results[4].Name)
}

// TestCallToolsParallel_ConcurrencyLimit tests that maxConcurrency is respected.
func TestCallToolsParallel_ConcurrencyLimit(t *testing.T) {
	const maxConcurrency = 2
	activeCalls := 0
	maxActive := 0
	var mu sync.Mutex

	// Channel to signal when work starts and completes
	workStarted := make(chan struct{}, 10)
	workComplete := make(chan struct{}, 10)

	harness := newMockHarness(func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
		mu.Lock()
		activeCalls++
		if activeCalls > maxActive {
			maxActive = activeCalls
		}
		mu.Unlock()

		// Signal work started
		workStarted <- struct{}{}

		// Simulate some work by waiting for signal
		<-workComplete

		mu.Lock()
		activeCalls--
		mu.Unlock()

		return map[string]any{"status": "ok"}, nil
	})

	ctx := context.Background()
	calls := []agent.ToolCall{
		{Name: "tool1", Input: map[string]any{}},
		{Name: "tool2", Input: map[string]any{}},
		{Name: "tool3", Input: map[string]any{}},
		{Name: "tool4", Input: map[string]any{}},
		{Name: "tool5", Input: map[string]any{}},
	}

	// Start the parallel calls in background
	done := make(chan struct{})
	go func() {
		harness.CallToolsParallel(ctx, calls, maxConcurrency)
		close(done)
	}()

	// Wait for first batch to start
	<-workStarted
	<-workStarted

	// Check that we have exactly maxConcurrency active
	mu.Lock()
	currentActive := activeCalls
	mu.Unlock()
	assert.Equal(t, maxConcurrency, currentActive, "Should have exactly maxConcurrency active")

	// Complete all work
	for i := 0; i < len(calls); i++ {
		workComplete <- struct{}{}
	}

	// Wait for completion
	<-done

	// Verify we never exceeded the concurrency limit
	mu.Lock()
	defer mu.Unlock()
	assert.LessOrEqual(t, maxActive, maxConcurrency, "Should never exceed max concurrency")
}

// TestCallToolsParallel_ContextCancellation tests context cancellation handling.
func TestCallToolsParallel_ContextCancellation(t *testing.T) {
	harness := newMockHarness(func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
		// Simulate slow operation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := []agent.ToolCall{
		{Name: "tool1", Input: map[string]any{}},
		{Name: "tool2", Input: map[string]any{}},
		{Name: "tool3", Input: map[string]any{}},
	}

	results, err := harness.CallToolsParallel(ctx, calls, 5)

	require.NoError(t, err)
	require.Len(t, results, 3)

	// All calls should have context error
	for i, result := range results {
		assert.Equal(t, calls[i].Name, result.Name)
		assert.Error(t, result.Error)
		assert.Equal(t, context.Canceled, result.Error)
	}
}

// TestCallToolsParallel_DefaultConcurrency tests that 0 maxConcurrency uses default.
func TestCallToolsParallel_DefaultConcurrency(t *testing.T) {
	callCount := 0
	var mu sync.Mutex

	harness := newMockHarness(func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		return map[string]any{"status": "ok"}, nil
	})

	ctx := context.Background()
	calls := make([]agent.ToolCall, 15)
	for i := range calls {
		calls[i] = agent.ToolCall{
			Name:  "tool",
			Input: map[string]any{"index": i},
		}
	}

	// Pass 0 to test default concurrency
	results, err := harness.CallToolsParallel(ctx, calls, 0)

	require.NoError(t, err)
	require.Len(t, results, 15)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 15, callCount, "All calls should complete")
}
