package graphrag

import "time"

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

	// RunMetadata contains run provenance information if requested via IncludeRunMetadata.
	// This field will be nil if the query did not request run metadata or if the node
	// has no mission context.
	RunMetadata *RunMetadata `json:"run_metadata,omitempty"`
}

// RunMetadata contains run provenance information for a graph node result.
// This provides visibility into when and where a node was discovered.
type RunMetadata struct {
	// MissionName is the name of the mission that discovered this node
	MissionName string `json:"mission_name"`

	// RunNumber is the sequential run number of the mission
	RunNumber int `json:"run_number"`

	// DiscoveredAt is the timestamp when the node was first created
	DiscoveredAt time.Time `json:"discovered_at"`
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
