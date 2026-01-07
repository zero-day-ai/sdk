package serve

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		originalGoal:           "Test mission goal",
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
	assert.Equal(t, "Test mission goal", ctx.OriginalGoal())
	assert.Equal(t, 2, ctx.CurrentStepIndex())
	assert.Equal(t, 5, ctx.TotalSteps())
	assert.Equal(t, []string{"step3", "step4", "step5"}, ctx.RemainingSteps())
	assert.Equal(t, 1000, ctx.StepBudget())
	assert.Equal(t, 5000, ctx.MissionBudgetRemaining())
}

// mockPlanningContext is a simple mock for testing PlanContext.
type mockPlanningContext struct {
	originalGoal           string
	currentStepIndex       int
	totalSteps             int
	remainingSteps         []string
	stepBudget             int
	missionBudgetRemaining int
}

func (m *mockPlanningContext) OriginalGoal() string {
	return m.originalGoal
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
