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
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/llm"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	protolib "google.golang.org/protobuf/proto"
)

// Note: The serve package's streamingHarness implements agent.StreamingHarness
// to provide bidirectional streaming capabilities. The agent.StreamingHarness
// interface is defined in agent/harness.go and uses Go-native types rather than
// proto types for a cleaner API surface.

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
	mode   proto.AgentMode
	modeMu sync.RWMutex

	// sequence is an atomic counter for event sequence numbers
	sequence int64

	// mu protects concurrent access to the stream
	mu sync.Mutex

	// logger for non-fatal error logging
	logger *slog.Logger
}

// NewStreamingHarness creates a new agent.StreamingHarness that wraps the given harness
// and emits events to the provided stream.
//
// Parameters:
//   - harness: The underlying agent.Harness to wrap
//   - stream: The bidirectional gRPC stream for sending AgentMessages
//   - steeringCh: Channel for receiving SteeringMessages from the client
//   - mode: Initial execution mode (autonomous or interactive)
//
// Returns an agent.StreamingHarness ready for use in streaming agent execution.
func NewStreamingHarness(
	harness agent.Harness,
	stream grpc.BidiStreamingServer[proto.ClientMessage, proto.AgentMessage],
	steeringCh <-chan *proto.SteeringMessage,
	mode proto.AgentMode,
) agent.StreamingHarness {
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

	msg := BuildToolCallEvent(toolName, input, callID, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitToolResult emits the result of a tool invocation.
// Implements agent.StreamingHarness interface.
func (h *streamingHarness) EmitToolResult(output map[string]any, err error, callID string) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	// Determine success based on whether an error was provided
	success := err == nil

	msg := BuildToolResultEvent(callID, output, success, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitFinding emits a security finding discovered during testing
func (h *streamingHarness) EmitFinding(f *finding.Finding) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	// Convert SDK finding to proto finding
	protoFinding := FindingToProto(f)

	msg := BuildFindingEvent(protoFinding, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// EmitStatus emits an agent status change.
// Implements agent.StreamingHarness interface.
func (h *streamingHarness) EmitStatus(status string, message string) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	// Convert string status to proto.AgentStatus
	protoStatus := stringToProtoStatus(status)

	msg := BuildStatusEvent(protoStatus, message, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// stringToProtoStatus converts a string status to proto.AgentStatus
func stringToProtoStatus(status string) proto.AgentStatus {
	switch status {
	case "running":
		return proto.AgentStatus_AGENT_STATUS_RUNNING
	case "paused":
		return proto.AgentStatus_AGENT_STATUS_PAUSED
	case "completed":
		return proto.AgentStatus_AGENT_STATUS_COMPLETED
	case "failed":
		return proto.AgentStatus_AGENT_STATUS_FAILED
	case "interrupted":
		return proto.AgentStatus_AGENT_STATUS_INTERRUPTED
	case "waiting":
		return proto.AgentStatus_AGENT_STATUS_WAITING_FOR_INPUT
	default:
		return proto.AgentStatus_AGENT_STATUS_RUNNING
	}
}

// EmitError emits an error event to the client.
// Implements agent.StreamingHarness interface.
func (h *streamingHarness) EmitError(err error, errContext string) error {
	ctx := context.Background()
	traceID, spanID := h.getTraceInfo(ctx)

	// Build error code from error type and message from error + context
	code := "ERROR"
	message := err.Error()
	if errContext != "" {
		message = errContext + ": " + message
	}
	fatal := false // agent.StreamingHarness.EmitError is for non-fatal errors

	msg := BuildErrorEvent(code, message, fatal, h.nextSequence(), traceID, spanID)
	return h.send(msg)
}

// Steering returns a receive-only channel for steering messages from the client.
// Implements agent.StreamingHarness interface.
// This returns a channel of agent.SteeringMessage which wraps proto messages.
func (h *streamingHarness) Steering() <-chan agent.SteeringMessage {
	// Create a channel to convert proto messages to agent messages
	if h.steeringCh == nil {
		return nil
	}

	// Create a wrapper channel that converts proto steering messages to agent ones
	agentCh := make(chan agent.SteeringMessage, 10)
	go func() {
		defer close(agentCh)
		for protoMsg := range h.steeringCh {
			agentMsg := agent.SteeringMessage{
				Content:  protoMsg.Content,
				Priority: false, // Proto SteeringMessage has no Priority field, default to false
			}
			agentCh <- agentMsg
		}
	}()
	return agentCh
}

// Mode returns the current execution mode.
// Implements agent.StreamingHarness interface.
func (h *streamingHarness) Mode() agent.ExecutionMode {
	h.modeMu.RLock()
	defer h.modeMu.RUnlock()
	return protoModeToAgentMode(h.mode)
}

// protoModeToAgentMode converts proto.AgentMode to agent.ExecutionMode
func protoModeToAgentMode(mode proto.AgentMode) agent.ExecutionMode {
	switch mode {
	case proto.AgentMode_AGENT_MODE_AUTONOMOUS:
		return agent.ExecutionModeAutonomous
	case proto.AgentMode_AGENT_MODE_INTERACTIVE:
		return agent.ExecutionModeManual
	default:
		return agent.ExecutionModeAutonomous
	}
}

// SetMode updates the current execution mode atomically.
// Used internally by the streaming framework.
func (h *streamingHarness) SetMode(mode proto.AgentMode) {
	h.modeMu.Lock()
	defer h.modeMu.Unlock()
	h.mode = mode
}

// CallToolProto overrides the base harness CallToolProto to emit events automatically
func (h *streamingHarness) CallToolProto(ctx context.Context, name string, request protolib.Message, response protolib.Message) error {
	// Generate unique call ID
	callID := uuid.New().String()

	// Convert proto request to map for emission
	var inputMap map[string]any
	if request != nil {
		// Serialize to JSON then to map for display
		if jsonBytes, err := json.Marshal(request); err == nil {
			_ = json.Unmarshal(jsonBytes, &inputMap)
		}
	}

	// Emit tool call event before invoking
	if err := h.EmitToolCall(name, inputMap, callID); err != nil {
		h.logger.Warn("failed to emit tool call event", "error", err, "tool", name)
	}

	// Delegate to underlying harness
	toolErr := h.Harness.CallToolProto(ctx, name, request, response)

	// Convert proto response to map for emission
	var outputMap map[string]any
	if response != nil {
		if jsonBytes, err := json.Marshal(response); err == nil {
			_ = json.Unmarshal(jsonBytes, &outputMap)
		}
	}

	// Emit tool result event after invocation
	if emitErr := h.EmitToolResult(outputMap, toolErr, callID); emitErr != nil {
		h.logger.Warn("failed to emit tool result event", "error", emitErr, "tool", name)
	}

	return toolErr
}

// SubmitFinding overrides the base harness SubmitFinding to emit events automatically
func (h *streamingHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	// Emit finding event before submitting
	if err := h.EmitFinding(f); err != nil {
		h.logger.Warn("failed to emit finding event", "error", err, "finding_id", f.ID)
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
