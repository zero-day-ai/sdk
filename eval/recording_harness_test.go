package eval

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/graphragpb"
	"github.com/zero-day-ai/sdk/api/gen/toolspb"
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
	protolib "google.golang.org/protobuf/proto"
)

// mockHarness is a minimal mock implementation of agent.Harness for testing.
type mockHarness struct {
	completeFunc      func(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error)
	callToolProtoFunc func(ctx context.Context, name string, request protolib.Message, response protolib.Message) error
	submitFindingFunc func(ctx context.Context, f *finding.Finding) error
	memStore          memory.Store
}

func (m *mockHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, slot, messages, opts...)
	}
	return &llm.CompletionResponse{Content: "mock response"}, nil
}

func (m *mockHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{Content: "mock tool response"}, nil
}

func (m *mockHarness) CompleteStructured(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return map[string]any{"result": "structured"}, nil
}

func (m *mockHarness) CompleteStructuredAny(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return m.CompleteStructured(ctx, slot, messages, schema)
}

func (m *mockHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	ch := make(chan llm.StreamChunk)
	close(ch)
	return ch, nil
}

func (m *mockHarness) CallToolProto(ctx context.Context, name string, request protolib.Message, response protolib.Message) error {
	if m.callToolProtoFunc != nil {
		return m.callToolProtoFunc(ctx, name, request, response)
	}
	return nil
}

func (m *mockHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	return []tool.Descriptor{}, nil
}

func (m *mockHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	return "mock plugin result", nil
}

func (m *mockHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	return []plugin.Descriptor{}, nil
}

func (m *mockHarness) DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error) {
	return agent.NewSuccessResult("mock delegation"), nil
}

func (m *mockHarness) ListAgents(ctx context.Context) ([]agent.Descriptor, error) {
	return []agent.Descriptor{}, nil
}

func (m *mockHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	if m.submitFindingFunc != nil {
		return m.submitFindingFunc(ctx, f)
	}
	return nil
}

func (m *mockHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (m *mockHarness) Memory() memory.Store {
	if m.memStore != nil {
		return m.memStore
	}
	return &mockMemoryStore{}
}

func (m *mockHarness) Mission() types.MissionContext {
	return types.MissionContext{}
}

func (m *mockHarness) Target() types.TargetInfo {
	return types.TargetInfo{}
}

func (m *mockHarness) Tracer() trace.Tracer {
	return noop.NewTracerProvider().Tracer("test")
}

func (m *mockHarness) Logger() *slog.Logger {
	return slog.Default()
}

func (m *mockHarness) TokenUsage() llm.TokenTracker {
	return nil
}

func (m *mockHarness) QueryNodes(ctx context.Context, query *graphragpb.GraphQuery) ([]*graphragpb.QueryResult, error) {
	return nil, nil
}

func (m *mockHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return []graphrag.Result{}, nil
}

func (m *mockHarness) QuerySemantic(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return []graphrag.Result{}, nil
}

func (m *mockHarness) QueryStructured(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return []graphrag.Result{}, nil
}

func (m *mockHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	return []graphrag.AttackPattern{}, nil
}

func (m *mockHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	return []graphrag.FindingNode{}, nil
}

func (m *mockHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	return []graphrag.AttackChain{}, nil
}

func (m *mockHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	return []graphrag.FindingNode{}, nil
}

func (m *mockHarness) StoreNode(ctx context.Context, node *graphragpb.GraphNode) (string, error) {
	return "node-123", nil
}

func (m *mockHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "node-123", nil
}

func (m *mockHarness) StoreSemantic(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "node-123", nil
}

func (m *mockHarness) StoreStructured(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "node-123", nil
}

func (m *mockHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	return nil
}

func (m *mockHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	return []string{"node-1", "node-2"}, nil
}

func (m *mockHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	return []graphrag.TraversalResult{}, nil
}

func (m *mockHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	return types.HealthStatus{Status: "healthy"}
}

// PlanContext returns the planning context for the current execution.
func (m *mockHarness) PlanContext() planning.PlanningContext {
	return nil
}

// ReportStepHints allows agents to provide feedback to the planning system.
func (m *mockHarness) ReportStepHints(ctx context.Context, hints *planning.StepHints) error {
	return nil
}

// Mission Execution Context methods - stubs for testing
func (m *mockHarness) MissionExecutionContext() types.MissionExecutionContext {
	return types.MissionExecutionContext{}
}

func (m *mockHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
	return []types.MissionRunSummary{}, nil
}

func (m *mockHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (m *mockHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

// MissionManager methods - stubs for testing
func (m *mockHarness) CreateMission(ctx context.Context, workflow any, targetID string, opts *mission.CreateMissionOpts) (*mission.MissionInfo, error) {
	return &mission.MissionInfo{
		ID:       "mock-mission-id",
		Name:     "mock-mission",
		Status:   mission.MissionStatusPending,
		TargetID: targetID,
	}, nil
}

func (m *mockHarness) RunMission(ctx context.Context, missionID string, opts *mission.RunMissionOpts) error {
	return nil
}

func (m *mockHarness) GetMissionStatus(ctx context.Context, missionID string) (*mission.MissionStatusInfo, error) {
	return &mission.MissionStatusInfo{
		Status:   mission.MissionStatusRunning,
		Progress: 0.5,
	}, nil
}

func (m *mockHarness) WaitForMission(ctx context.Context, missionID string, timeout time.Duration) (*mission.MissionResult, error) {
	return &mission.MissionResult{
		MissionID: missionID,
		Status:    mission.MissionStatusCompleted,
	}, nil
}

func (m *mockHarness) ListMissions(ctx context.Context, filter *mission.MissionFilter) ([]*mission.MissionInfo, error) {
	return []*mission.MissionInfo{}, nil
}

func (m *mockHarness) CancelMission(ctx context.Context, missionID string) error {
	return nil
}

func (m *mockHarness) GetMissionResults(ctx context.Context, missionID string) (*mission.MissionResult, error) {
	return &mission.MissionResult{
		MissionID: missionID,
		Status:    mission.MissionStatusCompleted,
	}, nil
}

// GetCredential returns a mock credential for testing
func (m *mockHarness) GetCredential(ctx context.Context, name string) (*types.Credential, error) {
	return &types.Credential{
		Name:   name,
		Type:   "api-key",
		Secret: "mock-secret-value",
	}, nil
}

// mockMemoryStore is a minimal mock implementation of memory.Store.
type mockMemoryStore struct {
	workingData map[string]any
}

func (m *mockMemoryStore) Working() memory.WorkingMemory {
	return &mockWorkingMemory{data: m.workingData}
}

func (m *mockMemoryStore) Mission() memory.MissionMemory {
	return &mockMissionMemory{}
}

func (m *mockMemoryStore) LongTerm() memory.LongTermMemory {
	return &mockLongTermMemory{}
}

// mockWorkingMemory implements memory.WorkingMemory
type mockWorkingMemory struct {
	data map[string]any
}

func (m *mockWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	if m.data == nil {
		return nil, memory.ErrNotFound
	}
	val, ok := m.data[key]
	if !ok {
		return nil, memory.ErrNotFound
	}
	return val, nil
}

func (m *mockWorkingMemory) Set(ctx context.Context, key string, value any) error {
	if m.data == nil {
		m.data = make(map[string]any)
	}
	m.data[key] = value
	return nil
}

func (m *mockWorkingMemory) Delete(ctx context.Context, key string) error {
	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func (m *mockWorkingMemory) Clear(ctx context.Context) error {
	m.data = make(map[string]any)
	return nil
}

func (m *mockWorkingMemory) Keys(ctx context.Context) ([]string, error) {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// mockMissionMemory implements memory.MissionMemory
type mockMissionMemory struct{}

func (m *mockMissionMemory) Get(ctx context.Context, key string) (*memory.Item, error) {
	return nil, memory.ErrNotFound
}

func (m *mockMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	return nil
}

func (m *mockMissionMemory) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockMissionMemory) Search(ctx context.Context, query string, limit int) ([]memory.Result, error) {
	return nil, nil
}

func (m *mockMissionMemory) History(ctx context.Context, limit int) ([]memory.Item, error) {
	return nil, nil
}

func (m *mockMissionMemory) GetPreviousRunValue(ctx context.Context, key string) (any, error) {
	return nil, memory.ErrNotFound
}

func (m *mockMissionMemory) GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error) {
	return nil, nil
}

func (m *mockMissionMemory) ContinuityMode() memory.MemoryContinuityMode {
	return memory.MemoryIsolated
}

// mockLongTermMemory implements memory.LongTermMemory
type mockLongTermMemory struct{}

func (m *mockLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	return "mock-id", nil
}

func (m *mockLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]memory.Result, error) {
	return nil, nil
}

func (m *mockLongTermMemory) Delete(ctx context.Context, id string) error {
	return nil
}

// TestRecordingHarnessBasics tests basic recording harness functionality.
func TestRecordingHarnessBasics(t *testing.T) {
	mock := &mockHarness{}
	recorder := NewRecordingHarness(mock)

	// Initially, trajectory should have no steps
	traj := recorder.Trajectory()
	assert.Len(t, traj.Steps, 0)
	assert.False(t, traj.StartTime.IsZero())

	// After reset, trajectory should be cleared
	recorder.Reset()
	traj = recorder.Trajectory()
	assert.Len(t, traj.Steps, 0)
}

// TestRecordingHarnessLLMCalls tests recording of LLM completion calls.
func TestRecordingHarnessLLMCalls(t *testing.T) {
	ctx := context.Background()
	mock := &mockHarness{}
	recorder := NewRecordingHarness(mock)

	// Call Complete
	messages := []llm.Message{{Role: "user", Content: "test"}}
	resp, err := recorder.Complete(ctx, "primary", messages)
	require.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify trajectory recorded the call
	traj := recorder.Trajectory()
	require.Len(t, traj.Steps, 1)

	step := traj.Steps[0]
	assert.Equal(t, "llm", step.Type)
	assert.Equal(t, "primary", step.Name)
	assert.NotNil(t, step.Input)
	assert.NotNil(t, step.Output)
	assert.Empty(t, step.Error)
	assert.False(t, step.StartTime.IsZero())
	assert.Greater(t, step.Duration, time.Duration(0))
}

// TestRecordingHarnessToolCalls tests recording of tool invocations.
func TestRecordingHarnessToolCalls(t *testing.T) {
	ctx := context.Background()
	mock := &mockHarness{}
	recorder := NewRecordingHarness(mock)

	// Call a tool with proto messages
	req := &toolspb.HttpxRequest{Targets: []string{"https://example.com"}}
	resp := &toolspb.HttpxResponse{}
	err := recorder.CallToolProto(ctx, "httpx", req, resp)
	require.NoError(t, err)

	// Verify trajectory recorded the call
	traj := recorder.Trajectory()
	require.Len(t, traj.Steps, 1)

	step := traj.Steps[0]
	assert.Equal(t, "tool", step.Type)
	assert.Equal(t, "httpx", step.Name)
	assert.NotNil(t, step.Input)
	assert.NotNil(t, step.Output)
	assert.Empty(t, step.Error)
	assert.Greater(t, step.Duration, time.Duration(0))
}

// TestRecordingHarnessErrorRecording tests recording of errors.
func TestRecordingHarnessErrorRecording(t *testing.T) {
	ctx := context.Background()

	expectedErr := errors.New("mock error")
	mock := &mockHarness{
		callToolProtoFunc: func(ctx context.Context, name string, request protolib.Message, response protolib.Message) error {
			return expectedErr
		},
	}
	recorder := NewRecordingHarness(mock)

	// Call a tool that returns an error
	req := &toolspb.HttpxRequest{Targets: []string{"example.com"}}
	resp := &toolspb.HttpxResponse{}
	err := recorder.CallToolProto(ctx, "failing-tool", req, resp)
	require.Error(t, err)

	// Verify trajectory recorded the error
	traj := recorder.Trajectory()
	require.Len(t, traj.Steps, 1)

	step := traj.Steps[0]
	assert.Equal(t, "tool", step.Type)
	assert.Equal(t, "failing-tool", step.Name)
	assert.Equal(t, "mock error", step.Error)
}

// TestRecordingHarnessFindingSubmission tests recording of finding submissions.
func TestRecordingHarnessFindingSubmission(t *testing.T) {
	ctx := context.Background()
	mock := &mockHarness{}
	recorder := NewRecordingHarness(mock)

	// Submit a finding
	f := &finding.Finding{
		ID:       "finding-1",
		Severity: finding.SeverityHigh,
		Category: "injection",
	}
	err := recorder.SubmitFinding(ctx, f)
	require.NoError(t, err)

	// Verify trajectory recorded the submission
	traj := recorder.Trajectory()
	require.Len(t, traj.Steps, 1)

	step := traj.Steps[0]
	assert.Equal(t, "finding", step.Type)
	assert.Equal(t, "submit", step.Name)
	assert.NotNil(t, step.Input)
	assert.Empty(t, step.Error)
}

// TestRecordingHarnessMemoryOperations tests recording of memory operations.
func TestRecordingHarnessMemoryOperations(t *testing.T) {
	ctx := context.Background()
	mock := &mockHarness{
		memStore: &mockMemoryStore{workingData: make(map[string]any)},
	}
	recorder := NewRecordingHarness(mock)

	mem := recorder.Memory().Working()

	// Set a value
	err := mem.Set(ctx, "test-key", "test-value")
	require.NoError(t, err)

	// Get the value
	val, err := mem.Get(ctx, "test-key")
	require.NoError(t, err)
	assert.Equal(t, "test-value", val)

	// Delete the value
	err = mem.Delete(ctx, "test-key")
	require.NoError(t, err)

	// Verify trajectory recorded all operations
	traj := recorder.Trajectory()
	require.Len(t, traj.Steps, 3)

	assert.Equal(t, "memory.working", traj.Steps[0].Type)
	assert.Equal(t, "set", traj.Steps[0].Name)

	assert.Equal(t, "memory.working", traj.Steps[1].Type)
	assert.Equal(t, "get", traj.Steps[1].Name)

	assert.Equal(t, "memory.working", traj.Steps[2].Type)
	assert.Equal(t, "delete", traj.Steps[2].Name)
}

// TestRecordingHarnessMultipleOperations tests recording of multiple operations.
func TestRecordingHarnessMultipleOperations(t *testing.T) {
	ctx := context.Background()
	mock := &mockHarness{}
	recorder := NewRecordingHarness(mock)

	// Perform multiple operations
	_, _ = recorder.Complete(ctx, "primary", []llm.Message{{Role: "user", Content: "test"}})
	_ = recorder.CallToolProto(ctx, "httpx", &toolspb.HttpxRequest{Targets: []string{"test"}}, &toolspb.HttpxResponse{})
	_ = recorder.SubmitFinding(ctx, &finding.Finding{ID: "f1"})

	// Verify all operations were recorded
	traj := recorder.Trajectory()
	require.Len(t, traj.Steps, 3)

	assert.Equal(t, "llm", traj.Steps[0].Type)
	assert.Equal(t, "tool", traj.Steps[1].Type)
	assert.Equal(t, "finding", traj.Steps[2].Type)

	// Verify end time is set
	assert.False(t, traj.EndTime.IsZero())
	assert.True(t, traj.EndTime.After(traj.StartTime) || traj.EndTime.Equal(traj.StartTime))
}

// TestRecordingHarnessReset tests resetting the trajectory.
func TestRecordingHarnessReset(t *testing.T) {
	ctx := context.Background()
	mock := &mockHarness{}
	recorder := NewRecordingHarness(mock)

	// Perform some operations
	_, _ = recorder.Complete(ctx, "primary", []llm.Message{{Role: "user", Content: "test"}})
	traj := recorder.Trajectory()
	require.Len(t, traj.Steps, 1)

	// Reset the trajectory
	recorder.Reset()
	traj = recorder.Trajectory()
	assert.Len(t, traj.Steps, 0)
	assert.False(t, traj.StartTime.IsZero())
}

// TestRecordingHarnessThreadSafety tests thread-safe trajectory recording.
func TestRecordingHarnessThreadSafety(t *testing.T) {
	ctx := context.Background()
	mock := &mockHarness{}
	recorder := NewRecordingHarness(mock)

	// Perform concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = recorder.Complete(ctx, "primary", []llm.Message{{Role: "user", Content: "test"}})
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all operations were recorded
	traj := recorder.Trajectory()
	assert.Len(t, traj.Steps, 10)
}
