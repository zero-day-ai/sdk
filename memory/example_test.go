package memory_test

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/zero-day-ai/sdk/memory"
)

// mockStore provides a simple example implementation for testing.
// In production, use a concrete implementation that matches your storage needs.
type exampleStore struct {
	working  memory.WorkingMemory
	mission  memory.MissionMemory
	longTerm memory.LongTermMemory
}

func (s *exampleStore) Working() memory.WorkingMemory {
	return s.working
}

func (s *exampleStore) Mission() memory.MissionMemory {
	return s.mission
}

func (s *exampleStore) LongTerm() memory.LongTermMemory {
	return s.longTerm
}

// Example demonstrates basic usage of the three-tier memory system.
func Example() {
	ctx := context.Background()

	// In production, create a real store implementation
	// store := mystore.New(config)

	// For this example, we'll use a mock
	store := &exampleStore{
		// Implementations would be injected here
	}

	// Working Memory - for temporary data
	working := store.Working()
	if working != nil {
		working.Set(ctx, "current_step", 1)
		step, _ := working.Get(ctx, "current_step")
		fmt.Printf("Current step: %v\n", step)
	}

	// Mission Memory - for mission-scoped persistence
	mission := store.Mission()
	if mission != nil {
		mission.Set(ctx, "user_pref", "dark_mode", map[string]any{
			"category": "ui",
		})

		item, _ := mission.Get(ctx, "user_pref")
		if item != nil {
			fmt.Printf("Preference: %v\n", item.Value)
		}
	}

	// Long-Term Memory - for semantic knowledge
	longTerm := store.LongTerm()
	if longTerm != nil {
		longTerm.Store(ctx, "Important fact to remember", map[string]any{
			"priority": "high",
		})

		results, _ := longTerm.Search(ctx, "important", 5, nil)
		fmt.Printf("Found %d results\n", len(results))
	}
}

// ExampleWorkingMemory demonstrates ephemeral key-value storage.
func ExampleWorkingMemory() {
	ctx := context.Background()

	// Obtain working memory from store
	var working memory.WorkingMemory
	// working = store.Working()

	// For demonstration, we'll check if it's initialized
	if working == nil {
		fmt.Println("Working memory provides fast, ephemeral storage")
		return
	}

	// Store temporary calculation state
	working.Set(ctx, "partial_sum", 42)
	working.Set(ctx, "iteration", 3)

	// Retrieve when needed
	sum, _ := working.Get(ctx, "partial_sum")
	fmt.Printf("Partial sum: %v\n", sum)

	// List all keys
	keys, _ := working.Keys(ctx)
	fmt.Printf("Active keys: %v\n", keys)

	// Clear when done
	working.Clear(ctx)

	// Output:
	// Working memory provides fast, ephemeral storage
}

// ExampleMissionMemory demonstrates persistent mission-scoped storage.
func ExampleMissionMemory() {
	ctx := context.Background()

	var mission memory.MissionMemory
	// mission = store.Mission()

	if mission == nil {
		fmt.Println("Mission memory provides persistent, structured storage")
		return
	}

	// Store with rich metadata
	mission.Set(ctx, "user_language", "en-US", map[string]any{
		"category":   "i18n",
		"updated_by": "user",
		"timestamp":  time.Now().Unix(),
	})

	// Retrieve with full context
	item, _ := mission.Get(ctx, "user_language")
	if item != nil {
		fmt.Printf("Language: %v\n", item.Value)
		fmt.Printf("Category: %v\n", item.Metadata["category"])
		fmt.Printf("Created: %v\n", item.CreatedAt)
	}

	// Search across all stored items
	results, _ := mission.Search(ctx, "language settings", 10)
	for _, result := range results {
		fmt.Printf("Found: %s (relevance: %.2f)\n", result.Key, result.Score)
	}

	// Get recent history
	history, _ := mission.History(ctx, 5)
	fmt.Printf("Recent items: %d\n", len(history))

	// Output:
	// Mission memory provides persistent, structured storage
}

// ExampleLongTermMemory demonstrates vector-based semantic storage.
func ExampleLongTermMemory() {
	ctx := context.Background()

	var longTerm memory.LongTermMemory
	// longTerm = store.LongTerm()

	if longTerm == nil {
		fmt.Println("Long-term memory provides semantic, vector-based storage")
		return
	}

	// Store knowledge with metadata
	id, _ := longTerm.Store(ctx,
		"Go uses goroutines for concurrency, which are lightweight threads",
		map[string]any{
			"category": "programming",
			"language": "go",
			"topic":    "concurrency",
		},
	)
	fmt.Printf("Stored with ID: %s\n", id)

	// Semantic search - finds related content even with different wording
	results, _ := longTerm.Search(ctx,
		"How does Go handle parallel execution?",
		5,
		nil,
	)

	for _, result := range results {
		fmt.Printf("Content: %s\n", result.Value)
		fmt.Printf("Relevance: %.2f\n", result.Score)
	}

	// Search with metadata filters
	goResults, _ := longTerm.Search(ctx,
		"concurrency patterns",
		10,
		map[string]any{"language": "go"},
	)
	fmt.Printf("Go-specific results: %d\n", len(goResults))

	// Output:
	// Long-term memory provides semantic, vector-based storage
}

// ExampleItem demonstrates working with memory items.
func ExampleItem() {
	now := time.Now()
	item := &memory.Item{
		Key:   "user_profile",
		Value: map[string]any{"name": "Alice", "role": "admin"},
		Metadata: map[string]any{
			"category": "user_data",
			"version":  1,
		},
		CreatedAt: now.Add(-24 * time.Hour),
		UpdatedAt: now,
	}

	// Access metadata safely
	if category, ok := item.GetMetadata("category"); ok {
		fmt.Printf("Category: %v\n", category)
	}

	// Add metadata
	item.SetMetadata("last_login", now.Unix())

	// Check if modified
	if item.IsModified() {
		duration := item.UpdatedAt.Sub(item.CreatedAt)
		fmt.Printf("Modified after: %.0fh\n", duration.Hours())
	}

	// Clone for safe modification
	clone := item.Clone()
	clone.Value = map[string]any{"name": "Bob"}
	// Original item is unchanged

	fmt.Printf("Age: %.0fh\n", item.Age().Hours())

	// Output:
	// Category: user_data
	// Modified after: 24h
	// Age: 24h
}

// ExampleResult demonstrates working with search results.
func ExampleResult() {
	result := &memory.Result{
		Item: memory.Item{
			Key:       "golang_concurrency",
			Value:     "Go uses goroutines for lightweight concurrency",
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},
		Score: 0.92,
	}

	fmt.Printf("Result: %s\n", result.Key)
	fmt.Printf("Relevance: %.0f%%\n", result.Score*100)
	fmt.Printf("Content: %v\n", result.Value)

	// Access item methods
	if result.IsModified() {
		fmt.Println("Content has been updated")
	}

	// Output:
	// Result: golang_concurrency
	// Relevance: 92%
	// Content: Go uses goroutines for lightweight concurrency
	// Content has been updated
}

// Example_errorHandling demonstrates proper error handling patterns.
func Example_errorHandling() {
	ctx := context.Background()

	var mission memory.MissionMemory
	// mission = store.Mission()

	if mission == nil {
		log.Println("Store not initialized")
		return
	}

	// Handle not found errors
	item, err := mission.Get(ctx, "nonexistent_key")
	if err != nil {
		if err == memory.ErrNotFound {
			fmt.Println("Item not found - creating new one")
			mission.Set(ctx, "nonexistent_key", "default_value", nil)
		} else {
			log.Printf("Unexpected error: %v", err)
		}
	} else {
		fmt.Printf("Found: %v\n", item.Value)
	}

	// Handle context timeouts
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err = mission.Search(timeoutCtx, "query", 10)
	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Println("Search timed out")
		}
	}
}

// Example_concurrentAccess demonstrates safe concurrent access patterns.
func Example_concurrentAccess() {
	ctx := context.Background()

	var working memory.WorkingMemory
	// working = store.Working()

	if working == nil {
		fmt.Println("Safe concurrent access is implementation-dependent")
		return
	}

	// Most implementations provide thread-safe operations
	done := make(chan bool)

	// Goroutine 1: Write data
	go func() {
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key_%d", i)
			working.Set(ctx, key, i)
		}
		done <- true
	}()

	// Goroutine 2: Read data
	go func() {
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("key_%d", i)
			working.Get(ctx, key)
		}
		done <- true
	}()

	// Wait for completion
	<-done
	<-done

	keys, _ := working.Keys(ctx)
	fmt.Printf("Final key count: %d\n", len(keys))

	// Output:
	// Safe concurrent access is implementation-dependent
}
