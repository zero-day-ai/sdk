package serve

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/planning"
	"github.com/zero-day-ai/sdk/types"
)

func TestNewLocalHarness(t *testing.T) {
	h := newLocalHarness()

	assert.NotNil(t, h)
	assert.NotNil(t, h.memory)
	assert.NotNil(t, h.logger)
	assert.NotNil(t, h.tracer)
	assert.NotNil(t, h.tokenTracker)
}

func TestLocalHarness_Memory(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// Test Set and Get on working memory
	err := h.Memory().Working().Set(ctx, "test-key", "test-value")
	require.NoError(t, err)

	val, err := h.Memory().Working().Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", val)

	// Test Delete
	err = h.Memory().Working().Delete(ctx, "test-key")
	require.NoError(t, err)

	_, err = h.Memory().Working().Get(ctx, "test-key")
	assert.Error(t, err)

	// Test Keys with filtering manually
	err = h.Memory().Working().Set(ctx, "prefix:key1", "val1")
	require.NoError(t, err)
	err = h.Memory().Working().Set(ctx, "prefix:key2", "val2")
	require.NoError(t, err)
	err = h.Memory().Working().Set(ctx, "other:key", "val3")
	require.NoError(t, err)

	keys, err := h.Memory().Working().Keys(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(keys), 3)

	// Test Clear
	err = h.Memory().Working().Clear(ctx)
	require.NoError(t, err)

	keys, err = h.Memory().Working().Keys(ctx)
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestLocalHarness_TokenTracker(t *testing.T) {
	h := newLocalHarness()

	// Test Add and Get
	usage := llm.TokenUsage{
		InputTokens:  100,
		OutputTokens: 50,
		TotalTokens:  150,
	}
	h.TokenUsage().Add("primary", usage)

	retrieved := h.TokenUsage().BySlot("primary")
	assert.Equal(t, usage, retrieved)

	// Test Total
	h.TokenUsage().Add("secondary", llm.TokenUsage{
		InputTokens:  200,
		OutputTokens: 100,
		TotalTokens:  300,
	})

	total := h.TokenUsage().Total()
	assert.Equal(t, 300, total.InputTokens)
	assert.Equal(t, 150, total.OutputTokens)
	assert.Equal(t, 450, total.TotalTokens)

	// Test Reset
	h.TokenUsage().Reset()
	total = h.TokenUsage().Total()
	assert.Equal(t, 0, total.TotalTokens)
}

func TestLocalHarness_Observability(t *testing.T) {
	h := newLocalHarness()

	assert.NotNil(t, h.Logger())
	assert.NotNil(t, h.Tracer())
	assert.NotNil(t, h.TokenUsage())
}

func TestLocalHarness_Context(t *testing.T) {
	h := newLocalHarness()

	// Mission and Target should return empty defaults
	mission := h.Mission()
	assert.Equal(t, types.MissionContext{}, mission)

	target := h.Target()
	assert.Equal(t, types.TargetInfo{}, target)
}

func TestLocalHarness_LLMOperations_NotAvailable(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// Complete should return error
	_, err := h.Complete(ctx, "primary", []llm.Message{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// CompleteWithTools should return error
	_, err = h.CompleteWithTools(ctx, "primary", []llm.Message{}, []llm.ToolDef{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// Stream should return error
	_, err = h.Stream(ctx, "primary", []llm.Message{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")
}

func TestLocalHarness_ToolOperations_NotAvailable(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// CallTool should return error
	_, err := h.CallTool(ctx, "http-client", map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// ListTools should return empty list
	tools, err := h.ListTools(ctx)
	assert.NoError(t, err)
	assert.Empty(t, tools)
}

func TestLocalHarness_PluginOperations_NotAvailable(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// QueryPlugin should return error
	_, err := h.QueryPlugin(ctx, "graphrag", "query", map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// ListPlugins should return empty list
	plugins, err := h.ListPlugins(ctx)
	assert.NoError(t, err)
	assert.Empty(t, plugins)
}

func TestLocalHarness_AgentOperations_NotAvailable(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// DelegateToAgent should return error
	_, err := h.DelegateToAgent(ctx, "other-agent", agent.Task{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// ListAgents should return empty list
	agents, err := h.ListAgents(ctx)
	assert.NoError(t, err)
	assert.Empty(t, agents)
}

func TestLocalHarness_FindingOperations_NotAvailable(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// Create a test finding
	testFinding := &finding.Finding{
		ID:          "test-id",
		Title:       "Test Finding",
		Description: "Test description",
		Severity:    finding.SeverityHigh,
		Category:    finding.CategoryPromptInjection,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// SubmitFinding should return error
	err := h.SubmitFinding(ctx, testFinding)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// GetFindings should return empty list
	findings, err := h.GetFindings(ctx, finding.Filter{})
	assert.NoError(t, err)
	assert.Empty(t, findings)
}

func TestLocalHarness_GraphRAGOperations_NotAvailable(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// QueryGraphRAG should return error
	_, err := h.QueryGraphRAG(ctx, graphrag.Query{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// FindSimilarAttacks should return error
	_, err = h.FindSimilarAttacks(ctx, "test", 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// FindSimilarFindings should return error
	_, err = h.FindSimilarFindings(ctx, "test-id", 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// GetAttackChains should return error
	_, err = h.GetAttackChains(ctx, "T1234", 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// GetRelatedFindings should return error
	_, err = h.GetRelatedFindings(ctx, "test-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// StoreGraphNode should return error
	_, err = h.StoreGraphNode(ctx, graphrag.GraphNode{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// CreateGraphRelationship should return error
	err = h.CreateGraphRelationship(ctx, graphrag.Relationship{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// StoreGraphBatch should return error
	_, err = h.StoreGraphBatch(ctx, graphrag.Batch{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// TraverseGraph should return error
	_, err = h.TraverseGraph(ctx, "node-id", graphrag.TraversalOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// GraphRAGHealth should return unhealthy
	health := h.GraphRAGHealth(ctx)
	assert.True(t, health.IsUnhealthy())
}

func TestLocalHarness_PlanningOperations_NotAvailable(t *testing.T) {
	h := newLocalHarness()
	ctx := context.Background()

	// PlanContext should return nil (interface)
	planCtx := h.PlanContext()
	assert.Nil(t, planCtx)

	// ReportStepHints should be no-op
	err := h.ReportStepHints(ctx, &planning.StepHints{})
	assert.NoError(t, err) // No-op is acceptable
}

func TestInMemoryStore_Concurrent(t *testing.T) {
	store := newInMemoryStore()
	ctx := context.Background()

	// Test concurrent writes via Working memory tier
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			key := "key-" + string(rune('a'+n))
			err := store.Working().Set(ctx, key, n)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all keys were written
	keys, err := store.Working().Keys(ctx)
	require.NoError(t, err)
	assert.Len(t, keys, 10)
}

func TestLocalTokenTracker_Concurrent(t *testing.T) {
	tracker := &localTokenTracker{usage: make(map[string]llm.TokenUsage)}

	// Test concurrent adds
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			tracker.Add("primary", llm.TokenUsage{
				InputTokens:  10,
				OutputTokens: 5,
				TotalTokens:  15,
			})
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify total
	total := tracker.Total()
	assert.Equal(t, 100, total.InputTokens)
	assert.Equal(t, 50, total.OutputTokens)
	assert.Equal(t, 150, total.TotalTokens)
}
