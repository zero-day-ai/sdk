package serve

import (
	"testing"
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
)

// TestBuildOutputEvent tests the BuildOutputEvent function with various inputs.
func TestBuildOutputEvent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		isReasoning bool
		seq         int64
		traceID     string
		spanID      string
	}{
		{
			name:        "standard output",
			content:     "test output",
			isReasoning: false,
			seq:         1,
			traceID:     "trace-123",
			spanID:      "span-456",
		},
		{
			name:        "reasoning output",
			content:     "thinking about the problem...",
			isReasoning: true,
			seq:         2,
			traceID:     "trace-789",
			spanID:      "span-012",
		},
		{
			name:        "empty content",
			content:     "",
			isReasoning: false,
			seq:         0,
			traceID:     "",
			spanID:      "",
		},
		{
			name:        "special characters",
			content:     "output with\nnewlines\tand\ttabs",
			isReasoning: false,
			seq:         100,
			traceID:     "trace-special",
			spanID:      "span-special",
		},
		{
			name:        "unicode content",
			content:     "测试 🔒 security",
			isReasoning: true,
			seq:         999,
			traceID:     "trace-unicode",
			spanID:      "span-unicode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UnixMilli()
			msg := BuildOutputEvent(tt.content, tt.isReasoning, tt.seq, tt.traceID, tt.spanID)
			after := time.Now().UnixMilli()

			// Verify message structure
			if msg == nil {
				t.Fatal("BuildOutputEvent returned nil")
			}

			// Verify sequence number
			if msg.Sequence != tt.seq {
				t.Errorf("sequence = %d, want %d", msg.Sequence, tt.seq)
			}

			// Verify trace/span IDs
			if msg.TraceId != tt.traceID {
				t.Errorf("traceId = %q, want %q", msg.TraceId, tt.traceID)
			}
			if msg.SpanId != tt.spanID {
				t.Errorf("spanId = %q, want %q", msg.SpanId, tt.spanID)
			}

			// Verify timestamp is populated and reasonable
			if msg.TimestampMs < before || msg.TimestampMs > after {
				t.Errorf("timestampMs = %d, want between %d and %d", msg.TimestampMs, before, after)
			}

			// Verify payload type
			output, ok := msg.Payload.(*proto.AgentMessage_Output)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_Output", msg.Payload)
			}

			// Verify output fields
			if output.Output.Content != tt.content {
				t.Errorf("content = %q, want %q", output.Output.Content, tt.content)
			}
			if output.Output.IsReasoning != tt.isReasoning {
				t.Errorf("isReasoning = %v, want %v", output.Output.IsReasoning, tt.isReasoning)
			}
		})
	}
}

// TestBuildToolCallEvent tests the BuildToolCallEvent function with various inputs.
func TestBuildToolCallEvent(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    map[string]any
		callID   string
		seq      int64
		traceID  string
		spanID   string
	}{
		{
			name:     "standard tool call",
			toolName: "nmap",
			input:    map[string]any{"target": "192.168.1.1"},
			callID:   "call-123",
			seq:      1,
			traceID:  "trace-123",
			spanID:   "span-456",
		},
		{
			name:     "complex input",
			toolName: "sqlmap",
			input:    map[string]any{"target": "http://example.com", "params": map[string]any{"depth": 3, "risk": 2}},
			callID:   "call-456",
			seq:      5,
			traceID:  "trace-789",
			spanID:   "span-012",
		},
		{
			name:     "empty fields",
			toolName: "",
			input:    nil,
			callID:   "",
			seq:      0,
			traceID:  "",
			spanID:   "",
		},
		{
			name:     "special characters in tool name",
			toolName: "tool-with-dashes_and_underscores",
			input:    map[string]any{"key": "value"},
			callID:   "call-special",
			seq:      42,
			traceID:  "trace-special",
			spanID:   "span-special",
		},
		{
			name:     "large sequence number",
			toolName: "test-tool",
			input:    map[string]any{},
			callID:   "call-999",
			seq:      9223372036854775807, // max int64
			traceID:  "trace-max",
			spanID:   "span-max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UnixMilli()
			msg := BuildToolCallEvent(tt.toolName, tt.input, tt.callID, tt.seq, tt.traceID, tt.spanID)
			after := time.Now().UnixMilli()

			// Verify message structure
			if msg == nil {
				t.Fatal("BuildToolCallEvent returned nil")
			}

			// Verify sequence number
			if msg.Sequence != tt.seq {
				t.Errorf("sequence = %d, want %d", msg.Sequence, tt.seq)
			}

			// Verify trace/span IDs
			if msg.TraceId != tt.traceID {
				t.Errorf("traceId = %q, want %q", msg.TraceId, tt.traceID)
			}
			if msg.SpanId != tt.spanID {
				t.Errorf("spanId = %q, want %q", msg.SpanId, tt.spanID)
			}

			// Verify timestamp is populated and reasonable
			if msg.TimestampMs < before || msg.TimestampMs > after {
				t.Errorf("timestampMs = %d, want between %d and %d", msg.TimestampMs, before, after)
			}

			// Verify payload type
			toolCall, ok := msg.Payload.(*proto.AgentMessage_ToolCall)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_ToolCall", msg.Payload)
			}

			// Verify tool call fields
			if toolCall.ToolCall.ToolName != tt.toolName {
				t.Errorf("toolName = %q, want %q", toolCall.ToolCall.ToolName, tt.toolName)
			}
			// Verify Input is a TypedMap (not checking exact values, just structure)
			if tt.input != nil && toolCall.ToolCall.Input == nil {
				t.Error("expected Input to be set")
			}
			if toolCall.ToolCall.CallId != tt.callID {
				t.Errorf("callId = %q, want %q", toolCall.ToolCall.CallId, tt.callID)
			}
		})
	}
}

// TestBuildToolResultEvent tests the BuildToolResultEvent function with various inputs.
func TestBuildToolResultEvent(t *testing.T) {
	tests := []struct {
		name    string
		callID  string
		output  any
		success bool
		seq     int64
		traceID string
		spanID  string
	}{
		{
			name:    "successful tool result",
			callID:  "call-123",
			output:  map[string]any{"result": "success", "data": "value"},
			success: true,
			seq:     2,
			traceID: "trace-123",
			spanID:  "span-456",
		},
		{
			name:    "failed tool result",
			callID:  "call-456",
			output:  map[string]any{"error": "connection timeout"},
			success: false,
			seq:     3,
			traceID: "trace-789",
			spanID:  "span-012",
		},
		{
			name:    "nil output",
			callID:  "call-789",
			output:  nil,
			success: true,
			seq:     4,
			traceID: "trace-empty",
			spanID:  "span-empty",
		},
		{
			name:    "complex nested output",
			callID:  "call-complex",
			output:  map[string]any{"ports": []any{80, 443}, "services": map[string]any{"80": "http", "443": "https"}},
			success: true,
			seq:     10,
			traceID: "trace-complex",
			spanID:  "span-complex",
		},
		{
			name:    "error with special characters",
			callID:  "call-error",
			output:  map[string]any{"error": "Failed: \n\tPermission denied"},
			success: false,
			seq:     50,
			traceID: "trace-error",
			spanID:  "span-error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UnixMilli()
			msg := BuildToolResultEvent(tt.callID, tt.output, tt.success, tt.seq, tt.traceID, tt.spanID)
			after := time.Now().UnixMilli()

			// Verify message structure
			if msg == nil {
				t.Fatal("BuildToolResultEvent returned nil")
			}

			// Verify sequence number
			if msg.Sequence != tt.seq {
				t.Errorf("sequence = %d, want %d", msg.Sequence, tt.seq)
			}

			// Verify trace/span IDs
			if msg.TraceId != tt.traceID {
				t.Errorf("traceId = %q, want %q", msg.TraceId, tt.traceID)
			}
			if msg.SpanId != tt.spanID {
				t.Errorf("spanId = %q, want %q", msg.SpanId, tt.spanID)
			}

			// Verify timestamp is populated and reasonable
			if msg.TimestampMs < before || msg.TimestampMs > after {
				t.Errorf("timestampMs = %d, want between %d and %d", msg.TimestampMs, before, after)
			}

			// Verify payload type
			toolResult, ok := msg.Payload.(*proto.AgentMessage_ToolResult)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_ToolResult", msg.Payload)
			}

			// Verify tool result fields
			if toolResult.ToolResult.CallId != tt.callID {
				t.Errorf("callId = %q, want %q", toolResult.ToolResult.CallId, tt.callID)
			}
			// Output is now a TypedValue, just verify it exists for non-nil outputs
			if tt.output != nil && toolResult.ToolResult.Output == nil {
				t.Error("expected Output to be set for non-nil input")
			}
			if toolResult.ToolResult.Success != tt.success {
				t.Errorf("success = %v, want %v", toolResult.ToolResult.Success, tt.success)
			}
		})
	}
}

// TestBuildFindingEvent tests the BuildFindingEvent function with various inputs.
func TestBuildFindingEvent(t *testing.T) {
	tests := []struct {
		name    string
		finding *proto.Finding
		seq     int64
		traceID string
		spanID  string
	}{
		{
			name: "standard finding",
			finding: &proto.Finding{
				Severity: proto.FindingSeverity_FINDING_SEVERITY_HIGH,
				Title:    "SQL Injection",
			},
			seq:     1,
			traceID: "trace-123",
			spanID:  "span-456",
		},
		{
			name: "detailed finding with evidence",
			finding: &proto.Finding{
				Severity:    proto.FindingSeverity_FINDING_SEVERITY_CRITICAL,
				Title:       "Prompt Injection",
				Description: "Evidence: SELECT * FROM users",
			},
			seq:     10,
			traceID: "trace-789",
			spanID:  "span-012",
		},
		{
			name:    "nil finding",
			finding: nil,
			seq:     0,
			traceID: "",
			spanID:  "",
		},
		{
			name: "finding with MITRE mapping",
			finding: &proto.Finding{
				Severity:    proto.FindingSeverity_FINDING_SEVERITY_MEDIUM,
				MitreAttack: &proto.MitreMapping{TechniqueId: "T1059.001"},
			},
			seq:     25,
			traceID: "trace-mitre",
			spanID:  "span-mitre",
		},
		{
			name: "finding with unicode",
			finding: &proto.Finding{
				Title:    "SQL注入漏洞",
				Severity: proto.FindingSeverity_FINDING_SEVERITY_HIGH,
			},
			seq:     100,
			traceID: "trace-unicode",
			spanID:  "span-unicode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UnixMilli()
			msg := BuildFindingEvent(tt.finding, tt.seq, tt.traceID, tt.spanID)
			after := time.Now().UnixMilli()

			// Verify message structure
			if msg == nil {
				t.Fatal("BuildFindingEvent returned nil")
			}

			// Verify sequence number
			if msg.Sequence != tt.seq {
				t.Errorf("sequence = %d, want %d", msg.Sequence, tt.seq)
			}

			// Verify trace/span IDs
			if msg.TraceId != tt.traceID {
				t.Errorf("traceId = %q, want %q", msg.TraceId, tt.traceID)
			}
			if msg.SpanId != tt.spanID {
				t.Errorf("spanId = %q, want %q", msg.SpanId, tt.spanID)
			}

			// Verify timestamp is populated and reasonable
			if msg.TimestampMs < before || msg.TimestampMs > after {
				t.Errorf("timestampMs = %d, want between %d and %d", msg.TimestampMs, before, after)
			}

			// Verify payload type
			findingEvent, ok := msg.Payload.(*proto.AgentMessage_Finding)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_Finding", msg.Payload)
			}

			// Verify finding is properly set
			if tt.finding != nil && findingEvent.Finding.Finding == nil {
				t.Error("expected Finding to be set for non-nil input")
			}
		})
	}
}

// TestBuildStatusEvent tests the BuildStatusEvent function with various inputs.
func TestBuildStatusEvent(t *testing.T) {
	tests := []struct {
		name    string
		status  proto.AgentStatus
		message string
		seq     int64
		traceID string
		spanID  string
	}{
		{
			name:    "running status",
			status:  proto.AgentStatus_AGENT_STATUS_RUNNING,
			message: "agent started",
			seq:     1,
			traceID: "trace-123",
			spanID:  "span-456",
		},
		{
			name:    "completed status",
			status:  proto.AgentStatus_AGENT_STATUS_COMPLETED,
			message: "execution finished successfully",
			seq:     100,
			traceID: "trace-789",
			spanID:  "span-012",
		},
		{
			name:    "failed status",
			status:  proto.AgentStatus_AGENT_STATUS_FAILED,
			message: "critical error occurred",
			seq:     50,
			traceID: "trace-fail",
			spanID:  "span-fail",
		},
		{
			name:    "waiting for input status with empty message",
			status:  proto.AgentStatus_AGENT_STATUS_WAITING_FOR_INPUT,
			message: "",
			seq:     0,
			traceID: "",
			spanID:  "",
		},
		{
			name:    "paused status with special characters",
			status:  proto.AgentStatus_AGENT_STATUS_PAUSED,
			message: "waiting for user input:\n\t- choice A\n\t- choice B",
			seq:     75,
			traceID: "trace-paused",
			spanID:  "span-paused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UnixMilli()
			msg := BuildStatusEvent(tt.status, tt.message, tt.seq, tt.traceID, tt.spanID)
			after := time.Now().UnixMilli()

			// Verify message structure
			if msg == nil {
				t.Fatal("BuildStatusEvent returned nil")
			}

			// Verify sequence number
			if msg.Sequence != tt.seq {
				t.Errorf("sequence = %d, want %d", msg.Sequence, tt.seq)
			}

			// Verify trace/span IDs
			if msg.TraceId != tt.traceID {
				t.Errorf("traceId = %q, want %q", msg.TraceId, tt.traceID)
			}
			if msg.SpanId != tt.spanID {
				t.Errorf("spanId = %q, want %q", msg.SpanId, tt.spanID)
			}

			// Verify timestamp is populated and reasonable
			if msg.TimestampMs < before || msg.TimestampMs > after {
				t.Errorf("timestampMs = %d, want between %d and %d", msg.TimestampMs, before, after)
			}

			// Verify payload type
			statusChange, ok := msg.Payload.(*proto.AgentMessage_Status)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_Status", msg.Payload)
			}

			// Verify status fields
			if statusChange.Status.Status != tt.status {
				t.Errorf("status = %v, want %v", statusChange.Status.Status, tt.status)
			}
			if statusChange.Status.Message != tt.message {
				t.Errorf("message = %q, want %q", statusChange.Status.Message, tt.message)
			}
		})
	}
}

// TestBuildErrorEvent tests the BuildErrorEvent function with various inputs.
func TestBuildErrorEvent(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		message string
		fatal   bool
		seq     int64
		traceID string
		spanID  string
	}{
		{
			name:    "fatal error",
			code:    "E001",
			message: "out of memory",
			fatal:   true,
			seq:     1,
			traceID: "trace-123",
			spanID:  "span-456",
		},
		{
			name:    "non-fatal warning",
			code:    "W002",
			message: "connection retry successful",
			fatal:   false,
			seq:     5,
			traceID: "trace-789",
			spanID:  "span-012",
		},
		{
			name:    "empty error",
			code:    "",
			message: "",
			fatal:   false,
			seq:     0,
			traceID: "",
			spanID:  "",
		},
		{
			name:    "error with newlines",
			code:    "E500",
			message: "stack trace:\n  at func1()\n  at func2()",
			fatal:   true,
			seq:     99,
			traceID: "trace-stack",
			spanID:  "span-stack",
		},
		{
			name:    "validation error",
			code:    "VALIDATION_ERROR",
			message: "invalid input: expected string, got number",
			fatal:   false,
			seq:     42,
			traceID: "trace-validation",
			spanID:  "span-validation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UnixMilli()
			msg := BuildErrorEvent(tt.code, tt.message, tt.fatal, tt.seq, tt.traceID, tt.spanID)
			after := time.Now().UnixMilli()

			// Verify message structure
			if msg == nil {
				t.Fatal("BuildErrorEvent returned nil")
			}

			// Verify sequence number
			if msg.Sequence != tt.seq {
				t.Errorf("sequence = %d, want %d", msg.Sequence, tt.seq)
			}

			// Verify trace/span IDs
			if msg.TraceId != tt.traceID {
				t.Errorf("traceId = %q, want %q", msg.TraceId, tt.traceID)
			}
			if msg.SpanId != tt.spanID {
				t.Errorf("spanId = %q, want %q", msg.SpanId, tt.spanID)
			}

			// Verify timestamp is populated and reasonable
			if msg.TimestampMs < before || msg.TimestampMs > after {
				t.Errorf("timestampMs = %d, want between %d and %d", msg.TimestampMs, before, after)
			}

			// Verify payload type
			errorEvent, ok := msg.Payload.(*proto.AgentMessage_Error)
			if !ok {
				t.Fatalf("payload type = %T, want *proto.AgentMessage_Error", msg.Payload)
			}

			// Verify error fields - Code is now an ErrorCode enum, not a string
			expectedCode := StringToProtoErrorCode(tt.code)
			if errorEvent.Error.Code != expectedCode {
				t.Errorf("code = %v, want %v", errorEvent.Error.Code, expectedCode)
			}
			if errorEvent.Error.Message != tt.message {
				t.Errorf("message = %q, want %q", errorEvent.Error.Message, tt.message)
			}
			if errorEvent.Error.Fatal != tt.fatal {
				t.Errorf("fatal = %v, want %v", errorEvent.Error.Fatal, tt.fatal)
			}
		})
	}
}

// TestBuildEventsTimestampConsistency verifies that timestamps are always non-zero
// and reasonably close to current time.
func TestBuildEventsTimestampConsistency(t *testing.T) {
	// Build several events in quick succession
	before := time.Now().UnixMilli()

	msg1 := BuildOutputEvent("test", false, 1, "trace", "span")
	msg2 := BuildToolCallEvent("tool", map[string]any{}, "call", 2, "trace", "span")
	msg3 := BuildToolResultEvent("call", map[string]any{}, true, 3, "trace", "span")
	msg4 := BuildFindingEvent(&proto.Finding{Title: "test"}, 4, "trace", "span")
	msg5 := BuildStatusEvent(proto.AgentStatus_AGENT_STATUS_RUNNING, "running", 5, "trace", "span")
	msg6 := BuildErrorEvent("ERR", "error", false, 6, "trace", "span")

	after := time.Now().UnixMilli()

	messages := []*proto.AgentMessage{msg1, msg2, msg3, msg4, msg5, msg6}

	for i, msg := range messages {
		// Verify timestamp is non-zero
		if msg.TimestampMs == 0 {
			t.Errorf("message %d has zero timestamp", i+1)
		}

		// Verify timestamp is within reasonable range
		if msg.TimestampMs < before || msg.TimestampMs > after {
			t.Errorf("message %d timestamp %d is outside expected range [%d, %d]",
				i+1, msg.TimestampMs, before, after)
		}
	}
}

// TestBuildEventsSequenceNumbers verifies that sequence numbers are correctly set.
func TestBuildEventsSequenceNumbers(t *testing.T) {
	sequences := []int64{0, 1, 100, 999999, -1, 9223372036854775807}

	for _, seq := range sequences {
		msg := BuildOutputEvent("test", false, seq, "trace", "span")
		if msg.Sequence != seq {
			t.Errorf("BuildOutputEvent: sequence = %d, want %d", msg.Sequence, seq)
		}

		msg = BuildToolCallEvent("tool", map[string]any{}, "call", seq, "trace", "span")
		if msg.Sequence != seq {
			t.Errorf("BuildToolCallEvent: sequence = %d, want %d", msg.Sequence, seq)
		}

		msg = BuildToolResultEvent("call", map[string]any{}, true, seq, "trace", "span")
		if msg.Sequence != seq {
			t.Errorf("BuildToolResultEvent: sequence = %d, want %d", msg.Sequence, seq)
		}

		msg = BuildFindingEvent(&proto.Finding{Title: "test"}, seq, "trace", "span")
		if msg.Sequence != seq {
			t.Errorf("BuildFindingEvent: sequence = %d, want %d", msg.Sequence, seq)
		}

		msg = BuildStatusEvent(proto.AgentStatus_AGENT_STATUS_RUNNING, "running", seq, "trace", "span")
		if msg.Sequence != seq {
			t.Errorf("BuildStatusEvent: sequence = %d, want %d", msg.Sequence, seq)
		}

		msg = BuildErrorEvent("ERR", "error", false, seq, "trace", "span")
		if msg.Sequence != seq {
			t.Errorf("BuildErrorEvent: sequence = %d, want %d", msg.Sequence, seq)
		}
	}
}

// TestBuildEventsTraceSpanIDs verifies that trace and span IDs are correctly set.
func TestBuildEventsTraceSpanIDs(t *testing.T) {
	testCases := []struct {
		traceID string
		spanID  string
	}{
		{"", ""},
		{"trace-123", "span-456"},
		{"very-long-trace-id-with-many-characters", "very-long-span-id-with-many-characters"},
		{"trace:with:colons", "span:with:colons"},
		{"trace/with/slashes", "span/with/slashes"},
	}

	for _, tc := range testCases {
		msg := BuildOutputEvent("test", false, 1, tc.traceID, tc.spanID)
		if msg.TraceId != tc.traceID {
			t.Errorf("BuildOutputEvent: traceId = %q, want %q", msg.TraceId, tc.traceID)
		}
		if msg.SpanId != tc.spanID {
			t.Errorf("BuildOutputEvent: spanId = %q, want %q", msg.SpanId, tc.spanID)
		}
	}
}
