// Package queue provides Redis-based work queue primitives for distributed tool execution.
//
// The queue package enables horizontal scaling of tool execution by decoupling
// work submission from execution. Agents submit work items to Redis queues,
// workers consume and execute them, and results flow back through Redis pub/sub.
//
// # Core Components
//
// Client: Interface for interacting with Redis queues. Provides methods for:
//   - Push/Pop operations for work queues
//   - Publish/Subscribe for result delivery
//   - Tool registration and discovery
//   - Health monitoring and worker tracking
//
// WorkItem: A unit of work containing tool name, input data, and trace context.
//
// Result: The outcome of executing a WorkItem, including output or error.
//
// ToolMeta: Metadata about a registered tool for discovery and routing.
//
// # Redis Key Schema
//
// The queue system uses a structured key naming convention:
//   - tool:<name>:queue - List for work items (LPUSH/BRPOP)
//   - tool:<name>:meta - Hash for tool metadata
//   - tool:<name>:health - String with 30s TTL for heartbeat
//   - tool:<name>:workers - Integer counter for active workers
//   - tools:available - Set of all registered tool names
//   - results:<jobID> - Pub/Sub channel for job results
//
// # Usage
//
// Creating a queue client:
//
//	client := queue.NewRedisClient(queue.RedisOptions{
//		URL: "redis://localhost:6379",
//		TLS: nil,
//		ConnectTimeout: 5 * time.Second,
//	})
//
// Pushing work to a queue:
//
//	err := client.Push(ctx, "tool:nmap:queue", queue.WorkItem{
//		JobID: "job-123",
//		Index: 0,
//		Total: 1,
//		Tool: "nmap",
//		InputJSON: `{"target":"192.168.1.1"}`,
//		InputType: "gibson.tools.nmap.v1.ScanRequest",
//		SubmittedAt: time.Now().UnixMilli(),
//	})
//
// Popping work from a queue (blocking):
//
//	item, err := client.Pop(ctx, "tool:nmap:queue")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// Process item...
//
// Publishing results:
//
//	err := client.Publish(ctx, "results:job-123", queue.Result{
//		JobID: "job-123",
//		Index: 0,
//		OutputJSON: `{"hosts":[...]}`,
//		CompletedAt: time.Now().UnixMilli(),
//	})
//
// Subscribing to results:
//
//	results, err := client.Subscribe(ctx, "results:job-123")
//	if err != nil {
//		log.Fatal(err)
//	}
//	for result := range results {
//		fmt.Printf("Received result %d/%d\n", result.Index, result.Total)
//	}
//
// Registering a tool:
//
//	err := client.RegisterTool(ctx, queue.ToolMeta{
//		Name: "nmap",
//		Version: "1.0.0",
//		Description: "Network scanner",
//		InputMessageType: "gibson.tools.nmap.v1.ScanRequest",
//		OutputMessageType: "gibson.tools.nmap.v1.ScanResponse",
//		Tags: []string{"discovery", "network"},
//	})
//
// Listing available tools:
//
//	tools, err := client.ListTools(ctx)
//	for _, tool := range tools {
//		fmt.Printf("Tool: %s v%s\n", tool.Name, tool.Version)
//	}
//
// Sending heartbeats:
//
//	ticker := time.NewTicker(10 * time.Second)
//	for range ticker.C {
//		if err := client.Heartbeat(ctx, "nmap"); err != nil {
//			log.Printf("Heartbeat failed: %v", err)
//		}
//	}
//
// # Error Handling
//
// All methods return errors for Redis connection failures, serialization
// errors, or context cancellation. Clients should implement retry logic
// with exponential backoff for transient failures.
//
// # Thread Safety
//
// RedisClient is safe for concurrent use by multiple goroutines.
package queue
