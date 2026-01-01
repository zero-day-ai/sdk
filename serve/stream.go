package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamingAgent is an optional interface for agents that support streaming execution.
// Agents implementing this interface will have ExecuteStreaming called instead of Execute
// when connected via StreamExecute.
//
// Agents that do not implement this interface will still work with StreamExecute - the
// framework will automatically wrap their Execute method and emit events by intercepting
// harness method calls (CallTool, SubmitFinding, Complete, Stream).
type StreamingAgent interface {
	agent.Agent
	// ExecuteStreaming runs the agent with streaming event emission support.
	// The StreamingHarness provides methods to emit events during execution:
	// - EmitOutput: Emit LLM reasoning output chunks
	// - EmitToolCall: Emit tool call start events
	// - EmitToolResult: Emit tool call result events
	// - EmitFinding: Emit security finding events
	// - EmitStatus: Emit status change events
	// - EmitError: Emit error events
	//
	// The harness also provides access to steering messages via Steering() and
	// the current execution mode via Mode().
	ExecuteStreaming(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error)
}

// StreamExecute handles bidirectional streaming RPC for agent execution.
// It enables real-time event streaming, steering messages, mode switching,
// interrupts, and resume with guidance.
//
// Message flow:
//  1. Client sends StartExecutionRequest to begin execution
//  2. Agent emits real-time events (OutputChunk, ToolCallEvent, etc.)
//  3. Client can send SteeringMessage, InterruptRequest, SetModeRequest, or ResumeRequest
//  4. Agent completes with StatusChange(COMPLETED or FAILED)
//
// The method spawns a receive loop goroutine to handle incoming client messages
// while the main goroutine executes the agent and sends events.
func (s *agentServiceServer) StreamExecute(stream proto.AgentService_StreamExecuteServer) error {
	ctx := stream.Context()

	// Wait for the initial StartExecutionRequest
	firstMsg, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			return status.Error(codes.InvalidArgument, "stream closed before receiving start request")
		}
		return status.Errorf(codes.Internal, "failed to receive start request: %v", err)
	}

	startReq, ok := firstMsg.Payload.(*proto.ClientMessage_Start)
	if !ok {
		return status.Error(codes.InvalidArgument, "first message must be StartExecutionRequest")
	}

	// Parse task from JSON
	var task agent.Task
	if err := json.Unmarshal([]byte(startReq.Start.TaskJson), &task); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid task JSON: %v", err)
	}

	// Get initial mode (default to autonomous if not specified)
	initialMode := startReq.Start.InitialMode
	if initialMode == proto.AgentMode_AGENT_MODE_AUTONOMOUS && startReq.Start.InitialMode == 0 {
		// Proto default is 0, which maps to AUTONOMOUS - this is fine
	}

	// Create channels for bidirectional communication
	steeringCh := make(chan *proto.SteeringMessage, 10)
	interruptCh := make(chan string, 1)
	resumeCh := make(chan string, 1)
	doneCh := make(chan struct{})

	// Track current mode with mutex for thread-safe access
	var modeMu sync.RWMutex
	currentMode := initialMode
	interrupted := false

	// Create a read-only view of the steering channel for the harness
	steeringReadCh := (<-chan *proto.SteeringMessage)(steeringCh)

	// Create StreamingHarness wrapping nil harness (for now)
	// TODO: In a full implementation, this would wrap a real harness
	// provided by the framework with LLM provisioning, tool/plugin access, etc.
	streamingHarness := NewStreamingHarness(nil, stream, steeringReadCh, currentMode)

	// Spawn receive loop goroutine to handle incoming client messages
	go func() {
		defer close(doneCh)
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					// Client closed the stream
					return
				}
				// Log the error but don't crash - the main execution will handle cleanup
				s.agent.Health(ctx) // Just to access logger via agent if needed
				return
			}

			// Route message to appropriate handler based on payload type
			switch payload := msg.Payload.(type) {
			case *proto.ClientMessage_Start:
				// Ignore subsequent start messages (already started)
				continue

			case *proto.ClientMessage_Steering:
				// Forward steering message to the steering channel
				select {
				case steeringCh <- payload.Steering:
					// Send acknowledgment back to client
					ack := &proto.AgentMessage{
						Payload: &proto.AgentMessage_SteeringAck{
							SteeringAck: &proto.SteeringAck{
								MessageId: payload.Steering.Id,
								Response:  "received",
							},
						},
						Sequence:    0, // Acks don't need sequence numbers
						TimestampMs: 0,
					}
					if sendErr := stream.Send(ack); sendErr != nil {
						// Log error but continue
					}
				case <-ctx.Done():
					return
				}

			case *proto.ClientMessage_Interrupt:
				// Signal interruption
				modeMu.Lock()
				interrupted = true
				modeMu.Unlock()

				// Emit PAUSED status to indicate execution is paused
				_ = streamingHarness.EmitStatus("paused",
					fmt.Sprintf("Execution paused: %s", payload.Interrupt.Reason))

				select {
				case interruptCh <- payload.Interrupt.Reason:
				default:
					// Channel full, drop message
				}

			case *proto.ClientMessage_SetMode:
				// Update the current mode atomically
				modeMu.Lock()
				currentMode = payload.SetMode.Mode
				modeMu.Unlock()

				// Update the mode in the streaming harness as well
				// SetMode is internal - use type assertion to access concrete type
				type concreteHarness interface {
					SetMode(proto.AgentMode)
				}
				if sh, ok := streamingHarness.(concreteHarness); ok {
					sh.SetMode(payload.SetMode.Mode)
				}

			case *proto.ClientMessage_Resume:
				// Resume execution with optional guidance
				modeMu.Lock()
				interrupted = false
				modeMu.Unlock()

				// Emit RUNNING status to indicate execution is resuming
				guidanceMsg := "Execution resumed"
				if payload.Resume.Guidance != "" {
					guidanceMsg = fmt.Sprintf("Execution resumed with guidance: %s", payload.Resume.Guidance)
				}
				_ = streamingHarness.EmitStatus("running", guidanceMsg)

				select {
				case resumeCh <- payload.Resume.Guidance:
				default:
					// Channel full, drop message
				}
			}
		}
	}()

	// Emit RUNNING status to indicate execution has started
	if err := streamingHarness.EmitStatus("running", "Starting execution"); err != nil {
		return status.Errorf(codes.Internal, "failed to emit running status: %v", err)
	}

	// Detect if the agent supports streaming execution via type assertion
	var result agent.Result
	var execErr error

	if streamingAgent, ok := s.agent.(StreamingAgent); ok {
		// Agent implements StreamingAgent interface - use native streaming execution
		// This allows the agent to explicitly emit events during execution
		result, execErr = streamingAgent.ExecuteStreaming(ctx, streamingHarness, task)
	} else {
		// Agent does not implement StreamingAgent - fall back to wrapped Execute
		// The StreamingHarness will automatically emit events by intercepting:
		// - CallTool -> emits ToolCallEvent before and ToolResultEvent after
		// - SubmitFinding -> emits FindingEvent before delegating
		// - Complete/Stream -> emits OutputChunk events for LLM responses
		result, execErr = s.agent.Execute(ctx, streamingHarness, task)
	}

	// Determine final status based on execution result
	var finalStatusStr string
	var finalMessage string
	if execErr != nil {
		finalStatusStr = "failed"
		finalMessage = fmt.Sprintf("Execution failed: %v", execErr)

		// Emit a fatal error event for execution failure
		// Use concrete type to access internal send method with proper error code
		type concreteStreaming interface {
			getTraceInfo(ctx context.Context) (string, string)
			nextSequence() int64
			send(msg *proto.AgentMessage) error
		}
		if sh, ok := streamingHarness.(concreteStreaming); ok {
			ctx := context.Background()
			traceID, spanID := sh.getTraceInfo(ctx)
			errMsg := BuildErrorEvent("EXECUTION_ERROR", execErr.Error(), true, sh.nextSequence(), traceID, spanID)
			if emitErr := sh.send(errMsg); emitErr != nil {
				// Log but don't fail the RPC
			}
		}
	} else {
		// Check if we were interrupted
		modeMu.RLock()
		wasInterrupted := interrupted
		modeMu.RUnlock()

		if wasInterrupted {
			finalStatusStr = "interrupted"
			finalMessage = "Execution interrupted by client"
		} else {
			finalStatusStr = "completed"
			// Extract status from result if available
			if result.Status != "" {
				finalMessage = string(result.Status)
			} else {
				finalMessage = "Execution completed successfully"
			}
		}
	}

	// Emit final status
	if err := streamingHarness.EmitStatus(finalStatusStr, finalMessage); err != nil {
		return status.Errorf(codes.Internal, "failed to emit final status: %v", err)
	}

	// Wait a moment for the receive loop to finish processing any pending messages
	select {
	case <-doneCh:
		// Receive loop finished
	case <-ctx.Done():
		// Context cancelled
	}

	return nil
}
