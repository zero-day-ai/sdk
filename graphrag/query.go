package graphrag

import (
	"errors"
	"fmt"
)

// MissionScope defines the scope for GraphRAG queries
type MissionScope int

const (
	// ScopeMissionRun queries only the current mission run (default)
	ScopeMissionRun MissionScope = iota

	// ScopeMission queries all runs of the current mission
	ScopeMission

	// ScopeGlobal queries all missions (no filtering)
	ScopeGlobal

	// Legacy aliases for backward compatibility
	ScopeCurrentRun  = ScopeMissionRun
	ScopeSameMission = ScopeMission
	ScopeAll         = ScopeGlobal
)

// DefaultMissionScope is the default scope for queries
const DefaultMissionScope = ScopeMissionRun

// String returns the string representation of the MissionScope
func (s MissionScope) String() string {
	switch s {
	case ScopeMissionRun:
		return "mission_run"
	case ScopeMission:
		return "mission"
	case ScopeGlobal:
		return "global"
	default:
		return fmt.Sprintf("MissionScope(%d)", s)
	}
}

// IsValid returns true if the scope is a valid value
func (s MissionScope) IsValid() bool {
	return s >= ScopeMissionRun && s <= ScopeGlobal
}

// ParseMissionScope parses a string into a MissionScope value
func ParseMissionScope(s string) (MissionScope, error) {
	switch s {
	case "mission_run", "current_run":
		return ScopeMissionRun, nil
	case "mission", "same_mission":
		return ScopeMission, nil
	case "global", "all":
		return ScopeGlobal, nil
	default:
		return 0, fmt.Errorf("invalid mission scope: %s", s)
	}
}

// Validate returns an error if the mission scope is invalid
func (s MissionScope) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid mission scope value: %d", s)
	}
	return nil
}

// AllMissionScopes returns all valid mission scope values
func AllMissionScopes() []MissionScope {
	return []MissionScope{ScopeMissionRun, ScopeMission, ScopeGlobal}
}

// Query represents a GraphRAG query with fluent builder pattern.
// It supports both natural language text queries (which will be embedded)
// and pre-computed embeddings, along with various filtering and scoring options.
type Query struct {
	// Text is the natural language query that will be embedded
	Text string `json:"text,omitempty"`

	// Embedding is a pre-computed embedding vector (alternative to Text)
	Embedding []float64 `json:"embedding,omitempty"`

	// TopK is the number of results to return
	TopK int `json:"top_k"`

	// MaxHops is the maximum graph traversal depth
	MaxHops int `json:"max_hops"`

	// MinScore is the minimum similarity threshold
	MinScore float64 `json:"min_score"`

	// NodeTypes filters results by node types
	NodeTypes []string `json:"node_types,omitempty"`

	// MissionID filters results by mission
	MissionID string `json:"mission_id,omitempty"`

	// VectorWeight is the weight for semantic similarity scoring
	VectorWeight float64 `json:"vector_weight"`

	// GraphWeight is the weight for graph structure scoring
	GraphWeight float64 `json:"graph_weight"`

	// Scope determines what data is visible (default: ScopeMissionRun)
	// This is the new mission-scoped storage scope system
	Scope MissionScope `json:"scope,omitempty"`

	// MissionRunID is set by harness (not agent) for mission-run scoped queries
	MissionRunID string `json:"-"`

	// Legacy fields (to be migrated in Phase 2)
	// MissionName is the mission name for legacy scope queries
	MissionName string `json:"mission_name,omitempty"`

	// RunNumber specifies a specific mission run number (legacy)
	RunNumber *int `json:"run_number,omitempty"`

	// IncludeRunMetadata indicates whether to include run provenance (legacy)
	IncludeRunMetadata bool `json:"include_run_metadata,omitempty"`
}

// NewQuery creates a new Query with the given text and sensible defaults.
// Default values:
//   - TopK: 10
//   - MaxHops: 3
//   - MinScore: 0.7
//   - VectorWeight: 0.6
//   - GraphWeight: 0.4
func NewQuery(text string) *Query {
	return &Query{
		Text:         text,
		TopK:         10,
		MaxHops:      3,
		MinScore:     0.7,
		VectorWeight: 0.6,
		GraphWeight:  0.4,
	}
}

// NewQueryFromEmbedding creates a new Query from a pre-computed embedding
// with sensible defaults.
// Default values:
//   - TopK: 10
//   - MaxHops: 3
//   - MinScore: 0.7
//   - VectorWeight: 0.6
//   - GraphWeight: 0.4
func NewQueryFromEmbedding(embedding []float64) *Query {
	return &Query{
		Embedding:    embedding,
		TopK:         10,
		MaxHops:      3,
		MinScore:     0.7,
		VectorWeight: 0.6,
		GraphWeight:  0.4,
	}
}

// NewStructuredQuery creates a new Query for pure structured (non-semantic) queries.
// This is used when you want to query the graph by node types, mission ID, or other
// structured filters WITHOUT semantic/vector search.
//
// Default values:
//   - TopK: 100 (return many results since we're not ranking by semantic similarity)
//   - MaxHops: 0 (no graph traversal by default)
//   - MinScore: 0.0 (no similarity threshold)
//   - VectorWeight: 0.0 (no semantic component)
//   - GraphWeight: 1.0 (pure graph structure)
//
// Use this when you want to retrieve nodes by type/attributes without semantic search.
// Example: Query all hosts and ports discovered in a mission without a text query.
func NewStructuredQuery() *Query {
	return &Query{
		TopK:         100,
		MaxHops:      0,
		MinScore:     0.0,
		VectorWeight: 0.0,
		GraphWeight:  1.0,
	}
}

// WithTopK sets the number of results to return.
// Returns the Query for method chaining.
func (q *Query) WithTopK(k int) *Query {
	q.TopK = k
	return q
}

// WithMaxHops sets the maximum graph traversal depth.
// Returns the Query for method chaining.
func (q *Query) WithMaxHops(hops int) *Query {
	q.MaxHops = hops
	return q
}

// WithMinScore sets the minimum similarity threshold.
// Returns the Query for method chaining.
func (q *Query) WithMinScore(score float64) *Query {
	q.MinScore = score
	return q
}

// WithNodeTypes sets the node types to filter by.
// Returns the Query for method chaining.
func (q *Query) WithNodeTypes(types ...string) *Query {
	q.NodeTypes = types
	return q
}

// WithMission sets the mission ID to filter by.
// Returns the Query for method chaining.
func (q *Query) WithMission(missionID string) *Query {
	q.MissionID = missionID
	return q
}

// WithWeights sets the vector and graph weights for scoring.
// Returns the Query for method chaining.
func (q *Query) WithWeights(vector, graph float64) *Query {
	q.VectorWeight = vector
	q.GraphWeight = graph
	return q
}

// WithScope sets the query scope for mission-scoped storage.
// Returns the Query for method chaining.
func (q *Query) WithScope(scope MissionScope) *Query {
	q.Scope = scope
	return q
}

// WithMissionRun queries a specific mission run by ID.
// Returns the Query for method chaining.
func (q *Query) WithMissionRun(runID string) *Query {
	q.MissionRunID = runID
	return q
}

// WithMissionName sets the mission name for legacy scope queries.
// Returns the Query for method chaining.
func (q *Query) WithMissionName(name string) *Query {
	q.MissionName = name
	return q
}

// WithRunNumber sets a specific run number to query.
// Returns the Query for method chaining.
func (q *Query) WithRunNumber(num int) *Query {
	q.RunNumber = &num
	return q
}

// WithIncludeRunMetadata sets whether to include run provenance information in results.
// Returns the Query for method chaining.
func (q *Query) WithIncludeRunMetadata(include bool) *Query {
	q.IncludeRunMetadata = include
	return q
}

// Validate ensures the Query is properly configured.
// Returns an error if:
//   - Both Text and Embedding are provided
//   - Neither Text nor Embedding is provided (UNLESS NodeTypes are specified for structured queries)
//   - Text is empty when provided
//   - Embedding is empty when provided
//   - TopK is less than or equal to 0
//   - MaxHops is less than 0
//   - MinScore is not between 0 and 1
//   - VectorWeight is negative
//   - GraphWeight is negative
//   - VectorWeight + GraphWeight does not equal 1.0
func (q *Query) Validate() error {
	// Check that exactly one of Text or Embedding is provided
	hasText := q.Text != ""
	hasEmbedding := len(q.Embedding) > 0

	if hasText && hasEmbedding {
		return errors.New("query must have either Text or Embedding, not both")
	}

	// Allow structured queries without Text/Embedding if NodeTypes are specified
	if !hasText && !hasEmbedding && len(q.NodeTypes) == 0 {
		return errors.New("query must have either Text, Embedding, or NodeTypes")
	}

	// Validate TopK
	if q.TopK <= 0 {
		return fmt.Errorf("TopK must be greater than 0, got %d", q.TopK)
	}

	// Validate MaxHops (0 is valid for structured queries without traversal)
	if q.MaxHops < 0 {
		return fmt.Errorf("MaxHops must be non-negative, got %d", q.MaxHops)
	}

	// Validate MinScore
	if q.MinScore < 0.0 || q.MinScore > 1.0 {
		return fmt.Errorf("MinScore must be between 0.0 and 1.0, got %f", q.MinScore)
	}

	// Validate VectorWeight
	if q.VectorWeight < 0.0 {
		return fmt.Errorf("VectorWeight must be non-negative, got %f", q.VectorWeight)
	}

	// Validate GraphWeight
	if q.GraphWeight < 0.0 {
		return fmt.Errorf("GraphWeight must be non-negative, got %f", q.GraphWeight)
	}

	// Validate that weights sum to 1.0 (with small epsilon for floating point)
	const epsilon = 0.0001
	weightSum := q.VectorWeight + q.GraphWeight
	if weightSum < 1.0-epsilon || weightSum > 1.0+epsilon {
		return fmt.Errorf("VectorWeight + GraphWeight must equal 1.0, got %f", weightSum)
	}

	// Note: Scope validation not implemented yet - will be added in Phase 2
	// when the int-based MissionScope type is fully integrated

	// Validate RunNumber if set (nil is valid - means all runs)
	if q.RunNumber != nil && *q.RunNumber < 1 {
		return fmt.Errorf("RunNumber must be greater than 0, got %d", *q.RunNumber)
	}

	return nil
}
