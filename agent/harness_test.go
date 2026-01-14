package agent

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

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

// mockHarness is a test implementation of the Harness interface.
type mockHarness struct {
	completeFunc          func(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error)
	completeWithToolsFunc func(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error)
	streamFunc            func(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error)
	callToolFunc          func(ctx context.Context, name string, input map[string]any) (map[string]any, error)
	listToolsFunc         func(ctx context.Context) ([]tool.Descriptor, error)
	queryPluginFunc       func(ctx context.Context, name string, method string, params map[string]any) (any, error)
	listPluginsFunc       func(ctx context.Context) ([]plugin.Descriptor, error)
	delegateToAgentFunc   func(ctx context.Context, name string, task Task) (Result, error)
	listAgentsFunc        func(ctx context.Context) ([]Descriptor, error)
	submitFindingFunc     func(ctx context.Context, f *finding.Finding) error
	getFindingsFunc       func(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)
	memoryStore           memory.Store
	mission               types.MissionContext
	target                types.TargetInfo
	tracer                trace.Tracer
	logger                *slog.Logger
	tokenUsage            llm.TokenTracker
}

func (m *mockHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, slot, messages, opts...)
	}
	return &llm.CompletionResponse{
		Content:      "mock response",
		FinishReason: "stop",
	}, nil
}

func (m *mockHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	if m.completeWithToolsFunc != nil {
		return m.completeWithToolsFunc(ctx, slot, messages, tools)
	}
	return &llm.CompletionResponse{
		Content:      "mock tool response",
		FinishReason: "stop",
	}, nil
}

func (m *mockHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, slot, messages)
	}
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Delta: "mock stream", FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func (m *mockHarness) CompleteStructured(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return map[string]any{"result": "structured"}, nil
}

func (m *mockHarness) CompleteStructuredAny(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return m.CompleteStructured(ctx, slot, messages, schema)
}

func (m *mockHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, name, input)
	}
	return map[string]any{"result": "success"}, nil
}

func (m *mockHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	if m.listToolsFunc != nil {
		return m.listToolsFunc(ctx)
	}
	return []tool.Descriptor{
		{Name: "tool1", Description: "Test tool 1"},
		{Name: "tool2", Description: "Test tool 2"},
	}, nil
}

func (m *mockHarness) CallToolsParallel(ctx context.Context, calls []ToolCall, maxConcurrency int) ([]ToolResult, error) {
	results := make([]ToolResult, len(calls))
	for i, call := range calls {
		output, err := m.CallTool(ctx, call.Name, call.Input)
		results[i] = ToolResult{
			Name:   call.Name,
			Output: output,
			Error:  err,
		}
	}
	return results, nil
}

func (m *mockHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	if m.queryPluginFunc != nil {
		return m.queryPluginFunc(ctx, name, method, params)
	}
	return map[string]any{"result": "plugin response"}, nil
}

func (m *mockHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	if m.listPluginsFunc != nil {
		return m.listPluginsFunc(ctx)
	}
	return []plugin.Descriptor{
		{Name: "plugin1", Description: "Test plugin", Version: "1.0.0"},
	}, nil
}

func (m *mockHarness) DelegateToAgent(ctx context.Context, name string, task Task) (Result, error) {
	if m.delegateToAgentFunc != nil {
		return m.delegateToAgentFunc(ctx, name, task)
	}
	return NewSuccessResult("delegated result"), nil
}

func (m *mockHarness) ListAgents(ctx context.Context) ([]Descriptor, error) {
	if m.listAgentsFunc != nil {
		return m.listAgentsFunc(ctx)
	}
	return []Descriptor{
		{Name: "agent1", Version: "1.0.0", Description: "Test agent"},
	}, nil
}

func (m *mockHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	if m.submitFindingFunc != nil {
		return m.submitFindingFunc(ctx, f)
	}
	return nil
}

func (m *mockHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	if m.getFindingsFunc != nil {
		return m.getFindingsFunc(ctx, filter)
	}
	return []*finding.Finding{}, nil
}

func (m *mockHarness) Memory() memory.Store {
	if m.memoryStore != nil {
		return m.memoryStore
	}
	return &mockMemoryStore{}
}

func (m *mockHarness) Mission() types.MissionContext {
	return m.mission
}

func (m *mockHarness) Target() types.TargetInfo {
	return m.target
}

func (m *mockHarness) Tracer() trace.Tracer {
	if m.tracer != nil {
		return m.tracer
	}
	return noop.NewTracerProvider().Tracer("test")
}

func (m *mockHarness) Logger() *slog.Logger {
	if m.logger != nil {
		return m.logger
	}
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (m *mockHarness) TokenUsage() llm.TokenTracker {
	if m.tokenUsage != nil {
		return m.tokenUsage
	}
	return llm.NewTokenTracker()
}

// GraphRAG methods - stubs for testing
func (m *mockHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}

func (m *mockHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	return nil, nil
}

func (m *mockHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (m *mockHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	return nil, nil
}

func (m *mockHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (m *mockHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "", nil
}

func (m *mockHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	return nil
}

func (m *mockHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	return nil, nil
}

func (m *mockHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	return nil, nil
}

func (m *mockHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("mock healthy")
}

// Planning methods - stubs for testing
func (m *mockHarness) PlanContext() planning.PlanningContext {
	return nil
}

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

func (m *mockHarness) QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error) {
	return nil, nil
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

// mockMemoryStore implements memory.Store for testing.
type mockMemoryStore struct {
	working *mockWorkingMemory
}

func (m *mockMemoryStore) Working() memory.WorkingMemory {
	if m.working == nil {
		m.working = &mockWorkingMemory{data: make(map[string]any)}
	}
	return m.working
}

func (m *mockMemoryStore) Mission() memory.MissionMemory {
	return &stubMissionMemory{}
}

func (m *mockMemoryStore) LongTerm() memory.LongTermMemory {
	return &stubLongTermMemory{}
}

// mockWorkingMemory implements memory.WorkingMemory for testing.
type mockWorkingMemory struct {
	data map[string]any
}

func (m *mockWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	if m.data == nil {
		return nil, errors.New("key not found")
	}
	val, ok := m.data[key]
	if !ok {
		return nil, errors.New("key not found")
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
	if m.data == nil {
		return nil
	}
	delete(m.data, key)
	return nil
}

func (m *mockWorkingMemory) Clear(ctx context.Context) error {
	m.data = make(map[string]any)
	return nil
}

func (m *mockWorkingMemory) Keys(ctx context.Context) ([]string, error) {
	if m.data == nil {
		return []string{}, nil
	}
	var keys []string
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// stubMissionMemory is a stub for mission memory
type stubMissionMemory struct{}

func (s *stubMissionMemory) Get(ctx context.Context, key string) (*memory.Item, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	return memory.ErrNotImplemented
}

func (s *stubMissionMemory) Delete(ctx context.Context, key string) error {
	return memory.ErrNotImplemented
}

func (s *stubMissionMemory) Search(ctx context.Context, query string, limit int) ([]memory.Result, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) History(ctx context.Context, limit int) ([]memory.Item, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) GetPreviousRunValue(ctx context.Context, key string) (any, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) ContinuityMode() memory.MemoryContinuityMode {
	return memory.MemoryIsolated
}

// stubLongTermMemory is a stub for long-term memory
type stubLongTermMemory struct{}

func (s *stubLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	return "", memory.ErrNotImplemented
}

func (s *stubLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]memory.Result, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubLongTermMemory) Delete(ctx context.Context, id string) error {
	return memory.ErrNotImplemented
}

func TestMockHarness_Complete(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "test prompt"},
	}

	resp, err := harness.Complete(ctx, "primary", messages)
	if err != nil {
		t.Errorf("Complete() error = %v, want nil", err)
	}
	if resp == nil {
		t.Fatal("Complete() returned nil response")
	}
	if resp.Content != "mock response" {
		t.Errorf("Complete() content = %s, want 'mock response'", resp.Content)
	}
}

func TestMockHarness_CompleteWithCustomFunc(t *testing.T) {
	callCount := 0
	harness := &mockHarness{
		completeFunc: func(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
			callCount++
			return &llm.CompletionResponse{Content: "custom response"}, nil
		},
	}

	ctx := context.Background()
	messages := []llm.Message{{Role: llm.RoleUser, Content: "test"}}

	resp, err := harness.Complete(ctx, "primary", messages)
	if err != nil {
		t.Errorf("Complete() error = %v", err)
	}
	if resp.Content != "custom response" {
		t.Errorf("Complete() content = %s, want 'custom response'", resp.Content)
	}
	if callCount != 1 {
		t.Errorf("custom function called %d times, want 1", callCount)
	}
}

func TestMockHarness_CompleteWithTools(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	messages := []llm.Message{{Role: llm.RoleUser, Content: "test"}}
	tools := []llm.ToolDef{
		{Name: "test-tool", Description: "A test tool"},
	}

	resp, err := harness.CompleteWithTools(ctx, "primary", messages, tools)
	if err != nil {
		t.Errorf("CompleteWithTools() error = %v", err)
	}
	if resp == nil {
		t.Fatal("CompleteWithTools() returned nil")
	}
}

func TestMockHarness_Stream(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	messages := []llm.Message{{Role: llm.RoleUser, Content: "test"}}

	ch, err := harness.Stream(ctx, "primary", messages)
	if err != nil {
		t.Errorf("Stream() error = %v", err)
	}

	chunk := <-ch
	if chunk.Delta != "mock stream" {
		t.Errorf("Stream() chunk = %s, want 'mock stream'", chunk.Delta)
	}
}

func TestMockHarness_CallTool(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	result, err := harness.CallTool(ctx, "test-tool", map[string]any{"param": "value"})
	if err != nil {
		t.Errorf("CallTool() error = %v", err)
	}
	if result == nil {
		t.Fatal("CallTool() returned nil")
	}
}

func TestMockHarness_ListTools(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	tools, err := harness.ListTools(ctx)
	if err != nil {
		t.Errorf("ListTools() error = %v", err)
	}
	if len(tools) != 2 {
		t.Errorf("ListTools() returned %d tools, want 2", len(tools))
	}
}

func TestMockHarness_QueryPlugin(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	result, err := harness.QueryPlugin(ctx, "plugin1", "method1", map[string]any{})
	if err != nil {
		t.Errorf("QueryPlugin() error = %v", err)
	}
	if result == nil {
		t.Fatal("QueryPlugin() returned nil")
	}
}

func TestMockHarness_ListPlugins(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	plugins, err := harness.ListPlugins(ctx)
	if err != nil {
		t.Errorf("ListPlugins() error = %v", err)
	}
	if len(plugins) != 1 {
		t.Errorf("ListPlugins() returned %d plugins, want 1", len(plugins))
	}
}

func TestMockHarness_DelegateToAgent(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	task := NewTask("task-1")
	result, err := harness.DelegateToAgent(ctx, "agent1", *task)
	if err != nil {
		t.Errorf("DelegateToAgent() error = %v", err)
	}
	if result.Status != StatusSuccess {
		t.Errorf("DelegateToAgent() status = %v, want %v", result.Status, StatusSuccess)
	}
}

func TestMockHarness_ListAgents(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	agents, err := harness.ListAgents(ctx)
	if err != nil {
		t.Errorf("ListAgents() error = %v", err)
	}
	if len(agents) != 1 {
		t.Errorf("ListAgents() returned %d agents, want 1", len(agents))
	}
}

func TestMockHarness_SubmitFinding(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	f := &finding.Finding{
		ID:       "finding-1",
		Severity: finding.SeverityHigh,
		Category: finding.CategoryJailbreak,
	}

	err := harness.SubmitFinding(ctx, f)
	if err != nil {
		t.Errorf("SubmitFinding() error = %v", err)
	}
}

func TestMockHarness_GetFindings(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	filter := finding.Filter{
		MissionID: "mission-1",
	}

	findings, err := harness.GetFindings(ctx, filter)
	if err != nil {
		t.Errorf("GetFindings() error = %v", err)
	}
	if findings == nil {
		t.Fatal("GetFindings() returned nil")
	}
}

func TestMockHarness_Memory(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	mem := harness.Memory()
	if mem == nil {
		t.Fatal("Memory() returned nil")
	}

	// Test working memory operations through the tiered API
	working := mem.Working()
	if working == nil {
		t.Fatal("Memory().Working() returned nil")
	}

	err := working.Set(ctx, "key1", "value1")
	if err != nil {
		t.Errorf("Memory.Working().Set() error = %v", err)
	}

	val, err := working.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Memory.Working().Get() error = %v", err)
	}
	if val != "value1" {
		t.Errorf("Memory.Working().Get() = %v, want 'value1'", val)
	}

	err = working.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Memory.Working().Delete() error = %v", err)
	}

	_, err = working.Get(ctx, "key1")
	if err == nil {
		t.Error("Memory.Working().Get() after delete should return error")
	}
}

func TestMockHarness_MissionAndTarget(t *testing.T) {
	mission := types.MissionContext{
		ID:   "mission-1",
		Name: "Test Mission",
	}
	target := types.TargetInfo{
		ID:   "target-1",
		Name: "Test Target",
		Type: string("llm_chat"),
	}

	harness := &mockHarness{
		mission: mission,
		target:  target,
	}

	if harness.Mission().ID != "mission-1" {
		t.Errorf("Mission().ID = %s, want mission-1", harness.Mission().ID)
	}
	if harness.Target().ID != "target-1" {
		t.Errorf("Target().ID = %s, want target-1", harness.Target().ID)
	}
}

func TestMockHarness_Observability(t *testing.T) {
	harness := &mockHarness{}

	tracer := harness.Tracer()
	if tracer == nil {
		t.Error("Tracer() returned nil")
	}

	logger := harness.Logger()
	if logger == nil {
		t.Error("Logger() returned nil")
	}

	tracker := harness.TokenUsage()
	if tracker == nil {
		t.Error("TokenUsage() returned nil")
	}
}

func TestMockMemoryStore_Keys(t *testing.T) {
	mem := &mockWorkingMemory{data: make(map[string]any)}
	ctx := context.Background()

	// Set some keys
	mem.Set(ctx, "app:key1", "value1")
	mem.Set(ctx, "app:key2", "value2")
	mem.Set(ctx, "other:key3", "value3")

	// Get all keys
	keys, err := mem.Keys(ctx)
	if err != nil {
		t.Errorf("Keys() error = %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}
}
