package memory

import (
	"errors"
	"fmt"
	"time"
)

// MemoryContinuityMode defines how memory state is handled across multiple
// agent execution runs within a mission.
//
// Memory continuity controls whether runs can access or share memory state
// from previous runs, enabling different collaboration and isolation patterns:
//
//   - MemoryIsolated: Each run has completely isolated memory (default)
//   - MemoryInherit: New runs can read prior run's memory in a copy-on-write manner
//   - MemoryShared: All runs share the same memory namespace with full read/write access
//
// The default mode is MemoryIsolated to ensure backwards compatibility and
// prevent unintended state sharing between runs.
//
// Example usage:
//
//	mode := MemoryInherit
//	if err := mode.Validate(); err != nil {
//	    return err
//	}
//
//	// Use mode in mission configuration
//	cfg := mission.NewConfig().
//	    SetMemoryContinuity(mode).
//	    Build()
type MemoryContinuityMode string

const (
	// MemoryIsolated indicates that each run has completely isolated memory.
	// This is the default mode and provides the strongest isolation guarantees.
	//
	// Use cases:
	//   - Independent parallel runs that should not interfere with each other
	//   - Testing scenarios where clean state is required
	//   - Security-sensitive operations requiring strict isolation
	//
	// Behavior:
	//   - Each run starts with empty memory state
	//   - Changes made during a run are not visible to other runs
	//   - Previous run memory is not accessible
	MemoryIsolated MemoryContinuityMode = "isolated"

	// MemoryInherit indicates that new runs can read memory from prior runs
	// but cannot modify the shared state (copy-on-write semantics).
	//
	// Use cases:
	//   - Sequential runs building on previous results
	//   - Progressive refinement workflows
	//   - Learning from historical run data
	//
	// Behavior:
	//   - New runs start with a read-only view of the previous run's memory
	//   - Writes create copies in the current run's namespace
	//   - Historical queries can access prior run values
	//   - Previous runs' memory remains immutable
	MemoryInherit MemoryContinuityMode = "inherit"

	// MemoryShared indicates that all runs share the same memory namespace
	// with full read and write access.
	//
	// Use cases:
	//   - Collaborative multi-agent scenarios
	//   - Shared state accumulation across runs
	//   - Real-time coordination between concurrent runs
	//
	// Behavior:
	//   - All runs read and write to the same memory namespace
	//   - Changes made by one run are immediately visible to all other runs
	//   - No isolation between runs
	//   - Concurrent access requires coordination to avoid race conditions
	MemoryShared MemoryContinuityMode = "shared"
)

// DefaultMemoryContinuity is the default memory continuity mode.
// It is set to MemoryIsolated to ensure backwards compatibility and
// prevent unintended state sharing between runs.
const DefaultMemoryContinuity = MemoryIsolated

// Common errors returned by continuity operations.
var (
	// ErrInvalidMode is returned when a MemoryContinuityMode value is not recognized.
	ErrInvalidMode = errors.New("memory: invalid continuity mode")
)

// String returns the string representation of the MemoryContinuityMode.
// This implements the fmt.Stringer interface.
func (m MemoryContinuityMode) String() string {
	return string(m)
}

// IsValid returns true if the MemoryContinuityMode is one of the defined constants.
// This method is useful for validation in configuration parsing and API handlers.
//
// Example:
//
//	mode := MemoryContinuityMode("isolated")
//	if !mode.IsValid() {
//	    return fmt.Errorf("invalid mode: %s", mode)
//	}
func (m MemoryContinuityMode) IsValid() bool {
	switch m {
	case MemoryIsolated, MemoryInherit, MemoryShared:
		return true
	default:
		return false
	}
}

// Validate returns an error if the MemoryContinuityMode is not valid.
// This is a convenience method that wraps IsValid() with a descriptive error.
//
// Example:
//
//	mode := MemoryContinuityMode(userInput)
//	if err := mode.Validate(); err != nil {
//	    return fmt.Errorf("invalid continuity mode: %w", err)
//	}
func (m MemoryContinuityMode) Validate() error {
	if !m.IsValid() {
		return fmt.Errorf("%w: %q (must be one of: isolated, inherit, shared)", ErrInvalidMode, m)
	}
	return nil
}

// HistoricalValue represents a value retrieved from a previous run's memory.
// This is used when querying memory history in MemoryInherit mode to access
// values stored by prior runs.
//
// Historical values include metadata about when and where they were stored,
// enabling agents to understand the provenance of inherited memory.
//
// Example:
//
//	// Query historical values for a key
//	history, err := mission.GetHistory(ctx, "user_preference")
//	for _, hv := range history {
//	    fmt.Printf("Run %d: %v (stored at %v)\n", hv.RunNumber, hv.Value, hv.StoredAt)
//	}
type HistoricalValue struct {
	// Value is the actual data stored in memory.
	// The type depends on what was originally stored and may require
	// type assertion to use.
	Value any `json:"value"`

	// RunNumber is the sequential run number within the mission where
	// this value was stored (1-based indexing).
	// This helps track the temporal sequence of memory changes.
	RunNumber int `json:"run_number"`

	// MissionID is the unique identifier of the mission this value belongs to.
	// This ensures values are not confused across different missions.
	MissionID string `json:"mission_id"`

	// StoredAt is the timestamp when this value was stored.
	// This provides precise temporal information beyond run numbers.
	StoredAt time.Time `json:"stored_at"`
}
