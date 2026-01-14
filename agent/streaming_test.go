package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/types"
)

func TestSetStreamingExecuteFunc(t *testing.T) {
	t.Run("agent with streaming func", func(t *testing.T) {
		streamingCalled := false
		cfg := NewConfig().
			SetName("test-agent").
			SetVersion("1.0.0").
			SetDescription("Test agent with streaming").
			SetExecuteFunc(func(ctx context.Context, harness Harness, task Task) (Result, error) {
				return NewSuccessResult("regular"), nil
			}).
			SetStreamingExecuteFunc(func(ctx context.Context, harness StreamingHarness, task Task) (Result, error) {
				streamingCalled = true
				return NewSuccessResult("streaming"), nil
			})

		agent, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Verify agent has ExecuteStreaming method
		sdkAgent, ok := agent.(*sdkAgent)
		require.True(t, ok, "agent should be *sdkAgent")
		assert.NotNil(t, sdkAgent.streamingExecuteFunc, "streamingExecuteFunc should be set")

		// Test that ExecuteStreaming works
		mockHarness := &mockStreamingHarness{}
		task := NewTask("test")
		result, err := sdkAgent.ExecuteStreaming(context.Background(), mockHarness, *task)
		require.NoError(t, err)
		assert.True(t, streamingCalled, "streaming func should be called")
		assert.Equal(t, StatusSuccess, result.Status)
		assert.Equal(t, "streaming", result.Output)
	})

	t.Run("agent without streaming func", func(t *testing.T) {
		cfg := NewConfig().
			SetName("test-agent").
			SetVersion("1.0.0").
			SetDescription("Test agent without streaming").
			SetExecuteFunc(func(ctx context.Context, harness Harness, task Task) (Result, error) {
				return NewSuccessResult("regular"), nil
			})

		agent, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, agent)

		// Verify agent doesn't have streaming func
		sdkAgent, ok := agent.(*sdkAgent)
		require.True(t, ok, "agent should be *sdkAgent")
		assert.Nil(t, sdkAgent.streamingExecuteFunc, "streamingExecuteFunc should be nil")

		// Test that ExecuteStreaming returns error
		mockHarness := &mockStreamingHarness{}
		task := NewTask("test")
		result, err := sdkAgent.ExecuteStreaming(context.Background(), mockHarness, *task)
		require.Error(t, err)
		assert.Equal(t, StatusFailed, result.Status)
		assert.Contains(t, err.Error(), "streaming execute function not configured")
	})
}

// mockStreamingHarness is a minimal mock implementation of StreamingHarness for testing
type mockStreamingHarness struct {
	mockHarness
}

func (m *mockStreamingHarness) EmitOutput(content string, isReasoning bool) error {
	return nil
}

func (m *mockStreamingHarness) EmitToolCall(toolName string, input map[string]any, callID string) error {
	return nil
}

func (m *mockStreamingHarness) EmitToolResult(output map[string]any, err error, callID string) error {
	return nil
}

func (m *mockStreamingHarness) EmitFinding(f *finding.Finding) error {
	return nil
}

func (m *mockStreamingHarness) EmitStatus(status string, message string) error {
	return nil
}

func (m *mockStreamingHarness) EmitError(err error, context string) error {
	return nil
}

func (m *mockStreamingHarness) Steering() <-chan SteeringMessage {
	return make(<-chan SteeringMessage)
}

func (m *mockStreamingHarness) Mode() ExecutionMode {
	return ExecutionModeAutonomous
}

// GraphRAG stub implementations (not used in these tests but required by interface)
func (m *mockStreamingHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}

func (m *mockStreamingHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	return nil, nil
}

func (m *mockStreamingHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (m *mockStreamingHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	return nil, nil
}

func (m *mockStreamingHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (m *mockStreamingHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "", nil
}

func (m *mockStreamingHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	return nil
}

func (m *mockStreamingHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	return nil, nil
}

func (m *mockStreamingHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	return nil, nil
}

func (m *mockStreamingHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("ok")
}
