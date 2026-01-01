package serve

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/llm"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

// StreamingHarness extends agent.Harness with bidirectional streaming capabilities.
// It provides methods for emitting real-time events to clients during agent execution
// and receiving steering messages for interactive control.
//
// StreamingHarness is designed for agents that need to provide live feedback,
// such as streaming reasoning steps, tool invocations, and findings as they occur,
// rather than just returning a final result.
type StreamingHarness interface {
	// Embed the base Harness interface to inherit all standard capabilities
	// (LLM access, tools, plugins, findings, memory, etc.)
	agent.Harness

	// Event Emission Methods
	//
	// These methods enable agents to emit real-time events during execution.
	// All events are sent to the connected client via the gRPC stream.

	// EmitOutput emits a text output chunk to the client.
	// Use isReasoning=true for internal reasoning/thinking output,
	// or isReasoning=false for final user-facing output.
	//
	// Example:
	//   h.EmitOutput("Analyzing RBAC permissions...", true)  // reasoning
	//   h.EmitOutput("Found vulnerable service account", false)  // result
	EmitOutput(content string, isReasoning bool) error

	// EmitToolCall emits an event indicating a tool invocation is starting.
	// The callID should be a unique identifier for correlating with the result.
	//
	// Example:
	//   h.EmitToolCall("kubectl", map[string]any{"args": []string{"get", "pods"}}, "call-123")
	EmitToolCall(toolName string, input map[string]any, callID string) error

	// EmitToolResult emits the result of a tool invocation.
	// The callID must match the ID from the corresponding EmitToolCall.
	// Set success=true if the tool executed successfully, false if it failed.
	//
	// Example:
	//   h.EmitToolResult("call-123", map[string]any{"output": "..."}, true)
	EmitToolResult(callID string, output map[string]any, success bool) error

	// EmitFinding emits a security finding discovered during testing.
	// The finding will be both streamed to the client and recorded via SubmitFinding.
	//
	// Example:
	//   h.EmitFinding(myFinding)
	EmitFinding(finding agent.Finding) error

	// EmitStatus emits an agent status change (running, paused, waiting, etc.).
	// The message provides additional context about the status change.
	//
	// Example:
	//   h.EmitStatus(proto.AgentStatus_AGENT_STATUS_WAITING_FOR_INPUT, "Awaiting user approval")
	EmitStatus(status proto.AgentStatus, message string) error

	// EmitError emits an error event to the client.
	// Set fatal=true if the error should terminate execution, false for recoverable errors.
	//
	// Example:
	//   h.EmitError("RBAC_DENIED", "Insufficient permissions to list pods", false)
	EmitError(code string, message string, fatal bool) error

	// Steering and Mode Methods
	//
	// These methods enable bidirectional communication with the client.

	// Steering returns a receive-only channel for steering messages from the client.
	// Agents can listen on this channel to receive user input, approvals, or interrupts.
	//
	// Example:
	//   select {
	//   case msg := <-h.Steering():
	//       // Handle steering message
	//   case <-ctx.Done():
	//       return ctx.Err()
	//   }
	Steering() <-chan *proto.SteeringMessage

	// Mode returns the current execution mode (autonomous or interactive).
	// Agents can check this to adjust their behavior.
	//
	// Example:
	//   if h.Mode() == proto.AgentMode_AGENT_MODE_INTERACTIVE {
	//       // Wait for user approval before proceeding
	//   }
	Mode() proto.AgentMode

	// SetMode updates the current execution mode atomically.
	// This is called by the streaming framework when the client sends a SetModeRequest.
	// Agents generally should not call this directly.
	SetMode(mode proto.AgentMode)
}

// streamingHarness is the concrete implementation of StreamingHarness.
// It wraps an underlying agent.Harness and adds streaming event emission capabilities.
type streamingHarness struct {
	// Embedded harness for delegation of all agent.Harness methods
	agent.Harness

	// stream is the bidirectional gRPC stream for sending events to the client
	stream grpc.BidiStreamingServer[proto.ClientMessage, proto.AgentMessage]

	// steeringCh receives steering messages from the client
	steeringCh <-chan *proto.SteeringMessage

	// mode tracks the current execution mode (autonomous or interactive)
	// Protected by modeMu for thread-safe access
	mode proto.AgentMode
	modeMu sync.RWMutex

	// sequence is an atomic counter for event sequence numbers
	sequence int64

	// mu protects concurrent access to the stream
	mu sync.Mutex

	// logger for non-fatal error logging
	logger *slog.Logger
}

// NewStreamingHarness creates a new StreamingHarness that wraps the given harness
// and emits events to the provided stream.
//
// Parameters:
//   - harness: The underlying agent.Harness to wrap
//   - stream: The bidirectional gRPC stream for sending AgentMessages
//   - steeringCh: Channel for receiving SteeringMessages from the client
//   - mode: Initial execution mode (autonomous or interactive)
//
// Returns a StreamingHarness ready for use in streaming agent execution.
func NewStreamingHarness(
	harness agent.Harness,
	stream grpc.BidiStreamingServer[proto.ClientMessage, proto.AgentMessage],
	steeringCh <-chan *proto.SteeringMessage,
	mode proto.AgentMode,
) StreamingHarness {
	var logger *slog.Logger
	if harness != nil {
		logger = harness.Logger()
	} else {
		// Use default logger if harness is nil
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return &streamingHarness{
		Harness:    harness,
		stream:     stream,
		steeringCh: steeringCh,
		mode:       mode,
		sequence:   0,
		logger:     logger,
	}
}

// nextSequence atomically increments and returns the next sequence number
func (h *streamingHarness) nextSequence() int64 {
	return atomic.AddInt64(&h.sequence, 1)
}

// getTraceInfo extracts trace and span IDs from the harness tracer if available
func (h *streamingHarness) getTraceInfo(ctx context.Context) (traceID, spanID string) {
	// Check if harness is nil first
	if h.Harness == nil {
		return "", ""
	}

	if h.Tracer() == nil {
		return "", ""
	}

	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return "", ""
	}

	return span.SpanContext().TraceID().String(), span.SpanContext().SpanID().String()
}

// send safely sends a message to the stream with proper locking
// Returns an error if the send fails, but does not panic
func (h *streamingHarness) send(msg *proto.AgentMessage) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := h.stream.Send(msg); err != nil {
		h.logger.Error("failed to send streaming event",
			"error", err,
			"sequence", msg.Sequence,
			"event_type", msg.Payload,
		)
		return err
	}

	return nil
}

// EmitOutput emits a text output chunk to the client
func (h *streamingHarness) EmitOutput(content string, isReasoning bool) error {
	// Create context for trace info extraction (using background as fallback)
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	msg := BuildOutputEvent(content, isReasoning, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitToolCall emits an event indicating a tool invocation is starting
func (h *streamingHarness) EmitToolCall(toolName string, input map[string]any, callID string) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	// Serialize input to JSON
	inputJSON, err := json.Marshal(input)
	if err != nil {
		h.logger.Error("failed to marshal tool input", "error", err, "tool", toolName)
		return err
	}

	msg := BuildToolCallEvent(toolName, string(inputJSON), callID, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitToolResult emits the result of a tool invocation
func (h *streamingHarness) EmitToolResult(callID string, output map[string]any, success bool) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	// Serialize output to JSON
	outputJSON, err := json.Marshal(output)
	if err != nil {
		h.logger.Error("failed to marshal tool output", "error", err, "call_id", callID)
		return err
	}

	msg := BuildToolResultEvent(callID, string(outputJSON), success, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitFinding emits a security finding discovered during testing
func (h *streamingHarness) EmitFinding(finding agent.Finding) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	// Serialize finding to JSON
	findingJSON, err := json.Marshal(finding)
	if err != nil {
		h.logger.Error("failed to marshal finding", "error", err, "finding_id", finding.ID())
		return err
	}

	msg := BuildFindingEvent(string(findingJSON), h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitStatus emits an agent status change
func (h *streamingHarness) EmitStatus(status proto.AgentStatus, message string) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	msg := BuildStatusEvent(status, message, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitError emits an error event to the client
func (h *streamingHarness) EmitError(code string, message string, fatal bool) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	msg := BuildErrorEvent(code, message, fatal, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// Steering returns a receive-only channel for steering messages from the client
func (h *streamingHarness) Steering() <-chan *proto.SteeringMessage {
	return h.steeringCh
}

// Mode returns the current execution mode
func (h *streamingHarness) Mode() proto.AgentMode {
	h.modeMu.RLock()
	defer h.modeMu.RUnlock()
	return h.mode
}

// SetMode updates the current execution mode atomically
func (h *streamingHarness) SetMode(mode proto.AgentMode) {
	h.modeMu.Lock()
	defer h.modeMu.Unlock()
	h.mode = mode
}

// CallTool overrides the base harness CallTool to emit events automatically
func (h *streamingHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	// Generate unique call ID
	callID := uuid.New().String()

	// Emit tool call event before invoking
	if err := h.EmitToolCall(name, input, callID); err != nil {
		h.logger.Warn("failed to emit tool call event", "error", err, "tool", name)
	}

	// Delegate to underlying harness
	output, err := h.Harness.CallTool(ctx, name, input)

	// Emit tool result event after invocation
	success := err == nil
	if emitErr := h.EmitToolResult(callID, output, success); emitErr != nil {
		h.logger.Warn("failed to emit tool result event", "error", emitErr, "tool", name)
	}

	return output, err
}

// SubmitFinding overrides the base harness SubmitFinding to emit events automatically
func (h *streamingHarness) SubmitFinding(ctx context.Context, f agent.Finding) error {
	// Emit finding event before submitting
	if err := h.EmitFinding(f); err != nil {
		h.logger.Warn("failed to emit finding event", "error", err, "finding_id", f.ID())
	}

	// Delegate to underlying harness
	return h.Harness.SubmitFinding(ctx, f)
}

// Complete overrides the base harness Complete to emit output events automatically
func (h *streamingHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	// Delegate to underlying harness for actual LLM completion
	resp, err := h.Harness.Complete(ctx, slot, messages, opts...)
	if err != nil {
		return nil, err
	}

	// Emit output chunk event with the response content
	// isReasoning=false since Complete is for final output
	if resp.Content != "" {
		if emitErr := h.EmitOutput(resp.Content, false); emitErr != nil {
			h.logger.Warn("failed to emit output event", "error", emitErr, "slot", slot)
		}
	}

	return resp, nil
}

// Stream overrides the base harness Stream to emit output events for each chunk
func (h *streamingHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	// Get the underlying stream channel
	chunkCh, err := h.Harness.Stream(ctx, slot, messages)
	if err != nil {
		return nil, err
	}

	// Create a new channel for wrapped chunks
	wrappedCh := make(chan llm.StreamChunk, 1)

	// Spawn goroutine to forward chunks and emit events
	go func() {
		defer close(wrappedCh)

		for chunk := range chunkCh {
			// Emit output chunk event for each chunk with content
			// Determine if this is reasoning based on chunk characteristics
			// For now, we'll use isReasoning=false for all streamed content
			// (agents can use EmitOutput directly with isReasoning=true for internal thoughts)
			if chunk.HasContent() {
				if emitErr := h.EmitOutput(chunk.Delta, false); emitErr != nil {
					h.logger.Warn("failed to emit stream chunk event", "error", emitErr, "slot", slot)
				}
			}

			// Forward the chunk to the caller
			wrappedCh <- chunk
		}
	}()

	return wrappedCh, nil
}
