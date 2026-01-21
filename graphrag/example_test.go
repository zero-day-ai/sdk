package graphrag_test

import (
	"fmt"

	"github.com/zero-day-ai/sdk/graphrag"
)

// Example demonstrates basic GraphRAG node creation and querying.
func Example() {
	// Create a finding node
	finding := graphrag.NewGraphNode("finding").
		WithID("finding-001").
		WithContent("SQL injection vulnerability in login endpoint").
		WithProperty("severity", "critical").
		WithProperty("cvss_score", 9.8)

	// Validate the node
	if err := finding.Validate(); err != nil {
		fmt.Printf("failed to validate node: %v\n", err)
		return
	}

	// Create a query
	query := graphrag.NewQuery("SQL injection vulnerabilities").
		WithTopK(10).
		WithNodeTypes("finding").
		WithMinScore(0.7)

	// Validate the query
	if err := query.Validate(); err != nil {
		fmt.Printf("failed to validate query: %v\n", err)
		return
	}

	fmt.Println("Node created:", finding.Type)
	fmt.Println("Query configured for top", query.TopK, "results")
	// Output:
	// Node created: finding
	// Query configured for top 10 results
}

// ExampleNewGraphNode demonstrates creating a GraphNode with properties.
func ExampleNewGraphNode() {
	node := graphrag.NewGraphNode("finding").
		WithID("finding-123").
		WithContent("Cross-Site Scripting (XSS) in search parameter").
		WithProperty("severity", "high").
		WithProperty("confidence", 0.92).
		WithProperty("endpoint", "/api/search")

	fmt.Println("Type:", node.Type)
	fmt.Println("ID:", node.ID)
	fmt.Println("Severity:", node.Properties["severity"])
	// Output:
	// Type: finding
	// ID: finding-123
	// Severity: high
}

// ExampleNewQuery demonstrates creating a GraphRAG query.
func ExampleNewQuery() {
	query := graphrag.NewQuery("authentication bypass vulnerabilities").
		WithTopK(5).
		WithMaxHops(2).
		WithMinScore(0.75).
		WithNodeTypes("finding", "technique")

	fmt.Println("Query text:", query.Text)
	fmt.Println("Top K:", query.TopK)
	fmt.Println("Max hops:", query.MaxHops)
	fmt.Println("Min score:", query.MinScore)
	// Output:
	// Query text: authentication bypass vulnerabilities
	// Top K: 5
	// Max hops: 2
	// Min score: 0.75
}

// ExampleNewQueryFromEmbedding demonstrates creating a query from pre-computed embeddings.
func ExampleNewQueryFromEmbedding() {
	// Pre-computed embedding vector
	embedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}

	query := graphrag.NewQueryFromEmbedding(embedding).
		WithWeights(0.7, 0.3). // 70% vector, 30% graph
		WithTopK(10)

	fmt.Println("Embedding dimension:", len(query.Embedding))
	fmt.Println("Vector weight:", query.VectorWeight)
	fmt.Println("Graph weight:", query.GraphWeight)
	// Output:
	// Embedding dimension: 5
	// Vector weight: 0.7
	// Graph weight: 0.3
}

// ExampleNewRelationship demonstrates creating relationships between nodes.
func ExampleNewRelationship() {
	// Create a relationship linking a finding to a MITRE ATT&CK technique
	rel := graphrag.NewRelationship(
		"finding-123",
		"T1190",
		"USES_TECHNIQUE",
	).WithProperty("confidence", 0.95)

	fmt.Println("From:", rel.FromID)
	fmt.Println("To:", rel.ToID)
	fmt.Println("Type:", rel.Type)
	fmt.Println("Confidence:", rel.Properties["confidence"])
	// Output:
	// From: finding-123
	// To: T1190
	// Type: USES_TECHNIQUE
	// Confidence: 0.95
}

// ExampleRelationship_WithBidirectional demonstrates creating bidirectional relationships.
func ExampleRelationship_WithBidirectional() {
	// Create a bidirectional similarity relationship
	rel := graphrag.NewRelationship(
		"finding-123",
		"finding-456",
		"SIMILAR_TO",
	).WithProperty("similarity", 0.87).
		WithBidirectional(true)

	fmt.Println("Bidirectional:", rel.Bidirectional)
	fmt.Println("Type:", rel.Type)
	// Output:
	// Bidirectional: true
	// Type: SIMILAR_TO
}

// ExampleNewBatch demonstrates batch operations for efficient bulk storage.
func ExampleNewBatch() {
	// Create nodes
	finding1 := *graphrag.NewGraphNode("finding").
		WithID("finding-1").
		WithContent("SQL injection in login")

	finding2 := *graphrag.NewGraphNode("finding").
		WithID("finding-2").
		WithContent("XSS in search")

	technique := *graphrag.NewGraphNode("technique").
		WithID("T1190").
		WithContent("Exploit Public-Facing Application")

	// Create relationships
	rel1 := *graphrag.NewRelationship("finding-1", "T1190", "USES_TECHNIQUE")
	rel2 := *graphrag.NewRelationship("finding-2", "T1190", "USES_TECHNIQUE")

	// Build batch
	batch := graphrag.NewBatch().
		AddNode(finding1).
		AddNode(finding2).
		AddNode(technique).
		AddRelationship(rel1).
		AddRelationship(rel2)

	fmt.Println("Nodes in batch:", len(batch.Nodes))
	fmt.Println("Relationships in batch:", len(batch.Relationships))
	// Output:
	// Nodes in batch: 3
	// Relationships in batch: 2
}

// ExampleNewTraversalOptions demonstrates configuring graph traversal.
func ExampleNewTraversalOptions() {
	opts := graphrag.NewTraversalOptions().
		WithMaxDepth(3).
		WithRelationshipTypes([]string{"USES_TECHNIQUE", "SIMILAR_TO"}).
		WithNodeTypes([]string{"finding", "technique"}).
		WithDirection("both")

	fmt.Println("Max depth:", opts.MaxDepth)
	fmt.Println("Direction:", opts.Direction)
	fmt.Println("Relationship types:", len(opts.RelationshipTypes))
	// Output:
	// Max depth: 3
	// Direction: both
	// Relationship types: 2
}

// ExampleQuery_WithWeights demonstrates configuring hybrid scoring weights.
func ExampleQuery_WithWeights() {
	// Emphasize semantic similarity
	semanticQuery := graphrag.NewQuery("privilege escalation").
		WithWeights(0.8, 0.2)

	// Emphasize graph structure
	graphQuery := graphrag.NewQuery("attack chain").
		WithWeights(0.3, 0.7)

	fmt.Println("Semantic query - Vector:", semanticQuery.VectorWeight, "Graph:", semanticQuery.GraphWeight)
	fmt.Println("Graph query - Vector:", graphQuery.VectorWeight, "Graph:", graphQuery.GraphWeight)
	// Output:
	// Semantic query - Vector: 0.8 Graph: 0.2
	// Graph query - Vector: 0.3 Graph: 0.7
}

// Example_storingAttackData demonstrates storing attack findings and techniques.
func Example_storingAttackData() {
	// Create finding node
	finding := graphrag.NewGraphNode("finding").
		WithID("finding-001").
		WithContent("SQL injection in /api/login parameter 'username'").
		WithProperty("severity", "critical").
		WithProperty("cvss_score", 9.8).
		WithProperty("endpoint", "/api/login").
		WithProperty("parameter", "username")

	// Create technique node
	technique := graphrag.NewGraphNode("technique").
		WithID("T1190").
		WithContent("Exploit Public-Facing Application").
		WithProperty("tactic", "Initial Access")

	// Link finding to technique
	rel := graphrag.NewRelationship(
		finding.ID,
		"T1190",
		"USES_TECHNIQUE",
	).WithProperty("confidence", 0.95)

	// Validate all components
	if err := finding.Validate(); err != nil {
		fmt.Printf("failed to validate finding: %v\n", err)
		return
	}
	if err := technique.Validate(); err != nil {
		fmt.Printf("failed to validate technique: %v\n", err)
		return
	}
	if err := rel.Validate(); err != nil {
		fmt.Printf("failed to validate relationship: %v\n", err)
		return
	}

	fmt.Println("Finding severity:", finding.Properties["severity"])
	fmt.Println("Technique ID:", technique.ID)
	fmt.Println("Relationship confidence:", rel.Properties["confidence"])
	// Output:
	// Finding severity: critical
	// Technique ID: T1190
	// Relationship confidence: 0.95
}

// Example_queryingSimilarPatterns demonstrates querying for similar attack patterns.
func Example_queryingSimilarPatterns() {
	// Query for similar findings
	findingQuery := graphrag.NewQuery("SQL injection vulnerabilities").
		WithTopK(10).
		WithNodeTypes("finding").
		WithMinScore(0.75).
		WithMission("mission-123")

	// Query for related techniques
	techniqueQuery := graphrag.NewQuery("privilege escalation techniques").
		WithNodeTypes("technique").
		WithMaxHops(2).
		WithWeights(0.7, 0.3) // Emphasize semantic similarity

	// Validate queries
	if err := findingQuery.Validate(); err != nil {
		fmt.Printf("failed to validate finding query: %v\n", err)
		return
	}
	if err := techniqueQuery.Validate(); err != nil {
		fmt.Printf("failed to validate technique query: %v\n", err)
		return
	}

	fmt.Println("Finding query - Top K:", findingQuery.TopK)
	fmt.Println("Finding query - Mission:", findingQuery.MissionID)
	fmt.Println("Technique query - Max hops:", techniqueQuery.MaxHops)
	fmt.Println("Technique query - Vector weight:", techniqueQuery.VectorWeight)
	// Output:
	// Finding query - Top K: 10
	// Finding query - Mission: mission-123
	// Technique query - Max hops: 2
	// Technique query - Vector weight: 0.7
}

// Example_buildingAttackChain demonstrates creating an attack chain from multiple findings.
func Example_buildingAttackChain() {
	// Create findings representing attack steps
	recon := graphrag.NewGraphNode("finding").
		WithID("step-1").
		WithContent("Port scan detected open SSH port 22")

	bruteforce := graphrag.NewGraphNode("finding").
		WithID("step-2").
		WithContent("SSH bruteforce successful with weak credentials")

	privesc := graphrag.NewGraphNode("finding").
		WithID("step-3").
		WithContent("Privilege escalation via sudo misconfiguration")

	// Link findings in sequence to form attack chain
	rel1 := graphrag.NewRelationship(recon.ID, bruteforce.ID, "LEADS_TO").
		WithProperty("sequence", 1).
		WithProperty("chain_id", "attack-chain-001")

	rel2 := graphrag.NewRelationship(bruteforce.ID, privesc.ID, "LEADS_TO").
		WithProperty("sequence", 2).
		WithProperty("chain_id", "attack-chain-001")

	fmt.Println("Step 1:", recon.ID, "->", rel1.ToID)
	fmt.Println("Step 2:", bruteforce.ID, "->", rel2.ToID)
	fmt.Println("Chain ID:", rel1.Properties["chain_id"])
	// Output:
	// Step 1: step-1 -> step-2
	// Step 2: step-2 -> step-3
	// Chain ID: attack-chain-001
}

// Example_errorHandling demonstrates proper error handling with sentinel errors.
func Example_errorHandling() {
	// Create an invalid query (missing text and embedding)
	invalidQuery := &graphrag.Query{
		TopK:         10,
		MaxHops:      3,
		MinScore:     0.7,
		VectorWeight: 0.6,
		GraphWeight:  0.4,
	}

	// Validate will return ErrInvalidQuery
	err := invalidQuery.Validate()
	if err != nil {
		fmt.Println("Query validation failed:", err)
	}

	// Create an invalid node (missing type)
	invalidNode := &graphrag.GraphNode{
		ID:      "node-1",
		Content: "Some content",
	}

	err = invalidNode.Validate()
	if err != nil {
		fmt.Println("Node validation failed:", err)
	}

	// Output:
	// Query validation failed: query must have either Text, Embedding, or NodeTypes
	// Node validation failed: node type is required
}

// Example_missionIsolation demonstrates mission-scoped knowledge graphs.
func Example_missionIsolation() {
	// Create nodes in a mission context
	finding1 := graphrag.NewGraphNode("finding").
		WithContent("Vulnerability in mission alpha target")
	// MissionID would be auto-populated by Gibson harness

	finding2 := graphrag.NewGraphNode("finding").
		WithContent("Vulnerability in mission beta target")
	// MissionID would be auto-populated by Gibson harness

	// Query within specific mission context
	missionQuery := graphrag.NewQuery("vulnerabilities in authentication").
		WithMission("mission-alpha").
		WithNodeTypes("finding")

	fmt.Println("Finding 1 type:", finding1.Type)
	fmt.Println("Finding 2 type:", finding2.Type)
	fmt.Println("Query mission:", missionQuery.MissionID)
	// Output:
	// Finding 1 type: finding
	// Finding 2 type: finding
	// Query mission: mission-alpha
}

// ExampleMissionScope demonstrates using mission scopes to control query result filtering.
func ExampleMissionScope() {
	// Default scope queries current mission run only
	defaultScope := graphrag.DefaultMissionScope
	fmt.Println("Default scope:", defaultScope.String())

	// Query only current mission run (same as default)
	missionRunScope := graphrag.ScopeMissionRun
	fmt.Println("Mission run scope:", missionRunScope.String())
	fmt.Println("Mission run valid:", missionRunScope.IsValid())

	// Query all runs of same mission
	missionScope := graphrag.ScopeMission
	fmt.Println("Mission scope:", missionScope.String())

	// Query across all missions (global)
	globalScope := graphrag.ScopeGlobal
	fmt.Println("Global scope:", globalScope.String())

	// Output:
	// Default scope: mission_run
	// Mission run scope: mission_run
	// Mission run valid: true
	// Mission scope: mission
	// Global scope: global
}

// ExampleParseMissionScope demonstrates parsing and validating mission scopes.
func ExampleParseMissionScope() {
	// Parse valid scope (both old and new names are accepted)
	scope, err := graphrag.ParseMissionScope("mission_run")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Println("Parsed scope:", scope.String())

	// Validate scope
	if err := scope.Validate(); err != nil {
		fmt.Printf("validation error: %v\n", err)
		return
	}
	fmt.Println("Scope is valid:", scope.IsValid())

	// Get all available scopes
	allScopes := graphrag.AllMissionScopes()
	fmt.Println("Available scopes:", len(allScopes))

	// Output:
	// Parsed scope: mission_run
	// Scope is valid: true
	// Available scopes: 3
}
