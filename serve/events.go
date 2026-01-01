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
// toolName is the name of the tool being called, inputJSON is the serialized input,
// callID is a unique identifier for this call, seq is the sequence number,
// and traceID/spanID are for distributed tracing.
func BuildToolCallEvent(toolName, inputJSON, callID string, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_ToolCall{
			ToolCall: &proto.ToolCallEvent{
				ToolName:  toolName,
				InputJson: inputJSON,
				CallId:    callID,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// BuildToolResultEvent constructs an AgentMessage containing a ToolResultEvent.
// callID matches the call_id from the corresponding ToolCallEvent, outputJSON
// is the serialized tool output, success indicates if the tool call succeeded,
// seq is the sequence number, and traceID/spanID are for distributed tracing.
func BuildToolResultEvent(callID, outputJSON string, success bool, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_ToolResult{
			ToolResult: &proto.ToolResultEvent{
				CallId:     callID,
				OutputJson: outputJSON,
				Success:    success,
			},
		},
		TraceId:     traceID,
		SpanId:      spanID,
		Sequence:    seq,
		TimestampMs: time.Now().UnixMilli(),
	}
}

// BuildFindingEvent constructs an AgentMessage containing a FindingEvent.
// findingJSON is the serialized finding data, seq is the sequence number,
// and traceID/spanID are for distributed tracing.
func BuildFindingEvent(findingJSON string, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_Finding{
			Finding: &proto.FindingEvent{
				FindingJson: findingJSON,
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
// code is the error code, message describes the error, fatal indicates if
// this error should terminate execution, seq is the sequence number,
// and traceID/spanID are for distributed tracing.
func BuildErrorEvent(code, message string, fatal bool, seq int64, traceID, spanID string) *proto.AgentMessage {
	return &proto.AgentMessage{
		Payload: &proto.AgentMessage_Error{
			Error: &proto.ErrorEvent{
				Code:    code,
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
