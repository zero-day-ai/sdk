package graphrag

import (
	"errors"
	"fmt"
)

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

// Validate ensures the Query is properly configured.
// Returns an error if:
//   - Both Text and Embedding are provided
//   - Neither Text nor Embedding is provided
//   - Text is empty when provided
//   - Embedding is empty when provided
//   - TopK is less than or equal to 0
//   - MaxHops is less than or equal to 0
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

	if !hasText && !hasEmbedding {
		return errors.New("query must have either Text or Embedding")
	}

	// Validate TopK
	if q.TopK <= 0 {
		return fmt.Errorf("TopK must be greater than 0, got %d", q.TopK)
	}

	// Validate MaxHops
	if q.MaxHops <= 0 {
		return fmt.Errorf("MaxHops must be greater than 0, got %d", q.MaxHops)
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

	return nil
}
