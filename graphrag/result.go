package graphrag

// Result represents a single result from a GraphRAG query.
// It contains the matched node along with scoring information and path details.
type Result struct {
	// Node is the matched graph node
	Node GraphNode `json:"node"`

	// Score is the combined similarity score (0.0 to 1.0)
	Score float64 `json:"score"`

	// VectorScore is the semantic similarity score from embeddings (0.0 to 1.0)
	VectorScore float64 `json:"vector_score"`

	// GraphScore is the graph structure score based on connectivity (0.0 to 1.0)
	GraphScore float64 `json:"graph_score"`

	// Path is the sequence of node IDs from the query origin to this result
	Path []string `json:"path,omitempty"`

	// Distance is the number of hops from the query origin to this node
	Distance int `json:"distance"`
}

// TraversalResult represents a single result from a graph traversal operation.
// It contains the visited node and information about the path taken to reach it.
type TraversalResult struct {
	// Node is the visited graph node
	Node GraphNode `json:"node"`

	// Path is the sequence of node IDs from the traversal origin to this result
	Path []string `json:"path,omitempty"`

	// Distance is the number of hops from the traversal origin to this node
	Distance int `json:"distance"`
}
