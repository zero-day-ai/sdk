// Package memory provides a three-tier memory system for Gibson agents.
//
// The memory system is organized into three distinct tiers, each serving
// a different purpose and lifecycle:
//
//   - Working Memory: Ephemeral, in-memory key-value storage for temporary data
//     that only needs to persist during a single execution context.
//
//   - Mission Memory: Persistent storage scoped to a specific mission, allowing
//     agents to maintain state and context across multiple interactions within
//     the same mission.
//
//   - Long-Term Memory: Vector-based semantic storage for knowledge that persists
//     across missions, enabling agents to learn and recall information over time.
//
// # Working Memory
//
// Working memory provides a simple key-value interface for ephemeral data:
//
//	store := // ... obtain Store implementation
//	working := store.Working()
//
//	// Store temporary data
//	err := working.Set(ctx, "current_step", 3)
//
//	// Retrieve data
//	value, err := working.Get(ctx, "current_step")
//	step := value.(int)
//
//	// Clear when done
//	err = working.Clear(ctx)
//
// Working memory is typically cleared between agent executions and is not persisted
// to disk. It's ideal for tracking intermediate state during complex operations.
//
// # Mission Memory
//
// Mission memory provides persistent, structured storage for mission-scoped data:
//
//	mission := store.Mission()
//
//	// Store with metadata
//	err := mission.Set(ctx, "user_preference", "dark_mode", map[string]any{
//	    "category": "ui",
//	    "priority": 1,
//	})
//
//	// Retrieve with metadata
//	item, err := mission.Get(ctx, "user_preference")
//	fmt.Printf("Value: %v, Created: %v\n", item.Value, item.CreatedAt)
//
//	// Search by content
//	results, err := mission.Search(ctx, "dark mode settings", 10)
//	for _, result := range results {
//	    fmt.Printf("Match: %s (score: %.2f)\n", result.Key, result.Score)
//	}
//
//	// Get recent history
//	history, err := mission.History(ctx, 20)
//
// Mission memory persists data to disk and maintains metadata including creation
// and update timestamps. It supports full-text search across stored values.
//
// # Long-Term Memory
//
// Long-term memory uses vector embeddings for semantic search and retrieval:
//
//	longTerm := store.LongTerm()
//
//	// Store information with semantic embeddings
//	id, err := longTerm.Store(ctx, "The capital of France is Paris", map[string]any{
//	    "category": "geography",
//	    "source": "wikipedia",
//	})
//
//	// Search semantically
//	results, err := longTerm.Search(ctx, "What is the capital of France?", 5, nil)
//	for _, result := range results {
//	    fmt.Printf("Content: %s (score: %.2f)\n", result.Value, result.Score)
//	}
//
//	// Search with filters
//	results, err = longTerm.Search(ctx, "European capitals", 10, map[string]any{
//	    "category": "geography",
//	})
//
// Long-term memory persists across missions and enables agents to build up
// knowledge over time. The vector-based search allows for semantic matching
// even when query terms don't exactly match stored content.
//
// # Store Access
//
// The Store interface provides unified access to all three memory tiers:
//
//	type Store interface {
//	    Working() WorkingMemory
//	    Mission() MissionMemory
//	    LongTerm() LongTermMemory
//	}
//
// Implementations of Store are responsible for managing the lifecycle and
// persistence of each memory tier according to their respective semantics.
//
// # Context and Cancellation
//
// All memory operations accept a context.Context parameter, allowing for
// proper timeout handling and cancellation:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	item, err := mission.Get(ctx, "key")
//	if err != nil {
//	    if errors.Is(err, context.DeadlineExceeded) {
//	        // Handle timeout
//	    }
//	}
//
// # Error Handling
//
// Memory operations return errors that can be checked with standard error
// handling patterns. Common errors include:
//
//   - Item not found
//   - Invalid key or value
//   - Storage backend failures
//   - Context cancellation or timeout
//
// Check specific error conditions using errors.Is() or errors.As():
//
//	item, err := mission.Get(ctx, "missing_key")
//	if err != nil {
//	    if errors.Is(err, ErrNotFound) {
//	        // Handle missing item
//	    }
//	}
package memory
