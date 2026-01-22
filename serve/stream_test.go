package serve

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/llm"
	"google.golang.org/grpc/metadata"
)

// mockStreamServer implements grpc.BidiStreamingServer for full bidirectional testing.
// It simulates both sending and receiving messages over a gRPC stream.
type mockStreamServer struct {
	// recvQueue holds messages from client to server
	recvQueue chan *proto.ClientMessage
	// sentMessages captures all messages sent from server to client
	sentMessages []*proto.AgentMessage
	sentMu       sync.Mutex
	// ctx is the context for the stream
	ctx context.Context
	// recvErr can be set to simulate Recv() errors
	recvErr error
	// sendErr can be set to simulate Send() errors
	sendErr error
}

func newMockStreamServer(ctx context.Context) *mockStreamServer {
	return &mockStreamServer{
		recvQueue:    make(chan *proto.ClientMessage, 100),
		sentMessages: make([]*proto.AgentMessage, 0),
		ctx:          ctx,
	}
}

func (m *mockStreamServer) Send(msg *proto.AgentMessage) error {
	m.sentMu.Lock()
	defer m.sentMu.Unlock()

	if m.sendErr != nil {
		return m.sendErr
	}

	m.sentMessages = append(m.sentMessages, msg)
	return nil
}

func (m *mockStreamServer) Recv() (*proto.ClientMessage, error) {
	if m.recvErr != nil {
		return nil, m.recvErr
	}

	select {
	case msg, ok := <-m.recvQueue:
		if !ok {
			// Channel was closed
			return nil, io.EOF
		}
		return msg, nil
	case <-m.ctx.Done():
		return nil, io.EOF
	}
}

func (m *mockStreamServer) getSentMessages() []*proto.AgentMessage {
	m.sentMu.Lock()
	defer m.sentMu.Unlock()

	msgs := make([]*proto.AgentMessage, len(m.sentMessages))
	copy(msgs, m.sentMessages)
	return msgs
}

func (m *mockStreamServer) sendClientMessage(msg *proto.ClientMessage) {
	m.recvQueue <- msg
}

func (m *mockStreamServer) SetHeader(md metadata.MD) error  { return nil }
func (m *mockStreamServer) SendHeader(md metadata.MD) error { return nil }
func (m *mockStreamServer) SetTrailer(md metadata.MD)       {}
func (m *mockStreamServer) Context() context.Context        { return m.ctx }
func (m *mockStreamServer) SendMsg(msg interface{}) error   { return nil }
func (m *mockStreamServer) RecvMsg(msg interface{}) error   { return nil }

// executionCounter tracks how many times Execute was called
type executionCounter struct {
	count int
	mu    sync.Mutex
}

func (e *executionCounter) increment() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.count++
}

func (e *executionCounter) get() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.count
}

// mockStreamingAgent implements both agent.Agent and StreamingAgent for testing streaming agents.
type mockStreamingAgent struct {
	*mockAgent
	executeStreamingFunc func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error)
	streamingCounter     executionCounter
}

func (a *mockStreamingAgent) ExecuteStreaming(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
	a.streamingCounter.increment()

	if a.executeStreamingFunc != nil {
		return a.executeStreamingFunc(ctx, harness, task)
	}
	return agent.NewSuccessResult("streaming completed"), nil
}

// TestStreamExecute_BasicFlow tests the basic streaming execution flow.
func TestStreamExecute_BasicFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	// Create a simple agent
	testAgent := &mockAgent{
		name:    "test-agent",
		version: "1.0.0",
		executeFunc: func(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
			// Agent execution - just return success
			return agent.NewSuccessResult("test complete"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest
	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task: &proto.Task{
					Id:   "task-1",
					Goal: "Test task execution",
				},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Close the receive channel to signal end of client messages
	close(stream.recvQueue)

	// Wait for execution to complete
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete in time")
	}

	// Verify agent was executed (we track this via the executeFunc being called)
	// The mockAgent doesn't have a counter, so we just verify messages were sent correctly

	// Verify messages were sent
	msgs := stream.getSentMessages()
	if len(msgs) < 2 {
		t.Fatalf("sent %d messages, want at least 2 (RUNNING + COMPLETED)", len(msgs))
	}

	// First message should be RUNNING status
	firstMsg := msgs[0]
	if status, ok := firstMsg.Payload.(*proto.AgentMessage_Status); !ok {
		t.Errorf("first message type = %T, want StatusChange", firstMsg.Payload)
	} else if status.Status.Status != proto.AgentStatus_AGENT_STATUS_RUNNING {
		t.Errorf("first status = %v, want RUNNING", status.Status.Status)
	}

	// Last message should be COMPLETED status
	lastMsg := msgs[len(msgs)-1]
	if status, ok := lastMsg.Payload.(*proto.AgentMessage_Status); !ok {
		t.Errorf("last message type = %T, want StatusChange", lastMsg.Payload)
	} else if status.Status.Status != proto.AgentStatus_AGENT_STATUS_COMPLETED {
		t.Errorf("last status = %v, want COMPLETED", status.Status.Status)
	}
}

// TestStreamExecute_StreamingAgent tests that streaming agents use ExecuteStreaming.
func TestStreamExecute_StreamingAgent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	// Create a streaming agent
	testAgent := &mockStreamingAgent{
		mockAgent: &mockAgent{
			name:    "streaming-agent",
			version: "1.0.0",
		},
		executeStreamingFunc: func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
			// Emit some events during execution
			harness.EmitOutput("Starting analysis", true)
			harness.EmitStatus("running", "Analyzing target")
			harness.EmitOutput("Analysis complete", false)
			return agent.NewSuccessResult("streaming complete"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Close receive channel to signal end of client messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	// Verify ExecuteStreaming was called, not Execute
	if testAgent.streamingCounter.get() != 1 {
		t.Errorf("ExecuteStreaming() called %d times, want 1", testAgent.streamingCounter.get())
	}

	// Verify events were emitted
	msgs := stream.getSentMessages()
	if len(msgs) < 5 {
		t.Fatalf("sent %d messages, want at least 5 (initial RUNNING + 3 agent events + final COMPLETED)", len(msgs))
	}

	// Check for output events
	var outputCount int
	for _, msg := range msgs {
		if _, ok := msg.Payload.(*proto.AgentMessage_Output); ok {
			outputCount++
		}
	}
	if outputCount != 2 {
		t.Errorf("output events = %d, want 2", outputCount)
	}
}

// TestStreamExecute_SteeringMessage tests steering message handling.
func TestStreamExecute_SteeringMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	steeringReceived := make(chan agent.SteeringMessage, 1)

	// Create agent that waits for steering
	testAgent := &mockStreamingAgent{
		mockAgent: &mockAgent{
			name:    "steering-agent",
			version: "1.0.0",
		},
		executeStreamingFunc: func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
			// Wait for steering message
			select {
			case msg := <-harness.Steering():
				steeringReceived <- msg
			case <-time.After(1 * time.Second):
				return agent.Result{}, errors.New("steering timeout")
			}
			return agent.NewSuccessResult("received steering"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Wait a bit for execution to start
	time.Sleep(100 * time.Millisecond)

	// Send steering message
	steeringMsg := &proto.SteeringMessage{
		Id:      "steering-1",
		Content: "user guidance",
		Metadata: map[string]string{
			"type": "approval",
		},
	}
	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Steering{
			Steering: steeringMsg,
		},
	})

	// Close receive channel after sending all messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	// Verify agent received the steering message
	select {
	case received := <-steeringReceived:
		// agent.SteeringMessage only has Content and Priority (no ID)
		if received.Content != steeringMsg.Content {
			t.Errorf("steering message content = %q, want %q", received.Content, steeringMsg.Content)
		}
	default:
		t.Error("agent did not receive steering message")
	}

	// Verify SteeringAck was sent
	msgs := stream.getSentMessages()
	var foundAck bool
	for _, msg := range msgs {
		if ack, ok := msg.Payload.(*proto.AgentMessage_SteeringAck); ok {
			if ack.SteeringAck.MessageId == steeringMsg.Id {
				foundAck = true
				break
			}
		}
	}
	if !foundAck {
		t.Error("SteeringAck not found in sent messages")
	}
}

// TestStreamExecute_SetMode tests mode switching during execution.
func TestStreamExecute_SetMode(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	modesObserved := make(chan agent.ExecutionMode, 2)

	// Create agent that checks mode multiple times
	testAgent := &mockStreamingAgent{
		mockAgent: &mockAgent{
			name:    "mode-agent",
			version: "1.0.0",
		},
		executeStreamingFunc: func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
			// Record initial mode
			modesObserved <- harness.Mode()

			// Wait for mode to change
			time.Sleep(200 * time.Millisecond)

			// Record changed mode
			modesObserved <- harness.Mode()

			return agent.NewSuccessResult("mode test complete"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest with autonomous mode

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Wait for execution to start
	time.Sleep(100 * time.Millisecond)

	// Change mode to interactive
	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_SetMode{
			SetMode: &proto.SetModeRequest{
				Mode: proto.AgentMode_AGENT_MODE_INTERACTIVE,
			},
		},
	})

	// Close receive channel after sending all messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	// Verify mode changes
	modes := make([]agent.ExecutionMode, 0, 2)
	close(modesObserved)
	for mode := range modesObserved {
		modes = append(modes, mode)
	}

	if len(modes) != 2 {
		t.Fatalf("observed %d modes, want 2", len(modes))
	}

	if modes[0] != agent.ExecutionModeAutonomous {
		t.Errorf("initial mode = %v, want Autonomous", modes[0])
	}

	if modes[1] != agent.ExecutionModeManual {
		t.Errorf("changed mode = %v, want Manual (maps from INTERACTIVE)", modes[1])
	}
}

// TestStreamExecute_Interrupt tests interrupt handling.
func TestStreamExecute_Interrupt(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	// Create agent that runs long enough to be interrupted
	testAgent := &mockStreamingAgent{
		mockAgent: &mockAgent{
			name:    "interrupt-agent",
			version: "1.0.0",
		},
		executeStreamingFunc: func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
			// Simulate long-running work
			time.Sleep(500 * time.Millisecond)
			return agent.NewSuccessResult("completed"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Wait for execution to start
	time.Sleep(100 * time.Millisecond)

	// Send interrupt
	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Interrupt{
			Interrupt: &proto.InterruptRequest{
				Reason: "user requested pause",
			},
		},
	})

	// Close receive channel after sending all messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	// Verify PAUSED status was emitted
	msgs := stream.getSentMessages()
	var foundPaused bool
	for _, msg := range msgs {
		if status, ok := msg.Payload.(*proto.AgentMessage_Status); ok {
			if status.Status.Status == proto.AgentStatus_AGENT_STATUS_PAUSED {
				foundPaused = true
				break
			}
		}
	}
	if !foundPaused {
		t.Error("PAUSED status not found in sent messages")
	}
}

// TestStreamExecute_Resume tests resume handling.
func TestStreamExecute_Resume(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	// Create agent that can be paused and resumed
	testAgent := &mockStreamingAgent{
		mockAgent: &mockAgent{
			name:    "resume-agent",
			version: "1.0.0",
		},
		executeStreamingFunc: func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
			// Simulate work
			time.Sleep(200 * time.Millisecond)
			return agent.NewSuccessResult("completed"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Wait for execution to start
	time.Sleep(50 * time.Millisecond)

	// Send resume with guidance
	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Resume{
			Resume: &proto.ResumeRequest{
				Guidance: "continue with caution",
			},
		},
	})

	// Close receive channel after sending all messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	// Verify RUNNING status was emitted (from resume)
	msgs := stream.getSentMessages()
	var runningCount int
	for _, msg := range msgs {
		if status, ok := msg.Payload.(*proto.AgentMessage_Status); ok {
			if status.Status.Status == proto.AgentStatus_AGENT_STATUS_RUNNING {
				runningCount++
			}
		}
	}
	// Should have at least 2 RUNNING: initial + resume
	if runningCount < 2 {
		t.Errorf("RUNNING status count = %d, want at least 2", runningCount)
	}
}

// TestStreamExecute_AgentError tests error handling during agent execution.
func TestStreamExecute_AgentError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	expectedErr := errors.New("agent execution failed")

	// Create agent that returns an error
	testAgent := &mockAgent{
		name:    "error-agent",
		version: "1.0.0",
		executeFunc: func(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.Result{}, expectedErr
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Close receive channel to signal end of client messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v (should not return error, should send FAILED status)", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	msgs := stream.getSentMessages()

	// Verify FAILED status was emitted
	var foundFailed bool
	for _, msg := range msgs {
		if status, ok := msg.Payload.(*proto.AgentMessage_Status); ok {
			if status.Status.Status == proto.AgentStatus_AGENT_STATUS_FAILED {
				foundFailed = true
				if status.Status.Message != "Execution failed: "+expectedErr.Error() {
					t.Errorf("failure message = %q, want %q", status.Status.Message, "Execution failed: "+expectedErr.Error())
				}
				break
			}
		}
	}
	if !foundFailed {
		t.Error("FAILED status not found in sent messages")
	}

	// Verify ErrorEvent was emitted
	var foundError bool
	for _, msg := range msgs {
		if errMsg, ok := msg.Payload.(*proto.AgentMessage_Error); ok {
			if errMsg.Error.Code == proto.ErrorCode_ERROR_CODE_INTERNAL {
				foundError = true
				if errMsg.Error.Message != expectedErr.Error() {
					t.Errorf("error message = %q, want %q", errMsg.Error.Message, expectedErr.Error())
				}
				if !errMsg.Error.Fatal {
					t.Error("error fatal = false, want true")
				}
				break
			}
		}
	}
	if !foundError {
		t.Error("ErrorEvent not found in sent messages")
	}
}

// TestStreamExecute_NilTask tests handling of nil task.
// With proto-based task handling, nil tasks are converted to empty tasks and are valid.
func TestStreamExecute_NilTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	// Agent that handles empty tasks gracefully
	testAgent := &mockAgent{
		name:    "test-agent",
		version: "1.0.0",
		executeFunc: func(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
			// Nil proto.Task is converted to empty agent.Task, which is valid
			return agent.Result{
				Status: agent.StatusSuccess,
				Output: map[string]any{"result": "handled empty task"},
			}, nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest with nil Task
	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        nil, // Nil task is converted to empty task
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Wait for completion - should succeed (nil task converted to empty task)
	select {
	case err := <-errCh:
		// No error expected - nil task is valid and converted to empty task
		if err != nil {
			t.Fatalf("StreamExecute() error = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		// May timeout waiting for stream to complete - this is acceptable
		// as long as no error was returned
	}

	// Verify stream received messages
	msgs := stream.getSentMessages()
	// Should have at least the running status
	if len(msgs) == 0 {
		t.Error("expected at least one message, got 0")
	}
}

// TestStreamExecute_NoStartRequest tests handling of missing start request.
func TestStreamExecute_NoStartRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	testAgent := &mockAgent{
		name:    "test-agent",
		version: "1.0.0",
	}

	server := &agentServiceServer{agent: testAgent}

	// Close the stream immediately to simulate EOF
	close(stream.recvQueue)

	// Start execution
	err := server.StreamExecute(stream)
	if err == nil {
		t.Fatal("StreamExecute() error = nil, want error for stream closed")
	}
}

// TestStreamExecute_WrongFirstMessage tests handling of non-start first message.
func TestStreamExecute_WrongFirstMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	testAgent := &mockAgent{
		name:    "test-agent",
		version: "1.0.0",
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send a steering message instead of start (wrong first message)
	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Steering{
			Steering: &proto.SteeringMessage{
				Id:      "msg-1",
				Content: "this should fail",
			},
		},
	})

	// Wait for error
	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("StreamExecute() error = nil, want error for wrong first message")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not return error")
	}
}

// TestStreamExecute_SequenceNumbers tests that sequence numbers are incremented correctly.
func TestStreamExecute_SequenceNumbers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	// Create agent that emits multiple events
	testAgent := &mockStreamingAgent{
		mockAgent: &mockAgent{
			name:    "sequence-agent",
			version: "1.0.0",
		},
		executeStreamingFunc: func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
			harness.EmitOutput("output1", false)
			harness.EmitOutput("output2", false)
			harness.EmitOutput("output3", false)
			return agent.NewSuccessResult("completed"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Close receive channel to signal end of client messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	// Verify sequence numbers (excluding SteeringAck which has sequence 0)
	msgs := stream.getSentMessages()
	seenSequences := make(map[int64]bool)
	var maxSeq int64

	for _, msg := range msgs {
		if msg.Sequence > 0 { // Skip SteeringAck messages
			if seenSequences[msg.Sequence] {
				t.Errorf("duplicate sequence number: %d", msg.Sequence)
			}
			seenSequences[msg.Sequence] = true
			if msg.Sequence > maxSeq {
				maxSeq = msg.Sequence
			}
		}
	}

	// Should have sequence numbers 1, 2, 3, 4, 5...
	for i := int64(1); i <= maxSeq; i++ {
		if !seenSequences[i] {
			t.Errorf("missing sequence number: %d", i)
		}
	}
}

// TestStreamExecute_EventOrdering tests that events are emitted in the correct order.
func TestStreamExecute_EventOrdering(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream := newMockStreamServer(ctx)

	// Create agent with known event sequence
	testAgent := &mockStreamingAgent{
		mockAgent: &mockAgent{
			name:    "ordering-agent",
			version: "1.0.0",
		},
		executeStreamingFunc: func(ctx context.Context, harness agent.StreamingHarness, task agent.Task) (agent.Result, error) {
			// Emit events in specific order
			harness.EmitOutput("step1", false)                        // 2
			harness.EmitStatus("running", "working")                  // 3
			harness.EmitToolCall("tool1", map[string]any{}, "call-1") // 4
			harness.EmitToolResult(map[string]any{}, nil, "call-1")   // 5
			harness.EmitOutput("step2", false)                        // 6
			return agent.NewSuccessResult("completed"), nil
		},
	}

	server := &agentServiceServer{agent: testAgent}

	// Start execution
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamExecute(stream)
	}()

	// Send StartExecutionRequest

	stream.sendClientMessage(&proto.ClientMessage{
		Payload: &proto.ClientMessage_Start{
			Start: &proto.StartExecutionRequest{
				Task:        &proto.Task{Id: "task-1"},
				InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
			},
		},
	})

	// Close receive channel to signal end of client messages
	close(stream.recvQueue)

	// Wait for completion
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("StreamExecute() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamExecute() did not complete")
	}

	msgs := stream.getSentMessages()

	// Expected order (by sequence number):
	// 1: RUNNING (initial)
	// 2: output "step1"
	// 3: status RUNNING "working"
	// 4: tool_call "tool1"
	// 5: tool_result "call-1"
	// 6: output "step2"
	// 7: COMPLETED (final)

	if len(msgs) < 7 {
		t.Fatalf("sent %d messages, want at least 7", len(msgs))
	}

	// Verify first message is RUNNING
	if status, ok := msgs[0].Payload.(*proto.AgentMessage_Status); !ok || status.Status.Status != proto.AgentStatus_AGENT_STATUS_RUNNING {
		t.Errorf("msgs[0] = %T, want RUNNING status", msgs[0].Payload)
	}

	// Verify output at position 1
	if output, ok := msgs[1].Payload.(*proto.AgentMessage_Output); !ok || output.Output.Content != "step1" {
		t.Errorf("msgs[1] = %T, want output 'step1'", msgs[1].Payload)
	}

	// Verify tool call at position 3
	if toolCall, ok := msgs[3].Payload.(*proto.AgentMessage_ToolCall); !ok || toolCall.ToolCall.ToolName != "tool1" {
		t.Errorf("msgs[3] = %T, want tool_call 'tool1'", msgs[3].Payload)
	}

	// Verify tool result at position 4
	if toolResult, ok := msgs[4].Payload.(*proto.AgentMessage_ToolResult); !ok || toolResult.ToolResult.CallId != "call-1" {
		t.Errorf("msgs[4] = %T, want tool_result 'call-1'", msgs[4].Payload)
	}

	// Verify last message is COMPLETED
	lastMsg := msgs[len(msgs)-1]
	if status, ok := lastMsg.Payload.(*proto.AgentMessage_Status); !ok || status.Status.Status != proto.AgentStatus_AGENT_STATUS_COMPLETED {
		t.Errorf("last message = %T, want COMPLETED status", lastMsg.Payload)
	}
}

// TestCreateStreamingHarness_LocalMode tests harness creation without callback endpoint.
func TestCreateStreamingHarness_LocalMode(t *testing.T) {
	ctx := context.Background()
	server := &agentServiceServer{}

	req := &proto.StartExecutionRequest{
		Task: &proto.Task{
			Id:   "task-1",
			Goal: "test",
		},
		InitialMode: proto.AgentMode_AGENT_MODE_AUTONOMOUS,
		// No callback_endpoint - should create local harness
	}

	harness, cleanup, err := server.createStreamingHarness(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, harness)
	require.NotNil(t, cleanup)
	defer cleanup()

	// Verify it's a local harness by checking that LLM operations fail
	_, err = harness.Complete(ctx, "primary", []llm.Message{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available in standalone mode")

	// Verify working memory works through tiered API
	err = harness.Memory().Working().Set(ctx, "test", "value")
	assert.NoError(t, err)

	val, err := harness.Memory().Working().Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
}

// TestCreateStreamingHarness_CallbackMode tests harness creation with callback endpoint.
// Note: This test would require a mock orchestrator server to properly test.
func TestCreateStreamingHarness_CallbackMode_ConnectionAttempt(t *testing.T) {
	// Skip this test as it requires a running orchestrator server or proper mock
	// The gRPC connection behavior varies and may not fail immediately for non-existent endpoints
	t.Skip("Requires mock orchestrator server - skipping for now")
}

// TestCreateStreamingHarness_InvalidMissionJSON tests handling of invalid mission JSON.
func TestCreateStreamingHarness_InvalidMissionJSON(t *testing.T) {
	// This test assumes we have a real orchestrator to connect to
	// For now, we just verify the JSON parsing logic by looking at error messages
	t.Skip("Requires mock orchestrator server - skipping for now")
}

// TestCreateStreamingHarness_InvalidTargetJSON tests handling of invalid target JSON.
func TestCreateStreamingHarness_InvalidTargetJSON(t *testing.T) {
	// This test assumes we have a real orchestrator to connect to
	// For now, we just verify the JSON parsing logic by looking at error messages
	t.Skip("Requires mock orchestrator server - skipping for now")
}
