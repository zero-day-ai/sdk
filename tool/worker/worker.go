package worker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/zero-day-ai/sdk/component"
	"github.com/zero-day-ai/sdk/queue"
	"github.com/zero-day-ai/sdk/tool"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// Options configures the worker behavior.
type Options struct {
	// RedisURL is the Redis connection string (e.g., "redis://localhost:6379")
	RedisURL string

	// Concurrency is the number of worker goroutines to start.
	// If 0, uses value from component.yaml or default (4).
	Concurrency int

	// ShutdownTimeout is the time to wait for graceful shutdown.
	// If 0, uses value from component.yaml or default (30s).
	ShutdownTimeout time.Duration

	// Logger is the structured logger for worker operations.
	// If nil, a default logger will be created.
	Logger *slog.Logger

	// ComponentConfig is the parsed component.yaml configuration.
	// If nil, the worker will attempt to load it from the current directory.
	// Set to an empty config to skip component.yaml loading.
	ComponentConfig *component.Config

	// ConfigPath is the path to component.yaml.
	// If empty and ComponentConfig is nil, searches from current directory.
	ConfigPath string
}

// Run starts the worker loop for the given tool with the specified options.
// It connects to Redis, registers the tool, starts N worker goroutines based on Concurrency,
// maintains a heartbeat, and handles graceful shutdown on SIGTERM/SIGINT.
//
// Configuration priority (highest to lowest):
//  1. Explicit Options values (if non-zero)
//  2. component.yaml worker section
//  3. Default values
//
// Each worker goroutine:
//  1. Pops a work item from the queue
//  2. Executes the tool with the work item input
//  3. Publishes the result back to Redis
//
// The function blocks until a shutdown signal is received or an error occurs.
// On shutdown, it waits for all workers to finish processing their current items
// before returning.
//
// Returns an error if Redis connection fails or if graceful shutdown times out.
func Run(t tool.Tool, opts Options) error {
	// Load component.yaml if not provided
	componentCfg := opts.ComponentConfig
	if componentCfg == nil {
		var err error
		if opts.ConfigPath != "" {
			componentCfg, err = component.Load(opts.ConfigPath)
		} else {
			componentCfg, err = component.LoadFromCurrentDir()
		}
		if err != nil {
			// component.yaml is optional - just use defaults
			componentCfg = nil
		}
	}

	// Apply configuration with priority: explicit opts > component.yaml > defaults
	opts = applyComponentConfig(opts, componentCfg)

	// Set remaining defaults
	if opts.RedisURL == "" {
		opts.RedisURL = "redis://localhost:6379"
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	// Generate unique worker ID (hostname + PID + UUID)
	workerID := generateWorkerID()

	logger := opts.Logger.With(
		"tool", t.Name(),
		"version", t.Version(),
		"worker_id", workerID,
	)

	logger.Info("worker starting",
		"concurrency", opts.Concurrency,
		"redis_url", opts.RedisURL,
	)

	// Connect to Redis
	redisClient, err := queue.NewRedisClient(queue.RedisOptions{
		URL: opts.RedisURL,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	defer redisClient.Close()

	// Create context for worker lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register tool with Redis
	meta := queue.ToolMeta{
		Name:              t.Name(),
		Version:           t.Version(),
		Description:       t.Description(),
		InputMessageType:  t.InputMessageType(),
		OutputMessageType: t.OutputMessageType(),
		Tags:              t.Tags(),
		Schema:            "", // Schema() is not available on tool.Tool interface
		WorkerCount:       0,  // Updated separately
	}

	logger.Info("registering tool",
		"name", meta.Name,
		"version", meta.Version,
		"input_type", meta.InputMessageType,
		"output_type", meta.OutputMessageType,
	)

	if err := redisClient.RegisterTool(ctx, meta); err != nil {
		logger.Error("failed to register tool", "error", err)
		return fmt.Errorf("failed to register tool: %w", err)
	}

	logger.Info("tool registered successfully")

	// Increment worker count on startup
	if err := redisClient.IncrementWorkerCount(ctx, t.Name()); err != nil {
		logger.Error("failed to increment worker count", "error", err)
	}

	// Ensure worker count is decremented on exit (even on crash)
	defer func() {
		// Use background context for cleanup since ctx may be cancelled
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if err := redisClient.DecrementWorkerCount(cleanupCtx, t.Name()); err != nil {
			logger.Error("failed to decrement worker count", "error", err)
		}
	}()

	// Start heartbeat goroutine
	heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
	defer stopHeartbeat()
	go runHeartbeat(heartbeatCtx, redisClient, t.Name(), logger)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start worker goroutines
	var wg sync.WaitGroup
	queueName := fmt.Sprintf("tool:%s:queue", t.Name())

	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func(workerNum int) {
			defer wg.Done()
			workerLoop(ctx, workerNum, t, redisClient, queueName, workerID, logger)
		}(i)
	}

	logger.Info("worker started",
		"workers", opts.Concurrency,
		"queue", queueName,
	)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("received signal, initiating graceful shutdown", "signal", sig)

	// Cancel context to stop workers and heartbeat
	cancel()

	// Wait for workers to finish with timeout
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		logger.Info("worker shutdown complete")
	case <-time.After(opts.ShutdownTimeout):
		logger.Warn("worker shutdown timeout exceeded", "timeout", opts.ShutdownTimeout)
	}

	return nil
}

// runHeartbeat sends periodic heartbeats to maintain tool health status.
// It runs in a goroutine and stops when the context is cancelled.
func runHeartbeat(ctx context.Context, client queue.Client, toolName string, logger *slog.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	logger.Debug("heartbeat goroutine started")

	for {
		select {
		case <-ctx.Done():
			logger.Debug("heartbeat goroutine stopped")
			return
		case <-ticker.C:
			if err := client.Heartbeat(ctx, toolName); err != nil {
				// Log at debug level to avoid noise - heartbeat failures are transient
				logger.Debug("heartbeat failed", "error", err)
			}
		}
	}
}

// workerLoop is the main loop for a single worker goroutine.
// It continuously pops work items from the queue, processes them,
// and publishes results until the context is cancelled.
func workerLoop(ctx context.Context, workerNum int, t tool.Tool, client queue.Client, queueName, workerID string, logger *slog.Logger) {
	logger = logger.With("worker_num", workerNum)
	logger.Debug("worker loop started", "queue", queueName)

	for {
		// Check if context is cancelled before popping
		select {
		case <-ctx.Done():
			logger.Debug("worker loop stopped", "reason", "context_cancelled")
			return
		default:
		}

		// Pop work item from queue (blocking with context)
		item, err := client.Pop(ctx, queueName)
		if err != nil {
			// Check if context was cancelled during Pop
			if ctx.Err() != nil {
				logger.Debug("worker loop stopped", "reason", "context_error")
				return
			}
			// Log error and continue
			logger.Error("failed to pop work item", "error", err)
			continue
		}

		// Check if Pop returned nil (shouldn't happen but handle it)
		if item == nil {
			continue
		}

		logger.Info("received work item",
			"job_id", item.JobID,
			"index", item.Index,
			"total", item.Total,
			"tool", item.Tool,
		)

		// Process work item
		result := processWorkItem(ctx, t, *item, workerID, logger)

		// Publish result to job-specific channel
		resultChannel := fmt.Sprintf("results:%s", item.JobID)
		if err := client.Publish(ctx, resultChannel, result); err != nil {
			logger.Error("failed to publish result", "error", err)
		}
	}
}

// processWorkItem processes a single work item and returns a result.
// It handles all errors at each step and ensures a result is always returned.
func processWorkItem(ctx context.Context, t tool.Tool, item queue.WorkItem, workerID string, logger *slog.Logger) queue.Result {
	startedAt := time.Now().UnixMilli()

	result := queue.Result{
		JobID:       item.JobID,
		Index:       item.Index,
		OutputType:  item.OutputType,
		WorkerID:    workerID,
		StartedAt:   startedAt,
		CompletedAt: 0, // Set later
	}

	// Find the input proto message type
	inputMsgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(item.InputType))
	if err != nil {
		result.Error = fmt.Sprintf("unknown input type: %s", item.InputType)
		result.CompletedAt = time.Now().UnixMilli()
		logger.Error("unknown input type", "input_type", item.InputType, "error", err)
		return result
	}

	// Create a new instance of the input message
	inputMsg := inputMsgType.New().Interface()

	// Unmarshal JSON to proto
	if err := protojson.Unmarshal([]byte(item.InputJSON), inputMsg); err != nil {
		result.Error = fmt.Sprintf("failed to unmarshal input: %v", err)
		result.CompletedAt = time.Now().UnixMilli()
		logger.Error("failed to unmarshal input", "error", err)
		return result
	}

	// Execute tool
	outputMsg, err := t.ExecuteProto(ctx, inputMsg)
	if err != nil {
		result.Error = err.Error()
		result.CompletedAt = time.Now().UnixMilli()
		logger.Error("tool execution failed", "error", err)
		return result
	}

	// Marshal output to JSON
	outputJSON, err := protojson.Marshal(outputMsg)
	if err != nil {
		result.Error = fmt.Sprintf("failed to marshal output: %v", err)
		result.CompletedAt = time.Now().UnixMilli()
		logger.Error("failed to marshal output", "error", err)
		return result
	}

	result.OutputJSON = string(outputJSON)
	result.CompletedAt = time.Now().UnixMilli()

	logger.Info("work item completed",
		"job_id", item.JobID,
		"index", item.Index,
		"duration_ms", result.CompletedAt-result.StartedAt,
	)

	return result
}

// generateWorkerID creates a unique identifier for this worker instance.
// Uses hostname + PID + UUID for uniqueness.
func generateWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	pid := os.Getpid()

	// Add UUID suffix for additional uniqueness
	id := uuid.New().String()[:8]

	return fmt.Sprintf("%s-%d-%s", hostname, pid, id)
}

// applyComponentConfig applies component.yaml settings to Options.
// Explicit Options values take priority over component.yaml values.
func applyComponentConfig(opts Options, cfg *component.Config) Options {
	if cfg == nil || cfg.Worker == nil {
		// No component.yaml or no worker section - use defaults
		if opts.Concurrency <= 0 {
			opts.Concurrency = 4
		}
		if opts.ShutdownTimeout == 0 {
			opts.ShutdownTimeout = 30 * time.Second
		}
		return opts
	}

	// Apply component.yaml values only if opts doesn't have explicit values
	if opts.Concurrency <= 0 {
		opts.Concurrency = cfg.Worker.GetConcurrency()
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = cfg.Worker.GetShutdownTimeout()
	}

	return opts
}
