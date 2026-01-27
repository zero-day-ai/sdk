package serve

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/api/gen/toolspb"
	"google.golang.org/grpc/metadata"
	protolib "google.golang.org/protobuf/proto"
)

// mockStream implements proto.ToolService_StreamExecuteServer for testing
type mockStream struct {
	mu       sync.Mutex
	messages []*proto.ToolMessage
	recvCh   chan *proto.ToolClientMessage
	ctx      context.Context
}

func newMockStream(ctx context.Context) *mockStream {
	return &mockStream{
		messages: make([]*proto.ToolMessage, 0),
		recvCh:   make(chan *proto.ToolClientMessage, 10),
		ctx:      ctx,
	}
}

func (m *mockStream) Send(msg *proto.ToolMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockStream) Recv() (*proto.ToolClientMessage, error) {
	select {
	case msg := <-m.recvCh:
		return msg, nil
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	}
}

func (m *mockStream) Context() context.Context {
	return m.ctx
}

func (m *mockStream) SetHeader(metadata.MD) error   { return nil }
func (m *mockStream) SendHeader(metadata.MD) error  { return nil }
func (m *mockStream) SetTrailer(metadata.MD)        {}
func (m *mockStream) SendMsg(msg interface{}) error { return nil }
func (m *mockStream) RecvMsg(msg interface{}) error { return nil }

func (m *mockStream) getMessages() []*proto.ToolMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to avoid race conditions
	result := make([]*proto.ToolMessage, len(m.messages))
	copy(result, m.messages)
	return result
}

// TestToolStreamImpl_Progress tests that Progress() produces correct ToolMessage with ToolProgress payload
func TestToolStreamImpl_Progress(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:       stream,
		cancelCh:     make(chan struct{}),
		executionID:  "test-exec-123",
		traceID:      "test-trace-456",
		parentSpanID: "test-span-789",
		sequence:     0,
	}

	// Emit progress
	err := toolStream.Progress(42, "scanning", "Scanning ports 1-1000")
	require.NoError(t, err)

	// Verify message
	messages := stream.getMessages()
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, int64(1), msg.Sequence, "sequence should be 1 for first message")
	assert.NotZero(t, msg.TimestampMs, "timestamp should be set")
	assert.Equal(t, "test-trace-456", msg.TraceId)
	assert.NotEmpty(t, msg.SpanId, "span ID should be generated")

	// Check payload
	progress, ok := msg.Payload.(*proto.ToolMessage_Progress)
	require.True(t, ok, "payload should be ToolProgress")
	assert.Equal(t, int32(42), progress.Progress.Percent)
	assert.Equal(t, "scanning", progress.Progress.Stage)
	assert.Equal(t, "Scanning ports 1-1000", progress.Progress.Message)
}

// TestToolStreamImpl_Partial tests that Partial() marshals proto and produces correct ToolPartialResult
func TestToolStreamImpl_Partial(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Create a partial result proto
	partialResult := &toolspb.NmapResponse{
		Hosts: []*toolspb.NmapHost{
			{
				Ip:    "192.168.1.1",
				State: "up",
			},
		},
	}

	// Emit partial result (incremental)
	err := toolStream.Partial(partialResult, true)
	require.NoError(t, err)

	// Verify message
	messages := stream.getMessages()
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, int64(1), msg.Sequence)
	assert.NotZero(t, msg.TimestampMs)
	assert.Equal(t, "test-trace-456", msg.TraceId)

	// Check payload
	partial, ok := msg.Payload.(*proto.ToolMessage_Partial)
	require.True(t, ok, "payload should be ToolPartialResult")
	assert.NotEmpty(t, partial.Partial.OutputJson, "output JSON should be set")
	assert.Contains(t, partial.Partial.OutputJson, "192.168.1.1", "JSON should contain the host address")
	assert.Contains(t, partial.Partial.Description, "incremental=true", "description should indicate incremental mode")
}

// TestToolStreamImpl_Warning tests that Warning() produces correct ToolWarning
func TestToolStreamImpl_Warning(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Emit warning
	err := toolStream.Warning("Connection timeout", "host_192.168.1.1")
	require.NoError(t, err)

	// Verify message
	messages := stream.getMessages()
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, int64(1), msg.Sequence)
	assert.NotZero(t, msg.TimestampMs)
	assert.Equal(t, "test-trace-456", msg.TraceId)

	// Check payload
	warning, ok := msg.Payload.(*proto.ToolMessage_Warning)
	require.True(t, ok, "payload should be ToolWarning")
	assert.Equal(t, "Connection timeout", warning.Warning.Message)
	assert.Equal(t, "host_192.168.1.1", warning.Warning.Code, "context is stored in Code field")
}

// TestToolStreamImpl_Complete tests that Complete() produces correct ToolComplete
func TestToolStreamImpl_Complete(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Create final result proto
	finalResult := &toolspb.NmapResponse{
		Hosts: []*toolspb.NmapHost{
			{
				Ip:    "192.168.1.1",
				State: "up",
			},
			{
				Ip:    "192.168.1.2",
				State: "up",
			},
		},
		ScanDuration: 15.5,
	}

	// Emit completion
	err := toolStream.Complete(finalResult)
	require.NoError(t, err)

	// Verify message
	messages := stream.getMessages()
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, int64(1), msg.Sequence)
	assert.NotZero(t, msg.TimestampMs)
	assert.Equal(t, "test-trace-456", msg.TraceId)

	// Check payload
	complete, ok := msg.Payload.(*proto.ToolMessage_Complete)
	require.True(t, ok, "payload should be ToolComplete")
	assert.NotEmpty(t, complete.Complete.OutputJson, "output JSON should be set")
	assert.Contains(t, complete.Complete.OutputJson, "192.168.1.1")
	assert.Contains(t, complete.Complete.OutputJson, "192.168.1.2")
	assert.Contains(t, complete.Complete.OutputJson, "15.5", "should contain scan duration")
}

// TestToolStreamImpl_Error tests that Error() produces correct ToolError with fatal flag
func TestToolStreamImpl_Error(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	testErr := errors.New("nmap binary not found")

	// Test fatal error
	err := toolStream.Error(testErr, true)
	require.NoError(t, err)

	// Verify message
	messages := stream.getMessages()
	require.Len(t, messages, 1)

	msg := messages[0]
	assert.Equal(t, int64(1), msg.Sequence)
	assert.NotZero(t, msg.TimestampMs)
	assert.Equal(t, "test-trace-456", msg.TraceId)

	// Check payload
	toolError, ok := msg.Payload.(*proto.ToolMessage_Error)
	require.True(t, ok, "payload should be ToolError")
	assert.Equal(t, "TOOL_ERROR", toolError.Error.Error.Code)
	assert.Equal(t, "nmap binary not found", toolError.Error.Error.Message)
	assert.True(t, toolError.Error.Fatal, "error should be marked as fatal")

	// Test non-fatal error
	stream2 := newMockStream(ctx)
	toolStream2 := &toolStreamImpl{
		stream:      stream2,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-456",
		traceID:     "test-trace-789",
		sequence:    0,
	}

	err = toolStream2.Error(errors.New("minor issue"), false)
	require.NoError(t, err)

	messages2 := stream2.getMessages()
	require.Len(t, messages2, 1)
	toolError2, ok := messages2[0].Payload.(*proto.ToolMessage_Error)
	require.True(t, ok)
	assert.False(t, toolError2.Error.Fatal, "error should not be marked as fatal")
}

// TestToolStreamImpl_SequenceIncrement tests that sequence numbers increment atomically
func TestToolStreamImpl_SequenceIncrement(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Emit multiple messages
	_ = toolStream.Progress(10, "init", "Starting")
	_ = toolStream.Progress(25, "scanning", "Scanning")
	_ = toolStream.Warning("test warning", "test_context")
	_ = toolStream.Progress(50, "analyzing", "Analyzing")

	// Verify sequences
	messages := stream.getMessages()
	require.Len(t, messages, 4)

	for i, msg := range messages {
		assert.Equal(t, int64(i+1), msg.Sequence, "sequence should increment")
	}
}

// TestToolStreamImpl_SequenceAtomicity tests that sequence increment is atomic across goroutines
func TestToolStreamImpl_SequenceAtomicity(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Emit messages from multiple goroutines
	const numGoroutines = 10
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				_ = toolStream.Progress(id*10+j, "test", "concurrent progress")
			}
		}(i)
	}

	wg.Wait()

	// Verify all sequences are unique and sequential
	messages := stream.getMessages()
	require.Len(t, messages, numGoroutines*messagesPerGoroutine)

	sequences := make(map[int64]bool)
	for _, msg := range messages {
		assert.False(t, sequences[msg.Sequence], "sequence %d should be unique", msg.Sequence)
		sequences[msg.Sequence] = true
		assert.GreaterOrEqual(t, msg.Sequence, int64(1))
		assert.LessOrEqual(t, msg.Sequence, int64(numGoroutines*messagesPerGoroutine))
	}

	// All sequences from 1 to total should be present
	for i := int64(1); i <= int64(numGoroutines*messagesPerGoroutine); i++ {
		assert.True(t, sequences[i], "sequence %d should be present", i)
	}
}

// TestToolStreamImpl_Timestamps tests that timestamps are set correctly
func TestToolStreamImpl_Timestamps(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	beforeTime := time.Now().UnixMilli()

	// Emit messages with small delays
	_ = toolStream.Progress(10, "init", "Starting")
	time.Sleep(5 * time.Millisecond)
	_ = toolStream.Progress(20, "scanning", "Scanning")
	time.Sleep(5 * time.Millisecond)
	_ = toolStream.Progress(30, "analyzing", "Analyzing")

	afterTime := time.Now().UnixMilli()

	// Verify timestamps
	messages := stream.getMessages()
	require.Len(t, messages, 3)

	for i, msg := range messages {
		assert.GreaterOrEqual(t, msg.TimestampMs, beforeTime, "timestamp should be >= start time")
		assert.LessOrEqual(t, msg.TimestampMs, afterTime, "timestamp should be <= end time")

		// Timestamps should be monotonically increasing (or equal due to fast execution)
		if i > 0 {
			assert.GreaterOrEqual(t, msg.TimestampMs, messages[i-1].TimestampMs,
				"timestamps should be monotonically increasing")
		}
	}
}

// TestToolStreamImpl_TraceContext tests that trace context is propagated
func TestToolStreamImpl_TraceContext(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	traceID := "test-trace-abc123"
	parentSpanID := "parent-span-def456"

	toolStream := &toolStreamImpl{
		stream:       stream,
		cancelCh:     make(chan struct{}),
		executionID:  "test-exec-123",
		traceID:      traceID,
		parentSpanID: parentSpanID,
		sequence:     0,
	}

	// Emit various message types
	_ = toolStream.Progress(10, "init", "Starting")
	_ = toolStream.Warning("test warning", "test_context")
	_ = toolStream.Complete(&toolspb.NmapResponse{})

	// Verify trace context in all messages
	messages := stream.getMessages()
	require.Len(t, messages, 3)

	for _, msg := range messages {
		assert.Equal(t, traceID, msg.TraceId, "trace ID should be propagated")
		assert.NotEmpty(t, msg.SpanId, "span ID should be generated")
		// Each message should have a unique span ID
	}

	// Verify span IDs are unique
	spanIDs := make(map[string]bool)
	for _, msg := range messages {
		assert.False(t, spanIDs[msg.SpanId], "span ID %s should be unique", msg.SpanId)
		spanIDs[msg.SpanId] = true
	}
}

// TestToolStreamImpl_Cancelled tests the Cancelled() channel
func TestToolStreamImpl_Cancelled(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	cancelCh := make(chan struct{})
	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    cancelCh,
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Initially, channel should not be closed
	select {
	case <-toolStream.Cancelled():
		t.Fatal("Cancelled() channel should not be closed initially")
	default:
		// Expected
	}

	// Close the cancel channel
	close(cancelCh)

	// Now, channel should be closed
	select {
	case <-toolStream.Cancelled():
		// Expected - channel is closed
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Cancelled() channel should be closed")
	}
}

// TestToolStreamImpl_ExecutionID tests the ExecutionID() method
func TestToolStreamImpl_ExecutionID(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	executionID := "exec-unique-id-12345"
	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: executionID,
		traceID:     "test-trace-456",
		sequence:    0,
	}

	assert.Equal(t, executionID, toolStream.ExecutionID())
}

// TestToolStreamImpl_NilOutput tests handling of nil proto output
// Note: protojson.Marshal(nil) returns "{}" which is valid, so Complete(nil)
// actually succeeds and sends an empty object as the output.
func TestToolStreamImpl_NilOutput(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Nil proto marshals to "{}" in protojson, which is valid
	var nilProto protolib.Message = nil
	err := toolStream.Complete(nilProto)
	assert.NoError(t, err, "nil proto marshals to empty object in protojson")

	// Verify message was sent with empty JSON object
	messages := stream.getMessages()
	require.Len(t, messages, 1, "message should be sent")

	complete := messages[0].GetComplete()
	require.NotNil(t, complete, "should be a complete message")
	assert.Equal(t, "{}", complete.OutputJson, "nil proto should marshal to empty JSON object")
}

// TestToolStreamImpl_ConcurrentSend tests thread-safety of send()
func TestToolStreamImpl_ConcurrentSend(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Send messages concurrently from multiple goroutines
	const numGoroutines = 20
	const messagesPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				switch j % 3 {
				case 0:
					_ = toolStream.Progress(id*10+j, "test", "concurrent test")
				case 1:
					_ = toolStream.Warning("test warning", "concurrent")
				case 2:
					partialResult := &toolspb.NmapResponse{
						Hosts: []*toolspb.NmapHost{{Ip: "192.168.1.1"}},
					}
					_ = toolStream.Partial(partialResult, true)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages were sent
	messages := stream.getMessages()
	assert.Len(t, messages, numGoroutines*messagesPerGoroutine)

	// Verify sequences are unique and sequential
	sequences := make(map[int64]bool)
	for _, msg := range messages {
		assert.False(t, sequences[msg.Sequence], "sequence should be unique")
		sequences[msg.Sequence] = true
	}
}

// TestToolStreamImpl_EmptyPhaseAndMessage tests edge cases with empty strings
func TestToolStreamImpl_EmptyPhaseAndMessage(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Emit progress with empty phase and message
	err := toolStream.Progress(0, "", "")
	require.NoError(t, err)

	messages := stream.getMessages()
	require.Len(t, messages, 1)

	progress, ok := messages[0].Payload.(*proto.ToolMessage_Progress)
	require.True(t, ok)
	assert.Equal(t, int32(0), progress.Progress.Percent)
	assert.Equal(t, "", progress.Progress.Stage)
	assert.Equal(t, "", progress.Progress.Message)

	// Emit warning with empty strings
	err = toolStream.Warning("", "")
	require.NoError(t, err)

	messages = stream.getMessages()
	require.Len(t, messages, 2)

	warning, ok := messages[1].Payload.(*proto.ToolMessage_Warning)
	require.True(t, ok)
	assert.Equal(t, "", warning.Warning.Message)
	assert.Equal(t, "", warning.Warning.Code)
}

// TestToolStreamImpl_HighPercentage tests progress with percentage > 100
func TestToolStreamImpl_HighPercentage(t *testing.T) {
	ctx := context.Background()
	stream := newMockStream(ctx)

	toolStream := &toolStreamImpl{
		stream:      stream,
		cancelCh:    make(chan struct{}),
		executionID: "test-exec-123",
		traceID:     "test-trace-456",
		sequence:    0,
	}

	// Emit progress with > 100% (tools should clamp, but we allow it through)
	err := toolStream.Progress(150, "done", "Overachieving")
	require.NoError(t, err)

	messages := stream.getMessages()
	require.Len(t, messages, 1)

	progress, ok := messages[0].Payload.(*proto.ToolMessage_Progress)
	require.True(t, ok)
	assert.Equal(t, int32(150), progress.Progress.Percent, "percent is passed through as-is")
}

// TestToolStreamImpl_AtomicIncrement tests that atomic.AddInt64 works correctly
func TestToolStreamImpl_AtomicIncrement(t *testing.T) {
	var counter int64

	const numGoroutines = 100
	const incrementsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}

	wg.Wait()

	expected := int64(numGoroutines * incrementsPerGoroutine)
	assert.Equal(t, expected, counter, "atomic increment should be safe across goroutines")
}
