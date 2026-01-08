package serve

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCallbackClient tests the callback client constructor.
func TestNewCallbackClient(t *testing.T) {
	t.Run("valid endpoint", func(t *testing.T) {
		client, err := NewCallbackClient("localhost:50051")
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "localhost:50051", client.endpoint)
	})

	t.Run("empty endpoint", func(t *testing.T) {
		client, err := NewCallbackClient("")
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("with TLS config", func(t *testing.T) {
		tlsConf := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
		client, err := NewCallbackClient("localhost:50051", WithCallbackTLS(tlsConf))
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, tlsConf, client.tlsConf)
	})

	t.Run("with token", func(t *testing.T) {
		token := "test-token-123"
		client, err := NewCallbackClient("localhost:50051", WithCallbackToken(token))
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, token, client.token)
	})
}

// TestCallbackClientSetTaskContext tests the SetTaskContext method.
func TestCallbackClientSetTaskContext(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	client.SetTaskContext("task-123", "test-agent", "mission-abc", "trace-456", "span-789")

	assert.Equal(t, "task-123", client.taskID)
	assert.Equal(t, "test-agent", client.agentName)
	assert.Equal(t, "mission-abc", client.missionID)
	assert.Equal(t, "trace-456", client.traceID)
	assert.Equal(t, "span-789", client.spanID)
}

// TestCallbackClientContextInfo tests the contextInfo method.
func TestCallbackClientContextInfo(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	client.SetTaskContext("task-123", "test-agent", "mission-abc", "trace-456", "span-789")

	ctx := client.contextInfo()
	assert.NotNil(t, ctx)
	assert.Equal(t, "task-123", ctx.TaskId)
	assert.Equal(t, "test-agent", ctx.AgentName)
	assert.Equal(t, "mission-abc", ctx.MissionId)
	assert.Equal(t, "trace-456", ctx.TraceId)
	assert.Equal(t, "span-789", ctx.SpanId)
}

// TestCallbackClientConnectionLifecycle tests connect/close lifecycle.
func TestCallbackClientConnectionLifecycle(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	// Initially not connected
	assert.False(t, client.IsConnected())

	// Close without connecting should work
	err = client.Close()
	assert.NoError(t, err)

	// Cannot connect after close
	ctx := context.Background()
	err = client.Connect(ctx)
	assert.Error(t, err)
}

// TestCallbackClientClose tests the Close method.
func TestCallbackClientClose(t *testing.T) {
	t.Run("close without connect", func(t *testing.T) {
		client, err := NewCallbackClient("localhost:50051")
		require.NoError(t, err)

		err = client.Close()
		assert.NoError(t, err)
	})

	t.Run("double close", func(t *testing.T) {
		client, err := NewCallbackClient("localhost:50051")
		require.NoError(t, err)

		err = client.Close()
		assert.NoError(t, err)

		// Second close should still work
		err = client.Close()
		assert.NoError(t, err)
	})
}

// TestCallbackClientConcurrency tests concurrent access to the client.
func TestCallbackClientConcurrency(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	// Set context from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			taskID := time.Now().String()
			client.SetTaskContext(taskID, "agent", "", "", "")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Client should still be usable
	assert.NotNil(t, client)
}

// TestCallbackClientNotConnected tests that RPCs fail when not connected.
func TestCallbackClientNotConnected(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	ctx := context.Background()

	// LLMComplete should fail when not connected
	_, err = client.LLMComplete(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	// CallTool should fail when not connected
	_, err = client.CallTool(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// TestCallbackClientNotConnectedErrorMessages verifies that all methods return
// distinct error messages with method names when the client is not connected.
func TestCallbackClientNotConnectedErrorMessages(t *testing.T) {
	client, err := NewCallbackClient("localhost:50051")
	require.NoError(t, err)

	ctx := context.Background()

	// Test each method and verify it returns a distinct error message
	testCases := []struct {
		name        string
		call        func() error
		expectedMsg string
	}{
		{
			name: "LLMComplete",
			call: func() error {
				_, err := client.LLMComplete(ctx, nil)
				return err
			},
			expectedMsg: "LLMComplete: client not connected",
		},
		{
			name: "LLMCompleteWithTools",
			call: func() error {
				_, err := client.LLMCompleteWithTools(ctx, nil)
				return err
			},
			expectedMsg: "LLMCompleteWithTools: client not connected",
		},
		{
			name: "LLMStream",
			call: func() error {
				_, err := client.LLMStream(ctx, nil)
				return err
			},
			expectedMsg: "LLMStream: client not connected",
		},
		{
			name: "CallTool",
			call: func() error {
				_, err := client.CallTool(ctx, nil)
				return err
			},
			expectedMsg: "CallTool: client not connected",
		},
		{
			name: "ListTools",
			call: func() error {
				_, err := client.ListTools(ctx, nil)
				return err
			},
			expectedMsg: "ListTools: client not connected",
		},
		{
			name: "QueryPlugin",
			call: func() error {
				_, err := client.QueryPlugin(ctx, nil)
				return err
			},
			expectedMsg: "QueryPlugin: client not connected",
		},
		{
			name: "ListPlugins",
			call: func() error {
				_, err := client.ListPlugins(ctx, nil)
				return err
			},
			expectedMsg: "ListPlugins: client not connected",
		},
		{
			name: "DelegateToAgent",
			call: func() error {
				_, err := client.DelegateToAgent(ctx, nil)
				return err
			},
			expectedMsg: "DelegateToAgent: client not connected",
		},
		{
			name: "ListAgents",
			call: func() error {
				_, err := client.ListAgents(ctx, nil)
				return err
			},
			expectedMsg: "ListAgents: client not connected",
		},
		{
			name: "SubmitFinding",
			call: func() error {
				_, err := client.SubmitFinding(ctx, nil)
				return err
			},
			expectedMsg: "SubmitFinding: client not connected",
		},
		{
			name: "GetFindings",
			call: func() error {
				_, err := client.GetFindings(ctx, nil)
				return err
			},
			expectedMsg: "GetFindings: client not connected",
		},
		{
			name: "MemoryGet",
			call: func() error {
				_, err := client.MemoryGet(ctx, nil)
				return err
			},
			expectedMsg: "MemoryGet: client not connected",
		},
		{
			name: "MemorySet",
			call: func() error {
				_, err := client.MemorySet(ctx, nil)
				return err
			},
			expectedMsg: "MemorySet: client not connected",
		},
		{
			name: "MemoryDelete",
			call: func() error {
				_, err := client.MemoryDelete(ctx, nil)
				return err
			},
			expectedMsg: "MemoryDelete: client not connected",
		},
		{
			name: "MemoryList",
			call: func() error {
				_, err := client.MemoryList(ctx, nil)
				return err
			},
			expectedMsg: "MemoryList: client not connected",
		},
		{
			name: "GraphRAGQuery",
			call: func() error {
				_, err := client.GraphRAGQuery(ctx, nil)
				return err
			},
			expectedMsg: "GraphRAGQuery: client not connected",
		},
		{
			name: "FindSimilarAttacks",
			call: func() error {
				_, err := client.FindSimilarAttacks(ctx, nil)
				return err
			},
			expectedMsg: "FindSimilarAttacks: client not connected",
		},
		{
			name: "FindSimilarFindings",
			call: func() error {
				_, err := client.FindSimilarFindings(ctx, nil)
				return err
			},
			expectedMsg: "FindSimilarFindings: client not connected",
		},
		{
			name: "GetAttackChains",
			call: func() error {
				_, err := client.GetAttackChains(ctx, nil)
				return err
			},
			expectedMsg: "GetAttackChains: client not connected",
		},
		{
			name: "GetRelatedFindings",
			call: func() error {
				_, err := client.GetRelatedFindings(ctx, nil)
				return err
			},
			expectedMsg: "GetRelatedFindings: client not connected",
		},
		{
			name: "StoreGraphNode",
			call: func() error {
				_, err := client.StoreGraphNode(ctx, nil)
				return err
			},
			expectedMsg: "StoreGraphNode: client not connected",
		},
		{
			name: "CreateGraphRelationship",
			call: func() error {
				_, err := client.CreateGraphRelationship(ctx, nil)
				return err
			},
			expectedMsg: "CreateGraphRelationship: client not connected",
		},
		{
			name: "StoreGraphBatch",
			call: func() error {
				_, err := client.StoreGraphBatch(ctx, nil)
				return err
			},
			expectedMsg: "StoreGraphBatch: client not connected",
		},
		{
			name: "TraverseGraph",
			call: func() error {
				_, err := client.TraverseGraph(ctx, nil)
				return err
			},
			expectedMsg: "TraverseGraph: client not connected",
		},
		{
			name: "GraphRAGHealth",
			call: func() error {
				_, err := client.GraphRAGHealth(ctx, nil)
				return err
			},
			expectedMsg: "GraphRAGHealth: client not connected",
		},
		{
			name: "GetPlanContext",
			call: func() error {
				_, err := client.GetPlanContext(ctx, nil)
				return err
			},
			expectedMsg: "GetPlanContext: client not connected",
		},
		{
			name: "ReportStepHints",
			call: func() error {
				_, err := client.ReportStepHints(ctx, nil)
				return err
			},
			expectedMsg: "ReportStepHints: client not connected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			require.Error(t, err, "expected error for %s when not connected", tc.name)
			assert.Equal(t, tc.expectedMsg, err.Error(),
				"error message for %s should include method name", tc.name)
		})
	}
}
