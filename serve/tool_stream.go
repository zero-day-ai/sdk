package serve

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/enum"
	"github.com/zero-day-ai/sdk/tool"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	protolib "google.golang.org/protobuf/proto"
)

// StreamExecute implements bidirectional streaming execution for tools.
// It enables real-time progress reporting, partial results, and graceful cancellation.
//
// Message flow:
//  1. Client sends ToolStartRequest to begin execution
//  2. Tool emits real-time events (ToolProgress, ToolPartialResult, ToolWarning)
//  3. Client can send ToolCancelRequest to terminate execution
//  4. Tool completes with ToolComplete or ToolError
//
// The method spawns a receive loop goroutine to handle incoming client messages
// while the main goroutine executes the tool and sends events.
func (s *toolServiceServer) StreamExecute(stream proto.ToolService_StreamExecuteServer) error {
	ctx := stream.Context()

	// Wait for the initial ToolStartRequest
	firstMsg, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return status.Error(codes.InvalidArgument, "stream closed before receiving start request")
		}
		return status.Errorf(codes.Internal, "failed to receive start request: %v", err)
	}

	startReq, ok := firstMsg.Payload.(*proto.ToolClientMessage_Start)
	if !ok {
		return status.Error(codes.InvalidArgument, "first message must be ToolStartRequest")
	}

	// Check if tool implements StreamingTool interface
	streamingTool, isStreaming := s.tool.(tool.StreamingTool)
	if !isStreaming {
		// Fall back to unary execution wrapped as stream
		return s.executeUnaryAsStream(ctx, stream, startReq.Start)
	}

	// Parse proto input using the tool's input message type
	input, err := s.parseToolInput(startReq.Start.InputJson)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
	}

	// Create cancellation channel for client-initiated cancellation
	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})

	// Generate execution ID if not provided
	executionID := startReq.Start.TraceId
	if executionID == "" {
		executionID = uuid.New().String()
	}

	// Create toolStreamImpl implementing tool.ToolStream interface
	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    cancelCh,
		executionID: executionID,
		traceID:     startReq.Start.TraceId,
		parentSpanID: startReq.Start.ParentSpanId,
		sequence:    0,
	}

	// Spawn receive loop goroutine to handle cancel messages
	go func() {
		defer close(doneCh)
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Client closed the stream
					return
				}
				// Stream error - return and let main goroutine handle cleanup
				return
			}

			// Route message based on payload type
			switch msg.Payload.(type) {
			case *proto.ToolClientMessage_Cancel:
				// Signal cancellation via channel
				select {
				case <-cancelCh:
					// Already cancelled
				default:
					close(cancelCh)
				}
				return

			case *proto.ToolClientMessage_Start:
				// Ignore subsequent start messages (already started)
				continue
			}
		}
	}()

	// Apply timeout via context if specified
	if startReq.Start.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(startReq.Start.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	// Execute the streaming tool
	execErr := streamingTool.StreamExecuteProto(ctx, input, toolStream)
	if execErr != nil {
		// Tool returned an error - emit fatal error if not already sent
		// Check if the error is context cancellation (from timeout or cancel)
		if execErr == context.Canceled || execErr == context.DeadlineExceeded {
			// This is a cancellation/timeout - tool should have already emitted appropriate message
			// If not, emit a warning
			_ = toolStream.Warning("Tool execution cancelled", execErr.Error())
		} else {
			// Fatal error - emit error event
			_ = toolStream.Error(execErr, true)
		}
	}

	// Wait a moment for the receive loop to finish processing any pending messages
	select {
	case <-doneCh:
		// Receive loop finished
	case <-time.After(100 * time.Millisecond):
		// Timeout waiting for receive loop
	}

	return nil
}

// parseToolInput parses JSON input into the tool's proto message type.
func (s *toolServiceServer) parseToolInput(inputJSON string) (protolib.Message, error) {
	// Get the tool's input message type
	inputTypeName := s.tool.InputMessageType()
	if inputTypeName == "" {
		return nil, fmt.Errorf("tool does not specify InputMessageType")
	}

	// Find the proto message type in the global registry
	messageType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(inputTypeName))
	if err != nil {
		return nil, fmt.Errorf("failed to find message type %q: %w", inputTypeName, err)
	}

	// Create a new instance of the proto message
	protoMsg := messageType.New().Interface()

	// Apply enum normalization using the centralized enum.Normalize function
	normalizedJSON := enum.Normalize(s.tool.Name(), inputJSON)

	// Unmarshal JSON input into the proto message with lenient settings
	unmarshaler := protojson.UnmarshalOptions{
		DiscardUnknown: true, // Ignore unknown fields
	}
	if err := unmarshaler.Unmarshal([]byte(normalizedJSON), protoMsg); err != nil {
		return nil, fmt.Errorf("invalid input JSON for type %s: %w", inputTypeName, err)
	}

	return protoMsg, nil
}

// executeUnaryAsStream provides backward compatibility for non-streaming tools.
// It calls the regular ExecuteProto method and wraps the result in streaming messages.
func (s *toolServiceServer) executeUnaryAsStream(ctx context.Context, stream proto.ToolService_StreamExecuteServer, req *proto.ToolStartRequest) error {
	// Create a simple tool stream for sending messages
	executionID := req.TraceId
	if executionID == "" {
		executionID = uuid.New().String()
	}

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}), // Not used for unary execution
		executionID: executionID,
		traceID:     req.TraceId,
		parentSpanID: req.ParentSpanId,
		sequence:    0,
	}

	// Emit initial progress
	if err := toolStream.Progress(0, "executing", "Starting tool execution"); err != nil {
		return status.Errorf(codes.Internal, "failed to emit progress: %v", err)
	}

	// Parse proto input
	input, err := s.parseToolInput(req.InputJson)
	if err != nil {
		_ = toolStream.Error(err, true)
		return status.Errorf(codes.InvalidArgument, "invalid input: %v", err)
	}

	// Apply timeout if specified
	if req.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	// Call regular ExecuteProto
	output, execErr := s.tool.ExecuteProto(ctx, input)
	if execErr != nil {
		// Emit error and return
		_ = toolStream.Error(execErr, true)
		return nil // Don't return gRPC error - already sent via stream
	}

	// Emit completion with result
	if err := toolStream.Complete(output); err != nil {
		return status.Errorf(codes.Internal, "failed to emit completion: %v", err)
	}

	return nil
}

// toolStreamImpl implements tool.ToolStream interface.
// It wraps the gRPC stream and provides a clean API for tools to emit events.
type toolStreamImpl struct {
	// stream is the bidirectional gRPC stream for sending events to the client
	stream proto.ToolService_StreamExecuteServer

	// cancelCh is closed when cancellation is requested by the client
	cancelCh chan struct{}

	// executionID is the unique identifier for this tool execution
	executionID string

	// traceID is the distributed trace ID propagated from the caller
	traceID string

	// parentSpanID is the parent span ID for distributed tracing
	parentSpanID string

	// sequence is an atomic counter for event sequence numbers
	sequence int64

	// mu protects concurrent access to the stream
	mu sync.Mutex
}

// Progress emits a progress update indicating execution status.
// Implements tool.ToolStream interface.
func (s *toolStreamImpl) Progress(percent int, phase, message string) error {
	msg := &proto.ToolMessage{
		Payload: &proto.ToolMessage_Progress{
			Progress: &proto.ToolProgress{
				Percent: int32(percent),
				Stage:   phase,
				Message: message,
			},
		},
	}
	return s.send(msg)
}

// Partial emits a partial result before the tool completes.
// Implements tool.ToolStream interface.
func (s *toolStreamImpl) Partial(output protolib.Message, incremental bool) error {
	// Marshal proto output to JSON
	jsonBytes, err := protojson.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal partial output: %w", err)
	}

	msg := &proto.ToolMessage{
		Payload: &proto.ToolMessage_Partial{
			Partial: &proto.ToolPartialResult{
				OutputJson: string(jsonBytes),
				// Note: proto field is called Description in the proto but the design doc
				// suggests is_incremental. Checking the actual proto definition...
				Description: fmt.Sprintf("incremental=%v", incremental),
			},
		},
	}
	return s.send(msg)
}

// Warning emits a non-fatal warning that does not stop execution.
// Implements tool.ToolStream interface.
func (s *toolStreamImpl) Warning(message, context string) error {
	msg := &proto.ToolMessage{
		Payload: &proto.ToolMessage_Warning{
			Warning: &proto.ToolWarning{
				Message: message,
				Code:    context, // Note: proto field is 'code' not 'context'
			},
		},
	}
	return s.send(msg)
}

// Complete emits the final result and signals successful stream completion.
// Implements tool.ToolStream interface.
func (s *toolStreamImpl) Complete(output protolib.Message) error {
	// Marshal proto output to JSON
	jsonBytes, err := protojson.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal final output: %w", err)
	}

	msg := &proto.ToolMessage{
		Payload: &proto.ToolMessage_Complete{
			Complete: &proto.ToolComplete{
				OutputJson: string(jsonBytes),
			},
		},
	}
	return s.send(msg)
}

// Error emits an error event during execution.
// Implements tool.ToolStream interface.
func (s *toolStreamImpl) Error(err error, fatal bool) error {
	msg := &proto.ToolMessage{
		Payload: &proto.ToolMessage_Error{
			Error: &proto.ToolError{
				Error: &proto.Error{
					Code:      "TOOL_ERROR",
					Message:   err.Error(),
					Retryable: false,
				},
				Fatal: fatal,
			},
		},
	}
	return s.send(msg)
}

// Cancelled returns a channel that closes when cancellation is requested.
// Implements tool.ToolStream interface.
func (s *toolStreamImpl) Cancelled() <-chan struct{} {
	return s.cancelCh
}

// ExecutionID returns the unique execution ID for this tool invocation.
// Implements tool.ToolStream interface.
func (s *toolStreamImpl) ExecutionID() string {
	return s.executionID
}

// send safely sends a ToolMessage to the stream with proper sequencing and timestamps.
// It atomically increments the sequence number, sets the timestamp, and propagates trace context.
func (s *toolStreamImpl) send(msg *proto.ToolMessage) error {
	// Atomically increment sequence number
	msg.Sequence = atomic.AddInt64(&s.sequence, 1)

	// Set timestamp in milliseconds
	msg.TimestampMs = time.Now().UnixMilli()

	// Propagate trace context from the request
	msg.TraceId = s.traceID

	// Generate span ID for this event
	msg.SpanId = s.generateSpanID()

	// Lock for thread-safe stream access
	s.mu.Lock()
	defer s.mu.Unlock()

	// Send the message
	if err := s.stream.Send(msg); err != nil {
		return fmt.Errorf("failed to send tool message: %w", err)
	}

	return nil
}

// generateSpanID generates a new span ID for distributed tracing.
// This creates a child span under the parent span ID provided in the request.
func (s *toolStreamImpl) generateSpanID() string {
	// Generate a new random span ID for this event
	// SpanID is 8 bytes, so we use the first 8 bytes of a UUID
	id := uuid.New()
	var spanID trace.SpanID
	copy(spanID[:], id[:8])
	return spanID.String()
}
