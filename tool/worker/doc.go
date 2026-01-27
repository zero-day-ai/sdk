// Package worker provides the main loop for running Gibson tools as Redis queue workers.
//
// # Overview
//
// The worker package enables tools to be run as background workers that consume
// work items from Redis queues and publish results back. This allows the Gibson
// framework to scale tool execution horizontally across multiple worker processes
// or containers.
//
// # Architecture
//
// Workers operate in a producer-consumer pattern:
//   - Gibson daemon (producer): Pushes WorkItems to Redis queues
//   - Tool workers (consumers): Pop WorkItems, execute tools, publish Results
//   - Gibson daemon (collector): Subscribes to results and returns to agents
//
// # Usage
//
// To create a worker for your tool:
//
//	func main() {
//	    // Create your tool instance
//	    myTool := &MyTool{}
//
//	    // Configure worker options
//	    opts := worker.Options{
//	        RedisURL:        "redis://localhost:6379",
//	        Concurrency:     4,  // Number of worker goroutines
//	        ShutdownTimeout: 30 * time.Second,
//	    }
//
//	    // Run the worker (blocks until shutdown)
//	    if err := worker.Run(myTool, opts); err != nil {
//	        log.Fatalf("Worker failed: %v", err)
//	    }
//	}
//
// # Concurrency
//
// The Concurrency option controls how many goroutines will process work items
// in parallel. Higher concurrency allows processing more work items simultaneously
// but consumes more resources. Choose based on:
//   - Tool execution time (longer = more concurrency beneficial)
//   - Resource constraints (CPU, memory, file descriptors)
//   - Queue depth (more queued work = benefit from higher concurrency)
//
// # Graceful Shutdown
//
// Workers handle SIGTERM and SIGINT signals gracefully:
//  1. Signal received → context cancelled
//  2. Workers finish processing current items
//  3. No new items are popped from queue
//  4. Workers exit once current work completes
//  5. Run() returns (or times out after ShutdownTimeout)
//
// This ensures work items are not left in an inconsistent state during shutdown.
//
// # Redis Queue Schema
//
// Workers interact with Redis using the following key patterns:
//   - tool:<name>:queue - List containing WorkItems (LPUSH/BRPOP)
//   - tool:<name>:meta - Hash containing tool metadata
//   - tool:<name>:health - Key with TTL for health checks
//   - tool:<name>:workers - Counter for active worker count
//   - results:<jobID> - Pub/sub channel for result delivery
//
// # Error Handling
//
// The worker loop is designed to be resilient:
//   - Redis connection errors: Fatal, causes Run() to return
//   - Pop errors: Logged and loop continues
//   - Tool execution errors: Captured and published as error Results
//   - Context cancellation: Graceful shutdown initiated
//
// # Future Enhancements
//
// Task 2.2 will add:
//   - Tool registration on startup
//   - Heartbeat goroutine (every 10s)
//   - Worker count tracking
//
// Task 2.3 will add:
//   - Work item processing (proto unmarshaling)
//   - Tool execution
//   - Result publishing
package worker
