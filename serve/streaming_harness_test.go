package serve

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/memory"
	"github.com/zero-day-ai/sdk/planning"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/tool"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// mockBidiStream implements grpc.BidiStreamingServer for testing.
// It captures all messages sent via Send() for verification in tests.
type mockBidiStream struct {
	grpc.ServerStream // embed for Context() and other methods

	// sentMessages captures all messages sent via Send()
	sentMessages []*proto.AgentMessage
	mu           sync.Mutex

	// sendErr can be set to simulate Send() failures
	sendErr error
}

// Send captures the message for later inspection in tests.
func (m *mockBidiStream) Send(msg *proto.AgentMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.sendErr != nil {
		return m.sendErr
	}

	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

// Recv is not used in these tests but required by the interface.
func (m *mockBidiStream) Recv() (*proto.ClientMessage, error) {
	return nil, errors.New("not implemented")
}

// getSentMessages returns a copy of sent messages for thread-safe inspection.
func (m *mockBidiStream) getSentMessages() []*proto.AgentMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	msgs := make([]*proto.AgentMessage, len(m.sentMessages))
	copy(msgs, m.sentMessages)
	return msgs
}

// clearSentMessages clears the captured messages.
func (m *mockBidiStream) clearSentMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentMessages = nil
}

// mockServerStream provides a minimal ServerStream implementation.
type mockServerStream struct{}

func (m *mockServerStream) SetHeader(md metadata.MD) error  { return nil }
func (m *mockServerStream) SendHeader(md metadata.MD) error { return nil }
func (m *mockServerStream) SetTrailer(md metadata.MD)       {}
func (m *mockServerStream) Context() context.Context        { return context.Background() }
func (m *mockServerStream) SendMsg(msg interface{}) error   { return nil }
func (m *mockServerStream) RecvMsg(msg interface{}) error   { return nil }

// mockStreamHarness is a minimal implementation of agent.Harness for testing.
type mockStreamHarness struct {
	callToolFunc      func(ctx context.Context, name string, input map[string]any) (map[string]any, error)
	submitFindingFunc func(ctx context.Context, f *finding.Finding) error
	completeFunc      func(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error)
	streamFunc        func(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error)
	logger            *slog.Logger
	tracer            trace.Tracer
}

func (m *mockStreamHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, slot, messages, opts...)
	}
	return &llm.CompletionResponse{Content: "mock response", FinishReason: "stop"}, nil
}

func (m *mockStreamHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	return &llm.CompletionResponse{Content: "mock response", FinishReason: "stop"}, nil
}

func (m *mockStreamHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, slot, messages)
	}
	ch := make(chan llm.StreamChunk, 1)
	ch <- llm.StreamChunk{Delta: "mock stream", FinishReason: "stop"}
	close(ch)
	return ch, nil
}

func (m *mockStreamHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	if m.callToolFunc != nil {
		return m.callToolFunc(ctx, name, input)
	}
	return map[string]any{"result": "success"}, nil
}

func (m *mockStreamHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	return []tool.Descriptor{}, nil
}

func (m *mockStreamHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	return nil, nil
}

func (m *mockStreamHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	return []plugin.Descriptor{}, nil
}

func (m *mockStreamHarness) DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error) {
	return agent.NewSuccessResult("delegated"), nil
}

func (m *mockStreamHarness) ListAgents(ctx context.Context) ([]agent.Descriptor, error) {
	return []agent.Descriptor{}, nil
}

func (m *mockStreamHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	if m.submitFindingFunc != nil {
		return m.submitFindingFunc(ctx, f)
	}
	return nil
}

func (m *mockStreamHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (m *mockStreamHarness) Memory() memory.Store {
	return &mockStreamMemoryStore{}
}

func (m *mockStreamHarness) Mission() types.MissionContext {
	return types.MissionContext{ID: "mission-1"}
}

func (m *mockStreamHarness) Target() types.TargetInfo {
	return types.TargetInfo{ID: "target-1"}
}

func (m *mockStreamHarness) Tracer() trace.Tracer {
	if m.tracer != nil {
		return m.tracer
	}
	return noop.NewTracerProvider().Tracer("test")
}

func (m *mockStreamHarness) Logger() *slog.Logger {
	if m.logger != nil {
		return m.logger
	}
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func (m *mockStreamHarness) TokenUsage() llm.TokenTracker {
	return llm.NewTokenTracker()
}

func (m *mockStreamHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	return nil, nil
}

func (m *mockStreamHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	return nil, nil
}

func (m *mockStreamHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (m *mockStreamHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	return nil, nil
}

func (m *mockStreamHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	return nil, nil
}

func (m *mockStreamHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	return "", nil
}

func (m *mockStreamHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	return nil
}

func (m *mockStreamHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	return nil, nil
}

func (m *mockStreamHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	return nil, nil
}

func (m *mockStreamHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	return types.HealthStatus{}
}

func (m *mockStreamHarness) PlanContext() planning.PlanningContext {
	return nil
}

func (m *mockStreamHarness) ReportStepHints(ctx context.Context, hints *planning.StepHints) error {
	return nil
}

// Mission Execution Context methods - stubs for testing
func (m *mockStreamHarness) MissionExecutionContext() types.MissionExecutionContext {
	return types.MissionExecutionContext{}
}

func (m *mockStreamHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
	return []types.MissionRunSummary{}, nil
}

func (m *mockStreamHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (m *mockStreamHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	return []*finding.Finding{}, nil
}

func (m *mockStreamHarness) QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error) {
	return nil, nil
}

// mockStreamMemoryStore implements memory.Store for testing.
type mockStreamMemoryStore struct{}

func (m *mockStreamMemoryStore) Working() memory.WorkingMemory {
	return &mockStreamWorkingMemory{}
}

func (m *mockStreamMemoryStore) Mission() memory.MissionMemory {
	return nil
}

func (m *mockStreamMemoryStore) LongTerm() memory.LongTermMemory {
	return nil
}

// mockStreamWorkingMemory implements memory.WorkingMemory for testing.
type mockStreamWorkingMemory struct{}

func (m *mockStreamWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	return nil, errors.New("not found")
}

func (m *mockStreamWorkingMemory) Set(ctx context.Context, key string, value any) error {
	return nil
}

func (m *mockStreamWorkingMemory) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockStreamWorkingMemory) Clear(ctx context.Context) error {
	return nil
}

func (m *mockStreamWorkingMemory) Keys(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

// TestNewStreamingHarness verifies that NewStreamingHarness creates a valid harness.
func TestNewStreamingHarness(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)
	mode := proto.AgentMode_AGENT_MODE_AUTONOMOUS

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, mode)

	if sh == nil {
		t.Fatal("NewStreamingHarness returned nil")
	}

	// Verify it implements agent.StreamingHarness interface
	_, ok := sh.(agent.StreamingHarness)
	if !ok {
		t.Error("NewStreamingHarness does not implement agent.StreamingHarness interface")
	}

	// Verify mode is set correctly (proto autonomous maps to agent autonomous)
	if sh.Mode() != agent.ExecutionModeAutonomous {
		t.Errorf("Mode() = %v, want %v", sh.Mode(), agent.ExecutionModeAutonomous)
	}

	// Verify steering channel is accessible
	if sh.Steering() == nil {
		t.Error("Steering() returned nil channel")
	}
}

// TestEmitOutput verifies that EmitOutput sends correct events to the stream.
func TestEmitOutput(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		isReasoning bool
	}{
		{
			name:        "standard output",
			content:     "Analysis complete",
			isReasoning: false,
		},
		{
			name:        "reasoning output",
			content:     "Thinking about next steps...",
			isReasoning: true,
		},
		{
			name:        "empty content",
			content:     "",
			isReasoning: false,
		},
		{
			name:        "multiline content",
			content:     "Line 1\nLine 2\nLine 3",
			isReasoning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHarness := &mockStreamHarness{}
			stream := &mockBidiStream{ServerStream: &mockServerStream{}}
			steeringCh := make(chan *proto.SteeringMessage, 10)

			sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

			err := sh.EmitOutput(tt.content, tt.isReasoning)
			if err != nil {
				t.Errorf("EmitOutput() error = %v, want nil", err)
			}

			msgs := stream.getSentMessages()
			if len(msgs) != 1 {
				t.Fatalf("sent %d messages, want 1", len(msgs))
			}

			msg := msgs[0]
			if msg.Sequence != 1 {
				t.Errorf("sequence = %d, want 1", msg.Sequence)
			}

			output, ok := msg.Payload.(*proto.AgentMessage_Output)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_Output", msg.Payload)
			}

			if output.Output.Content != tt.content {
				t.Errorf("content = %q, want %q", output.Output.Content, tt.content)
			}

			if output.Output.IsReasoning != tt.isReasoning {
				t.Errorf("isReasoning = %v, want %v", output.Output.IsReasoning, tt.isReasoning)
			}
		})
	}
}

// TestEmitToolCall verifies that EmitToolCall sends correct events.
func TestEmitToolCall(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	toolName := "kubectl"
	input := map[string]any{"command": "get", "resource": "pods"}
	callID := "call-123"

	err := sh.EmitToolCall(toolName, input, callID)
	if err != nil {
		t.Errorf("EmitToolCall() error = %v, want nil", err)
	}

	msgs := stream.getSentMessages()
	if len(msgs) != 1 {
		t.Fatalf("sent %d messages, want 1", len(msgs))
	}

	msg := msgs[0]
	toolCall, ok := msg.Payload.(*proto.AgentMessage_ToolCall)
	if !ok {
		t.Fatalf("payload type = %T, want *proto.AgentMessage_ToolCall", msg.Payload)
	}

	if toolCall.ToolCall.ToolName != toolName {
		t.Errorf("toolName = %q, want %q", toolCall.ToolCall.ToolName, toolName)
	}

	if toolCall.ToolCall.CallId != callID {
		t.Errorf("callId = %q, want %q", toolCall.ToolCall.CallId, callID)
	}

	// Verify input was serialized to JSON
	if toolCall.ToolCall.InputJson == "" {
		t.Error("inputJson is empty")
	}
}

// TestEmitToolResult verifies that EmitToolResult sends correct events.
func TestEmitToolResult(t *testing.T) {
	tests := []struct {
		name    string
		callID  string
		output  map[string]any
		err     error
		success bool // expected success in proto
	}{
		{
			name:    "successful result",
			callID:  "call-123",
			output:  map[string]any{"status": "ok", "pods": []string{"pod-1", "pod-2"}},
			err:     nil,
			success: true,
		},
		{
			name:    "failed result",
			callID:  "call-456",
			output:  map[string]any{"error": "connection timeout"},
			err:     errors.New("connection timeout"),
			success: false,
		},
		{
			name:    "empty output",
			callID:  "call-789",
			output:  map[string]any{},
			err:     nil,
			success: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHarness := &mockStreamHarness{}
			stream := &mockBidiStream{ServerStream: &mockServerStream{}}
			steeringCh := make(chan *proto.SteeringMessage, 10)

			sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

			err := sh.EmitToolResult(tt.output, tt.err, tt.callID)
			if err != nil {
				t.Errorf("EmitToolResult() error = %v, want nil", err)
			}

			msgs := stream.getSentMessages()
			if len(msgs) != 1 {
				t.Fatalf("sent %d messages, want 1", len(msgs))
			}

			msg := msgs[0]
			toolResult, ok := msg.Payload.(*proto.AgentMessage_ToolResult)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_ToolResult", msg.Payload)
			}

			if toolResult.ToolResult.CallId != tt.callID {
				t.Errorf("callId = %q, want %q", toolResult.ToolResult.CallId, tt.callID)
			}

			if toolResult.ToolResult.Success != tt.success {
				t.Errorf("success = %v, want %v", toolResult.ToolResult.Success, tt.success)
			}

			if toolResult.ToolResult.OutputJson == "" {
				t.Error("outputJson is empty")
			}
		})
	}
}

// TestEmitFinding verifies that EmitFinding sends correct events.
func TestEmitFinding(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	f := &finding.Finding{
		ID:       "finding-123",
		Severity: finding.SeverityHigh,
		Category: finding.CategoryInformationDisclosure,
	}

	err := sh.EmitFinding(f)
	if err != nil {
		t.Errorf("EmitFinding() error = %v, want nil", err)
	}

	msgs := stream.getSentMessages()
	if len(msgs) != 1 {
		t.Fatalf("sent %d messages, want 1", len(msgs))
	}

	msg := msgs[0]
	findingMsg, ok := msg.Payload.(*proto.AgentMessage_Finding)
	if !ok {
		t.Fatalf("payload type = %T, want *proto.AgentMessage_Finding", msg.Payload)
	}

	if findingMsg.Finding.FindingJson == "" {
		t.Error("findingJson is empty")
	}
}

// TestEmitStatus verifies that EmitStatus sends correct events.
func TestEmitStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		message       string
		expectedProto proto.AgentStatus
	}{
		{
			name:          "running status",
			status:        "running",
			message:       "Starting RBAC analysis",
			expectedProto: proto.AgentStatus_AGENT_STATUS_RUNNING,
		},
		{
			name:          "completed status",
			status:        "completed",
			message:       "Scan complete",
			expectedProto: proto.AgentStatus_AGENT_STATUS_COMPLETED,
		},
		{
			name:          "paused status",
			status:        "paused",
			message:       "Waiting for user approval",
			expectedProto: proto.AgentStatus_AGENT_STATUS_PAUSED,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHarness := &mockStreamHarness{}
			stream := &mockBidiStream{ServerStream: &mockServerStream{}}
			steeringCh := make(chan *proto.SteeringMessage, 10)

			sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

			err := sh.EmitStatus(tt.status, tt.message)
			if err != nil {
				t.Errorf("EmitStatus() error = %v, want nil", err)
			}

			msgs := stream.getSentMessages()
			if len(msgs) != 1 {
				t.Fatalf("sent %d messages, want 1", len(msgs))
			}

			msg := msgs[0]
			statusMsg, ok := msg.Payload.(*proto.AgentMessage_Status)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_Status", msg.Payload)
			}

			if statusMsg.Status.Status != tt.expectedProto {
				t.Errorf("status = %v, want %v", statusMsg.Status.Status, tt.expectedProto)
			}

			if statusMsg.Status.Message != tt.message {
				t.Errorf("message = %q, want %q", statusMsg.Status.Message, tt.message)
			}
		})
	}
}

// TestEmitError verifies that EmitError sends correct events.
func TestEmitError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		errContext string
	}{
		{
			name:       "error with context",
			err:        errors.New("Insufficient permissions"),
			errContext: "RBAC check",
		},
		{
			name:       "error without context",
			err:        errors.New("Cannot connect to cluster"),
			errContext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseHarness := &mockStreamHarness{}
			stream := &mockBidiStream{ServerStream: &mockServerStream{}}
			steeringCh := make(chan *proto.SteeringMessage, 10)

			sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

			err := sh.EmitError(tt.err, tt.errContext)
			if err != nil {
				t.Errorf("EmitError() error = %v, want nil", err)
			}

			msgs := stream.getSentMessages()
			if len(msgs) != 1 {
				t.Fatalf("sent %d messages, want 1", len(msgs))
			}

			msg := msgs[0]
			errorMsg, ok := msg.Payload.(*proto.AgentMessage_Error)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_Error", msg.Payload)
			}

			// Verify the error message contains the expected content
			if tt.errContext != "" {
				if !strings.Contains(errorMsg.Error.Message, tt.errContext) {
					t.Errorf("message = %q, want to contain %q", errorMsg.Error.Message, tt.errContext)
				}
			}
			if !strings.Contains(errorMsg.Error.Message, tt.err.Error()) {
				t.Errorf("message = %q, want to contain %q", errorMsg.Error.Message, tt.err.Error())
			}
		})
	}
}

// TestCallToolInterception verifies that CallTool emits tool call and result events.
func TestCallToolInterception(t *testing.T) {
	toolCalled := false
	baseHarness := &mockStreamHarness{
		callToolFunc: func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
			toolCalled = true
			return map[string]any{"result": "success"}, nil
		},
	}

	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	ctx := context.Background()
	result, err := sh.CallTool(ctx, "kubectl", map[string]any{"command": "get"})
	if err != nil {
		t.Errorf("CallTool() error = %v, want nil", err)
	}

	if !toolCalled {
		t.Error("underlying CallTool was not called")
	}

	if result == nil {
		t.Fatal("CallTool() returned nil result")
	}

	msgs := stream.getSentMessages()
	if len(msgs) != 2 {
		t.Fatalf("sent %d messages, want 2 (tool call + tool result)", len(msgs))
	}

	// First message should be tool call
	toolCall, ok := msgs[0].Payload.(*proto.AgentMessage_ToolCall)
	if !ok {
		t.Errorf("first message type = %T, want *proto.AgentMessage_ToolCall", msgs[0].Payload)
	} else {
		if toolCall.ToolCall.ToolName != "kubectl" {
			t.Errorf("toolName = %q, want 'kubectl'", toolCall.ToolCall.ToolName)
		}
	}

	// Second message should be tool result
	toolResult, ok := msgs[1].Payload.(*proto.AgentMessage_ToolResult)
	if !ok {
		t.Errorf("second message type = %T, want *proto.AgentMessage_ToolResult", msgs[1].Payload)
	} else {
		if !toolResult.ToolResult.Success {
			t.Error("toolResult.Success = false, want true")
		}
	}

	// Verify call ID matches
	if toolCall != nil && toolResult != nil {
		if toolCall.ToolCall.CallId != toolResult.ToolResult.CallId {
			t.Errorf("callId mismatch: call=%q, result=%q", toolCall.ToolCall.CallId, toolResult.ToolResult.CallId)
		}
	}
}

// TestCallToolInterceptionWithError verifies that tool errors are properly emitted.
func TestCallToolInterceptionWithError(t *testing.T) {
	expectedErr := errors.New("tool execution failed")
	baseHarness := &mockStreamHarness{
		callToolFunc: func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
			return nil, expectedErr
		},
	}

	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	ctx := context.Background()
	_, err := sh.CallTool(ctx, "kubectl", map[string]any{"command": "get"})
	if err != expectedErr {
		t.Errorf("CallTool() error = %v, want %v", err, expectedErr)
	}

	msgs := stream.getSentMessages()
	if len(msgs) != 2 {
		t.Fatalf("sent %d messages, want 2", len(msgs))
	}

	// Tool result should indicate failure
	toolResult, ok := msgs[1].Payload.(*proto.AgentMessage_ToolResult)
	if !ok {
		t.Fatalf("second message type = %T, want *proto.AgentMessage_ToolResult", msgs[1].Payload)
	}

	if toolResult.ToolResult.Success {
		t.Error("toolResult.Success = true, want false")
	}
}

// TestSubmitFindingInterception verifies that SubmitFinding emits finding events.
func TestSubmitFindingInterception(t *testing.T) {
	findingSubmitted := false
	baseHarness := &mockStreamHarness{
		submitFindingFunc: func(ctx context.Context, f *finding.Finding) error {
			findingSubmitted = true
			return nil
		},
	}

	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	f := &finding.Finding{
		ID:       "finding-123",
		Severity: finding.SeverityHigh,
		Category: finding.CategoryInformationDisclosure,
	}

	ctx := context.Background()
	err := sh.SubmitFinding(ctx, f)
	if err != nil {
		t.Errorf("SubmitFinding() error = %v, want nil", err)
	}

	if !findingSubmitted {
		t.Error("underlying SubmitFinding was not called")
	}

	msgs := stream.getSentMessages()
	if len(msgs) != 1 {
		t.Fatalf("sent %d messages, want 1", len(msgs))
	}

	findingMsg, ok := msgs[0].Payload.(*proto.AgentMessage_Finding)
	if !ok {
		t.Fatalf("payload type = %T, want *proto.AgentMessage_Finding", msgs[0].Payload)
	}

	if findingMsg.Finding.FindingJson == "" {
		t.Error("findingJson is empty")
	}
}

// TestCompleteInterception verifies that Complete emits output events.
func TestCompleteInterception(t *testing.T) {
	baseHarness := &mockStreamHarness{
		completeFunc: func(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
			return &llm.CompletionResponse{
				Content:      "LLM response content",
				FinishReason: "stop",
			}, nil
		},
	}

	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	ctx := context.Background()
	resp, err := sh.Complete(ctx, "primary", []llm.Message{{Role: llm.RoleUser, Content: "test"}})
	if err != nil {
		t.Errorf("Complete() error = %v, want nil", err)
	}

	if resp == nil {
		t.Fatal("Complete() returned nil response")
	}

	msgs := stream.getSentMessages()
	if len(msgs) != 1 {
		t.Fatalf("sent %d messages, want 1", len(msgs))
	}

	output, ok := msgs[0].Payload.(*proto.AgentMessage_Output)
	if !ok {
		t.Fatalf("payload type = %T, want *proto.AgentMessage_Output", msgs[0].Payload)
	}

	if output.Output.Content != "LLM response content" {
		t.Errorf("content = %q, want 'LLM response content'", output.Output.Content)
	}

	if output.Output.IsReasoning {
		t.Error("isReasoning = true, want false for Complete()")
	}
}

// TestStreamInterception verifies that Stream emits output events for each chunk.
func TestStreamInterception(t *testing.T) {
	chunks := []llm.StreamChunk{
		{Delta: "chunk1", FinishReason: ""},
		{Delta: "chunk2", FinishReason: ""},
		{Delta: "chunk3", FinishReason: "stop"},
	}

	baseHarness := &mockStreamHarness{
		streamFunc: func(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
			ch := make(chan llm.StreamChunk, len(chunks))
			for _, chunk := range chunks {
				ch <- chunk
			}
			close(ch)
			return ch, nil
		},
	}

	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	ctx := context.Background()
	chunkCh, err := sh.Stream(ctx, "primary", []llm.Message{{Role: llm.RoleUser, Content: "test"}})
	if err != nil {
		t.Fatalf("Stream() error = %v, want nil", err)
	}

	// Consume all chunks
	receivedChunks := []llm.StreamChunk{}
	for chunk := range chunkCh {
		receivedChunks = append(receivedChunks, chunk)
	}

	if len(receivedChunks) != len(chunks) {
		t.Fatalf("received %d chunks, want %d", len(receivedChunks), len(chunks))
	}

	// Verify all chunks were emitted as events
	msgs := stream.getSentMessages()
	if len(msgs) != len(chunks) {
		t.Fatalf("sent %d messages, want %d", len(msgs), len(chunks))
	}

	for i, msg := range msgs {
		output, ok := msg.Payload.(*proto.AgentMessage_Output)
		if !ok {
			t.Errorf("message %d type = %T, want *proto.AgentMessage_Output", i, msg.Payload)
			continue
		}

		if output.Output.Content != chunks[i].Delta {
			t.Errorf("message %d content = %q, want %q", i, output.Output.Content, chunks[i].Delta)
		}
	}
}

// TestSequenceNumberAtomicity verifies that sequence numbers are unique under concurrent access.
func TestSequenceNumberAtomicity(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	// Emit events concurrently
	numGoroutines := 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			sh.EmitOutput("concurrent output", false)
		}(i)
	}

	wg.Wait()

	msgs := stream.getSentMessages()
	if len(msgs) != numGoroutines {
		t.Fatalf("sent %d messages, want %d", len(msgs), numGoroutines)
	}

	// Verify all sequence numbers are unique and in range [1, numGoroutines]
	seenSequences := make(map[int64]bool)
	for _, msg := range msgs {
		if msg.Sequence < 1 || msg.Sequence > int64(numGoroutines) {
			t.Errorf("sequence %d out of range [1, %d]", msg.Sequence, numGoroutines)
		}

		if seenSequences[msg.Sequence] {
			t.Errorf("duplicate sequence number: %d", msg.Sequence)
		}
		seenSequences[msg.Sequence] = true
	}

	if len(seenSequences) != numGoroutines {
		t.Errorf("got %d unique sequences, want %d", len(seenSequences), numGoroutines)
	}
}

// TestSteeringAndModeAccessors verifies Steering() and Mode() methods.
func TestSteeringAndModeAccessors(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)
	mode := proto.AgentMode_AGENT_MODE_INTERACTIVE

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, mode)

	// Test Mode() - returns agent.ExecutionMode (converted from proto)
	expectedMode := agent.ExecutionModeManual // INTERACTIVE maps to Manual
	if sh.Mode() != expectedMode {
		t.Errorf("Mode() = %v, want %v", sh.Mode(), expectedMode)
	}

	// Test SetMode() - access via type assertion to concrete type
	concreteHarness := sh.(*streamingHarness)
	newMode := proto.AgentMode_AGENT_MODE_AUTONOMOUS
	concreteHarness.SetMode(newMode)
	expectedNewMode := agent.ExecutionModeAutonomous
	if sh.Mode() != expectedNewMode {
		t.Errorf("Mode() after SetMode() = %v, want %v", sh.Mode(), expectedNewMode)
	}

	// Test Steering() channel
	steeringChan := sh.Steering()
	if steeringChan == nil {
		t.Fatal("Steering() returned nil channel")
	}

	// Send a message to the channel and verify we can receive it
	testMsg := &proto.SteeringMessage{
		Id:      "msg-123",
		Content: "test input",
	}

	// Send in a goroutine to avoid blocking
	go func() {
		steeringCh <- testMsg
	}()

	// Use a timeout since the message goes through an adapter goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	select {
	case msg := <-steeringChan:
		// agent.SteeringMessage is a value type, check if Content is set
		if msg.Content == "" {
			t.Error("received empty message from steering channel")
		}
		if msg.Content != "test input" {
			t.Errorf("expected content 'test input', got %q", msg.Content)
		}
	case <-ctx.Done():
		t.Error("timeout waiting for steering message")
	}
}

// TestSetModeConcurrency verifies that SetMode is thread-safe.
func TestSetModeConcurrency(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	// Access concrete type for SetMode (internal method)
	concreteHarness := sh.(*streamingHarness)

	protoModes := []proto.AgentMode{
		proto.AgentMode_AGENT_MODE_AUTONOMOUS,
		proto.AgentMode_AGENT_MODE_INTERACTIVE,
	}

	// Expected agent.ExecutionMode values for validation
	expectedModes := []agent.ExecutionMode{
		agent.ExecutionModeAutonomous,
		agent.ExecutionModeManual,
	}

	var wg sync.WaitGroup
	iterations := 1000

	// Concurrently set and read mode
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			concreteHarness.SetMode(protoModes[i%len(protoModes)])
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			_ = sh.Mode()
		}
	}()

	wg.Wait()

	// Final mode should be valid
	finalMode := sh.Mode()
	validMode := false
	for _, mode := range expectedModes {
		if finalMode == mode {
			validMode = true
			break
		}
	}

	if !validMode {
		t.Errorf("final mode %v is not in expected modes %v", finalMode, expectedModes)
	}
}

// TestSendErrorHandling verifies that send errors are logged but don't panic.
func TestSendErrorHandling(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{
		ServerStream: &mockServerStream{},
		sendErr:      errors.New("stream closed"),
	}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	// EmitOutput should return error but not panic
	err := sh.EmitOutput("test", false)
	if err == nil {
		t.Error("EmitOutput() with stream error returned nil, want error")
	}

	// Verify no messages were actually sent
	msgs := stream.getSentMessages()
	if len(msgs) != 0 {
		t.Errorf("sent %d messages despite send error", len(msgs))
	}
}

// TestEmitMethodsSequenceOrdering verifies that multiple emit calls maintain sequence order.
func TestEmitMethodsSequenceOrdering(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	// Emit various event types in sequence
	sh.EmitOutput("output1", false)
	sh.EmitStatus("running", "running message")
	sh.EmitToolCall("tool1", map[string]any{}, "call-1")
	sh.EmitToolResult(map[string]any{"result": "done"}, nil, "call-1")
	sh.EmitFinding(&finding.Finding{ID: "f1", Severity: finding.SeverityHigh, Category: finding.CategoryJailbreak})
	sh.EmitError(errors.New("error"), "ERR context")
	sh.EmitOutput("output2", true)

	msgs := stream.getSentMessages()
	if len(msgs) != 7 {
		t.Fatalf("sent %d messages, want 7", len(msgs))
	}

	// Verify sequences are strictly increasing
	for i := 0; i < len(msgs); i++ {
		expectedSeq := int64(i + 1)
		if msgs[i].Sequence != expectedSeq {
			t.Errorf("message %d sequence = %d, want %d", i, msgs[i].Sequence, expectedSeq)
		}
	}
}

// TestHarnessDelegation verifies that the streaming harness properly delegates to underlying harness.
func TestHarnessDelegation(t *testing.T) {
	baseHarness := &mockStreamHarness{}
	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	ctx := context.Background()

	// Test delegation of various harness methods
	_, err := sh.ListTools(ctx)
	if err != nil {
		t.Errorf("ListTools() error = %v", err)
	}

	_, err = sh.ListPlugins(ctx)
	if err != nil {
		t.Errorf("ListPlugins() error = %v", err)
	}

	_, err = sh.ListAgents(ctx)
	if err != nil {
		t.Errorf("ListAgents() error = %v", err)
	}

	mem := sh.Memory()
	if mem == nil {
		t.Error("Memory() returned nil")
	}

	mission := sh.Mission()
	if mission.ID != "mission-1" {
		t.Errorf("Mission().ID = %q, want 'mission-1'", mission.ID)
	}

	target := sh.Target()
	if target.ID != "target-1" {
		t.Errorf("Target().ID = %q, want 'target-1'", target.ID)
	}

	tracer := sh.Tracer()
	if tracer == nil {
		t.Error("Tracer() returned nil")
	}

	logger := sh.Logger()
	if logger == nil {
		t.Error("Logger() returned nil")
	}

	tracker := sh.TokenUsage()
	if tracker == nil {
		t.Error("TokenUsage() returned nil")
	}
}

// TestConcurrentEmitsAndDelegation verifies thread-safety of concurrent operations.
func TestConcurrentEmitsAndDelegation(t *testing.T) {
	var callCount int64
	baseHarness := &mockStreamHarness{
		callToolFunc: func(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
			atomic.AddInt64(&callCount, 1)
			return map[string]any{"result": "ok"}, nil
		},
	}

	stream := &mockBidiStream{ServerStream: &mockServerStream{}}
	steeringCh := make(chan *proto.SteeringMessage, 10)

	sh := NewStreamingHarness(baseHarness, stream, steeringCh, proto.AgentMode_AGENT_MODE_AUTONOMOUS)

	numGoroutines := 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // 3 operations per goroutine

	ctx := context.Background()

	for i := 0; i < numGoroutines; i++ {
		// Emit output
		go func() {
			defer wg.Done()
			sh.EmitOutput("test", false)
		}()

		// Call tool (triggers interception)
		go func() {
			defer wg.Done()
			sh.CallTool(ctx, "tool", map[string]any{})
		}()

		// Emit status
		go func() {
			defer wg.Done()
			sh.EmitStatus("running", "running message")
		}()
	}

	wg.Wait()

	// Verify all tool calls were made
	if atomic.LoadInt64(&callCount) != int64(numGoroutines) {
		t.Errorf("callCount = %d, want %d", callCount, numGoroutines)
	}

	// Verify all messages were sent (output + tool call + tool result + status) * numGoroutines
	msgs := stream.getSentMessages()
	expectedMsgs := numGoroutines*4 // 1 output + 1 tool call + 1 tool result + 1 status
	if len(msgs) != expectedMsgs {
		t.Errorf("sent %d messages, want %d", len(msgs), expectedMsgs)
	}

	// Verify all sequence numbers are unique
	seenSequences := make(map[int64]bool)
	for _, msg := range msgs {
		if seenSequences[msg.Sequence] {
			t.Errorf("duplicate sequence number: %d", msg.Sequence)
		}
		seenSequences[msg.Sequence] = true
	}
}
