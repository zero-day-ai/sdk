package agent

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// mockHarness is a test implementation of the Harness interface.
type mockHarness struct {
	completeFunc         func(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error)
	completeWithToolsFunc func(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error)
	streamFunc           func(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error)
	callToolFunc         func(ctx context.Context, name string, input map[string]any) (map[string]any, error)
	listToolsFunc        func(ctx context.Context) ([]ToolDescriptor, error)
	queryPluginFunc      func(ctx context.Context, name string, method string, params map[string]any) (any, error)
	listPluginsFunc      func(ctx context.Context) ([]PluginDescriptor, error)
	delegateToAgentFunc  func(ctx context.Context, name string, task Task) (Result, error)
	listAgentsFunc       func(ctx context.Context) ([]Descriptor, error)
	submitFindingFunc    func(ctx context.Context, f Finding) error
	getFindingsFunc      func(ctx context.Context, filter FindingFilter) ([]Finding, error)
	memory               MemoryStore
	mission              types.MissionContext
	target               types.TargetInfo
	tracer               trace.Tracer
	logger               *slog.Logger
	tokenUsage           llm.TokenTracker
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

func (m *mockHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, name, input)
	}
	return map[string]any{"result": "success"}, nil
}

func (m *mockHarness) ListTools(ctx context.Context) ([]ToolDescriptor, error) {
	if m.listToolsFunc != nil {
		return m.listToolsFunc(ctx)
	}
	return []ToolDescriptor{
		{Name: "tool1", Description: "Test tool 1"},
		{Name: "tool2", Description: "Test tool 2"},
	}, nil
}

func (m *mockHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	if m.queryPluginFunc != nil {
		return m.queryPluginFunc(ctx, name, method, params)
	}
	return map[string]any{"result": "plugin response"}, nil
}

func (m *mockHarness) ListPlugins(ctx context.Context) ([]PluginDescriptor, error) {
	if m.listPluginsFunc != nil {
		return m.listPluginsFunc(ctx)
	}
	return []PluginDescriptor{
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

func (m *mockHarness) SubmitFinding(ctx context.Context, f Finding) error {
	if m.submitFindingFunc != nil {
		return m.submitFindingFunc(ctx, f)
	}
	return nil
}

func (m *mockHarness) GetFindings(ctx context.Context, filter FindingFilter) ([]Finding, error) {
	if m.getFindingsFunc != nil {
		return m.getFindingsFunc(ctx, filter)
	}
	return []Finding{}, nil
}

func (m *mockHarness) Memory() MemoryStore {
	if m.memory != nil {
		return m.memory
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

// mockMemoryStore is a simple in-memory implementation of MemoryStore.
type mockMemoryStore struct {
	data map[string]any
}

func (m *mockMemoryStore) Get(ctx context.Context, key string) (any, error) {
	if m.data == nil {
		return nil, errors.New("key not found")
	}
	val, ok := m.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return val, nil
}

func (m *mockMemoryStore) Set(ctx context.Context, key string, value any) error {
	if m.data == nil {
		m.data = make(map[string]any)
	}
	m.data[key] = value
	return nil
}

func (m *mockMemoryStore) Delete(ctx context.Context, key string) error {
	if m.data == nil {
		return nil
	}
	delete(m.data, key)
	return nil
}

func (m *mockMemoryStore) List(ctx context.Context, prefix string) ([]string, error) {
	if m.data == nil {
		return []string{}, nil
	}
	var keys []string
	for k := range m.data {
		if len(prefix) == 0 || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// mockFinding is a simple implementation of Finding for testing.
type mockFinding struct {
	id       string
	severity string
	category string
}

func (m *mockFinding) ID() string       { return m.id }
func (m *mockFinding) Severity() string { return m.severity }
func (m *mockFinding) Category() string { return m.category }

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

	task := NewTask("task-1", "delegated task")
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

	finding := &mockFinding{
		id:       "finding-1",
		severity: "high",
		category: "jailbreak",
	}

	err := harness.SubmitFinding(ctx, finding)
	if err != nil {
		t.Errorf("SubmitFinding() error = %v", err)
	}
}

func TestMockHarness_GetFindings(t *testing.T) {
	harness := &mockHarness{}
	ctx := context.Background()

	filter := FindingFilter{
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

	// Test memory operations
	err := mem.Set(ctx, "key1", "value1")
	if err != nil {
		t.Errorf("Memory.Set() error = %v", err)
	}

	val, err := mem.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Memory.Get() error = %v", err)
	}
	if val != "value1" {
		t.Errorf("Memory.Get() = %v, want 'value1'", val)
	}

	err = mem.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Memory.Delete() error = %v", err)
	}

	_, err = mem.Get(ctx, "key1")
	if err == nil {
		t.Error("Memory.Get() after delete should return error")
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
		Type: types.TargetTypeLLMChat,
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

func TestMockMemoryStore_List(t *testing.T) {
	mem := &mockMemoryStore{}
	ctx := context.Background()

	// Set some keys
	mem.Set(ctx, "app:key1", "value1")
	mem.Set(ctx, "app:key2", "value2")
	mem.Set(ctx, "other:key3", "value3")

	// List with prefix
	keys, err := mem.List(ctx, "app:")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("List() returned %d keys, want 2", len(keys))
	}

	// List all
	allKeys, err := mem.List(ctx, "")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(allKeys) != 3 {
		t.Errorf("List() returned %d keys, want 3", len(allKeys))
	}
}
