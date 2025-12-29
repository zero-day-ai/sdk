package memory

import (
	"encoding/json"
	"time"
)

// Item represents a memory item stored in mission or long-term memory.
// Items contain a value, optional metadata, and timestamps tracking creation
// and modification.
type Item struct {
	// Key is the unique identifier for this item within its memory store.
	Key string `json:"key"`

	// Value is the stored data. This can be any JSON-serializable value.
	Value any `json:"value"`

	// Metadata provides additional context about the item.
	// Common metadata fields include:
	//   - category: Categorization tag
	//   - source: Origin of the information
	//   - priority: Importance ranking
	//   - tags: List of associated tags
	Metadata map[string]any `json:"metadata,omitempty"`

	// CreatedAt is the timestamp when this item was first created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is the timestamp when this item was last modified.
	UpdatedAt time.Time `json:"updated_at"`
}

// Result represents a memory search result with a relevance score.
// Results are returned from search operations on mission and long-term memory,
// ordered by descending relevance score.
type Result struct {
	Item

	// Score represents the relevance of this result to the search query.
	// Higher scores indicate better matches.
	// - For mission memory full-text search: typically 0.0-1.0
	// - For long-term memory vector search: cosine similarity, typically 0.0-1.0
	Score float64 `json:"score"`
}

// String returns a human-readable representation of the Item.
func (i *Item) String() string {
	data, _ := json.MarshalIndent(i, "", "  ")
	return string(data)
}

// String returns a human-readable representation of the Result.
func (r *Result) String() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// GetMetadata retrieves a metadata value by key, returning the value and
// whether it was found. This is a type-safe helper for accessing metadata.
//
// Example:
//
//	if category, ok := item.GetMetadata("category"); ok {
//	    fmt.Printf("Category: %v\n", category)
//	}
func (i *Item) GetMetadata(key string) (any, bool) {
	if i.Metadata == nil {
		return nil, false
	}
	val, ok := i.Metadata[key]
	return val, ok
}

// SetMetadata sets a metadata value for the given key.
// If the metadata map is nil, it will be initialized.
//
// Example:
//
//	item.SetMetadata("priority", 1)
//	item.SetMetadata("tags", []string{"important", "user-facing"})
func (i *Item) SetMetadata(key string, value any) {
	if i.Metadata == nil {
		i.Metadata = make(map[string]any)
	}
	i.Metadata[key] = value
}

// HasMetadata checks if a metadata key exists.
//
// Example:
//
//	if item.HasMetadata("source") {
//	    fmt.Println("Item has source metadata")
//	}
func (i *Item) HasMetadata(key string) bool {
	if i.Metadata == nil {
		return false
	}
	_, ok := i.Metadata[key]
	return ok
}

// Clone creates a deep copy of the Item.
// This is useful when you need to modify an item without affecting the original.
//
// Example:
//
//	original := &Item{Key: "test", Value: "data"}
//	modified := original.Clone()
//	modified.Value = "new data"
//	// original.Value is still "data"
func (i *Item) Clone() *Item {
	clone := &Item{
		Key:       i.Key,
		Value:     cloneValue(i.Value),
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}

	if i.Metadata != nil {
		clone.Metadata = make(map[string]any, len(i.Metadata))
		for k, v := range i.Metadata {
			clone.Metadata[k] = cloneValue(v)
		}
	}

	return clone
}

// cloneValue creates a deep copy of a value using JSON marshaling.
// This works for any JSON-serializable value.
func cloneValue(v any) any {
	if v == nil {
		return nil
	}

	// Use JSON marshaling for deep copy
	data, err := json.Marshal(v)
	if err != nil {
		return v // Return original if can't marshal
	}

	var clone any
	if err := json.Unmarshal(data, &clone); err != nil {
		return v // Return original if can't unmarshal
	}

	return clone
}

// Age returns the duration since the item was created.
//
// Example:
//
//	if item.Age() > 24*time.Hour {
//	    fmt.Println("Item is more than a day old")
//	}
func (i *Item) Age() time.Duration {
	return time.Since(i.CreatedAt)
}

// TimeSinceUpdate returns the duration since the item was last updated.
//
// Example:
//
//	if item.TimeSinceUpdate() < time.Minute {
//	    fmt.Println("Recently updated")
//	}
func (i *Item) TimeSinceUpdate() time.Duration {
	return time.Since(i.UpdatedAt)
}

// IsModified returns true if the item has been updated since creation.
//
// Example:
//
//	if item.IsModified() {
//	    fmt.Printf("Item was modified %v after creation\n",
//	        item.UpdatedAt.Sub(item.CreatedAt))
//	}
func (i *Item) IsModified() bool {
	return !i.UpdatedAt.Equal(i.CreatedAt)
}
