package graphrag

import (
	"testing"
)

// ============================================================================
// Query Tests
// ============================================================================

func TestNewQuery(t *testing.T) {
	text := "What are the recent attack attempts?"
	query := NewQuery(text)

	if query.Text != text {
		t.Errorf("expected Text to be %q, got %q", text, query.Text)
	}

	if query.Embedding != nil {
		t.Error("expected Embedding to be nil")
	}

	// Check defaults
	if query.TopK != 10 {
		t.Errorf("expected default TopK to be 10, got %d", query.TopK)
	}

	if query.MaxHops != 3 {
		t.Errorf("expected default MaxHops to be 3, got %d", query.MaxHops)
	}

	if query.MinScore != 0.7 {
		t.Errorf("expected default MinScore to be 0.7, got %f", query.MinScore)
	}

	if query.VectorWeight != 0.6 {
		t.Errorf("expected default VectorWeight to be 0.6, got %f", query.VectorWeight)
	}

	if query.GraphWeight != 0.4 {
		t.Errorf("expected default GraphWeight to be 0.4, got %f", query.GraphWeight)
	}
}

func TestNewQueryFromEmbedding(t *testing.T) {
	embedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	query := NewQueryFromEmbedding(embedding)

	if query.Text != "" {
		t.Errorf("expected Text to be empty, got %q", query.Text)
	}

	if len(query.Embedding) != len(embedding) {
		t.Errorf("expected Embedding length to be %d, got %d", len(embedding), len(query.Embedding))
	}

	for i, val := range embedding {
		if query.Embedding[i] != val {
			t.Errorf("expected Embedding[%d] to be %f, got %f", i, val, query.Embedding[i])
		}
	}

	// Check defaults
	if query.TopK != 10 {
		t.Errorf("expected default TopK to be 10, got %d", query.TopK)
	}

	if query.MaxHops != 3 {
		t.Errorf("expected default MaxHops to be 3, got %d", query.MaxHops)
	}

	if query.MinScore != 0.7 {
		t.Errorf("expected default MinScore to be 0.7, got %f", query.MinScore)
	}

	if query.VectorWeight != 0.6 {
		t.Errorf("expected default VectorWeight to be 0.6, got %f", query.VectorWeight)
	}

	if query.GraphWeight != 0.4 {
		t.Errorf("expected default GraphWeight to be 0.4, got %f", query.GraphWeight)
	}
}

func TestQuery_WithTopK(t *testing.T) {
	query := NewQuery("test").WithTopK(20)

	if query.TopK != 20 {
		t.Errorf("expected TopK to be 20, got %d", query.TopK)
	}
}

func TestQuery_WithMaxHops(t *testing.T) {
	query := NewQuery("test").WithMaxHops(5)

	if query.MaxHops != 5 {
		t.Errorf("expected MaxHops to be 5, got %d", query.MaxHops)
	}
}

func TestQuery_WithMinScore(t *testing.T) {
	query := NewQuery("test").WithMinScore(0.85)

	if query.MinScore != 0.85 {
		t.Errorf("expected MinScore to be 0.85, got %f", query.MinScore)
	}
}

func TestQuery_WithNodeTypes(t *testing.T) {
	nodeTypes := []string{"AttackAttempt", "Conversation", "Artifact"}
	query := NewQuery("test").WithNodeTypes(nodeTypes...)

	if len(query.NodeTypes) != len(nodeTypes) {
		t.Errorf("expected NodeTypes length to be %d, got %d", len(nodeTypes), len(query.NodeTypes))
	}

	for i, typ := range nodeTypes {
		if query.NodeTypes[i] != typ {
			t.Errorf("expected NodeTypes[%d] to be %q, got %q", i, typ, query.NodeTypes[i])
		}
	}
}

func TestQuery_WithMission(t *testing.T) {
	missionID := "mission-123"
	query := NewQuery("test").WithMission(missionID)

	if query.MissionID != missionID {
		t.Errorf("expected MissionID to be %q, got %q", missionID, query.MissionID)
	}
}

func TestQuery_WithWeights(t *testing.T) {
	query := NewQuery("test").WithWeights(0.8, 0.2)

	if query.VectorWeight != 0.8 {
		t.Errorf("expected VectorWeight to be 0.8, got %f", query.VectorWeight)
	}

	if query.GraphWeight != 0.2 {
		t.Errorf("expected GraphWeight to be 0.2, got %f", query.GraphWeight)
	}
}

func TestQuery_BuilderChaining(t *testing.T) {
	// Test that all builder methods can be chained together
	query := NewQuery("test query").
		WithTopK(50).
		WithMaxHops(4).
		WithMinScore(0.8).
		WithNodeTypes("AttackAttempt", "Conversation").
		WithMission("mission-abc").
		WithWeights(0.7, 0.3)

	if query.Text != "test query" {
		t.Errorf("expected Text to be 'test query', got %q", query.Text)
	}

	if query.TopK != 50 {
		t.Errorf("expected TopK to be 50, got %d", query.TopK)
	}

	if query.MaxHops != 4 {
		t.Errorf("expected MaxHops to be 4, got %d", query.MaxHops)
	}

	if query.MinScore != 0.8 {
		t.Errorf("expected MinScore to be 0.8, got %f", query.MinScore)
	}

	if len(query.NodeTypes) != 2 {
		t.Errorf("expected NodeTypes length to be 2, got %d", len(query.NodeTypes))
	}

	if query.MissionID != "mission-abc" {
		t.Errorf("expected MissionID to be 'mission-abc', got %q", query.MissionID)
	}

	if query.VectorWeight != 0.7 {
		t.Errorf("expected VectorWeight to be 0.7, got %f", query.VectorWeight)
	}

	if query.GraphWeight != 0.3 {
		t.Errorf("expected GraphWeight to be 0.3, got %f", query.GraphWeight)
	}
}

func TestQuery_Validate(t *testing.T) {
	tests := []struct {
		name    string
		query   *Query
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid query with text",
			query:   NewQuery("test query"),
			wantErr: false,
		},
		{
			name:    "valid query with embedding",
			query:   NewQueryFromEmbedding([]float64{0.1, 0.2, 0.3}),
			wantErr: false,
		},
		{
			name: "valid query with custom weights",
			query: NewQuery("test").WithWeights(0.5, 0.5),
			wantErr: false,
		},
		{
			name: "both text and embedding",
			query: &Query{
				Text:         "test",
				Embedding:    []float64{0.1, 0.2},
				TopK:         10,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "query must have either Text or Embedding, not both",
		},
		{
			name: "neither text nor embedding",
			query: &Query{
				TopK:         10,
				MaxHops:      3,
				MinScore:     0.7,
				VectorWeight: 0.6,
				GraphWeight:  0.4,
			},
			wantErr: true,
			errMsg:  "query must have either Text or Embedding",
		},
		{
			name:    "zero TopK",
			query:   NewQuery("test").WithTopK(0),
			wantErr: true,
			errMsg:  "TopK must be greater than 0",
		},
		{
			name:    "negative TopK",
			query:   NewQuery("test").WithTopK(-5),
			wantErr: true,
			errMsg:  "TopK must be greater than 0",
		},
		{
			name:    "zero MaxHops",
			query:   NewQuery("test").WithMaxHops(0),
			wantErr: true,
			errMsg:  "MaxHops must be greater than 0",
		},
		{
			name:    "negative MaxHops",
			query:   NewQuery("test").WithMaxHops(-2),
			wantErr: true,
			errMsg:  "MaxHops must be greater than 0",
		},
		{
			name:    "MinScore below 0",
			query:   NewQuery("test").WithMinScore(-0.1),
			wantErr: true,
			errMsg:  "MinScore must be between 0.0 and 1.0",
		},
		{
			name:    "MinScore above 1",
			query:   NewQuery("test").WithMinScore(1.5),
			wantErr: true,
			errMsg:  "MinScore must be between 0.0 and 1.0",
		},
		{
			name:    "negative VectorWeight",
			query:   NewQuery("test").WithWeights(-0.2, 1.2),
			wantErr: true,
			errMsg:  "VectorWeight must be non-negative",
		},
		{
			name:    "negative GraphWeight",
			query:   NewQuery("test").WithWeights(1.2, -0.2),
			wantErr: true,
			errMsg:  "GraphWeight must be non-negative",
		},
		{
			name:    "weights don't sum to 1.0",
			query:   NewQuery("test").WithWeights(0.5, 0.3),
			wantErr: true,
			errMsg:  "VectorWeight + GraphWeight must equal 1.0",
		},
		{
			name:    "weights sum too high",
			query:   NewQuery("test").WithWeights(0.7, 0.7),
			wantErr: true,
			errMsg:  "VectorWeight + GraphWeight must equal 1.0",
		},
		{
			name:    "MinScore at 0.0 boundary",
			query:   NewQuery("test").WithMinScore(0.0),
			wantErr: false,
		},
		{
			name:    "MinScore at 1.0 boundary",
			query:   NewQuery("test").WithMinScore(1.0),
			wantErr: false,
		},
		{
			name:    "weights with small floating point error",
			query:   NewQuery("test").WithWeights(0.3333333, 0.6666667),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.query.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

// ============================================================================
// Batch Tests
// ============================================================================

func TestNewBatch(t *testing.T) {
	batch := NewBatch()

	if batch == nil {
		t.Fatal("expected NewBatch to return non-nil")
	}

	if batch.Nodes == nil {
		t.Error("expected Nodes slice to be initialized")
	}

	if batch.Relationships == nil {
		t.Error("expected Relationships slice to be initialized")
	}

	if len(batch.Nodes) != 0 {
		t.Errorf("expected Nodes length to be 0, got %d", len(batch.Nodes))
	}

	if len(batch.Relationships) != 0 {
		t.Errorf("expected Relationships length to be 0, got %d", len(batch.Relationships))
	}
}

func TestBatch_AddNode(t *testing.T) {
	batch := NewBatch()
	node := *NewGraphNode("TestNode").WithID("node-1")

	result := batch.AddNode(node)

	// Check method chaining returns the batch
	if result != batch {
		t.Error("expected AddNode to return the same batch instance for chaining")
	}

	if len(batch.Nodes) != 1 {
		t.Errorf("expected Nodes length to be 1, got %d", len(batch.Nodes))
	}

	if batch.Nodes[0].ID != "node-1" {
		t.Errorf("expected Nodes[0].ID to be 'node-1', got %q", batch.Nodes[0].ID)
	}

	if batch.Nodes[0].Type != "TestNode" {
		t.Errorf("expected Nodes[0].Type to be 'TestNode', got %q", batch.Nodes[0].Type)
	}
}

func TestBatch_AddRelationship(t *testing.T) {
	batch := NewBatch()
	rel := *NewRelationship("node1", "node2", "CONNECTS_TO")

	result := batch.AddRelationship(rel)

	// Check method chaining returns the batch
	if result != batch {
		t.Error("expected AddRelationship to return the same batch instance for chaining")
	}

	if len(batch.Relationships) != 1 {
		t.Errorf("expected Relationships length to be 1, got %d", len(batch.Relationships))
	}

	if batch.Relationships[0].FromID != "node1" {
		t.Errorf("expected Relationships[0].FromID to be 'node1', got %q", batch.Relationships[0].FromID)
	}

	if batch.Relationships[0].ToID != "node2" {
		t.Errorf("expected Relationships[0].ToID to be 'node2', got %q", batch.Relationships[0].ToID)
	}

	if batch.Relationships[0].Type != "CONNECTS_TO" {
		t.Errorf("expected Relationships[0].Type to be 'CONNECTS_TO', got %q", batch.Relationships[0].Type)
	}
}

func TestBatch_Chaining(t *testing.T) {
	// Test that AddNode and AddRelationship can be chained together
	node1 := *NewGraphNode("Node1").WithID("n1")
	node2 := *NewGraphNode("Node2").WithID("n2")
	rel := *NewRelationship("n1", "n2", "RELATED_TO")

	batch := NewBatch().
		AddNode(node1).
		AddNode(node2).
		AddRelationship(rel)

	if len(batch.Nodes) != 2 {
		t.Errorf("expected Nodes length to be 2, got %d", len(batch.Nodes))
	}

	if len(batch.Relationships) != 1 {
		t.Errorf("expected Relationships length to be 1, got %d", len(batch.Relationships))
	}

	// Verify nodes
	if batch.Nodes[0].ID != "n1" {
		t.Errorf("expected Nodes[0].ID to be 'n1', got %q", batch.Nodes[0].ID)
	}

	if batch.Nodes[1].ID != "n2" {
		t.Errorf("expected Nodes[1].ID to be 'n2', got %q", batch.Nodes[1].ID)
	}

	// Verify relationship
	if batch.Relationships[0].FromID != "n1" || batch.Relationships[0].ToID != "n2" {
		t.Errorf("expected relationship from 'n1' to 'n2', got from '%s' to '%s'",
			batch.Relationships[0].FromID, batch.Relationships[0].ToID)
	}
}

func TestBatch_MultipleAdditions(t *testing.T) {
	batch := NewBatch()

	// Add multiple nodes
	for i := 0; i < 5; i++ {
		node := *NewGraphNode("Node").WithID("node-" + string(rune('0'+i)))
		batch.AddNode(node)
	}

	// Add multiple relationships
	for i := 0; i < 3; i++ {
		rel := *NewRelationship("node-"+string(rune('0'+i)), "node-"+string(rune('1'+i)), "NEXT")
		batch.AddRelationship(rel)
	}

	if len(batch.Nodes) != 5 {
		t.Errorf("expected 5 nodes, got %d", len(batch.Nodes))
	}

	if len(batch.Relationships) != 3 {
		t.Errorf("expected 3 relationships, got %d", len(batch.Relationships))
	}
}

// ============================================================================
// TraversalOptions Tests
// ============================================================================

func TestNewTraversalOptions(t *testing.T) {
	opts := NewTraversalOptions()

	if opts == nil {
		t.Fatal("expected NewTraversalOptions to return non-nil")
	}

	// Check defaults
	if opts.MaxDepth != 3 {
		t.Errorf("expected default MaxDepth to be 3, got %d", opts.MaxDepth)
	}

	if opts.Direction != "outgoing" {
		t.Errorf("expected default Direction to be 'outgoing', got %q", opts.Direction)
	}

	if opts.RelationshipTypes == nil {
		t.Error("expected RelationshipTypes to be initialized")
	}

	if len(opts.RelationshipTypes) != 0 {
		t.Errorf("expected RelationshipTypes length to be 0, got %d", len(opts.RelationshipTypes))
	}

	if opts.NodeTypes == nil {
		t.Error("expected NodeTypes to be initialized")
	}

	if len(opts.NodeTypes) != 0 {
		t.Errorf("expected NodeTypes length to be 0, got %d", len(opts.NodeTypes))
	}
}

func TestTraversalOptions_WithMaxDepth(t *testing.T) {
	opts := NewTraversalOptions().WithMaxDepth(5)

	if opts.MaxDepth != 5 {
		t.Errorf("expected MaxDepth to be 5, got %d", opts.MaxDepth)
	}
}

func TestTraversalOptions_WithRelationshipTypes(t *testing.T) {
	types := []string{"CONNECTS_TO", "PART_OF", "SIMILAR_TO"}
	opts := NewTraversalOptions().WithRelationshipTypes(types)

	if len(opts.RelationshipTypes) != len(types) {
		t.Errorf("expected RelationshipTypes length to be %d, got %d", len(types), len(opts.RelationshipTypes))
	}

	for i, typ := range types {
		if opts.RelationshipTypes[i] != typ {
			t.Errorf("expected RelationshipTypes[%d] to be %q, got %q", i, typ, opts.RelationshipTypes[i])
		}
	}
}

func TestTraversalOptions_WithNodeTypes(t *testing.T) {
	types := []string{"AttackAttempt", "Conversation"}
	opts := NewTraversalOptions().WithNodeTypes(types)

	if len(opts.NodeTypes) != len(types) {
		t.Errorf("expected NodeTypes length to be %d, got %d", len(types), len(opts.NodeTypes))
	}

	for i, typ := range types {
		if opts.NodeTypes[i] != typ {
			t.Errorf("expected NodeTypes[%d] to be %q, got %q", i, typ, opts.NodeTypes[i])
		}
	}
}

func TestTraversalOptions_WithDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction string
	}{
		{"outgoing direction", "outgoing"},
		{"incoming direction", "incoming"},
		{"both directions", "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := NewTraversalOptions().WithDirection(tt.direction)

			if opts.Direction != tt.direction {
				t.Errorf("expected Direction to be %q, got %q", tt.direction, opts.Direction)
			}
		})
	}
}

func TestTraversalOptions_BuilderChaining(t *testing.T) {
	// Test that all builder methods can be chained together
	relTypes := []string{"CONNECTS_TO", "PART_OF"}
	nodeTypes := []string{"Node1", "Node2"}

	opts := NewTraversalOptions().
		WithMaxDepth(7).
		WithRelationshipTypes(relTypes).
		WithNodeTypes(nodeTypes).
		WithDirection("both")

	if opts.MaxDepth != 7 {
		t.Errorf("expected MaxDepth to be 7, got %d", opts.MaxDepth)
	}

	if len(opts.RelationshipTypes) != 2 {
		t.Errorf("expected RelationshipTypes length to be 2, got %d", len(opts.RelationshipTypes))
	}

	if len(opts.NodeTypes) != 2 {
		t.Errorf("expected NodeTypes length to be 2, got %d", len(opts.NodeTypes))
	}

	if opts.Direction != "both" {
		t.Errorf("expected Direction to be 'both', got %q", opts.Direction)
	}
}

func TestTraversalOptions_DirectionValues(t *testing.T) {
	// Test all valid direction values as documented
	validDirections := []string{"outgoing", "incoming", "both"}

	for _, direction := range validDirections {
		t.Run("direction_"+direction, func(t *testing.T) {
			opts := NewTraversalOptions().WithDirection(direction)

			if opts.Direction != direction {
				t.Errorf("expected Direction to be %q, got %q", direction, opts.Direction)
			}
		})
	}
}

// ============================================================================
// Result Tests
// ============================================================================

func TestResult_Structure(t *testing.T) {
	// Test that Result can be created with all fields
	node := NewGraphNode("TestType").WithID("node-1")
	result := Result{
		Node:        *node,
		Score:       0.95,
		VectorScore: 0.90,
		GraphScore:  0.85,
		Path:        []string{"origin", "intermediate", "node-1"},
		Distance:    2,
	}

	if result.Score != 0.95 {
		t.Errorf("expected Score to be 0.95, got %f", result.Score)
	}

	if result.VectorScore != 0.90 {
		t.Errorf("expected VectorScore to be 0.90, got %f", result.VectorScore)
	}

	if result.GraphScore != 0.85 {
		t.Errorf("expected GraphScore to be 0.85, got %f", result.GraphScore)
	}

	if result.Distance != 2 {
		t.Errorf("expected Distance to be 2, got %d", result.Distance)
	}

	if len(result.Path) != 3 {
		t.Errorf("expected Path length to be 3, got %d", len(result.Path))
	}

	if result.Node.ID != "node-1" {
		t.Errorf("expected Node.ID to be 'node-1', got %q", result.Node.ID)
	}
}

func TestResult_EmptyPath(t *testing.T) {
	// Test that Result works without a path
	node := NewGraphNode("TestType")
	result := Result{
		Node:        *node,
		Score:       0.80,
		VectorScore: 0.75,
		GraphScore:  0.70,
		Distance:    0,
	}

	if result.Path != nil {
		t.Errorf("expected Path to be nil, got %v", result.Path)
	}

	if result.Distance != 0 {
		t.Errorf("expected Distance to be 0, got %d", result.Distance)
	}
}

// ============================================================================
// TraversalResult Tests
// ============================================================================

func TestTraversalResult_Structure(t *testing.T) {
	// Test that TraversalResult can be created with all fields
	node := NewGraphNode("TestType").WithID("node-2")
	result := TraversalResult{
		Node:     *node,
		Path:     []string{"start", "node-2"},
		Distance: 1,
	}

	if result.Distance != 1 {
		t.Errorf("expected Distance to be 1, got %d", result.Distance)
	}

	if len(result.Path) != 2 {
		t.Errorf("expected Path length to be 2, got %d", len(result.Path))
	}

	if result.Path[0] != "start" {
		t.Errorf("expected Path[0] to be 'start', got %q", result.Path[0])
	}

	if result.Path[1] != "node-2" {
		t.Errorf("expected Path[1] to be 'node-2', got %q", result.Path[1])
	}

	if result.Node.ID != "node-2" {
		t.Errorf("expected Node.ID to be 'node-2', got %q", result.Node.ID)
	}
}

func TestTraversalResult_EmptyPath(t *testing.T) {
	// Test that TraversalResult works without a path
	node := NewGraphNode("TestType")
	result := TraversalResult{
		Node:     *node,
		Distance: 0,
	}

	if result.Path != nil {
		t.Errorf("expected Path to be nil, got %v", result.Path)
	}

	if result.Distance != 0 {
		t.Errorf("expected Distance to be 0, got %d", result.Distance)
	}
}

func TestTraversalResult_MultiHopPath(t *testing.T) {
	// Test a longer path scenario
	node := NewGraphNode("Destination")
	path := []string{"origin", "hop1", "hop2", "hop3", "destination"}
	result := TraversalResult{
		Node:     *node,
		Path:     path,
		Distance: 4,
	}

	if len(result.Path) != 5 {
		t.Errorf("expected Path length to be 5, got %d", len(result.Path))
	}

	if result.Distance != 4 {
		t.Errorf("expected Distance to be 4 (one less than path length), got %d", result.Distance)
	}

	// Verify path order
	expectedPath := []string{"origin", "hop1", "hop2", "hop3", "destination"}
	for i, expected := range expectedPath {
		if result.Path[i] != expected {
			t.Errorf("expected Path[%d] to be %q, got %q", i, expected, result.Path[i])
		}
	}
}
