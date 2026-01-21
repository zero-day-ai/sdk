package serve

import (
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
)

// BuildOutputEvent constructs an AgentMessage containing an OutputChunk event.
// content is the output text, isReasoning indicates if this is reasoning output,
// seq is the sequence number, and traceID/spanID are for distributed tracing.
func BuildOutputEvent(content string, isReasoning bool, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_Output{
			Output: &proto.OutputChunk{
				Content:     content,
				IsReasoning: isReasoning,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// BuildToolCallEvent constructs an AgentMessage containing a ToolCallEvent.
// toolName is the name of the tool being called, input is the map of input parameters,
// callID is a unique identifier for this call, seq is the sequence number,
// and traceID/spanID are for distributed tracing.
func BuildToolCallEvent(toolName string, input map[string]any, callID string, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_ToolCall{
			ToolCall: &proto.ToolCallEvent{
				ToolName: toolName,
				Input:    ToTypedMap(input),
				CallId:   callID,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// BuildToolResultEvent constructs an AgentMessage containing a ToolResultEvent.
// callID matches the call_id from the corresponding ToolCallEvent, output
// is the tool output value, success indicates if the tool call succeeded,
// seq is the sequence number, and traceID/spanID are for distributed tracing.
func BuildToolResultEvent(callID string, output any, success bool, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_ToolResult{
			ToolResult: &proto.ToolResultEvent{
				CallId:  callID,
				Output:  ToTypedValue(output),
				Success: success,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// BuildFindingEvent constructs an AgentMessage containing a FindingEvent.
// finding is the security finding data, seq is the sequence number,
// and traceID/spanID are for distributed tracing.
func BuildFindingEvent(finding *proto.Finding, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_Finding{
			Finding: &proto.FindingEvent{
				Finding: finding,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// BuildStatusEvent constructs an AgentMessage containing a StatusChange event.
// status is the new agent status, message provides context about the status change,
// seq is the sequence number, and traceID/spanID are for distributed tracing.
func BuildStatusEvent(status proto.AgentStatus, message string, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_Status{
			Status: &proto.StatusChange{
				Status:  status,
				Message: message,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// BuildErrorEvent constructs an AgentMessage containing an ErrorEvent.
// code is the error code string, message describes the error, fatal indicates if
// this error should terminate execution, seq is the sequence number,
// and traceID/spanID are for distributed tracing.
func BuildErrorEvent(code string, message string, fatal bool, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_Error{
			Error: &proto.ErrorEvent{
				Code:    StringToProtoErrorCode(code),
				Message: message,
				Fatal:   fatal,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}
