package memory

import (
	"context"
	"errors"
)

// Common errors returned by memory operations.
var (
	// ErrNotFound is returned when a requested item does not exist in the store.
	ErrNotFound = errors.New("memory: item not found")

	// ErrInvalidKey is returned when a key is empty or otherwise invalid.
	ErrInvalidKey = errors.New("memory: invalid key")

	// ErrInvalidValue is returned when a value cannot be stored (e.g., not serializable).
	ErrInvalidValue = errors.New("memory: invalid value")

	// ErrStorageFailed is returned when the underlying storage backend fails.
	ErrStorageFailed = errors.New("memory: storage operation failed")

	// ErrNotImplemented is returned when an operation is not supported by the implementation.
	ErrNotImplemented = errors.New("memory: operation not implemented")
)

// Store provides access to the three-tier memory system.
// Each tier has different characteristics and use cases:
//
//   - Working: Ephemeral, in-memory storage for temporary data
//   - Mission: Persistent storage scoped to a specific mission
//   - LongTerm: Vector-based semantic storage across missions
//
// Store implementations are responsible for managing the lifecycle
// and persistence strategy of each memory tier.
//
// Example usage:
//
//	store := // ... obtain Store implementation
//
//	// Use working memory for temporary state
//	working := store.Working()
//	working.Set(ctx, "temp_data", someValue)
//
//	// Use mission memory for mission-scoped persistence
//	mission := store.Mission()
//	mission.Set(ctx, "user_pref", value, metadata)
//
//	// Use long-term memory for semantic knowledge
//	longTerm := store.LongTerm()
//	longTerm.Store(ctx, "Important fact to remember", metadata)
type Store interface {
	// Working returns the working memory (ephemeral, in-memory).
	// Working memory is cleared between agent executions and provides
	// fast key-value access for temporary data.
	Working() WorkingMemory

	// Mission returns the mission memory (persistent per-mission).
	// Mission memory persists data across agent executions within
	// the same mission and supports search and history tracking.
	Mission() MissionMemory

	// LongTerm returns the long-term memory (vector-based).
	// Long-term memory persists across missions and uses vector
	// embeddings for semantic search and retrieval.
	LongTerm() LongTermMemory
}

// WorkingMemory provides ephemeral key-value storage for temporary data
// that only needs to persist during a single execution context.
//
// Working memory is typically:
//   - Fast: All operations are in-memory
//   - Ephemeral: Data is not persisted to disk
//   - Simple: Basic key-value operations
//
// Use working memory for:
//   - Intermediate calculation results
//   - Temporary state during multi-step operations
//   - Cache data that can be regenerated
//   - Scratch space for complex algorithms
//
// Example:
//
//	working := store.Working()
//
//	// Store temporary state
//	err := working.Set(ctx, "step", 1)
//	err = working.Set(ctx, "partial_result", []int{1, 2, 3})
//
//	// Retrieve when needed
//	step, err := working.Get(ctx, "step")
//	if err != nil {
//	    return err
//	}
//
//	// Clear when done
//	err = working.Clear(ctx)
type WorkingMemory interface {
	// Get retrieves a value by key.
	// Returns ErrNotFound if the key does not exist.
	Get(ctx context.Context, key string) (any, error)

	// Set stores a value with the given key.
	// If the key already exists, the value is replaced.
	// Returns ErrInvalidKey if the key is empty.
	Set(ctx context.Context, key string, value any) error

	// Delete removes a value by key.
	// Returns ErrNotFound if the key does not exist.
	Delete(ctx context.Context, key string) error

	// Clear removes all values from working memory.
	Clear(ctx context.Context) error

	// Keys returns all keys currently in working memory.
	// The returned slice may be empty if no keys exist.
	Keys(ctx context.Context) ([]string, error)
}

// MissionMemory provides persistent mission-scoped storage with metadata,
// search capabilities, and history tracking.
//
// Mission memory is typically:
//   - Persistent: Data is written to disk
//   - Structured: Items include metadata and timestamps
//   - Searchable: Full-text search across stored values
//   - Mission-scoped: Data persists for the duration of a mission
//
// Use mission memory for:
//   - User preferences and settings
//   - Conversation context and history
//   - Task state and progress
//   - Facts and observations from the current mission
//
// Example:
//
//	mission := store.Mission()
//
//	// Store with metadata
//	err := mission.Set(ctx, "user_pref", "dark_mode", map[string]any{
//	    "category": "ui",
//	    "updated_by": "user",
//	})
//
//	// Retrieve with full item details
//	item, err := mission.Get(ctx, "user_pref")
//	fmt.Printf("Value: %v, Created: %v\n", item.Value, item.CreatedAt)
//
//	// Search for relevant items
//	results, err := mission.Search(ctx, "user interface preferences", 10)
//
//	// Get recent history
//	history, err := mission.History(ctx, 20)
type MissionMemory interface {
	// Get retrieves an item by key with full metadata.
	// Returns ErrNotFound if the key does not exist.
	Get(ctx context.Context, key string) (*Item, error)

	// Set stores a value with the given key and metadata.
	// If the key already exists, the value and metadata are replaced
	// and UpdatedAt is set to the current time.
	// If the key is new, both CreatedAt and UpdatedAt are set.
	// Returns ErrInvalidKey if the key is empty.
	Set(ctx context.Context, key string, value any, metadata map[string]any) error

	// Delete removes an item by key.
	// Returns ErrNotFound if the key does not exist.
	Delete(ctx context.Context, key string) error

	// Search performs a full-text search across stored items,
	// returning up to 'limit' results ordered by relevance.
	// The query string is matched against both keys and values.
	// Returns an empty slice if no matches are found.
	Search(ctx context.Context, query string, limit int) ([]Result, error)

	// History returns the most recently updated items, up to 'limit'.
	// Items are ordered by UpdatedAt in descending order (most recent first).
	// Returns an empty slice if no items exist.
	History(ctx context.Context, limit int) ([]Item, error)
}

// LongTermMemory provides vector-based semantic storage that persists
// across missions, enabling agents to build up knowledge over time.
//
// Long-term memory is typically:
//   - Semantic: Uses vector embeddings for meaning-based search
//   - Persistent: Data persists across all missions
//   - Filterable: Metadata-based filtering of search results
//   - Scalable: Designed to handle large amounts of stored knowledge
//
// Use long-term memory for:
//   - Learned facts and knowledge
//   - Best practices and strategies
//   - Historical patterns and insights
//   - Reusable information across missions
//
// Example:
//
//	longTerm := store.LongTerm()
//
//	// Store knowledge with metadata
//	id, err := longTerm.Store(ctx,
//	    "Python's GIL prevents true parallelism for CPU-bound tasks",
//	    map[string]any{
//	        "category": "programming",
//	        "language": "python",
//	        "topic": "concurrency",
//	    },
//	)
//
//	// Search semantically
//	results, err := longTerm.Search(ctx,
//	    "How does Python handle parallel processing?",
//	    5,
//	    nil,
//	)
//
//	// Search with filters
//	results, err = longTerm.Search(ctx,
//	    "concurrency patterns",
//	    10,
//	    map[string]any{"language": "python"},
//	)
type LongTermMemory interface {
	// Store saves content with metadata and returns a unique identifier.
	// The content is converted to a vector embedding for semantic search.
	// Returns the unique ID of the stored item.
	Store(ctx context.Context, content string, metadata map[string]any) (string, error)

	// Search performs a semantic search using vector similarity.
	// Returns up to 'topK' results ordered by similarity score.
	// Optional filters restrict results to items with matching metadata.
	// Returns an empty slice if no matches are found.
	//
	// The query string is converted to a vector embedding and compared
	// against stored embeddings using cosine similarity or similar metrics.
	//
	// Filters are applied as exact matches on metadata fields:
	//   filters["category"] = "programming"  // Only items with category="programming"
	//   filters["language"] = "go"           // Only items with language="go"
	Search(ctx context.Context, query string, topK int, filters map[string]any) ([]Result, error)

	// Delete removes an item by its unique ID.
	// Returns ErrNotFound if the ID does not exist.
	Delete(ctx context.Context, id string) error
}
