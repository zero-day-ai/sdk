package graphrag

// Batch represents a collection of nodes and relationships to be created or updated together.
// It supports builder pattern methods for easy construction.
type Batch struct {
	// Nodes contains all nodes to be processed in this batch
	Nodes []GraphNode `json:"nodes"`

	// Relationships contains all relationships to be processed in this batch
	Relationships []Relationship `json:"relationships"`
}

// NewBatch creates a new empty Batch with initialized slices.
func NewBatch() *Batch {
	return &Batch{
		Nodes:         make([]GraphNode, 0),
		Relationships: make([]Relationship, 0),
	}
}

// AddNode adds a node to the batch and returns the batch for method chaining.
func (b *Batch) AddNode(n GraphNode) *Batch {
	b.Nodes = append(b.Nodes, n)
	return b
}

// AddRelationship adds a relationship to the batch and returns the batch for method chaining.
func (b *Batch) AddRelationship(r Relationship) *Batch {
	b.Relationships = append(b.Relationships, r)
	return b
}

// TraversalOptions defines parameters for graph traversal operations.
// It controls how the graph is traversed, including depth, filtering, and direction.
type TraversalOptions struct {
	// MaxDepth specifies the maximum number of hops to traverse from the starting node.
	// Default is 3.
	MaxDepth int `json:"max_depth"`

	// RelationshipTypes filters which relationship types to follow during traversal.
	// If empty, all relationship types are followed.
	RelationshipTypes []string `json:"relationship_types,omitempty"`

	// NodeTypes filters which node types to include in traversal results.
	// If empty, all node types are included.
	NodeTypes []string `json:"node_types,omitempty"`

	// Direction specifies the traversal direction:
	// - "outgoing": follow relationships from source to target (default)
	// - "incoming": follow relationships from target to source
	// - "both": follow relationships in both directions
	Direction string `json:"direction"`
}

// NewTraversalOptions creates a new TraversalOptions with sensible defaults.
// Default values: MaxDepth=3, Direction="outgoing"
func NewTraversalOptions() *TraversalOptions {
	return &TraversalOptions{
		MaxDepth:          3,
		Direction:         "outgoing",
		RelationshipTypes: make([]string, 0),
		NodeTypes:         make([]string, 0),
	}
}

// WithMaxDepth sets the maximum traversal depth and returns the options for chaining.
func (t *TraversalOptions) WithMaxDepth(depth int) *TraversalOptions {
	t.MaxDepth = depth
	return t
}

// WithRelationshipTypes sets the relationship type filter and returns the options for chaining.
func (t *TraversalOptions) WithRelationshipTypes(types []string) *TraversalOptions {
	t.RelationshipTypes = types
	return t
}

// WithNodeTypes sets the node type filter and returns the options for chaining.
func (t *TraversalOptions) WithNodeTypes(types []string) *TraversalOptions {
	t.NodeTypes = types
	return t
}

// WithDirection sets the traversal direction and returns the options for chaining.
// Valid values: "outgoing", "incoming", "both"
func (t *TraversalOptions) WithDirection(direction string) *TraversalOptions {
	t.Direction = direction
	return t
}
