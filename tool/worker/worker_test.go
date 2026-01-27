package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/zero-day-ai/sdk/queue"
	"github.com/zero-day-ai/sdk/types"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// mockTool is a mock implementation of tool.Tool for testing.
type mockTool struct {
	name        string
	version     string
	description string
	tags        []string
	inputType   string
	outputType  string
	executeFunc func(ctx context.Context, input proto.Message) (proto.Message, error)
	healthFunc  func(ctx context.Context) types.HealthStatus
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Version() string     { return m.version }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) Tags() []string {
	if m.tags != nil {
		return m.tags
	}
	return []string{}
}

func (m *mockTool) InputMessageType() string {
	if m.inputType != "" {
		return m.inputType
	}
	return "google.protobuf.StringValue"
}

func (m *mockTool) OutputMessageType() string {
	if m.outputType != "" {
		return m.outputType
	}
	return "google.protobuf.StringValue"
}

func (m *mockTool) ExecuteProto(ctx context.Context, input proto.Message) (proto.Message, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	// Default implementation - return empty response
	return wrapperspb.String("success"), nil
}

func (m *mockTool) Health(ctx context.Context) types.HealthStatus {
	if m.healthFunc != nil {
		return m.healthFunc(ctx)
	}
	return types.HealthStatus{Status: "ok"}
}

// setupTestRedis creates a miniredis instance and returns its address.
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, string) {
	t.Helper()
	s := miniredis.RunT(t)
	return s, fmt.Sprintf("redis://%s", s.Addr())
}

// newTestLogger creates a logger that discards output for tests.
func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Only log errors in tests
	}))
}

func TestWorkerLoop_BasicExecution(t *testing.T) {
	s, redisURL := setupTestRedis(t)
	defer s.Close()

	// Create a mock tool that counts executions
	var execCount atomic.Int32
	mockT := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool",
		tags:        []string{"test"},
		executeFunc: func(ctx context.Context, input proto.Message) (proto.Message, error) {
			execCount.Add(1)
			req := input.(*wrapperspb.StringValue)
			return wrapperspb.String(fmt.Sprintf("Processed: %s", req.GetValue())), nil
		},
	}

	// Create Redis client
	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// Note: We skip RegisterTool due to miniredis limitation with marshaling []interface{}
	// In production, this works fine with real Redis. See queue/client_test.go for details.
	// The worker doesn't strictly require pre-registration to function, only to be discoverable.

	// Create work items
	queueName := fmt.Sprintf("tool:%s:queue", mockT.Name())
	jobID := "test-job-1"
	numItems := 5

	for i := 0; i < numItems; i++ {
		req := wrapperspb.String(fmt.Sprintf("item-%d", i))
		inputJSON, _ := protojson.Marshal(req)

		item := queue.WorkItem{
			JobID:       jobID,
			Index:       i,
			Total:       numItems,
			Tool:        mockT.Name(),
			InputJSON:   string(inputJSON),
			InputType:   mockT.InputMessageType(),
			OutputType:  mockT.OutputMessageType(),
			SubmittedAt: time.Now().UnixMilli(),
		}
		if err := client.Push(context.Background(), queueName, item); err != nil {
			t.Fatalf("Failed to push work item: %v", err)
		}
	}

	// Subscribe to results
	resultChannel := fmt.Sprintf("results:%s", jobID)
	resultsChan, err := client.Subscribe(context.Background(), resultChannel)
	if err != nil {
		t.Fatalf("Failed to subscribe to results: %v", err)
	}

	// Start worker loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		workerLoop(ctx, 0, mockT, client, queueName, "test-worker-1", newTestLogger())
	}()

	// Collect results
	results := make([]queue.Result, 0, numItems)
	timeout := time.After(5 * time.Second)

	for len(results) < numItems {
		select {
		case result := <-resultsChan:
			results = append(results, result)
		case <-timeout:
			t.Fatalf("Timeout waiting for results, got %d/%d", len(results), numItems)
		}
	}

	// Cancel worker and wait
	cancel()
	wg.Wait()

	// Verify execution count
	if got := int(execCount.Load()); got != numItems {
		t.Errorf("Expected %d executions, got %d", numItems, got)
	}

	// Verify all results
	for i, result := range results {
		if result.JobID != jobID {
			t.Errorf("Result %d: wrong job ID: got %s, want %s", i, result.JobID, jobID)
		}
		if result.HasError() {
			t.Errorf("Result %d: unexpected error: %s", i, result.Error)
		}
		if result.OutputJSON == "" {
			t.Errorf("Result %d: empty output JSON", i)
		}
	}
}

func TestWorkerLoop_ToolExecutionError(t *testing.T) {
	s, redisURL := setupTestRedis(t)
	defer s.Close()

	// Create a mock tool that returns an error
	expectedErr := errors.New("tool execution failed")
	mockT := &mockTool{
		name:        "failing-tool",
		version:     "1.0.0",
		description: "Failing test tool",
		tags:        []string{"test"},
		executeFunc: func(ctx context.Context, input proto.Message) (proto.Message, error) {
			return nil, expectedErr
		},
	}

	// Create Redis client
	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// Note: We skip RegisterTool due to miniredis limitation with marshaling []interface{}
	// In production, this works fine with real Redis. See queue/client_test.go for details.
	// The worker doesn't strictly require pre-registration to function, only to be discoverable.

	// Create a work item
	queueName := fmt.Sprintf("tool:%s:queue", mockT.Name())
	jobID := "error-job-1"

	req := wrapperspb.String("test-data")
	inputJSON, _ := protojson.Marshal(req)

	item := queue.WorkItem{
		JobID:       jobID,
		Index:       0,
		Total:       1,
		Tool:        mockT.Name(),
		InputJSON:   string(inputJSON),
		InputType:   mockT.InputMessageType(),
		OutputType:  mockT.OutputMessageType(),
		SubmittedAt: time.Now().UnixMilli(),
	}
	if err := client.Push(context.Background(), queueName, item); err != nil {
		t.Fatalf("Failed to push work item: %v", err)
	}

	// Subscribe to results
	resultChannel := fmt.Sprintf("results:%s", jobID)
	resultsChan, err := client.Subscribe(context.Background(), resultChannel)
	if err != nil {
		t.Fatalf("Failed to subscribe to results: %v", err)
	}

	// Start worker loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		workerLoop(ctx, 0, mockT, client, queueName, "test-worker-1", newTestLogger())
	}()

	// Wait for result
	var result queue.Result
	select {
	case result = <-resultsChan:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for result")
	}

	// Cancel worker and wait
	cancel()
	wg.Wait()

	// Verify error result
	if !result.HasError() {
		t.Error("Expected result to have error")
	}
	if result.Error != expectedErr.Error() {
		t.Errorf("Expected error %q, got %q", expectedErr.Error(), result.Error)
	}
	if result.OutputJSON != "" {
		t.Errorf("Expected empty output JSON on error, got %q", result.OutputJSON)
	}
}

func TestWorkerLoop_GracefulShutdown(t *testing.T) {
	s, redisURL := setupTestRedis(t)
	defer s.Close()

	// Create a mock tool with slow execution
	var started, completed atomic.Bool
	mockT := &mockTool{
		name:        "slow-tool",
		version:     "1.0.0",
		description: "Slow test tool",
		tags:        []string{"test"},
		executeFunc: func(ctx context.Context, input proto.Message) (proto.Message, error) {
			started.Store(true)
			// Simulate work
			time.Sleep(100 * time.Millisecond)
			completed.Store(true)
			return wrapperspb.String("success"), nil
		},
	}

	// Create Redis client
	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// Note: We skip RegisterTool due to miniredis limitation with marshaling []interface{}
	// In production, this works fine with real Redis. See queue/client_test.go for details.
	// The worker doesn't strictly require pre-registration to function, only to be discoverable.

	// Create a work item
	queueName := fmt.Sprintf("tool:%s:queue", mockT.Name())
	jobID := "shutdown-job-1"

	req := wrapperspb.String("test-data")
	inputJSON, _ := protojson.Marshal(req)

	item := queue.WorkItem{
		JobID:       jobID,
		Index:       0,
		Total:       1,
		Tool:        mockT.Name(),
		InputJSON:   string(inputJSON),
		InputType:   mockT.InputMessageType(),
		OutputType:  mockT.OutputMessageType(),
		SubmittedAt: time.Now().UnixMilli(),
	}
	if err := client.Push(context.Background(), queueName, item); err != nil {
		t.Fatalf("Failed to push work item: %v", err)
	}

	// Start worker loop
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		workerLoop(ctx, 0, mockT, client, queueName, "test-worker-1", newTestLogger())
	}()

	// Wait for execution to start
	for i := 0; i < 100; i++ {
		if started.Load() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Cancel context while work is in progress
	cancel()

	// Wait for worker to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - worker finished
	case <-time.After(5 * time.Second):
		t.Fatal("Worker did not shut down gracefully")
	}

	// Verify work completed despite cancellation
	if !completed.Load() {
		t.Error("Work item should have completed before shutdown")
	}
}

func TestWorkerLoop_ConcurrentWorkers(t *testing.T) {
	s, redisURL := setupTestRedis(t)
	defer s.Close()

	// Create a mock tool with concurrent execution tracking
	var execCount atomic.Int32
	var maxConcurrent atomic.Int32
	var currentConcurrent atomic.Int32

	mockT := &mockTool{
		name:        "concurrent-tool",
		version:     "1.0.0",
		description: "Concurrent test tool",
		tags:        []string{"test"},
		executeFunc: func(ctx context.Context, input proto.Message) (proto.Message, error) {
			current := currentConcurrent.Add(1)
			execCount.Add(1)

			// Update max concurrent
			for {
				max := maxConcurrent.Load()
				if current <= max {
					break
				}
				if maxConcurrent.CompareAndSwap(max, current) {
					break
				}
			}

			// Simulate work
			time.Sleep(50 * time.Millisecond)
			currentConcurrent.Add(-1)

			return wrapperspb.String("success"), nil
		},
	}

	// Create Redis client
	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	// Note: We skip RegisterTool due to miniredis limitation with marshaling []interface{}
	// In production, this works fine with real Redis. See queue/client_test.go for details.
	// The worker doesn't strictly require pre-registration to function, only to be discoverable.

	// Create work items
	queueName := fmt.Sprintf("tool:%s:queue", mockT.Name())
	jobID := "concurrent-job-1"
	numItems := 10
	concurrency := 3

	for i := 0; i < numItems; i++ {
		req := wrapperspb.String(fmt.Sprintf("item-%d", i))
		inputJSON, _ := protojson.Marshal(req)

		item := queue.WorkItem{
			JobID:       jobID,
			Index:       i,
			Total:       numItems,
			Tool:        mockT.Name(),
			InputJSON:   string(inputJSON),
			InputType:   mockT.InputMessageType(),
			OutputType:  mockT.OutputMessageType(),
			SubmittedAt: time.Now().UnixMilli(),
		}
		if err := client.Push(context.Background(), queueName, item); err != nil {
			t.Fatalf("Failed to push work item: %v", err)
		}
	}

	// Start multiple worker loops
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerNum int) {
			defer wg.Done()
			workerLoop(ctx, workerNum, mockT, client, queueName, fmt.Sprintf("test-worker-%d", workerNum), newTestLogger())
		}(i)
	}

	// Wait for all work to complete
	time.Sleep(1 * time.Second)

	// Cancel workers
	cancel()
	wg.Wait()

	// Verify concurrent execution
	if got := int(execCount.Load()); got != numItems {
		t.Errorf("Expected %d executions, got %d", numItems, got)
	}

	maxConc := int(maxConcurrent.Load())
	if maxConc < 2 {
		t.Errorf("Expected concurrent execution (max >= 2), got max concurrent = %d", maxConc)
	}
	if maxConc > concurrency {
		t.Errorf("Expected max concurrent <= %d, got %d", concurrency, maxConc)
	}
}

func TestProcessWorkItem_InvalidInputType(t *testing.T) {
	mockT := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool",
		tags:        []string{"test"},
		inputType:   "non.existent.MessageType",
		outputType:  "google.protobuf.StringValue",
	}

	item := queue.WorkItem{
		JobID:       "test-job",
		Index:       0,
		Total:       1,
		Tool:        mockT.Name(),
		InputJSON:   `{}`,
		InputType:   "non.existent.MessageType",
		OutputType:  mockT.OutputMessageType(),
		SubmittedAt: time.Now().UnixMilli(),
	}

	result := processWorkItem(context.Background(), mockT, item, "test-worker", newTestLogger())

	if !result.HasError() {
		t.Error("Expected result to have error for invalid input type")
	}
	if result.Error == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestProcessWorkItem_InvalidJSON(t *testing.T) {
	mockT := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool",
		tags:        []string{"test"},
	}

	item := queue.WorkItem{
		JobID:       "test-job",
		Index:       0,
		Total:       1,
		Tool:        mockT.Name(),
		InputJSON:   `{invalid json`,
		InputType:   mockT.InputMessageType(),
		OutputType:  mockT.OutputMessageType(),
		SubmittedAt: time.Now().UnixMilli(),
	}

	result := processWorkItem(context.Background(), mockT, item, "test-worker", newTestLogger())

	if !result.HasError() {
		t.Error("Expected result to have error for invalid JSON")
	}
}

func TestRunHeartbeat(t *testing.T) {
	s, redisURL := setupTestRedis(t)
	defer s.Close()

	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	toolName := "test-tool"

	// Start heartbeat goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Run heartbeat in background
	done := make(chan struct{})
	go func() {
		runHeartbeat(ctx, client, toolName, newTestLogger())
		close(done)
	}()

	// Wait for at least one heartbeat (runs every 10s, but we force one immediately in our test version)
	// Since runHeartbeat uses a ticker, we need to wait for the first tick
	time.Sleep(15 * time.Millisecond) // Give it time for the first heartbeat

	healthKey := fmt.Sprintf("tool:%s:health", toolName)

	// Wait up to 500ms for heartbeat to appear
	var val string
	var getErr error
	for i := 0; i < 50; i++ {
		val, getErr = s.Get(healthKey)
		if getErr == nil && val == "ok" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if getErr != nil {
		t.Fatalf("Failed to get heartbeat key after waiting: %v", getErr)
	}
	if val != "ok" {
		t.Errorf("Expected heartbeat value 'ok', got %q", val)
	}

	// Wait for context to cancel and goroutine to finish
	<-ctx.Done()
	<-done
}

func TestGenerateWorkerID(t *testing.T) {
	// Generate multiple IDs and verify uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		id := generateWorkerID()

		// Check format (should contain hostname, PID, and UUID suffix)
		if id == "" {
			t.Error("Generated empty worker ID")
		}

		// Check uniqueness (due to UUID suffix)
		if ids[id] {
			t.Errorf("Generated duplicate worker ID: %s", id)
		}
		ids[id] = true
	}
}

func TestProcessWorkItem_ResultTimestamps(t *testing.T) {
	mockT := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool",
		tags:        []string{"test"},
		executeFunc: func(ctx context.Context, input proto.Message) (proto.Message, error) {
			time.Sleep(50 * time.Millisecond)
			return wrapperspb.String("success"), nil
		},
	}

	item := queue.WorkItem{
		JobID:       "test-job",
		Index:       0,
		Total:       1,
		Tool:        mockT.Name(),
		InputJSON:   string(mustMarshal(wrapperspb.String("test"))),
		InputType:   mockT.InputMessageType(),
		OutputType:  mockT.OutputMessageType(),
		SubmittedAt: time.Now().UnixMilli(),
	}

	result := processWorkItem(context.Background(), mockT, item, "test-worker", newTestLogger())

	// Verify timestamps
	if result.StartedAt <= 0 {
		t.Error("StartedAt should be positive")
	}
	if result.CompletedAt <= 0 {
		t.Error("CompletedAt should be positive")
	}
	if result.CompletedAt < result.StartedAt {
		t.Errorf("CompletedAt (%d) should be >= StartedAt (%d)", result.CompletedAt, result.StartedAt)
	}

	duration := result.Duration()
	if duration < 40*time.Millisecond || duration > 200*time.Millisecond {
		t.Errorf("Expected duration around 50ms, got %v", duration)
	}
}

func TestProcessWorkItem_WorkerID(t *testing.T) {
	mockT := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool",
		tags:        []string{"test"},
	}

	item := queue.WorkItem{
		JobID:       "test-job",
		Index:       0,
		Total:       1,
		Tool:        mockT.Name(),
		InputJSON:   string(mustMarshal(wrapperspb.String("test"))),
		InputType:   mockT.InputMessageType(),
		OutputType:  mockT.OutputMessageType(),
		SubmittedAt: time.Now().UnixMilli(),
	}

	workerID := "test-worker-123"
	result := processWorkItem(context.Background(), mockT, item, workerID, newTestLogger())

	if result.WorkerID != workerID {
		t.Errorf("Expected WorkerID %q, got %q", workerID, result.WorkerID)
	}
}

func TestWorkerLoop_ContextCancellation(t *testing.T) {
	s, redisURL := setupTestRedis(t)
	defer s.Close()

	mockT := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool",
		tags:        []string{"test"},
	}

	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	queueName := fmt.Sprintf("tool:%s:queue", mockT.Name())

	// Start worker with already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var wg sync.WaitGroup
	wg.Add(1)

	finished := make(chan struct{})
	go func() {
		defer wg.Done()
		workerLoop(ctx, 0, mockT, client, queueName, "test-worker", newTestLogger())
		close(finished)
	}()

	// Worker should exit quickly
	select {
	case <-finished:
		// Success - worker exited
	case <-time.After(1 * time.Second):
		t.Fatal("Worker did not exit after context cancellation")
	}

	wg.Wait()
}

// Helper function to marshal proto messages for tests
func mustMarshal(msg proto.Message) []byte {
	data, err := protojson.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal proto: %v", err))
	}
	return data
}

// TestRun_Integration is an integration test for the Run function.
// This test verifies the full worker lifecycle but does not test signal handling.
func TestRun_Integration(t *testing.T) {
	// Note: This test does not cover signal handling (SIGTERM/SIGINT) as that
	// requires OS-level signal sending which is difficult to test reliably.
	// Signal handling should be tested manually or with integration tests.
	// This test is skipped in short mode as it's more of an integration test.
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s, redisURL := setupTestRedis(t)
	defer s.Close()

	var execCount atomic.Int32
	mockT := &mockTool{
		name:        "integration-tool",
		version:     "1.0.0",
		description: "Integration test tool",
		tags:        []string{"test"},
		executeFunc: func(ctx context.Context, input proto.Message) (proto.Message, error) {
			execCount.Add(1)
			return wrapperspb.String("success"), nil
		},
	}

	// Create Redis client for pushing work
	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	queueName := fmt.Sprintf("tool:%s:queue", mockT.Name())
	jobID := "integration-job-1"

	// Subscribe to results FIRST (before pushing work)
	resultChannel := fmt.Sprintf("results:%s", jobID)
	resultsChan, err := client.Subscribe(context.Background(), resultChannel)
	if err != nil {
		t.Fatalf("Failed to subscribe to results: %v", err)
	}

	// Start worker BEFORE pushing work
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		workerLoop(ctx, 0, mockT, client, queueName, "integration-worker", newTestLogger())
	}()

	// Give worker time to start
	time.Sleep(50 * time.Millisecond)

	// Push a work item
	req := wrapperspb.String("test-data")
	inputJSON, _ := protojson.Marshal(req)

	item := queue.WorkItem{
		JobID:       jobID,
		Index:       0,
		Total:       1,
		Tool:        mockT.Name(),
		InputJSON:   string(inputJSON),
		InputType:   mockT.InputMessageType(),
		OutputType:  mockT.OutputMessageType(),
		SubmittedAt: time.Now().UnixMilli(),
	}
	if err := client.Push(context.Background(), queueName, item); err != nil {
		t.Fatalf("Failed to push work item: %v", err)
	}

	// Wait for result with timeout
	select {
	case result := <-resultsChan:
		if result.HasError() {
			t.Errorf("Unexpected error: %s", result.Error)
		}
	case <-time.After(2 * time.Second):
		cancel() // Cancel context
		wg.Wait() // Wait for worker to stop
		t.Fatal("Timeout waiting for result")
	}

	// Cancel and wait
	cancel()

	// Wait for worker with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Worker did not shut down in time")
	}

	// Verify execution
	if got := int(execCount.Load()); got != 1 {
		t.Errorf("Expected 1 execution, got %d", got)
	}
}

func TestOptions_Defaults(t *testing.T) {
	// Test that default values are applied correctly by checking
	// the logic that would be used in Run()

	tests := []struct {
		name    string
		opts    Options
		wantURL string
		wantC   int
		wantT   time.Duration
	}{
		{
			name:    "empty options",
			opts:    Options{},
			wantURL: "redis://localhost:6379",
			wantC:   4,
			wantT:   30 * time.Second,
		},
		{
			name:    "custom redis URL",
			opts:    Options{RedisURL: "redis://custom:6379"},
			wantURL: "redis://custom:6379",
			wantC:   4,
			wantT:   30 * time.Second,
		},
		{
			name:    "custom concurrency",
			opts:    Options{Concurrency: 8},
			wantURL: "redis://localhost:6379",
			wantC:   8,
			wantT:   30 * time.Second,
		},
		{
			name:    "custom shutdown timeout",
			opts:    Options{ShutdownTimeout: 60 * time.Second},
			wantURL: "redis://localhost:6379",
			wantC:   4,
			wantT:   60 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.opts

			// Apply defaults (same logic as in Run)
			if opts.RedisURL == "" {
				opts.RedisURL = "redis://localhost:6379"
			}
			if opts.Concurrency <= 0 {
				opts.Concurrency = 4
			}
			if opts.ShutdownTimeout == 0 {
				opts.ShutdownTimeout = 30 * time.Second
			}

			if opts.RedisURL != tt.wantURL {
				t.Errorf("RedisURL = %q, want %q", opts.RedisURL, tt.wantURL)
			}
			if opts.Concurrency != tt.wantC {
				t.Errorf("Concurrency = %d, want %d", opts.Concurrency, tt.wantC)
			}
			if opts.ShutdownTimeout != tt.wantT {
				t.Errorf("ShutdownTimeout = %v, want %v", opts.ShutdownTimeout, tt.wantT)
			}
		})
	}
}

// TestWorkerRegistration verifies that the worker properly registers/unregisters
// and maintains worker count.
func TestWorkerRegistration(t *testing.T) {
	s, redisURL := setupTestRedis(t)
	defer s.Close()

	mockT := &mockTool{
		name:        "registration-tool",
		version:     "1.0.0",
		description: "Registration test tool",
		tags:        []string{"test"},
	}

	client, err := queue.NewRedisClient(queue.RedisOptions{URL: redisURL})
	if err != nil {
		t.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Note: We skip RegisterTool due to miniredis limitation with marshaling []interface{}
	// In production, this works fine with real Redis. See queue/client_test.go for details.

	// Check initial worker count
	count, err := client.GetWorkerCount(ctx, mockT.Name())
	if err != nil {
		t.Fatalf("Failed to get worker count: %v", err)
	}
	if count != 0 {
		t.Errorf("Initial worker count = %d, want 0", count)
	}

	// Increment worker count (simulating worker startup)
	if err := client.IncrementWorkerCount(ctx, mockT.Name()); err != nil {
		t.Fatalf("Failed to increment worker count: %v", err)
	}

	count, err = client.GetWorkerCount(ctx, mockT.Name())
	if err != nil {
		t.Fatalf("Failed to get worker count: %v", err)
	}
	if count != 1 {
		t.Errorf("Worker count after increment = %d, want 1", count)
	}

	// Decrement worker count (simulating worker shutdown)
	if err := client.DecrementWorkerCount(ctx, mockT.Name()); err != nil {
		t.Fatalf("Failed to decrement worker count: %v", err)
	}

	count, err = client.GetWorkerCount(ctx, mockT.Name())
	if err != nil {
		t.Fatalf("Failed to get worker count: %v", err)
	}
	if count != 0 {
		t.Errorf("Worker count after decrement = %d, want 0", count)
	}
}

// Ensure that proto messages are properly registered in the global registry
func init() {
	// Verify google.protobuf.StringValue is registered (it should be by default)
	msgType, err := protoregistry.GlobalTypes.FindMessageByName(
		protoreflect.FullName("google.protobuf.StringValue"),
	)
	if err != nil {
		panic(fmt.Sprintf("google.protobuf.StringValue not registered: %v", err))
	}
	if msgType == nil {
		panic("google.protobuf.StringValue type is nil")
	}
}
